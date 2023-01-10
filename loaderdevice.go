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

package qemu

import (
	"fmt"
	"strings"
)

// LoaderDevice represents a qemu loader device.
type LoaderDevice struct {
	File string `yaml:"file"`
	ID   string `yaml:"id"`
}

// Valid returns true if there is a valid structure defined for LoaderDevice
func (dev LoaderDevice) Valid() error {
	if dev.File == "" {
		return fmt.Errorf("LoaderDevice has empty File field")
	}

	if dev.ID == "" {
		return fmt.Errorf("LoaderDevice has empty ID field")
	}

	return nil
}

// QemuParams returns the qemu parameters built out of this loader device.
func (dev LoaderDevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string

	deviceParams = append(deviceParams, "loader")
	deviceParams = append(deviceParams, fmt.Sprintf("file=%s", dev.File))
	deviceParams = append(deviceParams, fmt.Sprintf("id=%s", dev.ID))

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}
