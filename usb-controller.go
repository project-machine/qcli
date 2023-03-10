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

// USBController represents an USB controller device.
type USBControllerDevice struct {
	ID                   string       `yaml:"id"`
	Driver               DeviceDriver `yaml:"driver"`
	Addr                 string       `yaml:"addr,omitempty"`
	FailoverPairID       string       `yaml:"failover-pair-id,omitempty"`
	ROMFile              string       `yaml:"romfile,omitempty"`
	ROMBar               string       `yaml:"rombar,omitempty"`
	Multifunction        bool         `yaml:"multifunction,omitempty"`
	XPCIELinkStateDLLLA  bool         `yaml:"x-pcie-lnksta-dllla,omitempty"`
	XPCIeExternalCapInit bool         `yaml:"x-pcie-extcap-init,omitempty"`
	CommandSerrEnable    bool         `yaml:"command-seer-enable,omitempty"`
}

// Valid returns true if the USBController structure is valid and complete.
func (usbCon USBControllerDevice) Valid() error {
	if usbCon.ID == "" {
		return fmt.Errorf("USBController has empty ID field")
	}

	if usbCon.Driver == "" {
		return fmt.Errorf("USBController has empty Driver field")
	}
	return nil
}

// QemuParams returns the qemu parameters built out of this USBController device.
func (usbCon USBControllerDevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string

	driver := usbCon.deviceName(config)
	deviceParams = append(deviceParams, fmt.Sprintf("%s,id=%s", driver, usbCon.ID))
	addr := config.pciBusSlots.GetSlot(usbCon.Addr)
	if addr > 0 {
		deviceParams = append(deviceParams, fmt.Sprintf("addr=0x%02x", addr))
	}
	if usbCon.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", usbCon.ROMFile))
	}
	if usbCon.ROMBar != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("rombar=%s", usbCon.ROMBar))
	}
	if usbCon.Multifunction {
		deviceParams = append(deviceParams, "multifunction=on")
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))
	return qemuParams
}

// deviceName returns the QEMU device name for the current combination of
// driver and transport.
func (usbCon USBControllerDevice) deviceName(config *Config) string {
	return string(usbCon.Driver)
}
