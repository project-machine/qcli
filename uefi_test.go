package qcli

import (
	"strings"
	"testing"
	"runtime"
)

func TestUEFIFirmwareDeviceValid(t *testing.T) {
	udev := UEFIFirmwareDevice{}

	if err := udev.Valid(); err == nil {
		t.Fatalf("UEFIFirmwareDevice should not be valid when Code is empty")
	}

	udev.Code = "code.fd"
	if err := udev.Valid(); err == nil {
		t.Fatalf("UEFIFirmwareDevice should not be valid when Vars is empty")
	}

	udev.Vars = "vars.fd"
	if err := udev.Valid(); err != nil {
		t.Fatalf("UEFIFirmwareDevice should be valid when Code and Vars are set")
	}

}

func TestAppendUEFIFirmwareDevice(t *testing.T) {
	udev := UEFIFirmwareDevice{Code: "OVMF_CODE.fd", Vars: "OVMF_VARS.fd"}
	expected := "-drive if=pflash,format=raw,readonly=on,file=OVMF_CODE.fd -drive if=pflash,format=raw,file=OVMF_VARS.fd"

	testAppend(udev, expected, t)
}

func TestAppendUEFIFirmwareDeviceConfig(t *testing.T) {
	c := &Config{}
	udev := UEFIFirmwareDevice{Code: "OVMF_CODE.fd", Vars: "OVMF_VARS.fd"}
	c.UEFIFirmwareDevices = append(c.UEFIFirmwareDevices, udev)
	err := c.appendDevices()
	if err != nil {
		t.Fatalf("Failed to append UEFI firwmware device: %s", err)
	}
	if len(c.qemuParams) == 0 {
		t.Errorf("Expected non-empty qemuParams, found %s", c.qemuParams)
	}

	expected := "-drive if=pflash,format=raw,readonly=on,file=OVMF_CODE.fd -drive if=pflash,format=raw,file=OVMF_VARS.fd"
	result := strings.Join(c.qemuParams, " ")
	if expected != result {
		t.Fatalf("Failed to append parameters\nexpected[%s]\n!=\nfound   [%s]", expected, result)
	}
}

func TestNewFirmwareDev(t *testing.T) {
	secureBoot := true
	udev_ptr, err := NewSystemUEFIFirmwareDevice(secureBoot)
	if err != nil {
		t.Fatalf("Failed to find secure firmware blobs: %s", err)
	}
	udev := *udev_ptr
	switch runtime.GOARCH{
	case "amd64":
		if PathExists(UbuntuSecVars) {
			expected := "-drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE.secboot.fd -drive if=pflash,format=raw,file=/usr/share/OVMF/OVMF_VARS.ms.fd"
			testAppend(udev, expected, t)
		} else if PathExists(CentosSecVars){
			expected := "-drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE.secboot.fd -drive if=pflash,format=raw,file=/usr/share/OVMF/OVMF_VARS.secboot.fd"
			testAppend(udev, expected, t)
		} else {
			t.Fatalf("Failed to find secure firmware blobs")
		}
	case "arm64", "aarch64":
		expected := "-drive if=pflash,format=raw,readonly=on,file=/usr/share/AAVMF/AAVMF_CODE.ms.fd -drive if=pflash,format=raw,file=/usr/share/AAVMF/AAVMF_VARS.ms.fd"
		testAppend(udev, expected, t)
	}
}
// TODO: add system tests to handle different distros
