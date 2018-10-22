package macvlan

import (
	"flag"
	dockertypes "github.com/docker/engine-api/types"
	containerConfig "github.com/docker/engine-api/types/container"
	"github.com/stretchr/testify/assert"
	"k8s.io/kubernetes/pkg/kubelet/container"
	docker "k8s.io/kubernetes/pkg/kubelet/dockertools"
	"k8s.io/kubernetes/pkg/kubelet/network"
	"testing"
	"time"

	"fmt"
)

const (
	defaultRuntimeRequestTimeoutDuration = 1 * time.Minute
	defaultImagePullProgressDeadline     = 1 * time.Minute
	defaultDockerEndpoint                = "unix:///var/run/docker.sock"
)

func TestMacvlanPlugin(t *testing.T) {
	flag.Set("alsologtostderr", "true")
	flag.Set("v", "6")
	flag.Parse()

	dockerClient := docker.ConnectToDockerOrDie(defaultDockerEndpoint, defaultRuntimeRequestTimeoutDuration, defaultImagePullProgressDeadline)
	plug := NewPlugin("/etc/macvlan", dockerClient)
	if plug == nil {
		t.Fatalf("Nil network plugin!")
	}
	plug.Init(nil, "", "10.252.0.0/16", 1500)

	ipArr := []string{
		"eth123-20.1.1.103",
		"eth123-20.1.1.104",
		"eth123-20.1.1.105",
	}
	config := dockertypes.ContainerCreateConfig{
		Config: &containerConfig.Config{
			Image: "reg.dhdc.com/library/nginx:1.9",
		},
	}
	containerIDs := []string{}
	for i := 0; i < len(ipArr); i++ {
		resp, err := dockerClient.CreateContainer(config)
		assert.NoError(t, err)
		containerIDs = append(containerIDs, resp.ID)
		err = dockerClient.StartContainer(resp.ID)
		assert.NoError(t, err)
		containerID := container.BuildContainerID("", resp.ID)
		annotations := map[string]string{
			network.IPAnnotationKey:     ipArr[i],
			network.MaskAnnotationKey:   "eth123-16",
			network.RoutesAnnotationKey: "192.168.2.0/24",
		}
		err = plug.SetUpPod("fakeNamespace", fmt.Sprintf("fakeName-%d", i), containerID, annotations)
		assert.NoError(t, err)
	}
	time.Sleep(3 * time.Second)
	for i := 0; i < len(containerIDs); i++ {
		containerID := container.BuildContainerID("", containerIDs[i])
		err := plug.TearDownPod("fakeNamespace", fmt.Sprintf("fakeName-%d", i), containerID)
		assert.NoError(t, err)
		err = dockerClient.StopContainer(containerIDs[i], 2)
		assert.NoError(t, err)
		err = dockerClient.RemoveContainer(containerIDs[i], dockertypes.ContainerRemoveOptions{RemoveVolumes: true})
		assert.NoError(t, err)
	}

}
