package driver

import (
	"archive/tar"
	"bytes"
	"docker-machine-driver-utm/pkg/utm"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"

	_ "embed"
)

const (
	IsoFilename    = "boot2docker.iso"
	B2dURL         = "https://github.com/iIIusi0n/docker-machine-driver-utm/releases/download/v1.0.0/boot2docker.iso"
	DefaultSSHUser = "docker"
)

type Driver struct {
	*drivers.BaseDriver

	Memory         int
	Disk           int
	CPU            int
	Network        string
	HostInterface  string
	Boot2DockerURL string
	ISO            string
	DiskPath       string
	VM             *utm.VM
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     DefaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) Create() error {
	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Infof("Machine path: %s", d.ResolveStorePath("."))
	if err := os.MkdirAll(d.ResolveStorePath("."), 0755); err != nil {
		return err
	}

	if err := downloadBoot2DockerISO(d.ResolveStorePath(IsoFilename)); err != nil {
		return err
	}

	log.Infof("Generating disk image...")
	if err := d.generateDiskImage(d.Disk); err != nil {
		return err
	}

	conf := &utm.QemuConf{
		Name:         fmt.Sprintf("docker-machine-%s", d.MachineName),
		Architecture: "x86_64",
		Memory:       d.Memory,
		CPU:          d.CPU,
		UEFI:         false,
		Drives: []utm.QemuDriveConf{
			{
				Removable: true,
				Source:    utm.QemuDriveSource(d.ResolveStorePath(IsoFilename)),
			},
			{
				Raw:       true,
				Interface: utm.QemuDriveInterfaceIDE,
				Source:    utm.QemuDriveSource(d.ResolveStorePath(fmt.Sprintf("%s.img", d.MachineName))),
			},
		},
		Networks: []utm.QemuNetworkConf{
			{
				Mode: utm.QemuNetworkMode(d.Network),
			},
		},
	}
	if conf.Networks[0].Mode == utm.QemuNetworkModeBridged {
		conf.Networks[0].HostInterface = d.HostInterface
	}

	vm, err := utm.CreateQemuVM(conf)
	if err != nil {
		return err
	}
	d.VM = vm

	return d.Start()
}

func (d *Driver) DriverName() string {
	return "utm"
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.IntFlag{
			Name:  "utm-memory",
			Usage: "Memory size for the UTM VM in MB",
			Value: 1024,
		},
		mcnflag.IntFlag{
			Name:  "utm-disk",
			Usage: "Disk size for the UTM VM in MB",
			Value: 8192,
		},
		mcnflag.IntFlag{
			Name:  "utm-cpu",
			Usage: "Number of CPUs for the UTM VM",
			Value: 1,
		},
		mcnflag.StringFlag{
			Name:  "utm-network",
			Usage: "Network type for the UTM VM (emulated, shared, host, bridged)",
			Value: "shared",
		},
		mcnflag.StringFlag{
			Name:  "utm-host-interface",
			Usage: "Host interface for the UTM VM (bridged mode)",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "utm-boot2docker-url",
			Usage: "URL to the boot2docker ISO",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "utm-ssh-user",
			Usage: "SSH user for the UTM VM",
			Value: DefaultSSHUser,
		},
	}
}

func (d *Driver) GetIP() (string, error) {
	if err := d.validateVM(); err != nil {
		return "", err
	}

	sta, err := d.VM.GetStatus()
	if err != nil {
		return "", err
	}

	if sta == utm.VmStatusStarted {
		return d.VM.GetIP()
	}

	return "", nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	return d.ResolveStorePath("id_rsa")
}

func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}
	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = DefaultSSHUser
	}
	return d.SSHUser
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetState() (state.State, error) {
	if err := d.validateVM(); err != nil {
		return state.None, err
	}

	sta, err := d.VM.GetStatus()
	if err != nil {
		return state.None, err
	}

	ip, err := d.GetIP()
	if err != nil {
		return state.None, err
	}

	switch sta {
	case utm.VmStatusStopped, utm.VmStatusPaused:
		return state.Stopped, nil
	case utm.VmStatusStarting:
		return state.Starting, nil
	case utm.VmStatusStarted:
		if ip == "" {
			return state.Starting, nil
		}
		return state.Running, nil
	case utm.VmStatusResuming:
		return state.Starting, nil
	case utm.VmStatusStopping, utm.VmStatusPausing:
		return state.Stopping, nil
	default:
		return state.None, nil
	}
}

func (d *Driver) Kill() error {
	if err := d.validateVM(); err != nil {
		return err
	}
	return d.VM.Kill()
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Remove() error {
	err := d.VM.Stop()
	if err != nil {
		log.Infof("Error while stopping VM: %v", err)
	}

	time.Sleep(1 * time.Second)

	log.Infof("Removing UTM VM...")
	err = d.validateVM()
	if err != nil {
		log.Infof("Error while validating VM: %v", err)
	}

	err = utm.DeleteVmByID(d.VM.ID)
	if err != nil {
		log.Infof("Error while deleting VM: %v", err)
	}

	log.Infof("VM removed successfully")
	return nil
}

func (d *Driver) Restart() error {
	if err := d.validateVM(); err != nil {
		return err
	}

	err := d.Stop()
	if err != nil {
		return err
	}

	return d.Start()
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Memory = flags.Int("utm-memory")
	d.Disk = flags.Int("utm-disk")
	d.CPU = flags.Int("utm-cpu")
	d.Network = flags.String("utm-network")
	d.HostInterface = flags.String("utm-host-interface")
	d.Boot2DockerURL = flags.String("utm-boot2docker-url")

	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.ISO = d.ResolveStorePath(IsoFilename)
	d.SSHUser = flags.String("utm-ssh-user")
	d.SSHPort = 22
	d.DiskPath = d.ResolveStorePath(fmt.Sprintf("%s.img", d.MachineName))
	return nil
}

func (d *Driver) Start() error {
	log.Infof("Starting UTM VM...")
	if err := d.validateVM(); err != nil {
		return err
	}

	err := d.VM.Start()
	if err != nil {
		return err
	}

	log.Infof("Waiting for VM to get an IP address...")
	for i := 0; i < 120; i++ {
		ip, err := d.GetIP()
		if err == nil && ip != "" {
			log.Infof("VM started successfully with IP: %s", ip)
			return nil
		}
		time.Sleep(5 * time.Second)
	}

	return errors.New("timeout waiting for VM to get an IP address")
}

func (d *Driver) Stop() error {
	log.Infof("Stopping UTM VM...")
	if err := d.validateVM(); err != nil {
		return err
	}
	err := d.VM.Pause()
	if err != nil {
		return err
	}
	log.Infof("VM stopped successfully")
	return nil
}

func (d *Driver) validateVM() error {
	if d.VM == nil {
		vmName := fmt.Sprintf("docker-machine-%s", d.MachineName)
		vm, err := utm.GetVmByName(vmName)
		if err != nil {
			return err
		}
		d.VM = vm
	}

	return nil
}

func (d *Driver) generateDiskImage(size int) error {
	log.Debugf("Creating %d MB hard disk image...", size)

	magicString := "boot2docker, please format-me"

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	file := &tar.Header{Name: magicString, Size: int64(len(magicString))}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(magicString)); err != nil {
		return err
	}

	file = &tar.Header{Name: ".ssh", Typeflag: tar.TypeDir, Mode: 0700}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	pubKey, err := os.ReadFile(d.GetSSHKeyPath() + ".pub")
	if err != nil {
		return err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys2", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	raw := bytes.NewReader(buf.Bytes())
	return createDiskImage(d.DiskPath, size, raw)
}

func createDiskImage(dest string, size int, r io.Reader) error {
	sizeBytes := int64(size) << 20
	f, err := os.Create(dest)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}
	f.Seek(sizeBytes-1, 0)
	f.Write([]byte{0})
	return f.Close()
}

func downloadBoot2DockerISO(path string) error {
	resp, err := http.Get(B2dURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
