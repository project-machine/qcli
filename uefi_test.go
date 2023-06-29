package qcli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	expected := "-drive if=pflash,format=raw,readonly=on,file=OVMF_CODE.fd -drive if=pflash,format=raw,file=OVMF_VARS.fd"

	testAppend(udev, expected, t)
}

func createTree(basePath string, files []string) error {
	for _, file := range files {
		targetFile := filepath.Join(basePath, file)
		dir := filepath.Dir(targetFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("Failed to create directory %q for file %q: %s", dir, targetFile, err)
		}
		fh, err := os.OpenFile(targetFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("Failed to create file %q: %s", targetFile, err)
		}
		fh.Close()
		// fmt.Printf("touched %q\n", targetFile)
	}
	return nil
}

func ubuntuVMFFiles() []string {
	switch runtime.GOARCH {
	case "amd64", "x86_64":
		return []string{
			"OVMF_CODE_4M.fd",
			"OVMF_CODE_4M.ms.fd",
			"OVMF_CODE_4M.secboot.fd",
			"OVMF_CODE.fd",
			"OVMF_CODE.ms.fd",
			"OVMF_CODE.secboot.fd",
			"OVMF_VARS_4M.fd",
			"OVMF_VARS_4M.ms.fd",
			"OVMF_VARS.fd",
			"OVMF_VARS.ms.fd",
			"OVMF_VARS.snakeoil.fd",
		}
	case "arm64", "aarch64":
		return []string{
			"AAVMF32_CODE.fd",
			"AAVMF32_VARS.fd",
			"AAVMF_CODE.fd",
			"AAVMF_CODE.ms.fd",
			"AAVMF_CODE.snakeoil.fd",
			"AAVMF_VARS.fd",
			"AAVMF_VARS.ms.fd",
			"AAVMF_VARS.snakeoil.fd",
		}
	}
	return []string{}
}

func TestNewUEFIFIrmwareDeviceSecureBoot(t *testing.T) {
	origVMFHostPrefix := VMFHostPrefix
	VMFHostPrefix = t.TempDir()
	defer func() {
		VMFHostPrefix = origVMFHostPrefix
	}()

	basePath := VMFPathBase()
	files := ubuntuVMFFiles()
	if err := createTree(basePath, files); err != nil {
		t.Fatalf("Failed to create directory structure for test: %s", err)
	}

	secureBoot := true
	udev, err := NewSystemUEFIFirmwareDevice(secureBoot)
	if err != nil {
		t.Fatalf("Invalid New UEFI Firwmare device: %s", err)
	}
	codePath := ""
	varsPath := ""
	switch runtime.GOARCH {
	case "amd64", "x86_64":
		codePath = filepath.Join(basePath, "OVMF_CODE_4M.secboot.fd")
		varsPath = filepath.Join(basePath, "OVMF_VARS_4M.ms.fd")
	case "arm64", "aarch64":
		codePath = filepath.Join(basePath, "AAVMF_CODE.ms.fd")
		varsPath = filepath.Join(basePath, "AAVMF_VARS.ms.fd")
	}
	expected := fmt.Sprintf("-drive if=pflash,format=raw,readonly=on,file=%s -drive if=pflash,format=raw,file=%s", codePath, varsPath)
	testAppend(*udev, expected, t)
}

func TestNewUEFIFIrmwareDevice(t *testing.T) {
	origVMFHostPrefix := VMFHostPrefix
	VMFHostPrefix = t.TempDir()
	defer func() {
		VMFHostPrefix = origVMFHostPrefix
	}()

	basePath := VMFPathBase()
	files := ubuntuVMFFiles()
	if err := createTree(basePath, files); err != nil {
		t.Fatalf("Failed to create directory structure for test: %s", err)
	}

	secureBoot := false
	udev, err := NewSystemUEFIFirmwareDevice(secureBoot)
	if err != nil {
		t.Fatalf("Invalid New UEFI Firwmare device: %s", err)
	}
	codePath := ""
	varsPath := ""
	switch runtime.GOARCH {
	case "amd64", "x86_64":
		codePath = filepath.Join(basePath, "OVMF_CODE_4M.fd")
		varsPath = filepath.Join(basePath, "OVMF_VARS_4M.fd")
	case "arm64", "aarch64":
		codePath = filepath.Join(basePath, "AAVMF_CODE.fd")
		varsPath = filepath.Join(basePath, "AAVMF_VARS.fd")
	}
	expected := fmt.Sprintf("-drive if=pflash,format=raw,readonly=on,file=%s -drive if=pflash,format=raw,file=%s", codePath, varsPath)
	testAppend(*udev, expected, t)
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

// TODO: add system tests to handle different distros
