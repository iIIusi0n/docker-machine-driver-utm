package utm

type VmBackend string
type VmStatus string
type DirectoryShareMode string
type QemuDriveInterface string
type QemuDriveSource string
type QemuNetworkMode string
type QemuPortForwardingProtocol string

const (
	VmBackendApple      VmBackend = "apple"
	VmBackendQemu       VmBackend = "qemu"
	VmBackendUnavilable VmBackend = "unavilable"
)

const (
	VmStatusStopped  VmStatus = "stopped"
	VmStatusStarting VmStatus = "starting"
	VmStatusStarted  VmStatus = "started"
	VmStatusPausing  VmStatus = "pausing"
	VmStatusPaused   VmStatus = "paused"
	VmStatusResuming VmStatus = "resuming"
	VmStatusStopping VmStatus = "stopping"
)

const (
	DirectoryShareModeNone   DirectoryShareMode = "none"
	DirectoryShareModeWebDAV DirectoryShareMode = "WebDAV"
	DirectoryShareModeVirtFS DirectoryShareMode = "VirtFS"
)

const (
	QemuDriveInterfaceNone   QemuDriveInterface = "none"
	QemuDriveInterfaceIDE    QemuDriveInterface = "IDE"
	QemuDriveInterfaceSCSI   QemuDriveInterface = "SCSI"
	QemuDriveInterfaceSD     QemuDriveInterface = "SD"
	QemuDriveInterfaceMTD    QemuDriveInterface = "MTD"
	QemuDriveInterfaceFloppy QemuDriveInterface = "Floppy"
	QemuDriveInterfacePFlash QemuDriveInterface = "PFlash"
	QemuDriveInterfaceVirtIO QemuDriveInterface = "VirtIO"
	QemuDriveInterfaceNVMe   QemuDriveInterface = "NVMe"
	QemuDriveInterfaceUSB    QemuDriveInterface = "USB"
)

const (
	QemuNetworkModeEmulated QemuNetworkMode = "emulated"
	QemuNetworkModeShared   QemuNetworkMode = "shared"
	QemuNetworkModeHost     QemuNetworkMode = "host"
	QemuNetworkModeBridged  QemuNetworkMode = "bridged"
)

const (
	QemuPortForwardingProtocolTCP QemuPortForwardingProtocol = "TCP"
	QemuPortForwardingProtocolUDP QemuPortForwardingProtocol = "UDP"
)

type VM struct {
	ID      string    `applescript:"id"`
	Name    string    `applescript:"name"`
	Backend VmBackend `applescript:"backend"`
	Status  VmStatus  `applescript:"status"`
}

type QemuConf struct {
	Name           string             `applescript:"name"`
	Notes          string             `applescript:"notes"`
	Architecture   string             `applescript:"architecture"`
	Memory         int                `applescript:"memory"`
	CPU            int                `applescript:"cpu cores"`
	UEFI           bool               `applescript:"uefi"`
	DirectoryShare DirectoryShareMode `applescript:"directory share mode"`
	Drives         []QemuDriveConf    `applescript:"drives"`
	Networks       []QemuNetworkConf  `applescript:"network interfaces"`
}

type QemuDriveConf struct {
	ID        string             `applescript:"id"`
	Removable bool               `applescript:"removable"`
	Interface QemuDriveInterface `applescript:"interface"`
	HostSize  int                `applescript:"host size"`
	GuestSize int                `applescript:"guest size"`
	Raw       bool               `applescript:"raw"`
	Source    QemuDriveSource    `applescript:"source"`
}

type QemuNetworkConf struct {
	Index          int                      `applescript:"index"`
	Hardware       string                   `applescript:"hardware"`
	Mode           QemuNetworkMode          `applescript:"mode"`
	MAC            string                   `applescript:"address"`
	HostInterface  string                   `applescript:"host interface"`
	PortForwarding []QemuPortForwardingConf `applescript:"port forwarding"`
}

type QemuPortForwardingConf struct {
	Protocol  QemuPortForwardingProtocol `applescript:"protocol"`
	HostAddr  string                     `applescript:"host address"`
	HostPort  int                        `applescript:"host port"`
	GuestAddr string                     `applescript:"guest address"`
	GuestPort int                        `applescript:"guest port"`
}
