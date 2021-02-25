package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

type ExtendSystemInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExtendSystemInfoSpec   `json:"spec"`
	Status ExtendSystemInfoStatus `json:"status,omitempty"`
}

type ExtendSystemInfoSpec struct {
	CPU        CPUInfo    `json:"cpu,omitempty"`
	Memory     MemoryInfo `json:"memory,omitempty"`
	Uptime     float64    `json:"uptime,omitempty"`
	Networks   []Nic      `json:"networks,omitempty"`
	Disks      []Disk     `json:"disk,omitempty"`
	IPMI       IPMI       `json:"ipmi,omitempty"`
	PCIDevices PCIDevice  `json:"pci,omitempty"`
	GPU        GPUInfo    `json:"gpu,omitempty"`
}

type GPUInfo struct {
	Count int `json:"count"`
}

type CPUInfo struct {
	Arch        string `json:"arch,omitempty"`
	Model       string `json:"model,omitempty"`
	PhysicalNum int    `json:"physicalNum,omitempty"`
	Cores       int    `json:"cores,omitempty"`
	Count       int    `json:"count,omitempty"`
}

type MemoryInfo struct {
	Total int64 `json:"total,omitempty"`
}

type Nic struct {
	Name      string   `json:"name"`
	IP        []string `json:"ip,omitempty"`
	Mac       string   `json:"mac"`
	Speed     string   `json:"speed,omitempty"`
	IsLink    bool     `json:"isLink"`
	IsVirtual bool     `json:"isVirtual,omitempty"`
	IsLogic   bool     `json:"isLogic"`
	Gateway   string   `json:"gateway,omitempty"`
}

type Disk struct {
	Name  string `json:"name"`
	Size  uint64 `json:"size,omitempty"`
	IsSSD bool   `json:"isSSD"`
	// TODO: next version implement
	Type string `json:"type,omitempty"`
	// TODO: next version implement
	Serial string `json:"serial,omitempty"`
	// TODO: next version implement
	Vendor     string      `json:"vendor,omitempty"`
	Partitions []Partition `json:"partitions,omitempty"`
	// MountPoint is the path at which the block devices is mounted.
	MountPoint string `json:"mountPoint,omitempty"`
	// InUse indicates that the block device is in use (e.g. mounted).
	InUse bool `json:"inuse"`
	// UUID is a unique identifier for the filesystem on the block device.
	//
	// This will be empty if the block device does not have a filesystem,
	// or if the filesystem is not yet known to Juju.
	//
	// The UUID format is not necessarily uniform; for example, LVM UUIDs
	// differ in format to the standard v4 UUIDs.
	UUID string `json:"uuid,omitempty"`
	// HardwareId is the block device's hardware ID, which is composed of
	// a serial number, vendor and model name. Not all block devices have
	// these properties, so HardwareId may be empty. This is used to identify
	// a block device if it is available, in preference to UUID or device
	// name, as the hardware ID is immutable.
	HardwareId string `json:"hardwareID,omitempty"`

	HealthStatus string `json:"healthStatus,omitempty"`

	Error string `json:"error,omitempty"`
}

type Partition struct {
	Name       string `json:"name"`
	Size       uint64 `json:"size,omitempty"`
	MountPoint string `json:"mountPoint,omitempty"`
	Type       string `json:"type,omitempty"`
	// UUID is a unique identifier for the filesystem on the block device.
	//
	// This will be empty if the block device does not have a filesystem,
	// or if the filesystem is not yet known to Juju.
	//
	// The UUID format is not necessarily uniform; for example, LVM UUIDs
	// differ in format to the standard v4 UUIDs.
	UUID string `json:"uuid,omitempty"`
	// InUse indicates that the block device is in use (e.g. mounted).
	InUse bool `json:"inuse"`
}

type IPMI struct {
	IP      string `json:"ip,omitempty"`
	Mac     string `json:"mac,omitempty"`
	NetMask string `json:"netmask,omitempty"`
	Gateway string `json:"gateway,omitempty"`
}

type PCIDevice struct {
	Name string `json:"name,omitempty"`
}

type ExtendSystemInfoStatus struct {
	LastUpdateTime metav1.Time `json:"lastUpdateTime"`
	Message        string      `json:"message"`
}

// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ExtendSystemInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ExtendSystemInfo `json:"items"`
}
