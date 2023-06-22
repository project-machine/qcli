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
	"runtime"
)

const (
	TPMEmulatorDevice    = "emulator"
	TPMPassthroughDevice = "passthrough"
)

// TPM represents a qemu tpm device.
type TPMDevice struct {
	ID     string       `yaml:"id"`
	Driver DeviceDriver `yaml:"driver"`
	Type   string       `yaml:"type"`
	Path   string       `yaml:"path,omitempty"`
}

// Valid returns true if there is a valid structure defined for TPM device
func (tpm TPMDevice) Valid() error {
	if tpm.ID == "" {
		return fmt.Errorf("TPM device ID is not set")
	}

	if tpm.Driver == "" {
		return fmt.Errorf("TPM device Driver is not set")
	}

	if tpm.Path == "" {
		return fmt.Errorf("TPM device Path is not set")
	}

	if tpm.Type == "" {
		return fmt.Errorf("TPM device Type is not set")
	}

	switch tpm.Type {
	case TPMEmulatorDevice, TPMPassthroughDevice:
		break
	default:
		return fmt.Errorf("TPM device Type '%s' is unknown", tpm.Type)
	}

	return nil
}

// QemuParams returns the qemu parameters built out of this tpm device.
func (tpm TPMDevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string
	var tpmParams []string
	var chardevParams []string

	// -device tpm-tis,tpmdev=tpm0
	deviceParams = append(deviceParams, tpm.deviceName(), fmt.Sprintf("tpmdev=%s", tpm.ID))

	// -tpmdev emulator,id=tpm0,chardev=chrtpm0
	charDev := fmt.Sprintf("chr%s", tpm.ID)
	tpmParams = append(tpmParams, tpm.Type, fmt.Sprintf("id=%s", tpm.ID), fmt.Sprintf("chardev=%s", charDev))

	// -chardev socket,id=chrtpm0,path=tpm0.socket
	chardevParams = append(chardevParams, "socket", fmt.Sprintf("id=%s", charDev), fmt.Sprintf("path=%s", tpm.Path))

	qemuParams = append(qemuParams, "-chardev")
	qemuParams = append(qemuParams, strings.Join(chardevParams, ","))
	qemuParams = append(qemuParams, "-tpmdev")
	qemuParams = append(qemuParams, strings.Join(tpmParams, ","))
	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))

	return qemuParams
}

func (tpm TPMDevice) deviceName() string {
	switch tpm.Driver {
    case TPMTISDevice:
          if runtime.GOARCH == "aarch64" || runtime.GOARCH == "arm64" {
              return string(tpm.Driver + "-device")
          }
    }
    return string(tpm.Driver)
}
