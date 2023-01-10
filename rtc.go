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

// RTCBaseType is the qemu RTC base time type.
type RTCBaseType string

// RTCClock is the qemu RTC clock type.
type RTCClock string

// RTCDriftFix is the qemu RTC drift fix type.
type RTCDriftFix string

const (
	// UTC is the UTC base time for qemu RTC.
	UTC RTCBaseType = "utc"

	// LocalTime is the local base time for qemu RTC.
	LocalTime RTCBaseType = "localtime"
)

const (
	// Host is for using the host clock as a reference.
	Host RTCClock = "host"

	// RT is for using the host monotonic clock as a reference.
	RT RTCClock = "rt"

	// VM is for using the guest clock as a reference
	VM RTCClock = "vm"
)

const (
	// Slew is the qemu RTC Drift fix mechanism.
	Slew RTCDriftFix = "slew"

	// NoDriftFix means we don't want/need to fix qemu's RTC drift.
	NoDriftFix RTCDriftFix = "none"
)

// RTC represents a qemu Real Time Clock configuration.
type RTC struct {
	// Base is the RTC start time.
	Base RTCBaseType

	// Clock is the is the RTC clock driver.
	Clock RTCClock

	// DriftFix is the drift fixing mechanism.
	DriftFix RTCDriftFix
}

// Valid returns true if the RTC structure is valid and complete.
func (rtc RTC) Valid() bool {
	if rtc.Clock != Host && rtc.Clock != RT && rtc.Clock != VM {
		return false
	}

	if rtc.DriftFix != Slew && rtc.DriftFix != NoDriftFix {
		return false
	}

	return true
}

func (config *Config) appendRTC() {
	if !config.RTC.Valid() {
		return
	}

	var RTCParams []string

	RTCParams = append(RTCParams, fmt.Sprintf("base=%s", string(config.RTC.Base)))

	if config.RTC.DriftFix != "" {
		RTCParams = append(RTCParams, fmt.Sprintf("driftfix=%s", config.RTC.DriftFix))
	}

	if config.RTC.Clock != "" {
		RTCParams = append(RTCParams, fmt.Sprintf("clock=%s", config.RTC.Clock))
	}

	config.qemuParams = append(config.qemuParams, "-rtc")
	config.qemuParams = append(config.qemuParams, strings.Join(RTCParams, ","))
}
