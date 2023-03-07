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

// QMPSocketType is the type of socket used for QMP communication.
type QMPSocketType string

const (
	// Unix socket for QMP.
	Unix QMPSocketType = "unix"
)

// QMPSocket represents a qemu QMP socket configuration.
type QMPSocket struct {
	// Type is the socket type (e.g. "unix").
	Type QMPSocketType `yaml:"type" default:"unix"`

	// Name is the socket name.
	Name string `yaml:"name"`

	// Server tells if this is a server socket.
	Server bool `yaml:"server"`

	// NoWait tells if qemu should block waiting for a client to connect.
	NoWait bool `yaml:"no-wait"`
}

// Valid returns true if the QMPSocket structure is valid and complete.
func (qmp QMPSocket) Valid() error {
	if qmp.Type == "" {
		return fmt.Errorf("QMPSocket has empty Type field")
	}
	if qmp.Name == "" {
		return fmt.Errorf("QMPSocket has empty Name field")
	}
	if qmp.Type != Unix {
		return fmt.Errorf("QMPSocket has invalid Type field: %s", qmp.Type)
	}

	return nil
}

func (config *Config) appendQMPSockets() error {
	var errors []string
	for _, q := range config.QMPSockets {
		if err := q.Valid(); err != nil {
			errors = append(errors, err.Error())
			continue
		}

		qmpParams := append([]string{}, fmt.Sprintf("%s:%s", q.Type, q.Name))
		if q.Server {
			qmpParams = append(qmpParams, "server=on")
			if q.NoWait {
				qmpParams = append(qmpParams, "wait=off")
			}
		}

		config.qemuParams = append(config.qemuParams, "-qmp")
		config.qemuParams = append(config.qemuParams, strings.Join(qmpParams, ","))
	}

	if len(errors) > 0 {
		return fmt.Errorf("Failed to append %d QMPSocket(s):\n%s", len(errors), strings.Join(errors, "\n"))
	}

	return nil
}
