/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/volume"
	volumeutil "k8s.io/kubernetes/pkg/volume/util"
	"os/exec"
	"path/filepath"
)

type fcAttacher struct {
	host    volume.VolumeHost
	manager diskManager
}

var _ volume.Attacher = &fcAttacher{}

var _ volume.AttachableVolumePlugin = &fcPlugin{}

func (plugin *fcPlugin) NewAttacher() (volume.Attacher, error) {
	return &fcAttacher{
		host:    plugin.host,
		manager: &FCUtil{},
	}, nil
}

func (plugin *fcPlugin) GetDeviceMountRefs(deviceMountPath string) ([]string, error) {
	mounter := plugin.host.GetMounter(plugin.GetPluginName())
	return mount.GetMountRefs(mounter, deviceMountPath)
}

func (attacher *fcAttacher) Attach(spec *volume.Spec, nodeName types.NodeName) (string, error) {
	return "", nil
}

func (attacher *fcAttacher) VolumesAreAttached(specs []*volume.Spec, nodeName types.NodeName) (map[*volume.Spec]bool, error) {
	volumesAttachedCheck := make(map[*volume.Spec]bool)
	for _, spec := range specs {
		volumesAttachedCheck[spec] = true
	}

	return volumesAttachedCheck, nil
}

func (attacher *fcAttacher) WaitForAttach(spec *volume.Spec, devicePath string, pod *v1.Pod, timeout time.Duration) (string, error) {
	attacher.host.GetFcMutex().Lock()
	defer attacher.host.GetFcMutex().Unlock()
	if spec.Volume.FC.RemoteVolumeID == "" {
		return "", fmt.Errorf("Empty FC.RemoteVolumeID")
	}

	lun, wwns, _, err := Lock(attacher.host.GetRemoteVolumeServerAddress(), spec.Volume.FC.RemoteVolumeID, attacher.host.GetInstanceID(), string(pod.UID))
	if err != nil {
		glog.Errorf("fc: failed to setup: %v", err)
		glog.Errorf("diskSetUp_failed, so we must unlock volume:%v", spec.Volume.FC.RemoteVolumeID)
		if err.Error() != "no fc disk found" || err.Error() != "Is Likely Not Mount Point" {
			glog.Errorf("diskSetUp_failed, clean %v", spec.Volume.FC.RemoteVolumeID)
			if lun != "" && len(wwns) != 0 && spec.Volume.FC.RemoteVolumeID != "" {
				volumeID := spec.Volume.FC.RemoteVolumeID
				wwnsstr := strings.Join(wwns, ",")
				lun := lun
				glog.V(1).Infoln("RemoteDellVolume_Clean after disksetup failed")
				out, err2 := exec.Command("/bin/bash", "-c", "/usr/bin/clean_removal.sh "+wwnsstr+" "+lun+" "+volumeID).CombinedOutput()
				if err2 != nil {
					glog.V(1).Infof("RemoteDellVolume_Clean clean fc device failed, meet error: %v , info: %v,"+
						"volumeID=%v, wwns=%v, lun=%v", err, string(out), volumeID, wwnsstr, lun)
					glog.Errorf("RemoteDellVolume_Clean clean fc device failed, meet error: %v , info: %v", err, string(out))
					//err = fmt.Errorf("clean fc device failed, meet error: %v , info: %v", err, string(out))
					//return err
				}
				err1 := UnlockWhenSetupFailed(attacher.host.GetRemoteVolumeServerAddress(), spec.Volume.FC.RemoteVolumeID, attacher.host.GetInstanceID(), string(pod.UID))
				if err1 != nil {
					glog.Errorf("After failure of diskSetUp(%v), So we unlock this volume, but failed: %v", volumeID, err1)
					err = fmt.Errorf(err.Error() + "   " + err1.Error())
					glog.Errorf("fc: failed to setup: %v", err)
					return "", err
				}
				glog.V(1).Infoln("RemoteDellVolume_Clean after disksetup failed")
				exec.Command("/bin/bash", "-c", "/usr/bin/clean_removal.sh "+wwnsstr+" "+lun+" "+volumeID).CombinedOutput()
			}
		}
		glog.Errorf(err.Error())
		return "", err
	}

	WriteVolumeInfoInPluginDir(attacher.host.GetPodDir(string(pod.UID)), spec.Name(), spec.Volume.FC.RemoteVolumeID, lun, wwns)

	mounter, err := volumeSpecToMounter(spec, attacher.host)
	if err != nil {
		glog.Warningf("failed to get fc mounter: %v", err)
		return "", err
	}
	return attacher.manager.AttachDisk(*mounter)
}

func (attacher *fcAttacher) GetDeviceMountPath(
	spec *volume.Spec) (string, error) {
	mounter, err := volumeSpecToMounter(spec, attacher.host)
	if err != nil {
		glog.Warningf("failed to get fc mounter: %v", err)
		return "", err
	}

	return attacher.manager.MakeGlobalPDName(*mounter.fcDisk), nil
}

func (attacher *fcAttacher) MountDevice(spec *volume.Spec, devicePath string, deviceMountPath string) error {
	mounter := attacher.host.GetMounter(fcPluginName)
	notMnt, err := mounter.IsLikelyNotMountPoint(deviceMountPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(deviceMountPath, 0750); err != nil {
				return err
			}
			notMnt = true
		} else {
			return err
		}
	}

	volumeSource, readOnly, err := getVolumeSource(spec)
	if err != nil {
		return err
	}

	options := []string{}
	if readOnly {
		options = append(options, "ro")
	}
	if notMnt {
		diskMounter := &mount.SafeFormatAndMount{Interface: mounter, Exec: attacher.host.GetExec(fcPluginName)}
		mountOptions := volumeutil.MountOptionFromSpec(spec, options...)
		err = diskMounter.FormatAndMount(devicePath, deviceMountPath, volumeSource.FSType, mountOptions)
		if err != nil {
			os.Remove(deviceMountPath)
			return err
		}
	}
	return nil
}

type fcDetacher struct {
	mounter mount.Interface
	manager diskManager
	host    volume.VolumeHost
	podUID  string
}

var _ volume.Detacher = &fcDetacher{}

func (plugin *fcPlugin) NewDetacher() (volume.Detacher, error) {
	return &fcDetacher{
		mounter: plugin.host.GetMounter(plugin.GetPluginName()),
		manager: &FCUtil{},
		host:  plugin.host,
	}, nil
}

func (detacher *fcDetacher) Detach(volumeName string, nodeName types.NodeName) error {
	return nil
}

func (detacher *fcDetacher) UnmountDevice(deviceMountPath string) error {
	// Specify device name for DetachDisk later
	devName, _, err := mount.GetDeviceNameFromMount(detacher.mounter, deviceMountPath)
	if err != nil {
		glog.Errorf("fc: failed to get device from mnt: %s\nError: %v", deviceMountPath, err)
		return err
	}
	// Unmount for deviceMountPath(=globalPDPath)
	err = volumeutil.UnmountPath(deviceMountPath, detacher.mounter)
	if err != nil {
		return fmt.Errorf("fc: failed to unmount: %s\nError: %v", deviceMountPath, err)
	}
	unMounter := volumeSpecToUnmounter(detacher.mounter)
	err = detacher.manager.DetachDisk(*unMounter, devName)
	if err != nil {
		return fmt.Errorf("fc: failed to detach disk: %s\nError: %v", devName, err)
	}
	glog.V(4).Infof("fc: successfully detached disk: %s", devName)
	return nil
}

func volumeSpecToMounter(spec *volume.Spec, host volume.VolumeHost) (*fcDiskMounter, error) {
	fc, readOnly, err := getVolumeSource(spec)
	if err != nil {
		return nil, err
	}
	var lun string
	var wwids []string
	if fc.Lun != nil && len(fc.TargetWWNs) != 0 {
		lun = strconv.Itoa(int(*fc.Lun))
	} else if len(fc.WWIDs) != 0 {
		for _, wwid := range fc.WWIDs {
			wwids = append(wwids, strings.Replace(wwid, " ", "_", -1))
		}
	} else {
		return nil, fmt.Errorf("fc: no fc disk information found. failed to make a new mounter")
	}
	fcDisk := &fcDisk{
		plugin: &fcPlugin{
			host: host,
		},
		wwns:  fc.TargetWWNs,
		lun:   lun,
		wwids: wwids,
		io:    &osIOHandler{},
	}
	// TODO: remove feature gate check after no longer needed
	if utilfeature.DefaultFeatureGate.Enabled(features.BlockVolume) {
		volumeMode, err := volumeutil.GetVolumeMode(spec)
		if err != nil {
			return nil, err
		}
		glog.V(5).Infof("fc: volumeSpecToMounter volumeMode %s", volumeMode)
		return &fcDiskMounter{
			fcDisk:     fcDisk,
			fsType:     fc.FSType,
			volumeMode: volumeMode,
			readOnly:   readOnly,
			mounter:    volumeutil.NewSafeFormatAndMountFromHost(fcPluginName, host),
			deviceUtil: volumeutil.NewDeviceHandler(volumeutil.NewIOHandler()),
		}, nil
	}
	return &fcDiskMounter{
		fcDisk:     fcDisk,
		fsType:     fc.FSType,
		readOnly:   readOnly,
		mounter:    volumeutil.NewSafeFormatAndMountFromHost(fcPluginName, host),
		deviceUtil: volumeutil.NewDeviceHandler(volumeutil.NewIOHandler()),
	}, nil
}

func volumeSpecToUnmounter(mounter mount.Interface) *fcDiskUnmounter {
	return &fcDiskUnmounter{
		fcDisk: &fcDisk{
			io: &osIOHandler{},
		},
		mounter:    mounter,
		deviceUtil: volumeutil.NewDeviceHandler(volumeutil.NewIOHandler()),
	}
}

func UnlockWhenSetupFailed(remoteVolumeServerAddress, volumeID, instanceID, podID string) error {
	glog.V(1).Info("UnlockWhenSetupFailed FibreChannel Unlock Begin")
	glog.V(1).Info("UnlockWhenSetupFailed FibreChannel Unlock, Try to UnlockFromPod Begin")
	err1 := UnlockFromPod(remoteVolumeServerAddress, volumeID, podID)

	glog.V(1).Info("UnlockWhenSetupFailed FibreChannel Unlock, Try to RemoteDetach from Server")
	err2 := DetachFromServer(remoteVolumeServerAddress, instanceID, volumeID)
	if err2 != nil {
		glog.Errorf("UnlockWhenSetupFailed FibreChannel Unlock, RemoteDetach Failed: %v", err2)
		var err error
		if err1 != nil {
			err = fmt.Errorf(err1.Error() + " " + err2.Error())
		} else {
			err = err2
		}
		return err
	}
	return nil

}

func WriteVolumeInfoInPluginDir(rootpath, volName, volumeID, lun string, wwns []string) error {
	//rootpath := fc.GetVolumeIDFilePath()
	volumepath := filepath.Join(rootpath, "volumes", "kubernetes.io~fc", "dellvolumeinfo")
	glog.V(1).Infof("Write VolumeID: %v To %v", volName, volumepath)

	os.Remove(volumepath)
	os.Create(volumepath)

	f, err := os.OpenFile(volumepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		glog.Errorf("Create_Dellvolumeinfo error: %v", err)
		return err
	}
	defer f.Close()

	_, err = f.WriteString("volName=" + volumeID + "\n")
	if err != nil {
		glog.Errorf("Create_DellvolumeinfoFail to Write VolumeID: %v To %v , Meet %v", volumeID, volumepath, err)
		return fmt.Errorf("Create_Dellvolumeinfo Fail to Write VolumeID: %v To %v , Meet %v", volumeID, volumepath, err)
	}

	_, err = f.WriteString("wwns=" + strings.Join(wwns, ",") + "\n")
	if err != nil {
		glog.Errorf("Create_Dellvolumeinfo Fail to Write Wwns: %v To %v , Meet %v", wwns, volumepath, err)
		return fmt.Errorf("Fail to Write Wwns: %v To %v , Meet %v", wwns, volumepath, err)
	}

	_, err = f.WriteString("lun=" + lun + "\n")
	if err != nil {
		glog.Errorf("Create_Dellvolumeinfo Fail to Write Lun: %v To %v , Meet %v", lun, volumepath, err)
		return fmt.Errorf("Fail to Write Lun: %v To %v , Meet %v", lun, volumepath, err)
	}
	return nil
}
