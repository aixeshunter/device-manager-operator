# Device-Manager

[![Build Status](http://10.33.46.222:8000/api/badges/docker/device-manager/status.svg)](http://10.33.46.222:8000/docker/device-manager)

## Usage

* 负责监听CRD并对宿主机设备进行操作，目前主要包括：

1. 磁盘挂载
2. 磁盘卸载

* 注意点
1. 一个节点一份extend


## 开发记录

### 容器内磁盘挂载

```bash
ssh 127.0.0.1 mount /dev/sdd <mount path>
```

* 注意点
1. 宿主机上的`~/.ssh`文件夹需要挂载到容器内；

2. 宿主机上的`/etc/ssh/ssh_config`文件需要挂载进容器内；

3. 容器内需要安装ssh包；

### 需要用到的命令

1. lsblk
2. mount, umount
3. mkfs.xfs mkfs.ext4
4. ssh

### koscenter需要开发的接口

1. 磁盘列表（调用node-hardware-discovery接口）
2. 磁盘挂载（status置为available）、卸载、清理

### extenddevice示例

```yaml
apiVersion: device.k8s.io/v1alpha1
kind: ExtendDevice
metadata:
  creationTimestamp: "2021-03-04T03:35:43Z"
  generation: 802
  name: master1
  resourceVersion: "61443566"
  selfLink: /apis/device.k8s.io/v1alpha1/extenddevices/master1
  uid: 8c42ad96-4761-441d-b07a-0ec9987abae2
spec:
  node: master1
  disk:
  - name: sdc
    clean: false
    fsType: xfs
    mountPoint: /opt/mnt2
    uuid: b6895b0c-7408-45f5-bba9-ca1baf7aa44a
    formatting: true
    action: mount
    status: Available
    dump: false
  - name: sdb
    clean: false
    fsType: ext4
    mountPoint: /opt/mnt1
    uuid: bef7fc78-c463-4a8a-9aed-f0a40d725060
    formatting: true
    action: mount
    status: Available
    dump: true
status:
  lastUpdateTime: "2021-03-04T06:57:27Z"
  message: create
```

## Installation

### deploy with helm

```bash
helm install device-manager device-manager/ -n kube-system
```

### uninstall

```bash
helm delete device-manager  -n kube-system
```