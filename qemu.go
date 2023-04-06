/*
// Copyright contributors to the Virtual Machine Manager for Go project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
*/

// Package qemu provides methods and types for launching and managing QEMU
// instances.  Instances can be launched with the LaunchQemu function and
// managed thereafter via QMPStart and the QMP object that this function
// returns.  To manage a qemu instance after it has been launched you need
// to pass the -qmp option during launch requesting the qemu instance to create
// a QMP unix domain manageent socket, e.g.,
// -qmp unix:/tmp/qmp-socket,server,nowait.  For more information see the
// example below.

package qcli

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"context"

	"gopkg.in/yaml.v2"
)

func isDimmSupported(config *Config) bool {
	switch runtime.GOARCH {
	case "amd64", "386", "ppc64le", "arm64":
		if config != nil && config.Machine.Type == MachineTypeMicrovm {
			// microvm does not support NUMA
			return false
		}
		return true
	default:
		return false
	}
}

// SMP is the multi processors configuration structure.
type SMP struct {
	// CPUs is the number of VCPUs made available to qemu.
	CPUs uint32 `yaml:"cpus"`

	// Cores is the number of cores made available to qemu.
	Cores uint32 `yaml:"cores"`

	// Threads is the number of threads made available to qemu.
	Threads uint32 `yaml:"threads"`

	// Sockets is the number of sockets made available to qemu.
	Sockets uint32 `yaml:"sockets"`

	// MaxCPUs is the maximum number of VCPUs that a VM can have.
	// This value, if non-zero, MUST BE equal to or greater than CPUs
	MaxCPUs uint32 `yaml:"max-cpus"`
}

// Memory is the guest memory configuration structure.
type Memory struct {
	// Size is the amount of memory made available to the guest.
	// It should be suffixed with M or G for sizes in megabytes or
	// gigabytes respectively.
	Size string `yaml:"size-string"`

	// Slots is the amount of memory slots made available to the guest.
	Slots uint8 `yaml:"slots"`

	// MaxMem is the maximum amount of memory that can be made available
	// to the guest through e.g. hot pluggable memory.
	MaxMem string `yaml:"max-mem-string"`

	// Path is the file path of the memory device. It points to a local
	// file path used by FileBackedMem.
	Path string `yaml:"path"`
}

// Kernel is the guest kernel configuration structure.
type Kernel struct {
	// Path is the guest kernel path on the host filesystem.
	Path string `yaml:"path"`

	// InitrdPath is the guest initrd path on the host filesystem.
	InitrdPath string `yaml:"initrd-path"`

	// Params is the kernel parameters string.
	Params string `yaml:"params-string"`
}

// Knobs regroups a set of qemu boolean settings
type Knobs struct {
	// NoUserConfig prevents qemu from loading user config files.
	NoUserConfig bool `yaml:"no-user-config"`

	// NoDefaults prevents qemu from creating default devices.
	NoDefaults bool `yaml:"no-defaults"`

	// NoGraphic completely disables graphic output.
	NoGraphic bool `yaml:"no-graphic"`

	// Daemonize will turn the qemu process into a daemon
	Daemonize bool `yaml:"daemonize"`

	// Both HugePages and MemPrealloc require the Memory.Size of the VM
	// to be set, as they need to reserve the memory upfront in order
	// for the VM to boot without errors.
	//
	// HugePages always results in memory pre-allocation.
	// However the setup is different from normal pre-allocation.
	// Hence HugePages has precedence over MemPrealloc
	// HugePages will pre-allocate all the RAM from huge pages
	HugePages bool `yaml:"hugepages"`

	// MemPrealloc will allocate all the RAM upfront
	MemPrealloc bool `yaml:"memory-preallocate"`

	// FileBackedMem requires Memory.Size and Memory.Path of the VM to
	// be set.
	FileBackedMem bool `yaml:"file-backed-memory"`

	// MemShared will set the memory device as shared.
	MemShared bool `yaml:"mem-shared"`

	// Mlock will control locking of memory
	Mlock bool `yaml:"mlock"`

	// Stopped will not start guest CPU at startup
	Stopped bool `yaml:"create-but-do-not-start"`

	// Exit instead of rebooting
	// Prevents QEMU from rebooting in the event of a Triple Fault.
	NoReboot bool `yaml:"no-reboot"`

	// Don’t exit QEMU on guest shutdown, but instead only stop the emulation.
	NoShutdown bool `yaml:"no-shutdown"`

	// IOMMUPlatform will enable IOMMU for supported devices
	IOMMUPlatform bool `yaml:"iommu-platform-enable"`

	// Disable the HPET clocksource
	NoHPET bool `yaml:"no-hpet-clocksource"`

	// Snapshot will create temporary writable disks to avoid modifying originals
	Snapshot bool `yaml:"snapshot-enable"`
}

// IOThread allows IO to be performed on a separate thread.
type IOThread struct {
	ID string `yaml:"id"`
}

const (
	// MigrationFD is the migration incoming type based on open file descriptor.
	// Skip default 0 so that it must be set on purpose.
	MigrationFD = 1
	// MigrationExec is the migration incoming type based on commands.
	MigrationExec = 2
	// MigrationDefer is the defer incoming type
	MigrationDefer = 3
)

// Incoming controls migration source preparation
type Incoming struct {
	// Possible values are MigrationFD, MigrationExec
	MigrationType int `yaml:"type"`
	// Only valid if MigrationType == MigrationFD
	FD *os.File
	// Only valid if MigrationType == MigrationExec
	Exec string `yaml:"exec"`
}

// VMConfigContainer holds a single VM config
type VMConfigContainer struct {
	VMConfig Config `yaml:"config"`
}

// Config is the qemu configuration structure.
// It allows for passing custom settings and parameters to the qemu API.
type Config struct {
	// Path is the qemu binary path.
	Path string `yaml:"qemu-binary-path"`

	// StateDir is the directory where VM state will be stored
	StateDir string `yaml:"state-dir"`

	// Ctx is the context used when launching qemu.
	Ctx context.Context

	// User ID.
	Uid uint32 `yaml:"user-id,omitempty"`
	// Group ID.
	Gid uint32 `yaml:"group-id,omitempty"`
	// Supplementary group IDs.
	Groups []uint32 `yaml:"groups,omitempty"`

	// Name is the qemu guest name
	Name string `yaml:"name"`

	// UUID is the qemu process UUID.
	UUID string `yaml:"uuid"`

	// CPUModel is the CPU model to be used by qemu.
	CPUModel string `yaml:"cpu-model"`

	// CPUModelFlags auguments the capabilities of the cpu
	CPUModelFlags []string `yaml:"cpu-model-flags"`

	// SeccompSandbox is the qemu function which enables the seccomp feature
	SeccompSandbox string `yaml:"seccomp-sandbox"`

	// Machine
	Machine Machine `yaml:"machine"`

	// SMBIOS
	SMBIOS SMBIOSInfo `yaml:"smbios"`

	// QMPSockets is a slice of QMP socket description.
	QMPSockets []QMPSocket `yaml:"qmp-sockets"`

	// Devices is a list of devices for qemu to create and drive.
	devices []Device

	RngDevices            []RngDevice            `yaml:"rng-devices"`
	BlkDevices            []BlockDevice          `yaml:"blk-devices"`
	NetDevices            []NetDevice            `yaml:"net-devices"`
	CharDevices           []CharDevice           `yaml:"char-devices"`
	LegacySerialDevices   []LegacySerialDevice   `yaml:"legacy-serial-devices"`
	SerialDevices         []SerialDevice         `yaml:"serial-devices"`
	MonitorDevices        []MonitorDevice        `yaml:"monitor-devices"`
	PCIeRootPortDevices   []PCIeRootPortDevice   `yaml:"pcie-root-port-devices"`
	UEFIFirmwareDevices   []UEFIFirmwareDevice   `yaml:"uefi-firmware-devices"`
	SCSIControllerDevices []SCSIControllerDevice `yaml:"scsi-controller-devices"`
	IDEControllerDevices  []IDEControllerDevice  `yaml:"ide-controller-devices"`
	USBControllerDevices  []USBControllerDevice  `yaml:"usb-controller-devices"`

	// RTC is the qemu Real Time Clock configuration
	RTC RTC `yaml:"real-time-clock"`

	// VGA is the qemu VGA mode.
	VGA string `yaml:"vga-mode"`

	// SpiceDevice is the qemu spice protocol device for remote display
	SpiceDevice SpiceDevice `yaml:"spice"`

	// TPMDevice is a QEMU TPM device for guest OS use
	TPM TPMDevice `yaml:"tpm"`

	// Kernel is the guest kernel configuration.
	Kernel Kernel `yaml:"kernel"`

	// Memory is the guest memory configuration.
	Memory Memory `yaml:"memory"`

	// SMP is the quest multi processors configuration.
	SMP SMP `yaml:"smp"`

	// GlobalParams is for -global parameter
	GlobalParams []string `yaml:"global-params"`

	// Knobs is a set of qemu boolean settings.
	Knobs Knobs `yaml:"qemu-knobs"`

	// Bios is the -bios parameter
	Bios string `yaml:"bios-path"`

	// PFlash specifies the parallel flash images (-pflash parameter)
	PFlash []string `yaml:"pflash-images"`

	// Incoming controls migration source preparation
	Incoming Incoming `yaml:"incoming"`

	// fds is a list of open file descriptors to be passed to the spawned qemu process
	fds []*os.File

	// FwCfg is the -fw_cfg parameter
	FwCfg []FwCfg `yaml:"firmware-config"`

	IOThreads []IOThread `yaml:"iothreads"`

	// PidFile is the -pidfile parameter
	PidFile string `yaml:"pid-file"`

	// LogFile is the -D parameter
	LogFile string `yaml:"log-file"`

	// SM-BIOS Info TBD

	pciBusSlots PCIBus

	qemuParams []string
}

// appendFDs append a list of file descriptors to the qemu configuration and
// returns a slice of offset file descriptors that will be seen by the qemu process.
func (config *Config) appendFDs(fds []*os.File) []int {
	var fdInts []int

	oldLen := len(config.fds)

	config.fds = append(config.fds, fds...)

	// The magic 3 offset comes from https://golang.org/src/os/exec/exec.go:
	//     ExtraFiles specifies additional open files to be inherited by the
	//     new process. It does not include standard input, standard output, or
	//     standard error. If non-nil, entry i becomes file descriptor 3+i.
	for i := range fds {
		fdInts = append(fdInts, oldLen+3+i)
	}

	return fdInts
}

func (config *Config) appendSeccompSandbox() {
	if config.SeccompSandbox != "" {
		config.qemuParams = append(config.qemuParams, "-sandbox")
		config.qemuParams = append(config.qemuParams, config.SeccompSandbox)
	}
}

func (config *Config) appendName() {
	if config.Name != "" {
		config.qemuParams = append(config.qemuParams, "-name")
		config.qemuParams = append(config.qemuParams, config.Name)
	}
}

// ConfigFieldName, QemuParamName, ConfigFieldValue
func getConfigOnOff(paramName, paramKey, paramVal string) string {
	if paramVal != "" {
		switch paramVal {
		case "on", "off":
			return fmt.Sprintf("%s=%s", paramKey, paramVal)
		default:
			log.Fatalf("Invalid %s value: '%s', must be one of 'on', 'off'", paramName, paramVal)
		}
	}
	return ""
}

func (config *Config) appendCPUModel() {
	if config.CPUModel != "" {
		var cpuParams []string
		cpuParams = append(cpuParams, config.CPUModel)

		if len(config.CPUModelFlags) > 0 {
			cpuParams = append(cpuParams, config.CPUModelFlags...)
		}
		config.qemuParams = append(config.qemuParams, "-cpu")
		config.qemuParams = append(config.qemuParams, strings.Join(cpuParams, ","))
	}
}

func (config *Config) appendUUID() {
	if config.UUID != "" {
		config.qemuParams = append(config.qemuParams, "-uuid")
		config.qemuParams = append(config.qemuParams, config.UUID)
	}
}

func (config *Config) appendMemory() {
	// FIXME: handle normalizing size suffix into MiB
	if config.Memory.Size != "" {
		var memoryParams []string

		memoryParams = append(memoryParams, config.Memory.Size)

		if config.Memory.Slots > 0 {
			memoryParams = append(memoryParams, fmt.Sprintf("slots=%d", config.Memory.Slots))
		}

		if config.Memory.MaxMem != "" {
			memoryParams = append(memoryParams, fmt.Sprintf("maxmem=%s", config.Memory.MaxMem))
		}

		config.qemuParams = append(config.qemuParams, "-m")
		config.qemuParams = append(config.qemuParams, strings.Join(memoryParams, ","))
	}
}

func (config *Config) appendCPUs() error {
	if config.SMP.CPUs > 0 {
		var SMPParams []string

		SMPParams = append(SMPParams, fmt.Sprintf("%d", config.SMP.CPUs))

		if config.SMP.Cores > 0 {
			SMPParams = append(SMPParams, fmt.Sprintf("cores=%d", config.SMP.Cores))
		}

		if config.SMP.Threads > 0 {
			SMPParams = append(SMPParams, fmt.Sprintf("threads=%d", config.SMP.Threads))
		}

		if config.SMP.Sockets > 0 {
			SMPParams = append(SMPParams, fmt.Sprintf("sockets=%d", config.SMP.Sockets))
		}

		if config.SMP.MaxCPUs > 0 {
			if config.SMP.MaxCPUs < config.SMP.CPUs {
				return fmt.Errorf("MaxCPUs %d must be equal to or greater than CPUs %d",
					config.SMP.MaxCPUs, config.SMP.CPUs)
			}
			SMPParams = append(SMPParams, fmt.Sprintf("maxcpus=%d", config.SMP.MaxCPUs))
		}

		config.qemuParams = append(config.qemuParams, "-smp")
		config.qemuParams = append(config.qemuParams, strings.Join(SMPParams, ","))
	}

	return nil
}

func (config *Config) appendGlobalParams() {
	if len(config.GlobalParams) > 0 {
		for _, param := range config.GlobalParams {
			config.qemuParams = append(config.qemuParams, "-global")
			config.qemuParams = append(config.qemuParams, param)
		}
	}
}

func (config *Config) appendPFlashParam() {
	for _, p := range config.PFlash {
		config.qemuParams = append(config.qemuParams, "-pflash")
		config.qemuParams = append(config.qemuParams, p)
	}
}

func (config *Config) appendVGA() {
	if config.VGA != "" {
		config.qemuParams = append(config.qemuParams, "-vga")
		config.qemuParams = append(config.qemuParams, config.VGA)
	}
}

func (config *Config) appendSpice() {
	if config.SpiceDevice.Port != "" || config.SpiceDevice.TLSPort != "" {
		config.devices = append(config.devices, config.SpiceDevice)
	}
}

func (config *Config) appendTPM() {
	if config.TPM.ID != "" {
		config.devices = append(config.devices, config.TPM)
	}
}

func (config *Config) appendKernel() {
	if config.Kernel.Path != "" {
		config.qemuParams = append(config.qemuParams, "-kernel")
		config.qemuParams = append(config.qemuParams, config.Kernel.Path)

		if config.Kernel.InitrdPath != "" {
			config.qemuParams = append(config.qemuParams, "-initrd")
			config.qemuParams = append(config.qemuParams, config.Kernel.InitrdPath)
		}

		if config.Kernel.Params != "" {
			config.qemuParams = append(config.qemuParams, "-append")
			config.qemuParams = append(config.qemuParams, config.Kernel.Params)
		}
	}
}

func (config *Config) appendMemoryKnobs() {
	if config.Memory.Size == "" {
		return
	}
	var objMemParam, numaMemParam string
	dimmName := "dimm1"
	if config.Knobs.HugePages {
		objMemParam = "memory-backend-file,id=" + dimmName + ",size=" + config.Memory.Size + ",mem-path=/dev/hugepages"
		numaMemParam = "node,memdev=" + dimmName
	} else if config.Knobs.FileBackedMem && config.Memory.Path != "" {
		objMemParam = "memory-backend-file,id=" + dimmName + ",size=" + config.Memory.Size + ",mem-path=" + config.Memory.Path
		numaMemParam = "node,memdev=" + dimmName
	} else {
		objMemParam = "memory-backend-ram,id=" + dimmName + ",size=" + config.Memory.Size
		numaMemParam = "node,memdev=" + dimmName
	}

	if config.Knobs.MemShared {
		objMemParam += ",share=on"
	}
	if config.Knobs.MemPrealloc {
		objMemParam += ",prealloc=on"
	}
	config.qemuParams = append(config.qemuParams, "-object")
	config.qemuParams = append(config.qemuParams, objMemParam)

	if isDimmSupported(config) {
		config.qemuParams = append(config.qemuParams, "-numa")
		config.qemuParams = append(config.qemuParams, numaMemParam)
	} else {
		config.qemuParams = append(config.qemuParams, "-machine")
		config.qemuParams = append(config.qemuParams, "memory-backend="+dimmName)
	}
}

func (config *Config) appendKnobs() {

	config.appendMemoryKnobs()

	if config.Knobs.NoUserConfig {
		config.qemuParams = append(config.qemuParams, "-no-user-config")
	}

	if config.Knobs.NoDefaults {
		config.qemuParams = append(config.qemuParams, "-nodefaults")
	}

	if config.Knobs.NoGraphic {
		config.qemuParams = append(config.qemuParams, "-nographic")
	}

	if config.Knobs.NoReboot {
		config.qemuParams = append(config.qemuParams, "--no-reboot")
	}

	if config.Knobs.NoShutdown {
		config.qemuParams = append(config.qemuParams, "--no-shutdown")
	}

	if config.Knobs.Daemonize {
		config.qemuParams = append(config.qemuParams, "-daemonize")
	}

	if config.Knobs.Mlock {
		config.qemuParams = append(config.qemuParams, "-overcommit")
		config.qemuParams = append(config.qemuParams, "mem-lock=on")
	}

	if config.Knobs.Stopped {
		config.qemuParams = append(config.qemuParams, "-S")
	}

	if config.Knobs.NoHPET {
		config.qemuParams = append(config.qemuParams, "-no-hpet")
	}

	if config.Knobs.Snapshot {
		config.qemuParams = append(config.qemuParams, "-snapshot")
	}
}

func (config *Config) appendBios() {
	if config.Bios != "" {
		config.qemuParams = append(config.qemuParams, "-bios")
		config.qemuParams = append(config.qemuParams, config.Bios)
	}
}

func (config *Config) appendIOThreads() {
	for _, t := range config.IOThreads {
		if t.ID != "" {
			config.qemuParams = append(config.qemuParams, "-object")
			config.qemuParams = append(config.qemuParams, fmt.Sprintf("iothread,id=%s", t.ID))
		}
	}
}

func (config *Config) appendIncoming() {
	var uri string
	switch config.Incoming.MigrationType {
	case MigrationExec:
		uri = fmt.Sprintf("exec:%s", config.Incoming.Exec)
	case MigrationFD:
		chFDs := config.appendFDs([]*os.File{config.Incoming.FD})
		uri = fmt.Sprintf("fd:%d", chFDs[0])
	case MigrationDefer:
		uri = "defer"
	default:
		return
	}
	config.qemuParams = append(config.qemuParams, "-S", "-incoming", uri)
}

func (config *Config) appendPidFile() {
	if config.PidFile != "" {
		config.qemuParams = append(config.qemuParams, "-pidfile")
		config.qemuParams = append(config.qemuParams, config.PidFile)
	}
}

func (config *Config) appendLogFile() {
	if config.LogFile != "" {
		config.qemuParams = append(config.qemuParams, "-D")
		config.qemuParams = append(config.qemuParams, config.LogFile)
	}
}

// GetSocketPaths seaches config for Chardev,Serial,Monitor and QMP sockets
func GetSocketPaths(config *Config) ([]string, error) {
	var sockets []string

	for _, cdev := range config.CharDevices {
		if cdev.Backend == Socket {
			sockets = append(sockets, cdev.Path)
		}
	}

	for _, mdev := range config.MonitorDevices {
		if mdev.Backend == Socket {
			sockets = append(sockets, mdev.Path)
		}
	}

	for _, ldev := range config.LegacySerialDevices {
		if ldev.Backend == Socket {
			sockets = append(sockets, ldev.Path)
		}
	}

	for _, qdev := range config.QMPSockets {
		if qdev.Type == Unix {
			sockets = append(sockets, qdev.Name)
		}
	}

	return sockets, nil
}

func ConfigureParams(config *Config, logger QMPLog) ([]string, error) {
	var err error
	if logger == nil {
		logger = qmpNullLogger{}
	}
	config.appendName()
	config.appendUUID()
	config.appendMachine()
	config.appendCPUModel()
	config.appendSpice()
	config.appendTPM()
	if err := config.appendSMBIOSInfo(); err != nil {
		return []string{}, err
	}
	err = config.appendQMPSockets()
	if err != nil {
		return []string{}, err
	}
	config.appendMemory()
	err = config.appendDevices()
	if err != nil {
		return []string{}, err
	}
	config.appendRTC()
	config.appendGlobalParams()
	config.appendPFlashParam()
	config.appendVGA()
	config.appendKnobs()
	config.appendKernel()
	config.appendBios()
	config.appendIOThreads()
	config.appendIncoming()
	config.appendPidFile()
	config.appendLogFile()
	config.appendFwCfg(logger)
	config.appendSeccompSandbox()

	if err := config.appendCPUs(); err != nil {
		return []string{}, err
	}

	return config.qemuParams, nil
}

func ReadConfig(configFile string) (*Config, error) {
	content, err := ioutil.ReadFile(configFile)

	if err != nil {
		return nil, fmt.Errorf("Failed to read config file '%s':%s", configFile, err)
	}

	return UnmarshalConfig(content)
}

func WriteConfig(configFile string, config *Config) error {

	content, err := MarshalConfig(config)
	if err != nil {
		return fmt.Errorf("Failed to marshal qcli.Config: %s", err)
	}

	return ioutil.WriteFile(configFile, content, 0644)
}

func MarshalConfig(config *Config) ([]byte, error) {
	content, err := yaml.Marshal(config)
	if err != nil {
		return []byte{}, err
	}
	return content, nil
}

func UnmarshalConfig(content []byte) (*Config, error) {
	var cfg Config
	err := yaml.Unmarshal(content, &cfg)
	return &cfg, err
}

// LaunchQemu can be used to launch a new qemu instance.
//
// The Config parameter contains a set of qemu parameters and settings.
//
// This function writes its log output via logger parameter.
//
// The function will block until the launched qemu process exits.  "", nil
// will be returned if the launch succeeds.  Otherwise a string containing
// the contents of stderr + a Go error object will be returned.
func LaunchQemu(config *Config, logger QMPLog) (string, error) {

	if _, err := ConfigureParams(config, logger); err != nil {
		return "", err
	}

	if len(config.qemuParams) == 0 {
		return "", fmt.Errorf("Failed to configure qemu parameters")
	}

	ctx := config.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	attr := syscall.SysProcAttr{}
	attr.Credential = &syscall.Credential{
		Uid:    config.Uid,
		Gid:    config.Gid,
		Groups: config.Groups,
	}
	logger.Infof("Running VM as: uid=%d gid=%d", config.Uid, config.Gid)

	return LaunchCustomQemu(ctx, config.Path, config.qemuParams,
		config.fds, &attr, logger)
}

// LaunchCustomQemu can be used to launch a new qemu instance.
//
// The path parameter is used to pass the qemu executable path.
//
// params is a slice of options to pass to qemu-system-x86_64 and fds is a
// list of open file descriptors that are to be passed to the spawned qemu
// process.  The attrs parameter can be used to control aspects of the
// newly created qemu process, such as the user and group under which it
// runs.  It may be nil.
//
// This function writes its log output via logger parameter.
//
// The function will block until the launched qemu process exits.  "", nil
// will be returned if the launch succeeds.  Otherwise a string containing
// the contents of stderr + a Go error object will be returned.
func LaunchCustomQemu(ctx context.Context, path string, params []string, fds []*os.File,
	attr *syscall.SysProcAttr, logger QMPLog) (string, error) {
	if logger == nil {
		logger = qmpNullLogger{}
	}

	errStr := ""

	if path == "" {
		path = "qemu-system-x86_64"
	}

	/* #nosec */
	cmd := exec.CommandContext(ctx, path, params...)
	if len(fds) > 0 {
		logger.Infof("Adding extra file %v", fds)
		cmd.ExtraFiles = fds
	}

	// FIXME: non-root user can't run with this set?
	// cmd.SysProcAttr = attr

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	logger.Infof("launching %s with: %v", path, params)

	err := cmd.Run()
	if err != nil {
		logger.Errorf("Unable to launch %s: %v", path, err)
		errStr = stderr.String()
		logger.Errorf("%s", errStr)
	}
	logger.Infof("LaunchCustomQemu returns")
	return errStr, err
}
