# Device-Manager

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