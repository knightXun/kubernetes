
## minicloud events

- 分成三种层次的event：NodeLevel，PodLevel
- EventType分为两种: Normal, Warning

### PodLevel

- Kind表示为pod类型消息
- EventType为Normal(正常)或者Warning(报警)
- Reason固定为PodUpdate
- Message.PodName说明是哪个Pod
- Message.NodeName表示这个Pod所在节点
- Message.ResourceVersion（保留字段，用于表示event唯一性）
- Message.Status 表示Pod是否Ready（只有两个值NotReady和Ready）
- Message.Phase 表示Pod当前所处的状态 (Value值：Pending正在创建，Running正在运行，Delete删除，Creating kubelet已经接受到pod json并开始创建)
- Message.Pod这个Pod的Json文件
- Message.Time Event生成时间
- Message.EventMessage表示事件的日志说明
- Message.EventReason表示这个event产生的原因，分为:
    - DeadlineExceeded 删除Pod超时
    - FailedValidation Pod json文件格式错误
    - DNSConfigForming dns配置错误
    - NetworkNotReady pod容器网络异常
    - MissingClusterDNS pod的dns配置错误
    - Evicted pod被Kubelet驱逐
    - InspectImageFailed 镜像找不到
    - PulledImage 镜像拉取成功
    - ErrImageNeverPull 镜像拉取规则为Never，但是本地没有这个镜像
    - BackOffPullImage 镜像拉取超时
    - FailedKillPod 无法杀死Pod
    - PodDelete 删除pod
    - PodCreate 创建Pod
    - FailedToCreatePodContainer 无法创建Pod的容器
    - FailedToMakePodDataDirectories 无法创建Pod的数据目录
    - FailedMount Pod无法挂卷
    - PodReject kubelet拒绝创建这个Pod
    - InvalidEnvironmentVariableNames 无效的环境变量
    - CreatedContainer Pod创建容器成功
    - ContainerStartFailed Pod无法启动容器
    - ContainerStarted Pod启动容器成功
    - ContainerCreateFailed Pod创建容器失败
    - FailedPostStartHook Pod执行PostStartHook失败
    - FailedPreStopHook Pod执行PerStopHook失败
    - KillingContainer Pod杀死容器
    - SandboxChanged Pod的沙盘容器改变，Pod内部所有容器将被重新创建
    - FailedCreatePodSandBox 无法创建Pod的沙盘容器
    - FailedStatusPodSandBox 无法获取Pod沙盘容器状态
    - ContainerBackOff Pod容器反复重启
    - FailedSync 无法同步Pod状态
    - ExceededGracePeriod 在GraceTime内没有杀死容器
    - Preempting Kubelet准备驱逐Pod
    - ContainerUnhealthy Pod的容器不健康
    - PeriodStatusCheck Pod周期性状态检查

#### 示例：

```
{
	"taskCmd": "kubelet.PodLevel",
	"body": {
		"Kind": "Pod",
		"EventType": "Normal",
		"Reason": "PodUpdate",
		"Message": {
			"PodName": "ceph-go-10.30.100.101",
			"NodeName": "10.30.100.101",
			"EventAction": "xxxxx",
			"Time": xxxx,
			"EventReason": "",
			"EventMessage": "Event具体原因",
			"Pod": {
				"apiVersion": "v1",
				"kind": "Pod",
				"metadata": {
					"annotations": {
						"kubernetes.io/config.hash": "6d16421011a3be44bdde8035622d1135",
						"kubernetes.io/config.mirror": "6d16421011a3be44bdde8035622d1135",
						"kubernetes.io/config.seen": "2018-12-07T17:13:02.850212852+08:00",
						"kubernetes.io/config.source": "file"
					},
					"creationTimestamp": "2018-12-07T09:13:02Z",
					"name": "ceph-go-10.30.100.101",
					"namespace": "default",
					"resourceVersion": "167186359",
					"selfLink": "/api/v1/namespaces/default/pods/ceph-go-10.30.100.101",
					"uid": "4ca04f00-fa00-11e8-8ab7-0cc47a70d098"
				},
				"spec": {
					"containers": [{
						"image": "harbor.dahuatech.com/chenmiao/cephgo:20181012",
						"imagePullPolicy": "IfNotPresent",
						"livenessProbe": {
							"failureThreshold": 3,
							"httpGet": {
								"path": "/health",
								"port": "service-port",
								"scheme": "HTTP"
							},
							"initialDelaySeconds": 10,
							"periodSeconds": 10,
							"successThreshold": 1,
							"timeoutSeconds": 2
						},
						"name": "ceph-go",
						"ports": [{
							"containerPort": 9090,
							"hostPort": 9090,
							"name": "service-port",
							"protocol": "TCP"
						}],
						"resources": {
							"limits": {
								"cpu": "1",
								"memory": "512Mi"
							},
							"requests": {
								"cpu": "1",
								"memory": "512Mi"
							}
						},
						"terminationMessagePath": "/dev/termination-log",
						"terminationMessagePolicy": "File",
						"volumeMounts": [{
								"mountPath": "/data/cephGO/conf",
								"name": "conf"
							},
							{
								"mountPath": "/data/cephGO/logs",
								"name": "logs"
							},
							{
								"mountPath": "/data/cephGO/logs/opration",
								"name": "opration-log"
							}
						]
					}],
					"dnsPolicy": "ClusterFirst",
					"hostNetwork": true,
					"nodeName": "10.30.100.101",
					"restartPolicy": "Always",
					"schedulerName": "default-scheduler",
					"securityContext": {},
					"terminationGracePeriodSeconds": 30,
					"tolerations": [{
						"effect": "NoExecute",
						"operator": "Exists"
					}],
					"volumes": [{
							"hostPath": {
								"path": "/data/cephGO/conf",
								"type": ""
							},
							"name": "conf"
						},
						{
							"hostPath": {
								"path": "/data/cephGO/logs",
								"type": ""
							},
							"name": "logs"
						},
						{
							"hostPath": {
								"path": "/var/log/dhc/operationlog",
								"type": ""
							},
							"name": "opration-log"
						}
					]
				},
				"status": {
					"conditions": [{
							"lastProbeTime": null,
							"lastTransitionTime": "2018-12-07T09:13:02Z",
							"status": "True",
							"type": "Initialized"
						},
						{
							"lastProbeTime": null,
							"lastTransitionTime": "2018-12-19T01:03:25Z",
							"status": "True",
							"type": "Ready"
						},
						{
							"lastProbeTime": null,
							"lastTransitionTime": "2018-12-07T09:13:02Z",
							"status": "True",
							"type": "PodScheduled"
						}
					],
					"containerStatuses": [{
						"containerID": "docker://5693888224f63b434ceb63a50cf5c979225252e6a9eb693082ce7500081fa856",
						"image": "harbor.dahuatech.com/chenmiao/cephgo:20181012",
						"imageID": "docker-pullable://harbor.dahuatech.com/chenmiao/cephgo@sha256:76bb4eb4431c7c38aa1e19f31a1795e9b3069ed727c9b590a8639506a068a416",
						"lastState": {
							"terminated": {
								"containerID": "docker://5b3e7229be66fd50d2a343c0ac3e012674d7b8175e479c4c4e7a66c78aa1661e",
								"exitCode": 0,
								"finishedAt": "2018-12-19T01:02:51Z",
								"reason": "Completed",
								"startedAt": "2018-12-14T00:58:32Z"
							}
						},
						"name": "ceph-go",
						"ready": true,
						"restartCount": 3,
						"state": {
							"running": {
								"startedAt": "2018-12-19T01:03:25Z"
							}
						}
					}],
					"hostIP": "10.30.100.101",
					"phase": "Running",
					"podIP": "10.30.100.101",
					"qosClass": "Guaranteed",
					"startTime": "2018-12-07T09:13:02Z"
				}
			}
		}
	}
}
```

### NodeLevel

- Kind表示Node类型Event
- NodeName表示来自那个节点
- EventType有Normal,Warning,Normal代表正常，Warning表示报警信息
- Reason固定为"NodeUpdate"
- Message表示Event的具体信息
- Message.EventReason事件原因
- Message.EventMessage事件日志
- Message.EventAction可能取值为：
    - NodeDnsFailed(节点dns配置错误，报警)
    - NodeEvictStart(节点开始驱逐pod，报警)
    - InvalidDiskCapacity(节点磁盘报警event)
    - FreeDiskSpaceFailed(节点清理image失败)
    - ContainerGCFailed（节点ContainerGC失败）
    - KubeletSetupFailed（kubelet启动失败）
    - StartingKubelet（kubelet启动中，正常event）
    - NodeReady（kubelet正常运行）
    - NodeNotReady(kubelet不正常运行)
    - NodeHasInsufficientMemory(节点内存不足，报警)
    - NodeHasSufficientMemory(节点内存充足，正常)
    - NodeHasSufficientPID(节点PID充足)
    - NodeHasInsufficientPID(节点PID不足，报警)
    - NodeHasDiskPressure（节点磁盘压力大）
    - NodeHasNoDiskPressure(节点磁盘充足)
    - NodeRebooted（节点重启，报警）
    - NodeAllocatableEnforced(节点cgroup正常)
    - FailedNodeAllocatableEnforcement(节点cgroup不正常，报警)

#### 示例：

```
{
	"taskCmd": "kubelet.NodeLevel",
	"body": {
		"Kind": "Node",
		"NodeName": "NodeName_XXX",
		"EventType": "Normal"，
		"Reason": "NodeUpdate",
		"Message": {
		    "EventAction": "xxxxx",
		    "EventMessage": "Event具体原因",
		    "EventReason": "xxxx"
		}
	}
}
```

### 应用上线的event

正常流程, PodLevel级别Event

Pending正在创建，Running正在运行，Delete删除，Creating

```
graph LR
Creating --> Pending
Pending --> Running
Pending --> Delete
Running --> Pending
Running --> Delete
```

```
状态图
                  |---------------------------|
                  |                           |
Creating --> Pending <-----> Running ------> Delete
     |                                        |
     |----------------------------------------|
```

#### Creating消息例子
```
{
	"taskCmd": "kubelet.PodLevel",
	"body": {
		"Kind": "Pod",
		"EventType": "Normal",
		"Reason": "PodUpdate",
		"Message": {
		    "Time": xxxx,
			"PodName": "ceph-go",
		    "ResourceVersion": "",
		    "Status": "NotReady",
		    "Phase": "Creating",
			"NodeName": "10.30.100.101",
			"EventReason": "PodCreate",
			"EventMessage": "Pod Create",
			"Pod": {...}
		}
	}
}
```
#### Pending消息例子
```
{
	"taskCmd": "kubelet.PodLevel",
	"body": {
		"Kind": "Pod",
		"EventType": "Normal",
		"Reason": "PodUpdate",
		"Message": {
		    "Time": xxxx,
			"PodName": "ceph-go",
		    "ResourceVersion": "",
		    "Status": "NotReady",
		    "Phase": "Pending",
			"NodeName": "10.30.100.101",
			"EventReason": "xxx",
			"EventMessage": "xxxx",
			"Pod": {...}
		}
	}
}
```
#### Running消息例子
```
{
	"taskCmd": "kubelet.PodLevel",
	"body": {
		"Kind": "Pod",
		"EventType": "Normal",
		"Reason": "PodUpdate",
		"Message": {
		    "Time": xxxx,
			"PodName": "ceph-go",
		    "ResourceVersion": "",
		    "Status": "Ready",
		    "Phase": "Running",
			"NodeName": "10.30.100.101",
			"EventReason": "PeriodStatusCheck",
			"EventMessage": "PeriodStatusCheck PodReady",
			"Pod": {...}
		}
	}
}
```

#### Delete消息例子
```
{
	"taskCmd": "kubelet.PodLevel",
	"body": {
		"Kind": "Pod",
		"EventType": "Normal",
		"Reason": "PodUpdate",
		"Message": {
			"PodName": "ceph-go",
		    "ResourceVersion": "",
		    "Status": "NotReady",
		    "Phase": "Delete",
			"NodeName": "10.30.100.101",
			"EventReason": "PodDelete",
			"EventMessage": "Pod Delete Successfully",
			"Pod": {...}
		}
	}
}
```