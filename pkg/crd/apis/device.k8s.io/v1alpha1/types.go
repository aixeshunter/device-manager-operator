package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Disk action
	DiskMount  = "mount"
	DiskUmount = "umount"

	// Clean status
	CleanSuccess = "cleanSucceed"
	CleanFailed  = "cleanFailed"
	Cleaning     = "cleaning"

	// Disk mount status
	MountSuccess  = "mountSucceed"
	MountFailed   = "mountFailed"
	UmountSuccess = "umountSucceed"
	UmountFailed  = "umountFailed"
	// waiting
	MountAvail = "Available"
	// executing
	Pending = "Pending"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

type ExtendDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExtendDeviceSpec   `json:"spec"`
	Status ExtendDeviceStatus `json:"status,omitempty"`
}

type ExtendDeviceStatus struct {
	LastUpdateTime metav1.Time `json:"lastUpdateTime"`
	Message        string      `json:"message"`
}

type ExtendDeviceSpec struct {
	Disks []Disk `json:"disk,omitempty"`
	// usb is not supported
	USB []USB `json:"usb,omitempty"`

	// The node name that device exists
	Node string `json:"node"`
}

type Disk struct {
	// The name of block device, like sda, vda, hda.
	// The name regex: "(s|v|h)d([a-z]*)"
	Name string `json:"name"`
	// Clean disk data
	// if true: clean data of disk
	// if false: pass
	Clean bool `json:"clean"`

	// filesystem of block device, like xfs, ext4, ext3.
	FileSystemType string `json:"fsType"`

	// MountPoint is the path at which the block devices is mounted.
	MountPoint string `json:"mountPoint"`

	// UUID is a unique identifier for the filesystem on the block device.
	//
	// This will be empty if the block device does not have a filesystem,
	// or if the filesystem is not yet known to Juju.
	//
	// The UUID format is not necessarily uniform; for example, LVM UUIDs
	// differ in format to the standard v4 UUIDs.
	UUID string `json:"uuid"`

	// Whether to format the block device before mount.
	Formatting bool `json:"formatting"`

	// The action to disk, about mount and umount.
	Action string `json:"action"`

	// The status of block storage device
	Status string `json:"status"`

	// The clean status of block storage device
	CleanStatus string `json:"cleanStatus,omitempty"`

	Dump bool `json:"dump"`

	// The error message of action
	Error []Error `json:"errors"`
}

type Error struct {
	Err  string      `json:"error,omitempty"`
	Time metav1.Time `json:"ErrorTime"`
}

// usb is not supported
type USB struct {
	UUID string `json:"uuid,omitempty"`
}

// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ExtendDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ExtendDevice `json:"items"`
}
