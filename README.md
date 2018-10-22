# Kubernetes

[![Submit Queue Widget]][Submit Queue] [![GoDoc Widget]][GoDoc] [![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/569/badge)](https://bestpractices.coreinfrastructure.org/projects/569)

<img src="https://github.com/kubernetes/kubernetes/raw/master/logo/logo.png" width="100">

----

Kubernetes is an open source system for managing [containerized applications]
across multiple hosts; providing basic mechanisms for deployment, maintenance,
and scaling of applications.

Kubernetes builds upon a decade and a half of experience at Google running
production workloads at scale using a system called [Borg],
combined with best-of-breed ideas and practices from the community.

Kubernetes is hosted by the Cloud Native Computing Foundation ([CNCF]).
If you are a company that wants to help shape the evolution of
technologies that are container-packaged, dynamically-scheduled
and microservices-oriented, consider joining the CNCF.
For details about who's involved and how Kubernetes plays a role,
read the CNCF [announcement].

----

## To start using Kubernetes

See our documentation on [kubernetes.io].

Try our [interactive tutorial].

Take a free course on [Scalable Microservices with Kubernetes].

## To start developing Kubernetes

The [community repository] hosts all information about
building Kubernetes from source, how to contribute code
and documentation, who to contact about what, etc.

If you want to build Kubernetes right away there are two options:

##### You have a working [Go environment].

```
$ go get -d k8s.io/kubernetes
$ cd $GOPATH/src/k8s.io/kubernetes
$ make
```

##### You have a working [Docker environment].

```
$ git clone https://github.com/kubernetes/kubernetes
$ cd kubernetes
$ make quick-release
```

For the full story, head over to the [developer's documentation].

## Support

If you need support, start with the [troubleshooting guide],
and work your way through the process that we've outlined.

That said, if you have questions, reach out to us
[one way or another][communication].

## Dahua Cloud's Fork
此版本基于 v1.10.7。需要从v1.6.7合入以下功能：
1. kubelet Macvlan CNI插件
2. controller 根据IP的label和annotation 申请、释放IP
3. apiserver 的logdir admission插件，用来根据annotation 挂载hostpath做为日志目录
4. scheduler 根据节点的标签来过滤调度节点 https://gitlab.dahuatech.com/dhc/kubernetes/merge_requests/94/diffs
5. RC/Statefulset/Job的各类消息，包括： PodReady/PodNotReady/PodDelete/PodAdd/JobPodCompete。后来还加了RcReady/RcNotReady/StatefulsetReady/StatefulsetNotReady
6. FC 插件修改
7. kubelet提供rbd 卷resize接口
8. 释放IP失败时发送 warning 消息


## 注意事项
1. 改动代码尽量以插件形式，复用原有代码，尽量少自己定义新的字段。
2. 从v1.6.7 升级 v1.10.7 时，由于v1.6.7 冰利在 Pod status 定义了新的字段用于缩容，而v1.10.7 也在pod status中加了字段，两者的protobuf的序列号一致，但是类型不一致，所以会冲突。解决方法：在`staging/src/k8s.io/api/core/v1/generated.pb.go`中转换类型。见`3c5c61508f120a72ea6c38ed1ed2e343dce20975`

[announcement]: https://cncf.io/news/announcement/2015/07/new-cloud-native-computing-foundation-drive-alignment-among-container
[Borg]: https://research.google.com/pubs/pub43438.html
[CNCF]: https://www.cncf.io/about
[communication]: https://git.k8s.io/community/communication
[community repository]: https://git.k8s.io/community
[containerized applications]: https://kubernetes.io/docs/concepts/overview/what-is-kubernetes/
[developer's documentation]: https://git.k8s.io/community/contributors/devel#readme
[Docker environment]: https://docs.docker.com/engine
[Go environment]: https://golang.org/doc/install
[GoDoc]: https://godoc.org/k8s.io/kubernetes
[GoDoc Widget]: https://godoc.org/k8s.io/kubernetes?status.svg
[interactive tutorial]: http://kubernetes.io/docs/tutorials/kubernetes-basics
[kubernetes.io]: http://kubernetes.io
[Scalable Microservices with Kubernetes]: https://www.udacity.com/course/scalable-microservices-with-kubernetes--ud615
[Submit Queue]: http://submit-queue.k8s.io/#/ci
[Submit Queue Widget]: http://submit-queue.k8s.io/health.svg?v=1
[troubleshooting guide]: https://kubernetes.io/docs/tasks/debug-application-cluster/troubleshooting/

[![Analytics](https://kubernetes-site.appspot.com/UA-36037335-10/GitHub/README.md?pixel)]()
