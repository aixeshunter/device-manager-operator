package disks

import (
	"bufio"
	"fmt"
	"hikvision.com/cloud/device-manager/pkg/constants"
	nsv1alpha1 "hikvision.com/cloud/device-manager/pkg/crd/apis/device.k8s.io/v1alpha1"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"os/exec"
	"path"
	"strings"
)

func MountDisks(d BlockDevice, disk nsv1alpha1.Disk, chroot string) error {
	if d.InUse == true {
		if d.MountPoint != disk.MountPoint {
			klog.Errorf("mount: the disk %s was mounted on %s, but expect to %s.", disk.Name, d.MountPoint, disk.MountPoint)
		} else {
			klog.V(5).Infof("mount: the disk %s was mounted on %s already.", disk.Name, disk.MountPoint)

		}
	} else {
		if disk.Formatting == true {
			err := makeFileSystem(disk)
			if err != nil {
				return err
			}
		}

		if err := mount(chroot, disk); err != nil {
			return err
		}

		if disk.UUID == "" {
			blk, err := ListBlockDevices(chroot)
			if err != nil {
				klog.Errorf("mount: get lsblk failed for disk %s", disk.Name)
			} else {
				ds, ok := blk[disk.Name]
				if ok == true {
					if ds.UUID != "" {
						disk.UUID = ds.UUID
					}
				}
			}
		}

		if err := writeToFStab(chroot, disk); err != nil {
			return err
		}
	}

	return nil
}

func UmountDisks(d BlockDevice, disk nsv1alpha1.Disk, chroot string) error {
	if d.InUse == true {
		if d.MountPoint == "/" || disk.MountPoint == "/" {
			return fmt.Errorf("the mount point / can not umount")
		}

		if d.MountPoint != disk.MountPoint {
			klog.Errorf("umount: the disk %s was mounted on %s, but expect to %s.", disk.Name, d.MountPoint, disk.MountPoint)
			return fmt.Errorf("the disk %s was mounted on %s, but expect to %s", disk.Name, d.MountPoint, disk.MountPoint)
		} else {
			klog.Infof("umount: the disk %s was mounted on %s already, umount starting...", disk.Name, disk.MountPoint)
		}
		// Umount disk from mount point
		if err := umount(disk); err != nil {
			klog.Errorf("the disk %s was failed to umount: %s", disk.Name, err)
			return err
		}
		// delete contents from /etc/fstab
		if err := deleteFromFStab(chroot, disk, d); err != nil {
			return err
		}
	} else {
		klog.V(5).Infof("umount: the disk %s was not mounted already.", disk.Name)
	}

	return nil
}

func CleanDisk(lsblk map[string]BlockDevice, disk nsv1alpha1.Disk, chroot string) error {
	d := lsblk[disk.Name]
	if d.InUse == true {
		if d.MountPoint != disk.MountPoint {
			klog.Errorf("clean: the disk %s was mounted on %s, not %s, clean failed.", disk.Name, d.MountPoint, disk.MountPoint)
			return fmt.Errorf("clean: the disk %s was mounted on %s, not %s, clean failed", disk.Name, d.MountPoint, disk.MountPoint)
		} else {
			klog.V(5).Infof("clean: the disk %s was mounted on %s already.", disk.Name, disk.MountPoint)
			err := cleanData(chroot, disk)
			if err != nil {
				return err
			}
		}
	} else {
		klog.Errorf("clean: the disk %s was not mounted.", disk.Name)
	}

	return nil
}

func cleanData(chroot string, disk nsv1alpha1.Disk) error {
	p := chroot + disk.MountPoint
	dir, err := ioutil.ReadDir(p)
	if err != nil {
		return err
	}

	for _, d := range dir {
		_ = os.RemoveAll(path.Join([]string{p, d.Name()}...))
	}

	return nil
}

func makeFileSystem(disk nsv1alpha1.Disk) error {
	var cmd string
	switch disk.FileSystemType {
	case "xfs":
		cmd = "mkfs.xfs -f"
	case "ext4":
		cmd = "mkfs.ext4 -F"
	default:
		klog.Warningf("The filesystem type %s is not support, default to xfs.", disk.FileSystemType)
		cmd = "mkfs.xfs -f"
	}

	c := exec.Command(cmd, DevicePrefix+disk.Name)
	if err := c.Run(); err != nil {
		klog.Errorf("could not exec command %s to disk %s", cmd, disk.Name)
		return err
	}
	return nil
}

func umount(disk nsv1alpha1.Disk) error {
	klog.Infof("exec umount of disk %s.", disk.Name)
	cmd := exec.Command(
		sshCommand("umount"),
		DevicePrefix+disk.Name,
	)
	err := cmd.Run()
	if err != nil {
		klog.Errorf("the disk %s umount with mount point %s failed.", disk.Name, disk.MountPoint)
		return err
	}
	return nil
}

func deleteFromFStab(chroot string, disk nsv1alpha1.Disk, blk BlockDevice) error {
	var match string
	var output []byte
	var flag bool
	var err error
	if disk.UUID != "" {
		match = disk.UUID
	} else if blk.UUID != "" {
		match = blk.UUID
	}
	output, flag, err = constants.ReadFile(chroot+FStabPath, match, "")
	if err != nil {
		return err
	}

	if flag != true {
		match = DevicePrefix + disk.Name
		output, flag, err = constants.ReadFile(chroot+FStabPath, match, "")
		if err != nil {
			return err
		}
	}
	// write output to /etc/fstab
	if flag == true {
		err = constants.WriteToFile(chroot+FStabPath, output)
		if err != nil {
			return err
		}
	}

	return nil
}

func mount(chroot string, disk nsv1alpha1.Disk) error {
	klog.Infof("exec mount of disk %s.", disk.Name)
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
