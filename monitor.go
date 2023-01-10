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
	"strings"
)

// MonitorDevice represents a qemu legacy human monitor device.
type MonitorDevice struct {
	Name      string `yaml:"name"`
	ChardevID string `yaml:"chardev-id"`
}

// Valid returns true if the MonitorDevice structure is valid and complete.
func (dev MonitorDevice) Valid() error {
	// One must be set
	if dev.Name == "" && dev.ChardevID == "" {
		return fmt.Errorf("LegacySerialDevice requires either Name or ChardevID field to be set")
	}

	// Name and ChardevID are mutually exclusive
	if dev.Name != "" && dev.ChardevID != "" {
		return fmt.Errorf("LegacySerialDevice Name and ChardevID field are mutually exclusive")
	}

	return nil
}

// QemuParams returns the qemu parameters built out of this monitor device.
func (dev MonitorDevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var monParams []string

	if dev.Name != "" && dev.ChardevID == "" {
		monParams = append(monParams, dev.Name)
	}
	if dev.ChardevID != "" && dev.Name == "" {
		monParams = append(monParams, fmt.Sprintf("chardev:%s", dev.ChardevID))
	}

	qemuParams = append(qemuParams, "-monitor")
	qemuParams = append(qemuParams, strings.Join(monParams, ","))

	return qemuParams
}
