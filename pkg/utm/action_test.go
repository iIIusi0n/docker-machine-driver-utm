package utm

import (
	"log"
	"path/filepath"
	"testing"
)

func TestListVMs(t *testing.T) {
	vms, err := ListVMs()
	if err != nil {
		t.Fatal(err)
	}

	for _, vm := range vms {
		t.Logf("vm: %+v\n", vm)
	}
}

func TestCreateQemuVM(t *testing.T) {
	relativeIsoPath := "../../internal/driver/boot2docker.iso"
	isoPath, err := filepath.Abs(relativeIsoPath)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("isoPath: %s\n", isoPath)

	conf := &QemuConf{
		Name:         "boot2docker-test",
		Architecture: "x86_64",
		Memory:       1024,
		CPU:          1,
		UEFI:         false,
		Drives: []QemuDriveConf{
			{
				Removable: true,
				Source:    QemuDriveSource(isoPath),
			},
			{
				GuestSize: 8192,
			},
		},
		Networks: []QemuNetworkConf{
			{
				Mode: QemuNetworkModeShared,
			},
		},
	}
	vm, err := CreateQemuVM(conf)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("vm: %+v\n", vm)
}

func TestGetStatus(t *testing.T) {
	vm, err := GetVmByName("boot2docker-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("vm: %+v\n", vm)

	status, err := vm.GetStatus()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("status: %+v\n", status)

	if status != VmStatusStarted {
		err = vm.Start()
		if err != nil {
			t.Fatal(err)
		}
	}

	ip, err := vm.GetIP()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ip: %+v\n", ip)
}
