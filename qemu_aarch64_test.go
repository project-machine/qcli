package qcli

import (
	"strings"
	"testing"
)

var (
	fullUefiAarchVM = "-machine virt,accel=kvm -cpu host -m 1G -drive file=udisk.img,id=hd0,if=none,format=qcow2 -device virtio-blk-pci,drive=hd0,serial=hd0,disable-modern=false,addr=0x1e,bus=pcie.0,scsi=off,config-wce=off -drive file=ubuntu-22.04.2-live-server-arm64.iso,id=cdrom0,if=none,format=raw,media=cdrom,readonly=on -device virtio-blk-pci,drive=cdrom0,serial=cdrom0,bootindex=0,disable-modern=false,addr=0x1d,bus=pcie.0,scsi=off,config-wce=off -drive if=pflash,format=raw,readonly=on,file=/usr/share/AAVMF/AAVMF_CODE.ms.fd -drive if=pflash,format=raw,file=flash-vars-sec.img -object memory-backend-ram,id=dimm1,size=1G -numa node,memdev=dimm1 -nographic"
)

func fullVMConfigArch() *Config {
	c := &Config{
		Machine: Machine{
			Type:         MachineTypeVirt,
			Acceleration: MachineAccelerationKVM,
		},
		CPUModel:      "host", //cortex-a53, cortex-a57, cortex-a72
		Memory: Memory{
			Size: "1G",
		},
		BlkDevices: []BlockDevice{
			BlockDevice{
				Driver: VirtioBlock,
				ID: "hd0",
				File: "udisk.img",
				Format:	QCOW2,
				Interface: NoInterface,
			},
			BlockDevice {
				Driver: VirtioBlock,
				Interface: NoInterface,
				ID: "cdrom0",
				File: "ubuntu-22.04.2-live-server-arm64.iso",
				Format: RAW,
				ReadOnly: true,
				Media: "cdrom",
				BootIndex: "0",
			},
		},
		Knobs: Knobs{
			NoGraphic: true,
		},
	}
	return c
}

func TestFullUEFIMachineCommandArch(t *testing.T) {
	c := fullVMConfigArch()

	u := UEFIFirmwareDevice{
		Code: "/usr/share/AAVMF/AAVMF_CODE.ms.fd",
		Vars: UEFISecVarsFileNameArm,
	}

	c.UEFIFirmwareDevices = append(c.UEFIFirmwareDevices, u)

	expected := fullUefiAarchVM
	qemuParams, err := ConfigureParams(c, nil)
	if err != nil {
		t.Fatalf("Failed to Configure parameters, error: %s", err.Error())
	}
	result := strings.Join(qemuParams, " ")
	if expected != result {
		t.Fatalf("Failed to append parameters\nexpected[%s]\n!=\nfound    [%s]", expected, result)
	}
}
