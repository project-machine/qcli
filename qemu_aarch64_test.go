// +build aarch64 arm64

package qcli

import (
	"strings"
	"testing"
)

var (
	fullUefiAarch64VM = "-machine virt,accel=kvm -cpu host -m 1G -drive file=udisk.img,id=hd0,if=none,format=qcow2 -device virtio-blk-pci,drive=hd0,serial=hd0,disable-modern=false,addr=0x1e,bus=pcie.0,scsi=off,config-wce=off -drive file=ubuntu-22.04.2-live-server-arm64.iso,id=cdrom0,if=none,format=raw,media=cdrom,readonly=on -device virtio-blk-pci,drive=cdrom0,serial=cdrom0,bootindex=0,disable-modern=false,addr=0x1d,bus=pcie.0,scsi=off,config-wce=off -drive if=pflash,format=raw,readonly=on,file=/usr/share/AAVMF/AAVMF_CODE.ms.fd -drive if=pflash,format=raw,file=uefi-nvram.fd -object memory-backend-ram,id=dimm1,size=1G -numa node,memdev=dimm1 -nographic"
	fullUefiTPMAarch64VM = "-machine virt,accel=kvm -cpu host -m 1G -chardev socket,id=chrtpm0,path=tpm.sock -tpmdev emulator,id=tpm0,chardev=chrtpm0 -device tpm-tis-device,tpmdev=tpm0 -drive file=udisk.img,id=hd0,if=none,format=qcow2 -device virtio-blk-pci,drive=hd0,serial=hd0,disable-modern=false,addr=0x1e,bus=pcie.0,scsi=off,config-wce=off -drive file=ubuntu-22.04.2-live-server-arm64.iso,id=cdrom0,if=none,format=raw,media=cdrom,readonly=on -device virtio-blk-pci,drive=cdrom0,serial=cdrom0,bootindex=0,disable-modern=false,addr=0x1d,bus=pcie.0,scsi=off,config-wce=off -drive if=pflash,format=raw,readonly=on,file=/usr/share/AAVMF/AAVMF_CODE.ms.fd -drive if=pflash,format=raw,file=uefi-nvram.fd -object memory-backend-ram,id=dimm1,size=1G -numa node,memdev=dimm1 -nographic"
)

func fullVMConfigArch() *Config {
	c := &Config{
		Machine: Machine{
			Type:         MachineTypeVirt,
			Acceleration: MachineAccelerationKVM,
		},
		CPUModel:      "host",
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

func TestFullUEFIMachineCommandAarch64(t *testing.T) {
	c := fullVMConfigArch()

	u := UEFIFirmwareDevice{
		Code: "/usr/share/AAVMF/AAVMF_CODE.ms.fd",
		Vars: "uefi-nvram.fd",
	}

	c.UEFIFirmwareDevices = append(c.UEFIFirmwareDevices, u)

	expected := fullUefiAarch64VM
	qemuParams, err := ConfigureParams(c, nil)
	if err != nil {
		t.Fatalf("Failed to Configure parameters, error: %s", err.Error())
	}
	result := strings.Join(qemuParams, " ")
	if expected != result {
		t.Fatalf("Failed to append parameters\nexpected[%s]\n!=\nfound    [%s]", expected, result)
	}
}

func TestFullUEFITPMCommandAarch64(t *testing.T) {
	c := fullVMConfigArch()

	u := UEFIFirmwareDevice{
		Code: "/usr/share/AAVMF/AAVMF_CODE.ms.fd",
		Vars: "uefi-nvram.fd",
	}
	c.UEFIFirmwareDevices = append(c.UEFIFirmwareDevices, u)

	c.TPM = TPMDevice{
		ID:     "tpm0",
		Driver: TPMTISDeviceAarch64,
		Path:   "tpm.sock",
		Type:   TPMEmulatorDevice,
	}

	expected := fullUefiTPMAarch64VM
	qemuParams, err := ConfigureParams(c, nil)
	if err != nil {
		t.Fatalf("Failed to Configure parameters, error: %s", err.Error())
	}
	result := strings.Join(qemuParams, " ")
	if expected != result {
		t.Fatalf("Failed to append parameters when adding tpm dev to aarch64 setup\nexpected[%s]\n!=\nfound    [%s]", expected, result)
	}

}
