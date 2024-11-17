package utm

import (
	"docker-machine-driver-utm/pkg/applescript"
	"fmt"
	"strings"
)

func ListVMs() ([]*VM, error) {
	res, err := runUtmScript(
		`set vms to virtual machines`,
		`set output to ""`,
		`repeat with vm in vms`,
		`	set output to output & id of vm & "|#|" & name of vm & "|#|" & backend of vm & "|#|" & status of vm & "|&|"`,
		`end repeat`,
		`return output`,
	)
	if err != nil {
		return nil, err
	}

	res = strings.TrimSuffix(res, "|&|")
	lines := strings.Split(res, "|&|")
	vms := make([]*VM, len(lines))
	for i, line := range lines {
		fields := strings.Split(line, "|#|")
		if len(fields) != 4 {
			continue
		}

		vms[i] = &VM{ID: fields[0], Name: fields[1], Backend: VmBackend(fields[2]), Status: VmStatus(fields[3])}
	}
	return vms, nil
}

func GetVmByID(id string) (*VM, error) {
	vms, err := ListVMs()
	if err != nil {
		return nil, err
	}
	for _, vm := range vms {
		if vm.ID == id {
			return vm, nil
		}
	}
	return nil, fmt.Errorf("vm not found")
}

func GetVmByName(name string) (*VM, error) {
	vms, err := ListVMs()
	if err != nil {
		return nil, err
	}
	for _, vm := range vms {
		if vm.Name == name {
			return vm, nil
		}
	}
	return nil, fmt.Errorf("vm not found")
}

func (vm *VM) Start() error {
	_, err := runUtmScript(
		fmt.Sprintf(`set vm to virtual machine id "%s"`, vm.ID),
		`start vm`,
	)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VM) StartDisposable() error {
	_, err := runUtmScript(
		fmt.Sprintf(`set vm to virtual machine id "%s"`, vm.ID),
		`start vm without saving`,
	)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VM) Pause() error {
	_, err := runUtmScript(
		fmt.Sprintf(`set vm to virtual machine id "%s"`, vm.ID),
		`suspend vm`,
	)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VM) PauseAndSave() error {
	_, err := runUtmScript(
		fmt.Sprintf(`set vm to virtual machine id "%s"`, vm.ID),
		`suspend vm with saving`,
	)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VM) Stop() error {
	_, err := runUtmScript(
		fmt.Sprintf(`set vm to virtual machine id "%s"`, vm.ID),
		`stop vm`,
	)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VM) Shutdown() error {
	_, err := runUtmScript(
		fmt.Sprintf(`set vm to virtual machine id "%s"`, vm.ID),
		`stop vm by force`,
	)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VM) Kill() error {
	_, err := runUtmScript(
		fmt.Sprintf(`set vm to virtual machine id "%s"`, vm.ID),
		`stop vm by kill`,
	)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VM) GetStatus() (VmStatus, error) {
	res, err := runUtmScript(
		fmt.Sprintf(`set vm to virtual machine id "%s"`, vm.ID),
		`return status of vm`,
	)
	if err != nil {
		return "", err
	}
	return VmStatus(res), nil
}

func (vm *VM) GetIP() (string, error) {
	res, err := runUtmScript(
		fmt.Sprintf(`set vm to virtual machine id "%s"`, vm.ID),
		`return item 1 of (query ip of vm)`,
	)
	if err != nil {
		return "", err
	}
	return res, nil
}

func CreateQemuVM(conf *QemuConf) (*VM, error) {
	type sourceFiles struct {
		name string
		path string
	}
	files := []sourceFiles{}
	for i, drive := range conf.Drives {
		if drive.Source != "" {
			name := fmt.Sprintf("drive%d", i)
			files = append(files, sourceFiles{name: name, path: string(drive.Source)})
			conf.Drives[i].Source = QemuDriveSource(name)
		}
	}
	cmds := []string{}
	for _, file := range files {
		cmds = append(cmds, fmt.Sprintf(`set %s to POSIX file "%s"`, file.name, file.path))
	}

	res, err := applescript.Marshal(conf)
	if err != nil {
		return nil, err
	}
	output, err := runUtmScript(
		strings.Join(cmds, "\n"),
		fmt.Sprintf(
			`set vm to make new virtual machine with properties {backend: qemu, configuration: %s}`, string(res)),
		`set output to ""`,
		`set output to id of vm & "|#|" & name of vm & "|#|" & backend of vm & "|#|" & status of vm`,
		`return output`,
	)
	if err != nil {
		return nil, err
	}
	fields := strings.Split(output, "|#|")
	if len(fields) != 4 {
		return nil, fmt.Errorf("invalid response: %s", output)
	}

	return &VM{
		ID:      fields[0],
		Name:    fields[1],
		Backend: VmBackend(fields[2]),
		Status:  VmStatus(fields[3]),
	}, nil
}

func DeleteVmByID(id string) error {
	_, err := runUtmScript(
		fmt.Sprintf(`delete virtual machine id "%s"`, id),
	)
	return err
}

func DeleteVmByName(name string) error {
	_, err := runUtmScript(
		fmt.Sprintf(`delete virtual machine named "%s"`, name),
	)
	return err
}

func CopyToVM(vm *VM, src, dst string) error {
	_, err := runUtmScript(
		fmt.Sprintf(`set vm to virtual machine id "%s"`, vm.ID),
		fmt.Sprintf(`set input to POSIX file "%s"`, src),
		fmt.Sprintf(`push of (open file of vm at "%s" for writing) from input`, dst),
	)
	return err
}

func RunCommandOnVM(vm *VM, cmd string, args ...string) error {
	argStr := "{"
	for _, arg := range args {
		argStr += fmt.Sprintf("\"%s\", ", arg)
	}
	argStr = strings.TrimSuffix(argStr, ", ")
	argStr += "}"
	_, err := runUtmScript(
		fmt.Sprintf(`set vm to virtual machine id "%s"`, vm.ID),
		fmt.Sprintf(`execute of vm at "%s" with arguments %s`, cmd, argStr),
	)
	return err
}
