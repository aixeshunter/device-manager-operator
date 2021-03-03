package kube

import (
	"context"
	"fmt"
	nsv1alpha1 "hikvision.com/cloud/device-manager/pkg/crd/apis/device.k8s.io/v1alpha1"
	crdClient "hikvision.com/cloud/device-manager/pkg/crd/client/clientset/versioned"
	diskclient "hikvision.com/cloud/device-manager/pkg/devices/disks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient(kubeConfigFile string) (crdClient.Interface, error) {
	var clientConfig *rest.Config
	var err error
	if len(kubeConfigFile) > 0 {
		clientConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigFile)
		if err != nil {
			return nil, err
		}
	} else {
		clientConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	client, err := crdClient.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func GetExtendDevice(ctx context.Context, client crdClient.Interface, name string) (*nsv1alpha1.ExtendDevice, error) {
	ed, err := client.DeviceV1alpha1().ExtendDevices().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ed, nil
}

func UpdateExtendDevice(ctx context.Context, client crdClient.Interface, extendDevice *nsv1alpha1.ExtendDevice) (*nsv1alpha1.ExtendDevice, error) {
	ed, err := client.DeviceV1alpha1().ExtendDevices().Update(ctx, extendDevice)
	if err != nil {
		return nil, err
	}

	return ed, nil
}

func CreateExtendDevice(ctx context.Context, client crdClient.Interface, extendDevice *nsv1alpha1.ExtendDevice) (*nsv1alpha1.ExtendDevice, error) {
	ed, err := client.DeviceV1alpha1().ExtendDevices().Create(ctx, extendDevice)
	if err != nil {
		return nil, err
	}

	return ed, nil
}

func HandleExtendDevice(ctx context.Context, client crdClient.Interface, name string, chroot string) error {
	// Get old information of crd ExtendDevice
	ed, err := GetExtendDevice(ctx, client, name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		klog.V(4).Infof("Create empty extend device %s to cluster.", name)
		emptyED := GetEmptyExtendDevice(name)
		emptyED.Name = name
		if _, err := CreateExtendDevice(ctx, client, emptyED); err != nil {
			return err
		}
	} else {
		err := HandleDisks(ctx, client, ed, chroot)
		if err != nil {
			return err
		}
	}

	return nil
}

func HandleDisks(ctx context.Context, client crdClient.Interface, ed *nsv1alpha1.ExtendDevice, chroot string) error {
	lsblk, err := diskclient.ListBlockDevices(chroot)
	if err != nil {
		return err
	}

	disks := ed.Spec.Disks
	for i, d := range disks {
		if d.Status == nsv1alpha1.MountAvail || d.Status == "" {
			disks[i].Status = nsv1alpha1.Pending
		}
	}
	err = updateDiskStatus(ctx, client, ed, disks)
	if err != nil {
		klog.Error("Update disk status failed when starting.")
		return err
	}

	for i, d := range disks {
		bd, ok := lsblk[d.Name]
		if ok == false {
			klog.Errorf("disk %s not found in lsblk command, is it exist?", d.Name)
			disks[i].Status = nsv1alpha1.MountFailed
			disks[i].Error = append(disks[i].Error, fmt.Sprintf("disk %s not found in lsblk command, is it exist?", d.Name))
			_ = updateDiskStatus(ctx, client, ed, disks)
			continue
		}

		switch d.Action {
		case nsv1alpha1.DiskMount:
			// clean disk data
			if d.Status == nsv1alpha1.MountSuccess && d.Clean == true {
				disks[i].CleanStatus = nsv1alpha1.Cleaning
				_ = updateDiskStatus(ctx, client, ed, disks)
				if err := diskclient.CleanDisk(lsblk, d, chroot); err != nil {
					klog.Errorf("disk %s clean failed: %s", d.Name, err)
					disks[i].Error = append(disks[i].Error, fmt.Sprintf("disk clean failed: %s.", err))
					disks[i].CleanStatus = nsv1alpha1.CleanSuccess
				} else {
					disks[i].CleanStatus = nsv1alpha1.CleanFailed
				}
				disks[i].Clean = false
				_ = updateDiskStatus(ctx, client, ed, disks)
			}

			// mount disk
			if d.Status == nsv1alpha1.Pending || d.Status == nsv1alpha1.UmountSuccess {
				if err := diskclient.MountDisks(bd, d, chroot); err != nil {
					klog.Errorf("disk %s mount failed: %s", d.Name, err)
					disks[i].Status = nsv1alpha1.MountFailed
					disks[i].Error = append(disks[i].Error, fmt.Sprintf("disk umount failed: %s", err))
				} else {
					disks[i].Status = nsv1alpha1.MountSuccess
				}
				_ = updateDiskStatus(ctx, client, ed, disks)
			}
		case nsv1alpha1.DiskUmount:
			if d.Status == nsv1alpha1.MountSuccess {
				// update disk status
				disks[i].Status = nsv1alpha1.Pending
				_ = updateDiskStatus(ctx, client, ed, disks)
				if err := diskclient.UmountDisks(bd, d, chroot); err != nil {
					klog.Errorf("disk %s umount failed: %s", d.Name, err)
					disks[i].Status = nsv1alpha1.UmountFailed
					disks[i].Error = append(disks[i].Error, fmt.Sprintf("%s", err))
				} else {
					disks[i].Status = nsv1alpha1.UmountSuccess
				}
				_ = updateDiskStatus(ctx, client, ed, disks)
			}
		default:
			klog.Errorf("the action %s is not support of disk %s.", d.Action, d.Name)
			disks[i].Error = append(disks[i].Error, fmt.Sprintf("the action %s is not support of disk %s.", d.Action, d.Name))
			_ = updateDiskStatus(ctx, client, ed, disks)
		}
	}

	//ed.Spec.Disks = disks
	//ed.Status.LastUpdateTime = metav1.Now()
	//_, err = UpdateExtendDevice(ctx, client, ed)
	//if err != nil {
	//	return err
	//}
	return nil
}

func updateDiskStatus(ctx context.Context, client crdClient.Interface, ed *nsv1alpha1.ExtendDevice, disks []nsv1alpha1.Disk) error {
	ed.Spec.Disks = disks
	ed.Status.LastUpdateTime = metav1.Now()
	_, err := UpdateExtendDevice(ctx, client, ed)
	if err != nil {
		klog.Errorf("update disk %s status with k8s client failed.", disks)
		return err
	}

	return nil
}

func GetEmptyExtendDevice(name string) *nsv1alpha1.ExtendDevice {
	status := nsv1alpha1.ExtendDeviceStatus{
		LastUpdateTime: metav1.Now(),
		Message:        "create",
	}

	spec := nsv1alpha1.ExtendDeviceSpec{
		Node:  name,
		Disks: []nsv1alpha1.Disk{},
		USB:   []nsv1alpha1.USB{},
	}

	return &nsv1alpha1.ExtendDevice{
		// Extend Device default name is node name.
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec:   spec,
		Status: status,
	}
}
