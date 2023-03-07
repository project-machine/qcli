/*
Copyright © 2023 Ryan Harper <rharper@woxford.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package qcli

import (
	"fmt"
	"strings"
)

// IDEController represents an IDE controller device.
type IDEControllerDevice struct {
	ID                   string       `yaml:"id"`
	Driver               DeviceDriver `yaml:"driver"`
	Bus                  string       `yaml:"bus,omitempty"`
	Addr                 string       `yaml:"addr,omitempty"`
	FailoverPairID       string       `yaml:"failover-pair-id,omitempty"`
	ROMFile              string       `yaml:"romfile,omitempty"`
	ROMBar               string       `yaml:"rombar,omitempty"`
	Multifunction        bool         `yaml:"multifunction,omitempty"`
	XPCIELinkStateDLLLA  bool         `yaml:"x-pcie-lnksta-dllla,omitempty"`
	XPCIeExternalCapInit bool         `yaml:"x-pcie-extcap-init,omitempty"`
	CommandSerrEnable    bool         `yaml:"command-seer-enable,omitempty"`
}

// Valid returns true if the IDEController structure is valid and complete.
func (ideCon IDEControllerDevice) Valid() error {
	if ideCon.ID == "" {
		return fmt.Errorf("IDEController has empty ID field")
	}

	if ideCon.Driver == "" {
		return fmt.Errorf("IDEController has empty Driver field")
	}
	return nil
}

// QemuParams returns the qemu parameters built out of this IDEController device.
func (ideCon IDEControllerDevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string

	driver := ideCon.deviceName(config)
	deviceParams = append(deviceParams, fmt.Sprintf("%s,id=%s", driver, ideCon.ID))
	addr := config.pciBusSlots.GetSlot(ideCon.Addr)
	if addr > 0 {
		deviceParams = append(deviceParams, fmt.Sprintf("addr=0x%02x", addr))
		bus := "pcie.0"
		if ideCon.Bus != "" {
			bus = ideCon.Bus
		}
		deviceParams = append(deviceParams, fmt.Sprintf("bus=%s", bus))
	}
	if ideCon.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", ideCon.ROMFile))
	}
	if ideCon.ROMBar != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("rombar=%s", ideCon.ROMBar))
	}
	if ideCon.Multifunction {
		deviceParams = append(deviceParams, "multifunction=on")
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))
	return qemuParams
}

// deviceName returns the QEMU device name for the current combination of
// driver and transport.
func (ideCon IDEControllerDevice) deviceName(config *Config) string {
	return string(ideCon.Driver)
}
