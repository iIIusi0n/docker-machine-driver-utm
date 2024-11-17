package utm

import (
	"docker-machine-driver-utm/pkg/applescript"
	"testing"
)

func TestMarshal(t *testing.T) {
	vm := &VM{
		ID: "123",
	}

	res, err := applescript.Marshal(vm)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))

	conf := &QemuConf{
		Name:         "boot2docker",
		Architecture: "x86_64",
		Memory:       1024,
		CPU:          1,
		UEFI:         false,
		Drives: []QemuDriveConf{
			{
				Removable: true,
				Source:    "/tmp/boot2docker.iso",
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

	res, err = applescript.Marshal(conf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(res))
}
