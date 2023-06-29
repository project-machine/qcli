package qcli

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

type UEFIFirmwareDevice struct {
	Code string `yaml:"uefi-code"`
	Vars string `yaml:"uefi-vars"`
}

var VMFHostPrefix = "/usr/share"

const (
	UEFIVarsFileName = "uefi-nvram.fd"
	VMFCode          = "VMF_CODE" // OVMF_CODE , AAVMF_CODE
	VMFVars          = "VMF_VARS"
	VMFMs            = ".ms"
	VMFSecboot       = ".secboot"
	VMF4MB           = "_4M"
	VMFSuffix        = ".fd"
	VMF32Bit         = "32"
)

func VMFPrefix() string {
	switch runtime.GOARCH {
	case "aarch64", "arm64":
		return "AA"
	case "amd64", "x86_64":
		return "O"
	}
	return ""
}

func VMFPathBase() string {
	return filepath.Join(VMFHostPrefix, VMFPrefix()+"VMF")
}

func (u UEFIFirmwareDevice) Valid() error {
	if u.Code == "" {
		return fmt.Errorf("UEFIFirmwareDevice has empty Code field")
	}
	if u.Vars == "" {
		return fmt.Errorf("UEFIFirmwareDevice has empty Vars field")
	}
	return nil
}

func (u UEFIFirmwareDevice) QemuParams(config *Config) []string {
	var qemuParams []string

	if u.Code != "" {
		qemuParams = append(qemuParams, "-drive", "if=pflash,format=raw,readonly=on,file="+u.Code)
	}
	if u.Vars != "" {
		qemuParams = append(qemuParams, "-drive", "if=pflash,format=raw,file="+u.Vars)
	}

	return qemuParams
}

func (u UEFIFirmwareDevice) IsSecureBoot() bool {
	if strings.HasSuffix(u.Code, VMFSecboot) {
		return true
	}
	return false
}

func (u UEFIFirmwareDevice) Is4MB() bool {
	if strings.HasSuffix(u.Code, VMF4MB) {
		return true
	}
	return false
}

func (u UEFIFirmwareDevice) Exists() (bool, error) {
	if u.Code == "" {
		return false, fmt.Errorf("UEFIFirmwareDevice.Code is empty: %+v", u)
	}
	if u.Vars == "" {
		return false, fmt.Errorf("UEFIFirmwareDevice.Vars is empty: %+v", u)
	}
	codeFound := PathExists(u.Code)
	varsFound := PathExists(u.Vars)
	if codeFound && varsFound {
		return true, nil
	}
	missing := []string{}
	if !codeFound {
		missing = append(missing, fmt.Sprintf("Code not found at %q", u.Code))
	}
	if !varsFound {
		missing = append(missing, fmt.Sprintf("Vars not found at %q", u.Vars))
	}
	return false, fmt.Errorf("Failed to find UEFIFirmwareDevice paths: %s", strings.Join(missing, ", "))
}

// NewSystemUEFIFirmwareDevice looks at the local system to collect expected
// OVMF firmware files, callers will need to make a copy of the of the Vars
// template file before using it in a running VM.
func NewSystemUEFIFirmwareDevice(useSecureBoot bool) (*UEFIFirmwareDevice, error) {

	pfx := VMFPrefix()
	pathBase := VMFPathBase()

	// SecureBoot+4M
	secBoot4M := UEFIFirmwareDevice{
		Code: filepath.Join(pathBase, pfx+VMFCode+VMF4MB+VMFSecboot+VMFSuffix), // /usr/share/*VMF/*VMF_CODE_4M.secboot.fd
		Vars: filepath.Join(pathBase, pfx+VMFVars+VMF4MB+VMFSecboot+VMFSuffix), // OVMF_VARS_4M.secboot.fd
	}
	// SecureBoot+4M+MSVars
	secBoot4MVarsMs := UEFIFirmwareDevice{
		Code: filepath.Join(pathBase, pfx+VMFCode+VMF4MB+VMFSecboot+VMFSuffix), // OVMF_CODE_4M.secboot.fd
		Vars: filepath.Join(pathBase, pfx+VMFVars+VMF4MB+VMFMs+VMFSuffix),      // OVMF_VARS_4M.ms.fd
	}
	// SecureBoot
	secBoot := UEFIFirmwareDevice{
		Code: filepath.Join(pathBase, pfx+VMFCode+VMFSecboot+VMFSuffix), // OVMF_CODE.secboot.fd
		Vars: filepath.Join(pathBase, pfx+VMFVars+VMFSecboot+VMFSuffix), // OVMF_VARS.secboot.fd
	}
	// SecureBoot+MSVars (amd64 or arm64)
	secBootVarsMs := UEFIFirmwareDevice{
		Code: filepath.Join(pathBase, pfx+VMFCode+VMFMs+VMFSuffix), // {O,AA}VMF_CODE.ms.fd
		Vars: filepath.Join(pathBase, pfx+VMFVars+VMFMs+VMFSuffix), // {O,AA}VMF_VARS.ms.fd
	}
	// Insecure+4M
	insecureBoot4M := UEFIFirmwareDevice{
		Code: filepath.Join(pathBase, pfx+VMFCode+VMF4MB+VMFSuffix), // OVMF_CODE_4M.fd
		Vars: filepath.Join(pathBase, pfx+VMFVars+VMF4MB+VMFSuffix), // OVMF_Vars_4M.fd
	}

	// Insecure
	insecure := UEFIFirmwareDevice{
		Code: filepath.Join(pathBase, pfx+VMFCode+VMFSuffix), // {O,AA}VMF_CODE.fd
		Vars: filepath.Join(pathBase, pfx+VMFVars+VMFSuffix), // {O,AA}VMF_Vars.fd
	}

	var found bool
	var err error
	if useSecureBoot {
		// 4M  and .secboot variants are only on x86
		switch runtime.GOARCH {
		case "amd64", "x86_64":
			found, err = secBoot4MVarsMs.Exists()
			if err != nil {
				return &UEFIFirmwareDevice{}, fmt.Errorf("SecureBoot 4M MS Vars erorr: %s", err)
			}
			if found {
				return &secBoot4MVarsMs, nil
			}
			found, err = secBoot4M.Exists()
			if err != nil {
				return &UEFIFirmwareDevice{}, fmt.Errorf("SecureBoot 4M Vars error: %s", err)
			}
			if found {
				return &secBoot4M, nil
			}
			found, err = secBoot.Exists()
			if found {
				return &secBoot, nil
			}
			if err != nil {
				return &UEFIFirmwareDevice{}, fmt.Errorf("SecureBoot Vars error: %s", err)
			}

		}
		found, err = secBootVarsMs.Exists()
		if err != nil {
			return &UEFIFirmwareDevice{}, fmt.Errorf("SecureBoot MS Vars error: %s", err)
		}
		if found {
			return &secBootVarsMs, nil
		}

		return &UEFIFirmwareDevice{}, fmt.Errorf("%sVMF secureboot code,vars missing, check: %s", pfx, pathBase)
	}

	switch runtime.GOARCH {
	case "amd64", "x86_64":
		found, err = insecureBoot4M.Exists()
		if err != nil {
			return &UEFIFirmwareDevice{}, err
		}
		if found {
			return &insecureBoot4M, nil
		}
	}
	found, err = insecure.Exists()
	if found {
		return &insecure, nil
	}

	return &UEFIFirmwareDevice{}, fmt.Errorf("%sVMF code,vars missing, check: %s", pfx, pathBase)
}
