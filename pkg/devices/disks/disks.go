package disks

import (
	"bufio"
	"fmt"
	nsv1alpha1 "hikvision.com/cloud/device-manager/pkg/crd/apis/device.k8s.io/v1alpha1"
	"k8s.io/klog"
	"os"
	"os/exec"
	"strings"
)

func MountDisks(lsblk map[string]BlockDevice, disk nsv1alpha1.Disk, chroot string) error {
	d := lsblk[disk.Name]
	if d.InUse == true {
		if d.MountPoint != disk.MountPoint {
			klog.Errorf("the disk %s was mounted on %s, but expect to %s.", disk.Name, d.MountPoint, disk.MountPoint)
			// TODO
		} else {
			klog.V(5).Infof("the disk %s was mounted on %s already.", disk.Name, disk.MountPoint)
		}
	} else {
		if err := mount(chroot, disk); err != nil {
			return err
		}

	}
	return nil
}

func makeFileSystem(disk nsv1alpha1.Disk) error {
	var cmd string
	switch disk.FileSystemType {
	case "xfs":
		cmd = "mkfs.xfs"
	case "ext4":
		cmd = "mkfs.ext4"
	default:
		klog.Warningf("The filesystem type %s is not support, default to xfs.", disk.FileSystemType)
		cmd = "mkfs.xfs"
	}

	path, err := exec.LookPath(cmd)
	if err != nil {
		return err
	}
	c := exec.Command(path, "-F", DevicePrefix+disk.Name)
	if err := c.Run(); err != nil {
		klog.Errorf("could not exec command %s to disk %s", path, disk.Name)
		return err
	}
	return nil
}

func mount(chroot string, disk nsv1alpha1.Disk) error {
	klog.Infof("exec mount of disk %s.", disk)
	cmd := exec.Command(
		sshCommand("mount"),
		DevicePrefix+disk.Name,
		chroot+disk.MountPoint,
	)
	err := cmd.Run()
	if err != nil {
		klog.Errorf("the disk %s mounted on %s failed.", disk.Name, disk.MountPoint)
		return err
	}
	return nil
}

func writeToFStab(chroot string, disk nsv1alpha1.Disk) error {
	line := make([]string, 0, 6)
	if disk.UUID != "" {
		line = append(line, fmt.Sprintf("UUID=%s", disk.UUID))
	} else {
		line = append(line, DevicePrefix+disk.Name)
	}

	line = append(line, disk.MountPoint)
	line = append(line, disk.FileSystemType)
	line = append(line, "defaults")
	var dump string
	dump = "0"
	if disk.Dump == true {
		dump = "1"
	}
	line = append(line, dump)
	line = append(line, "0")

	file, err := os.OpenFile(chroot+FStabPath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	w := bufio.NewWriter(file)
	_, err = w.WriteString(strings.Join(line, " "))
	if err != nil {
		return err
	}

	err = w.Flush()
	if err != nil {
		return err
	}

	return nil
}

func sshCommand(cmd string) string {
	return "ssh 127.0.0.1 " + cmd
}
