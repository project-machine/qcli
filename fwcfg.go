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

// FwCfg allows QEMU to pass entries to the guest
// File and Str are mutually exclusive
type FwCfg struct {
	Name string `yaml:"name"`
	File string `yaml:"file"`
	Str  string `yaml:"string"`
}

// Valid returns true if the FwCfg structure is valid and complete.
func (fwcfg FwCfg) Valid() bool {
	if fwcfg.Name == "" {
		return false
	}

	if fwcfg.File != "" && fwcfg.Str != "" {
		return false
	}

	if fwcfg.File == "" && fwcfg.Str == "" {
		return false
	}

	return true
}

// QemuParams returns the qemu parameters built out of the FwCfg object
func (fwcfg FwCfg) QemuParams(config *Config) []string {
	var fwcfgParams []string
	var qemuParams []string

	for _, f := range config.FwCfg {
		if f.Name != "" {
			fwcfgParams = append(fwcfgParams, fmt.Sprintf("name=%s", f.Name))

			if f.File != "" {
				fwcfgParams = append(fwcfgParams, fmt.Sprintf("file=%s", f.File))
			}

			if f.Str != "" {
				fwcfgParams = append(fwcfgParams, fmt.Sprintf("string=%s", f.Str))
			}
		}

		qemuParams = append(qemuParams, "-fw_cfg")
		qemuParams = append(qemuParams, strings.Join(fwcfgParams, ","))
	}

	return qemuParams
}

func (config *Config) appendFwCfg(logger QMPLog) {
	if logger == nil {
		logger = qmpNullLogger{}
	}

	for _, f := range config.FwCfg {
		if !f.Valid() {
			logger.Errorf("fw_cfg is not valid: %+v", config.FwCfg)
			continue
		}

		config.qemuParams = append(config.qemuParams, f.QemuParams(config)...)
	}
}
