package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"hikvision.com/cloud/device-manager/pkg/constants"
	nsv1alpha1 "hikvision.com/cloud/device-manager/pkg/crd/apis/device.k8s.io/v1alpha1"
	crdClient "hikvision.com/cloud/device-manager/pkg/crd/client/clientset/versioned"
	diskclient "hikvision.com/cloud/device-manager/pkg/devices/disks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"time"

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
		klog.Infof("Create empty extend device resource %s to cluster.", name)
		emptyED := GetEmptyExtendDevice(name)
		emptyED.Name = name
		if _, err := CreateExtendDevice(ctx, client, emptyED); err != nil {
			return err
		}
	} else {
		klog.V(5).Infof("The extend device resource %s is existing in cluster.", name)
		err := HandleDisks(ctx, client, ed, chroot, name)
		if err != nil {
			klog.Errorf(fmt.Sprintf("error to handler disks: %s", err))
			return err
		}
	}

	return nil
}

func HandleDisks(ctx context.Context, client crdClient.Interface, ed *nsv1alpha1.ExtendDevice, chroot, name string) error {
	lsblk, err := diskclient.ListBlockDevices(chroot)
	if err != nil {
		return err
	}

	disks := ed.Spec.Disks
	change := false
	for i, d := range disks {
		if d.Status == nsv1alpha1.MountAvail || d.Status == "" {
			disks[i].Status = nsv1alpha1.Pending
			change = true
		}
	}
	if change == true {
		t := time.Now()
		ed.Status.Message = fmt.Sprintf("%s: There are some available disks should be handled.", t.Format(constants.TimeLayout))
		err = patchDisk(ctx, client, name, *ed, disks)
		if err != nil {
			klog.Error("Update disk status failed when starting.")
			return err
		}
	}

	for i, d := range disks {
		bd, ok := lsblk[d.Name]
		if ok == false {
			klog.Errorf("disk %s not found in lsblk command, is it exist?", d.Name)
			disks[i].Status = nsv1alpha1.MountFailed
			disks[i].Error = append(disks[i].Error, nsv1alpha1.Error{
				Err:  fmt.Sprintf("disk %s not found in lsblk command, is it exist?", d.Name),
				Time: metav1.Now(),
			})
			_ = patchDisk(ctx, client, name, *ed, disks)
			continue
		}

		switch d.Action {
		case nsv1alpha1.DiskMount:
			// clean disk data
			if d.Status == nsv1alpha1.MountSuccess && d.Clean == true {
				klog.Infof("disk %s clean starting...", d.Name)
				disks[i].CleanStatus = nsv1alpha1.Cleaning
				_ = patchDisk(ctx, client, name, *ed, disks)
				if err := diskclient.CleanDisk(lsblk, d, chroot); err != nil {
					klog.Errorf("disk %s clean failed: %s", d.Name, err)
					disks[i].Error = append(disks[i].Error, nsv1alpha1.Error{
						Err:  fmt.Sprintf("disk clean failed: %s.", err),
						Time: metav1.Now(),
					})
					disks[i].CleanStatus = nsv1alpha1.CleanSuccess
				} else {
					disks[i].CleanStatus = nsv1alpha1.CleanFailed
				}
				disks[i].Clean = false
				_ = patchDisk(ctx, client, name, *ed, disks)
			}

			// mount disk
			if d.Status == nsv1alpha1.Pending || d.Status == nsv1alpha1.UmountSuccess {
				klog.Infof("disk %s mount starting...", d.Name)
				if err := diskclient.MountDisks(bd, d, chroot); err != nil {
					klog.Errorf("disk %s mount failed: %s", d.Name, err)
					disks[i].Status = nsv1alpha1.MountFailed
					disks[i].Error = append(disks[i].Error, nsv1alpha1.Error{
						Err:  fmt.Sprintf("disk umount failed: %s", err),
						Time: metav1.Now(),
					})
					// umount mount disk
					lb, err := diskclient.ListBlockDevices(chroot)
					if err == nil {
					_:
						diskclient.UmountDisks(lb[d.Name], d, chroot)
					}
				} else {
					disks[i].Status = nsv1alpha1.MountSuccess
				}
				_ = patchDisk(ctx, client, name, *ed, disks)
			}
		case nsv1alpha1.DiskUmount:
			if d.Status == nsv1alpha1.MountSuccess {
				klog.Infof("disk %s umount starting...", d.Name)
				// update disk status
				disks[i].Status = nsv1alpha1.Pending
				_ = patchDisk(ctx, client, name, *ed, disks)
				if err := diskclient.UmountDisks(bd, d, chroot); err != nil {
					klog.Errorf("disk %s umount failed: %s", d.Name, err)
					disks[i].Status = nsv1alpha1.UmountFailed
					disks[i].Error = append(disks[i].Error, nsv1alpha1.Error{
						Err:  fmt.Sprintf("%s", err),
						Time: metav1.Now(),
					})
				} else {
					disks[i].Status = nsv1alpha1.UmountSuccess
				}
				_ = patchDisk(ctx, client, name, *ed, disks)
			}
		default:
			klog.Errorf("the action %s is not support of disk %s.", d.Action, d.Name)
			disks[i].Error = append(disks[i].Error, nsv1alpha1.Error{
				Err:  fmt.Sprintf("the action %s is not support of disk %s.", d.Action, d.Name),
				Time: metav1.Now(),
			})
			_ = patchDisk(ctx, client, name, *ed, disks)
		}
	}

	return nil
}

func updateDiskStatus(ctx context.Context, client crdClient.Interface, name string, disks []nsv1alpha1.Disk) error {
	ed, err := GetExtendDevice(ctx, client, name)
	if err != nil {
		return err
	}
	ed.Spec.Disks = disks
	ed.Status.LastUpdateTime = metav1.Now()
	_, err = UpdateExtendDevice(ctx, client, ed)
	if err != nil {
		klog.Errorf("update disk %v status with k8s client failed: %s.", disks, err)
		return err
	}
	return nil
}

func patchDisk(ctx context.Context, client crdClient.Interface, name string, old nsv1alpha1.ExtendDevice, disks []nsv1alpha1.Disk) error {
	old.Spec.Disks = disks
	old.Status.LastUpdateTime = metav1.Now()
	newData, err := json.Marshal(old)
	if err != nil {
		klog.Errorf("json.Marshal extenddevice resource %s failed: %s", old.Spec.Node, err)
		return err
	}

	_, err = client.DeviceV1alpha1().ExtendDevices().Patch(ctx, name, types.MergePatchType, newData)
	if err != nil {
		klog.Errorf("patch extenddevice resource %s failed: %s", old.Spec.Node, err)
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
