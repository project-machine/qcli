package main

import (
	"fmt"
	"os"
	"path"
	"qcli"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type QemuDisk struct {
	File      string `yaml:"file"`
	Format    string `yaml:"format"`
	Size      string `yaml:"size"`
	Attach    string `yaml:"attach"`
	Type      string `yaml:"type"`
	BlockSize int    `yaml:"blocksize"`
	BusAddr   string `yaml:"addr"`
	BootIndex string `yaml:"bootindex"`
	ReadOnly  bool   `yaml:"read-only"`
	Serial    string `yaml:"serial"`
}

func (q *QemuDisk) Sanitize(basedir string) error {
	validate := func(name string, found string, valid ...string) string {
		for _, i := range valid {
			if found == i {
				return ""
			}
		}
		return fmt.Sprintf("invalid %s: found %s expected %v", name, found, valid)
	}

	errors := []string{}

	if q.Format == "" {
		q.Format = "qcow2"
	}

	if q.Type == "" {
		q.Type = "ssd"
	}

	if q.Attach == "" {
		q.Attach = "scsi"
	}

	if q.File == "" {
		errors = append(errors, "empty File")
	}

	if !strings.Contains(q.File, "/") {
		q.File = path.Join(basedir, q.File)
	}

	if msg := validate("format", q.Format, "qcow2", "raw"); msg != "" {
		errors = append(errors, msg)
	}

	if msg := validate("attach", q.Attach, "scsi", "nvme", "virtio", "ide", "usb"); msg != "" {
		errors = append(errors, msg)
	}

	if msg := validate("type", q.Type, "hdd", "ssd", "cdrom"); msg != "" {
		errors = append(errors, msg)
	}

	if len(errors) != 0 {
		return fmt.Errorf("bad disk %#v: %s", q, strings.Join(errors, "\n"))
	}

	return nil
}

type VMDef struct {
	Name         string     `yaml:"name"`
	Cpus         uint32     `yaml:"cpus"`
	Memory       uint32     `yaml:"memory"`
	Serial       string     `yaml:"serial"`
	Nics         []NicDef   `yaml:"nics"`
	Disks        []QemuDisk `yaml:"disks"`
	Boot         string     `yaml:"boot"`
	Cdrom        string     `yaml:"cdrom"`
	UefiVars     string     `yaml:"uefi-vars"`
	TPM          bool       `yaml:"tpm"`
	TPMVersion   string     `yaml:"tpm-version"`
	KVMExtraOpts []string   `yaml:"extra-opts"`
	SecureBoot   bool       `yaml:"secure-boot"`
	Gui          bool       `yaml:"gui"`
}

type NicDef struct {
	BusAddr   string     `yaml:"addr"`
	Device    string     `yaml:"device"`
	ID        string     `yaml:"id"`
	Mac       string     `yaml:"mac"`
	Ports     []PortRule `yaml:"ports"`
	BootIndex string     `yaml:"bootindex"`
}

type VMNic struct {
	BusAddr    string
	DeviceType string
	HWAddr     string
	ID         string
	IFName     string
	NetIFName  string
	NetType    string
	NetAddr    string
	BootIndex  string
	Ports      []PortRule
}

type PortRule struct {
	Protocol string
	Host     Port
	Guest    Port
}

type Port struct {
	Address string
	Port    int
}

type Cluster struct {
	Machines []VMDef `yaml:"vms"`
}

func (p *PortRule) UnmarshalYAML(unmarshal func(interface{}) error) error {
	DefaultPortProtocol := "tcp"
	DefaultPortHostAddress := ""
	DefaultPortGuestAddress := ""
	var ruleVal map[string]string
	var err error

	if err = unmarshal(&ruleVal); err != nil {
		return err
	}

	for hostVal, guestVal := range ruleVal {
		hostToks := strings.Split(hostVal, ":")
		if len(hostToks) == 3 {
			p.Protocol = hostToks[0]
			p.Host.Address = hostToks[1]
			p.Host.Port, err = strconv.Atoi(hostToks[2])
			if err != nil {
				return err
			}
		} else if len(hostToks) == 2 {
			p.Protocol = DefaultPortProtocol
			p.Host.Address = hostToks[0]
			p.Host.Port, err = strconv.Atoi(hostToks[1])
			if err != nil {
				return err
			}
		} else {
			p.Protocol = DefaultPortProtocol
			p.Host.Address = DefaultPortHostAddress
			p.Host.Port, err = strconv.Atoi(hostToks[0])
			if err != nil {
				return err
			}
		}
		guestToks := strings.Split(guestVal, ":")
		if len(guestToks) == 2 {
			p.Guest.Address = guestToks[0]
			p.Guest.Port, err = strconv.Atoi(guestToks[1])
			if err != nil {
				return err
			}
		} else {
			p.Guest.Address = DefaultPortGuestAddress
			p.Guest.Port, err = strconv.Atoi(guestToks[0])
			if err != nil {
				return err
			}
		}
		break
	}
	if p.Protocol != "tcp" && p.Protocol != "udp" {
		return fmt.Errorf("Invalid PortRule.Protocol value: %s . Must be 'tcp' or 'udp'", p.Protocol)
	}
	return nil
}

func (p *PortRule) String() string {
	return fmt.Sprintf("%s:%s:%d-%s:%d", p.Protocol,
		p.Host.Address, p.Host.Port, p.Guest.Address, p.Guest.Port)
}

var QemuTypeIndex map[string]int

// Allocate the next number per Qemu Type string
// This is use to create unique, increasing index integers used to
// enumerate qemu id= parameters used to bind various objects together
// on the QEMU command line: e.g
//
// -object iothread,id=iothread2
// -drive id=drv1
// -device scsi-hd,drive=drv1,iothread=iothread2
func getNextQemuIndex(qtype string) int {
	currentIndex := 0
	ok := false
	if QemuTypeIndex == nil {
		QemuTypeIndex = make(map[string]int)
	}
	if currentIndex, ok = QemuTypeIndex[qtype]; !ok {
		currentIndex = -1
	}
	QemuTypeIndex[qtype] = currentIndex + 1
	return QemuTypeIndex[qtype]
}

func clearAllQemuIndex() {
	for key := range QemuTypeIndex {
		delete(QemuTypeIndex, key)
	}
}

func NewDefaultConfig(name string, numCpus, numMemMB uint32) (*qcli.Config, error) {
	smp := qcli.SMP{CPUs: numCpus}
	if numCpus < 1 {
		smp.CPUs = 4
	}

	mem := qcli.Memory{
		Size: fmt.Sprintf("%dm", numMemMB),
	}
	if numMemMB < 1 {
		mem.Size = "4096m"
	}

	c := &qcli.Config{
		Name: name,
		Machine: qcli.Machine{
			Type:         qcli.MachineTypePC35,
			Acceleration: qcli.MachineAccelerationKVM,
			SMM:          "on",
		},
		CPUModel:      "qemu64",
		CPUModelFlags: []string{"+x2apic"},
		SMP:           smp,
		Memory:        mem,
		RngDevices: []qcli.RngDevice{
			qcli.RngDevice{
				Driver:    qcli.VirtioRng,
				ID:        "rng0",
				Bus:       "pcie.0",
				Addr:      "3",
				Transport: qcli.TransportPCI,
				Filename:  qcli.RngDevUrandom,
			},
		},
		GlobalParams: []string{
			"ICH9-LPC.disable_s3=1",
			"driver=cfi.pflash01,property=secure,value=on",
		},
	}

	return c, nil
}

func (qd QemuDisk) QBlockDevice() (qcli.BlockDevice, error) {
	blk := qcli.BlockDevice{
		// Driver
		ID:   fmt.Sprintf("drive%d", getNextQemuIndex("drive")),
		File: qd.File,
		// Interface
		AIO:       qcli.Threads,
		BlockSize: qd.BlockSize,
		BusAddr:   qd.BusAddr,
		BootIndex: qd.BootIndex,
		ReadOnly:  qd.ReadOnly,
	}

	if qd.Format != "" {
		switch qd.Format {
		case "raw":
			blk.Format = qcli.RAW
		case "qcow2":
			blk.Format = qcli.QCOW2
		}
	} else {
		blk.Format = qcli.QCOW2
	}

	if qd.Attach == "" {
		qd.Attach = "virtio"
	}

	switch qd.Attach {
	case "scsi":
		blk.Driver = qcli.SCSIHD
	case "nvme":
		blk.Driver = qcli.NVME
	case "virtio":
		blk.Driver = qcli.VirtioBlock
	case "ide":
		if qd.Type == "cdrom" {
			blk.Driver = qcli.IDECDROM
			blk.Media = "cdrom"
		} else {
			blk.Driver = qcli.IDEHardDisk
		}
	case "usb":
		blk.Driver = qcli.USBStorage
	default:
		return blk, fmt.Errorf("Unknown Disk Attach type: %s", qd.Attach)
	}

	return blk, nil
}

func (nd NicDef) QNetDevice() (qcli.NetDevice, error) {
	//FIXME: how do we do bridge or socket/mcast types?
	ndev := qcli.NetDevice{
		Type:       qcli.USER,
		ID:         nd.ID,
		Addr:       nd.BusAddr,
		MACAddress: nd.Mac,
		User: qcli.NetDeviceUser{
			IPV4: true,
		},
		BootIndex: nd.BootIndex,
		Driver:    qcli.DeviceDriver(nd.Device),
	}
	return ndev, nil
}

func (v VMDef) GenQConfig(runDir string) (*qcli.Config, error) {
	c, err := NewDefaultConfig(v.Name, v.Cpus, v.Memory)
	if err != nil {
		return c, err
	}

	if v.Cdrom != "" {
		qd := QemuDisk{
			File:   v.Cdrom,
			Format: "raw",
			Attach: "ide",
			Type:   "cdrom",
		}
		v.Disks = append(v.Disks, qd)
	}

	for _, disk := range v.Disks {
		if err := disk.Sanitize(runDir); err != nil {
			if err != nil {
				return c, err
			}
		}
		qblk, err := disk.QBlockDevice()
		if err != nil {
			return c, err
		}
		c.BlkDevices = append(c.BlkDevices, qblk)
	}

	for _, nic := range v.Nics {
		qnet, err := nic.QNetDevice()
		if err != nil {
			return c, err
		}
		c.NetDevices = append(c.NetDevices, qnet)
	}

	return c, nil
}

func main() {
	var newCluster Cluster
	var clusterBytes []byte

	clusterBytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(clusterBytes, &newCluster)
	if err != nil {
		panic(err)
	}

	for _, vmdef := range newCluster.Machines {
		cfg, err := vmdef.GenQConfig(".")
		if err != nil {
			panic(err)
		}
		fname := fmt.Sprintf("machine-%s.yaml", vmdef.Name)
		err = qcli.WriteConfig(fname, cfg)
		fmt.Printf("Wrote %s config\n", fname)
		if err != nil {
			panic(err)
		}
	}

}
