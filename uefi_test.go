package qcli

import (
	"strings"
	"testing"
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
	expected := "-drive if=pflash,format=raw,readonly,file=OVMF_CODE.fd -drive if=pflash,format=raw,file=OVMF_VARS.fd"

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

	expected := "-drive if=pflash,format=raw,readonly,file=OVMF_CODE.fd -drive if=pflash,format=raw,file=OVMF_VARS.fd"
	result := strings.Join(c.qemuParams, " ")
	if expected != result {
		t.Fatalf("Failed to append parameters\nexpected[%s]\n!=\nfound   [%s]", expected, result)
	}
}

// TODO: add system tests to handle different distros
