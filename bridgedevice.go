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
	"strconv"
	"strings"
)

// BridgeType is the type of the bridge
type BridgeType uint

const (
	// PCIBridge is a pci bridge
	PCIBridge BridgeType = iota

	// PCIEBridge is a pcie bridge
	PCIEBridge
)

// BridgeDevice represents a qemu bridge device like pci-bridge, pxb, etc.
type BridgeDevice struct {
	// Type of the bridge
	Type BridgeType `yaml:"type"`

	// Bus number where the bridge is plugged, typically pci.0 or pcie.0
	Bus string `yaml:"bus"`

	// ID is used to identify the bridge in qemu
	ID string `yaml:"id"`

	// Chassis number
	Chassis int `yaml:"chassis"`

	// SHPC is used to enable or disable the standard hot plug controller
	SHPC bool `yaml:"standard-hotplug-controller"`

	// PCI Slot
	Addr string `yaml:"address"`

	// ROMFile specifies the ROM file being used for this device.
	ROMFile string `yaml:"rom-file"`

	// Address range reservations for devices behind the bridge
	// NB: strings seem an odd choice, but if they were integers,
	// they'd default to 0 by Go's rules in all the existing users
	// who don't set them.  0 is a valid value for certain cases,
	// but not you want by default.
	IOReserve     string `yaml:"io-reserve"`
	MemReserve    string `yaml:"mem-reserve"`
	Pref64Reserve string `yaml:"pref64-reserve"`
}

// Valid returns nil if the BridgeDevice structure is valid and complete.
func (bridgeDev BridgeDevice) Valid() error {
	if bridgeDev.Type != PCIBridge && bridgeDev.Type != PCIEBridge {
		return fmt.Errorf("BridgeDevice has invalid Type: %d", bridgeDev.Type)
	}

	if bridgeDev.Bus == "" {
		return fmt.Errorf("BridgeDevice missing Bus value")
	}

	if bridgeDev.ID == "" {
		return fmt.Errorf("BridgeDevice missing ID value")
	}

	return nil
}

// QemuParams returns the qemu parameters built out of this bridge device.
func (bridgeDev BridgeDevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string
	var driver DeviceDriver

	switch bridgeDev.Type {
	case PCIEBridge:
		driver = PCIePCIBridgeDriver
		deviceParams = append(deviceParams, fmt.Sprintf("%s,bus=%s,id=%s", driver, bridgeDev.Bus, bridgeDev.ID))
	default:
		driver = PCIBridgeDriver
		shpc := "off"
		if bridgeDev.SHPC {
			shpc = "on"
		}
		deviceParams = append(deviceParams, fmt.Sprintf("%s,bus=%s,id=%s,chassis_nr=%d,shpc=%s", driver, bridgeDev.Bus, bridgeDev.ID, bridgeDev.Chassis, shpc))
	}

	if bridgeDev.Addr != "" {
		addr, err := strconv.Atoi(bridgeDev.Addr)
		if err == nil && addr >= 0 {
			deviceParams = append(deviceParams, fmt.Sprintf("addr=%x", addr))
		}
	}

	var transport VirtioTransport
	if transport.isVirtioPCI(config) && bridgeDev.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", bridgeDev.ROMFile))
	}

	if bridgeDev.IOReserve != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("io-reserve=%s", bridgeDev.IOReserve))
	}
	if bridgeDev.MemReserve != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("mem-reserve=%s", bridgeDev.MemReserve))
	}
	if bridgeDev.Pref64Reserve != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("pref64-reserve=%s", bridgeDev.Pref64Reserve))
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}
