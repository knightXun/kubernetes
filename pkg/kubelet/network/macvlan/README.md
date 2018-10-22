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