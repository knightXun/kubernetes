# macvlan plugin

## Overview

[macvlan](http://backreference.org/2014/03/20/some-notes-on-macvlanmacvtap/) functions like a switch that is already connected to the host interface.
A host interface gets "enslaved" with the virtual interfaces sharing the physical device but having distinct MAC addresses.
Since each macvlan interface has its own MAC address, it makes it easy to use with existing DHCP servers already present on the network.

## Example configuration
```
# cat /etc/macvlan/k8s_macvlan.conf 
 {
        "Name": "eth0",
        "type": "bridge"
 }
```

## Code analysis
* 根据脚本可以给pause容器手动加IP和网卡。本CNI插件其实就是把这个过程自动化，利用`github.com/vishvananda/netlink`。
* 手动调脚本添加网卡： `/usr/local/bin/macvlanAddIP.sh pause容器的containerID` 

```/usr/local/bin/macvlanAddIP.sh
#!/bin/bash
PAUSE_CONTAINER_ID=$1
DEV=`docker inspect $PAUSE_CONTAINER_ID|grep annotation.ips|awk -F'[":-]' '{print $5}'`         # 从docker label中获取网卡名
MACVLAN_IP=`docker inspect $PAUSE_CONTAINER_ID|grep annotation.ips|awk -F'[":-]' '{print $6}'`  # 从从docker label中获取ip
DOCKER_PID=`docker inspect $PAUSE_CONTAINER_ID|grep "\"Pid\""|awk -F'[,:]' '{print $2}'`        # 获取docker的PID
echo "Pid of pause container is: $DOCKER_PID"
echo "Netcard is: $DEV"
echo "Macvlan IP is: $MACVLAN_IP"
if [ -z $MACVLAN_IP ]; then
  echo "no macvlan IP"
  exit 1
fi
DEFAULT_INTERFACE=`ip -4 route ls | grep default | grep -Po '(?<=dev )(\S+)'`                   # 获取宿主机网卡名
echo "host interface is $DEFAULT_INTERFACE"
ip link add $DEV link $DEFAULT_INTERFACE type macvlan mode bridge                               # 建立macvlan网卡并桥接到到宿主机网卡
ip link set dev $DEV netns $DOCKER_PID                                                          # 将macvlan网卡添加到容器的network namespace中
nsenter -t $DOCKER_PID -n ip link set $DEV up                                                   # 在network namespace中set up网卡
nsenter -t $DOCKER_PID -n ip addr add $MACVLAN_IP/16 dev $DEV                                   # 添加IP
nsenter -t $DOCKER_PID -n ip route add 172.25.0.0/16 via 10.30.96.1 dev $DEV                    # 添加路由
```
* `macvlan.go`中的`createMacvlan` 自动化了这一过程：
    1. `ip.RandomVethName()` 和 `netlink.LinkAdd(mv)` 建立macvlan网卡并桥接到到宿主机网卡
    2. `netlink.LinkSetUp(macvlanIface)` 在network namespace中set up网卡
    3. `netlink.AddrAdd(macvlanIface, ipaddr)`  添加IP
    4. `netlink.RouteAdd(&route)` 添加路由
    5. `netlink.RouteReplace(&netlink.Route{
       				LinkIndex: macvlanIface.Attrs().Index,
       				Scope:     netlink.SCOPE_UNIVERSE,
       				Gw:        macvlanGateway,
       			})` 替换默认路由，只在个人开发那边有。
       			
* 删除网卡过程比较简单，理论上，network namespace删除后，其中的macvlan网卡会自动删除，不需要额外操作

## 申请释放IP
CNI 插件需要从pod的annotation中获取IP、网卡名、掩码、路由信息。
```
	MaskAnnotationKey   = "mask"            # 掩码
	RoutesAnnotationKey = "routes"          # 路由
	IPAnnotationKey     = "ips"             # IP
	NetworkKey          = "network"         #网卡名及IP类型
	ChangeGateway       = "changeGateWay"   # 是否需要修改默认路由，适用于个人开发的网段
``` 
释放申请IP的代码在`pkg/util/iputils/`中， kube-controller-manager的`controller/controller_utils.go`会调用iputils，在创建pod前申请IP，删除pod后释放IP。