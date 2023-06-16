// +build aarch64
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

package qcli

import (
	"strings"
	"testing"
)

var (
	fullUefiAarch64VM = "-machine virt,accel=kvm -cpu host -m 1G -drive file=udisk.img,id=hd0,if=none,format=qcow2 -device virtio-blk-pci,drive=hd0,serial=hd0,disable-modern=false,addr=0x1e,bus=pcie.0,scsi=off,config-wce=off -drive file=ubuntu-22.04.2-live-server-arm64.iso,id=cdrom0,if=none,format=raw,media=cdrom,readonly=on -device virtio-blk-pci,drive=cdrom0,serial=cdrom0,bootindex=0,disable-modern=false,addr=0x1d,bus=pcie.0,scsi=off,config-wce=off -drive if=pflash,format=raw,readonly=on,file=/usr/share/AAVMF/AAVMF_CODE.ms.fd -drive if=pflash,format=raw,file=uefi-nvram.fd -object memory-backend-ram,id=dimm1,size=1G -numa node,memdev=dimm1 -nographic"
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

func TestFullUEFIMachineCommandArch(t *testing.T) {
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
