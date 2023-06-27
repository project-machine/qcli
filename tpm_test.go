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
	"testing"
	"runtime"
	"fmt"
)

var ValidTPM = TPMDevice{
	ID:     "tpm0",
	Driver: TPMTISDevice,
	Path:   "tpm.socket",
	Type:   TPMEmulatorDevice,
}

func TestTPMDevice(t *testing.T) {
	chardevName := "tpm-tis"
	if runtime.GOARCH == "aarch64" || runtime.GOARCH == "arm64" {
		chardevName = "tpm-tis-device"
	}
	chardevStr := fmt.Sprintf("-chardev socket,id=chrtpm0,path=tpm.socket -tpmdev emulator,id=tpm0,chardev=chrtpm0 -device %s,tpmdev=tpm0", chardevName)
		
	testCases := []struct {
		dev Device
		out string
	}{
		{ValidTPM, chardevStr},
	}

	for _, tc := range testCases {
		testAppend(tc.dev, tc.out, t)
	}
}

func TestTPMDeviceInvalid(t *testing.T) {
	dev := TPMDevice{}

	if err := dev.Valid(); err == nil {
		t.Fatalf("A TPMDevice with missing ID field is NOT valid")
	}
	dev.ID = "tpm0"

	if err := dev.Valid(); err == nil {
		t.Fatalf("A TPMDevice with missing Driver field is NOT valid")
	}
	dev.Driver = TPMTISDevice

	if err := dev.Valid(); err == nil {
		t.Fatalf("A TPMDevice with missing Path field is NOT valid")
	}
	dev.Path = "tpm.socket"

	if err := dev.Valid(); err == nil {
		t.Fatalf("A TPMDevice with missing Type field is NOT valid")
	}
	dev.Type = "foobar"

	if err := dev.Valid(); err == nil {
		t.Fatalf("A TPMDevice with unknown Type field is NOT valid")
	}
	dev.Type = TPMEmulatorDevice

}
