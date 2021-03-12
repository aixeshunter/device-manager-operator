# Device-Manager

[![Build Status](http://10.33.46.222:8000/api/badges/docker/device-manager/status.svg)](http://10.33.46.222:8000/docker/device-manager)

## Usage

* 负责监听CRD并对宿主机设备进行操作，目前主要包括：

1. 磁盘挂载
2. 磁盘卸载
3. 磁盘挂载时挂载点的文件清理

* 注意点
1. 一个节点一份extenddevice，名称为节点名


## 磁盘任务(action)和状态(status)

### 挂载任务及状态
在磁盘为卸载成功`umountSucceed`或等待中`Available`状态时可以执行挂载操作

* action: mount
  
* status:
1. mountSucceed：`挂载成功`
2. mountFailed：`挂载失败`
3. mounting: `挂载中`
4. Pending: 任务（挂载、卸载）等待中，可以显示`等待中`

### 卸载任务及状态
在磁盘非卸载成功`umountSucceed`状态时可以执行卸载操作

* action: umount

* status:
1. umountSucceed：`卸载成功`
2. umountFailed：`卸载失败`
3. umounting： `卸载中`
4. Pending: 任务（挂载、卸载）等待中，可以显示`等待中`

## 磁盘清理

在磁盘为挂载成功状态`MountSuccess`时可以执行清理操作。

## 执行清理

* action: clean
  
* cleanStatus:
1. cleanSucceed：`清理成功`
2. cleanFailed：`清理失败`
3. cleaning：`清理中`
4. Pending: `等待中`

## 其他状态

1. Available： `可用`

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
  node: worker2
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

## kocenter磁盘设备接口

### 获取disk设备信息


* Request
```
http://10.19.141.137:11443/v1/clusters/devices?name=master1&deviceType=disk
```

* Response
```json
{
  "data": {
    "devices": [
      {
        "name": "worker2",
        "disk": [
          {
            "name": "sda",
            "size": 600127266816,
            "sizeUnit": "558.9GB",
            "isSSD": false,
            "type": "",
            "serial": "5000039858391315",
            "vendor": "TOSHIBA",
            "partitions": [
              {
                "name": "sda1",
                "size": 1048576,
                "sizeUnit": "1MB",
                "mountPoint": "",
                "type": "",
                "uuid": "",
                "inuse": false,
                "status": "Available"
              },
              {
                "name": "sda2",
                "size": 1073741824,
                "sizeUnit": "1GB",
                "mountPoint": "/boot",
                "type": "ext4",
                "uuid": "4efc90f6-bbef-40f1-8818-8086341d66a3",
                "inuse": true,
                "action": "mount",
                "status": "mountSucceed"
              },
              {
                "name": "sda3",
                "size": 599050420224,
                "sizeUnit": "557.9GB",
                "mountPoint": "/",
                "type": "",
                "uuid": "a52c7B-KJr8-87Pm-d0Ox-5Qkh-5cSO-TTTJXs",
                "inuse": true,
                "action": "mount",
                "status": "mountSucceed"
              }
            ],
            "mountPoint": "",
            "inuse": true,
            "uuid": "",
            "hardwareID": "",
            "healthStatus": "pass",
            "healthError": null,
            "status": "Available"
          },
          {
            "name": "sdb",
            "size": 480103981056,
            "sizeUnit": "447.1GB",
            "isSSD": true,
            "type": "ext4",
            "serial": "BTDV741501GV480BGN",
            "vendor": "ATA",
            "partitions": [],
            "mountPoint": "",
            "inuse": false,
            "uuid": "1a689a11-919a-44ba-987e-5b144c0f2ec9",
            "hardwareID": "",
            "healthStatus": "pass",
            "healthError": null,
            "action": "mount",
            "status": "Pending"
          },
          {
            "name": "sdc",
            "size": 1800360124416,
            "sizeUnit": "1.6TB",
            "isSSD": false,
            "type": "xfs",
            "serial": "5000cca02c71c364",
            "vendor": "HGST",
            "partitions": [],
            "mountPoint": "",
            "inuse": false,
            "uuid": "974bc81d-ebb6-451f-b51a-96310679b0f8",
            "hardwareID": "",
            "healthStatus": "pass",
            "healthError": null,
            "action": "mount",
            "status": "mountSucceed"
          },
          {
            "name": "sdd",
            "size": 600127266816,
            "sizeUnit": "558.9GB",
            "isSSD": false,
            "type": "ext4",
            "serial": "5000cca02f419c4c",
            "vendor": "HGST",
            "partitions": [],
            "mountPoint": "",
            "inuse": false,
            "uuid": "0b5c6a03-7dd0-4e46-b969-129885d2f79b",
            "hardwareID": "",
            "healthStatus": "pass",
            "healthError": null,
            "status": "Available"
          }
        ]
      }
    ]
  },
  "message": "",
  "resultCode": "0"
}
```