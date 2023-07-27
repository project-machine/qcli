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
	"fmt"
	"log"
	"strings"
)

// Machine describes the machine type qemu will emulate.
type Machine struct {
	// Type is the machine type to be used by qemu.
	Type string `yaml:"type"`

	// Acceleration are the machine acceleration options to be used by qemu.
	Acceleration string `yaml:"acceleration"`

	// Options are options for the machine type
	// For example gic-version=host and usb=off
	// FIXME: remove this
	Options string `yaml:"options"`

	// on|off
	SMM string `yaml:"smm"`

	// KernelIRQChip controls accelerated IRQChip, value is on|off|split
	KernelIRQChip string `yaml:"kernel-irq-chip"`

	// Emulate VMPort, value is on|off|auto
	VMPort string `yaml:"vm-port"`

	KVMShadowMemSizeBytes int64 `yaml:"kvm-shadow-mem-size-bytes"`

	// on|off
	DumpGuestCore string `yaml:"dump-guest-core"`

	// on|off
	MemoryMerge string `yaml:"memory-merge"`

	// on|off
	IGDPassthrough string `yaml:"igd-passthrough"`

	// on|off
	AESKeyWrap string `yaml:"aes-key-wrap"`

	// on|off
	DEAKeyWrap string `yaml:"dea-key-wrap"`

	// on|off
	SuppressVMDescription string `yaml:"suppress-vm-description"`

	// on|off
	NVDIMM string `yaml:"nvdimm"`

	// on|off
	EnforceConfigSection string `yaml:"enforce-config-section"`
}

const (
	// MachineTypeMicrovm is the QEMU microvm machine type for amd64
	MachineTypeMicrovm string = "microvm"
	MachineTypePC35    string = "q35"
	MachineTypePC      string = "pc"
	MachineTypeVirt    string = "virt"

	MachineAccelerationKVM string = "kvm"
)

func (config *Config) appendMachine() {
	if config.Machine.Type != "" {
		var machineParams []string

		machineParams = append(machineParams, config.Machine.Type)

		if config.Machine.Acceleration != "" {
			machineParams = append(machineParams, fmt.Sprintf("accel=%s", config.Machine.Acceleration))
		}

		chip := config.Machine.KernelIRQChip
		if chip != "" {
			switch chip {
			case "on", "off", "split":
				machineParams = append(machineParams, fmt.Sprintf("kernel_irqchip=%s", chip))
			default:
				log.Fatalf("Invalid KernealIRQChip value: '%s', must be one of 'on', 'off', or 'split'", chip)
			}
		}

		vmport := config.Machine.VMPort
		if vmport != "" {
			switch vmport {
			case "on", "off", "auto":
				machineParams = append(machineParams, fmt.Sprintf("vmport=%s", vmport))
			default:
				log.Fatalf("Invalid VMPort value: '%s', must be one of 'on', 'off', or 'auto'", vmport)
			}
		}

		if config.Machine.KVMShadowMemSizeBytes > 0 {
			machineParams = append(machineParams, fmt.Sprintf("kvm_shadow_mem=%d", config.Machine.KVMShadowMemSizeBytes))
		}

		mParam := getConfigOnOff("SMM", "smm", config.Machine.SMM)
		if mParam != "" {
			machineParams = append(machineParams, mParam)
		}

		mParam = getConfigOnOff("DumpGuestCore", "dump-guest-core", config.Machine.DumpGuestCore)
		if mParam != "" {
			machineParams = append(machineParams, mParam)
		}

		mParam = getConfigOnOff("MemoryMerge", "mem-merge", config.Machine.MemoryMerge)
		if mParam != "" {
			machineParams = append(machineParams, mParam)
		}

		mParam = getConfigOnOff("IGDPassthrough", "igd-passthrough", config.Machine.IGDPassthrough)
		if mParam != "" {
			machineParams = append(machineParams, mParam)
		}

		mParam = getConfigOnOff("AESKeyWrap", "aes-key-wrap", config.Machine.AESKeyWrap)
		if mParam != "" {
			machineParams = append(machineParams, mParam)
		}

		mParam = getConfigOnOff("DEAKeyWrap", "dea-key-wrap", config.Machine.DEAKeyWrap)
		if mParam != "" {
			machineParams = append(machineParams, mParam)
		}

		mParam = getConfigOnOff("SuppresVMDescription", "suppress-vmdesc", config.Machine.SuppressVMDescription)
		if mParam != "" {
			machineParams = append(machineParams, mParam)
		}

		mParam = getConfigOnOff("NVDIMM", "nvdimm", config.Machine.NVDIMM)
		if mParam != "" {
			machineParams = append(machineParams, mParam)
		}

		mParam = getConfigOnOff("EnforceConfigSection", "enforce-config-section", config.Machine.EnforceConfigSection)
		if mParam != "" {
			machineParams = append(machineParams, mParam)
		}

		// FIXME: catch all for any options, might trigger duplicates though
		if config.Machine.Options != "" {
			machineParams = append(machineParams, config.Machine.Options)
		}

		config.qemuParams = append(config.qemuParams, "-machine")
		config.qemuParams = append(config.qemuParams, strings.Join(machineParams, ","))
	}
}
