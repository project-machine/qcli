package qcli

import (
	"fmt"
	"strings"
)

type UEFIFirmwareDevice struct {
	Code string `yaml:"uefi-code"`
	Vars string `yaml:"uefi-vars"`
}

const (
	UEFIVarsFileName = "uefi-nvram.fd"
	OVMFPathbase     = "/usr/share/OVMF/"
	OVMFCode         = "OVMF_CODE"
	OVMFVars         = "OVMF_VARS"
	OVMFVarsMs       = ".ms"
	OVMFSecboot      = ".secboot"
	OVMF4MB          = "_4M"
	OVMFSuffix       = ".fd"
)

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
		qemuParams = append(qemuParams, "-drive", "if=pflash,format=raw,readonly,file="+u.Code)
	}
	if u.Vars != "" {
		qemuParams = append(qemuParams, "-drive", "if=pflash,format=raw,file="+u.Vars)
	}

	return qemuParams
}

func (u UEFIFirmwareDevice) IsSecureBoot() bool {
	if strings.HasSuffix(u.Code, OVMFSecboot) {
		return true
	}
	return false
}

func (u UEFIFirmwareDevice) Is4MB() bool {
	if strings.HasSuffix(u.Code, OVMF4MB) {
		return true
	}
	return false
}

func (u UEFIFirmwareDevice) Exists() (bool, error) {
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
	// SecureBoot+4M
	secBoot4M := UEFIFirmwareDevice{
		Code: OVMFPathbase + OVMFCode + OVMF4MB + OVMFSecboot + OVMFSuffix, // OVMF_CODE_4M.secboot.fd
		Vars: OVMFPathbase + OVMFVars + OVMF4MB + OVMFSecboot + OVMFSuffix, // OVMF_VARS_4M.secboot.fd
	}
	// SecureBoot+4M+MSVars
	secBoot4MVarsMs := UEFIFirmwareDevice{
		Code: OVMFPathbase + OVMFCode + OVMF4MB + OVMFSecboot + OVMFSuffix, // OVMF_CODE_4M.secboot.fd
		Vars: OVMFPathbase + OVMFVars + OVMF4MB + OVMFVarsMs + OVMFSuffix,  // OVMF_VARS_4M.ms.fd
	}
	// SecureBoot
	secBoot := UEFIFirmwareDevice{
		Code: OVMFPathbase + OVMFCode + OVMFSecboot + OVMFSuffix, // OVMF_CODE.secboot.fd
		Vars: OVMFPathbase + OVMFVars + OVMFSecboot + OVMFSuffix, // OVMF_VARS.secboot.fd
	}
	// SecureBoot+MSVars
	secBootVarsMs := UEFIFirmwareDevice{
		Code: OVMFPathbase + OVMFCode + OVMFSecboot + OVMFSuffix, // OVMF_CODE.secboot.fd
		Vars: OVMFPathbase + OVMFVars + OVMFVarsMs + OVMFSuffix,  // OVMF_VARS.ms.fd
	}
	// Insecure+4M
	insecureBoot4M := UEFIFirmwareDevice{
		Code: OVMFPathbase + OVMFCode + OVMF4MB + OVMFSuffix, // OVMF_CODE_4M.fd
		Vars: OVMFPathbase + OVMFVars + OVMF4MB + OVMFSuffix, // OVMF_Vars_4M.fd
	}

	// Insecure
	insecure := UEFIFirmwareDevice{
		Code: OVMFPathbase + OVMFCode + OVMFSuffix, // OVMF_CODE.fd
		Vars: OVMFPathbase + OVMFVars + OVMFSuffix, // OVMF_Vars.fd
	}

	if useSecureBoot {
		found, err := secBoot4MVarsMs.Exists()
		if err != nil {
			return &UEFIFirmwareDevice{}, err
		}
		if found {
			return &secBoot4MVarsMs, nil
		}

		found, err = secBoot4M.Exists()
		if err != nil {
			return &UEFIFirmwareDevice{}, err
		}
		if found {
			return &secBoot4M, nil
		}

		found, err = secBootVarsMs.Exists()
		if err != nil {
			return &UEFIFirmwareDevice{}, err
		}
		if found {
			return &secBootVarsMs, nil
		}

		found, err = secBoot.Exists()
		if found {
			return &secBoot, nil
		}
		if err != nil {
			return &UEFIFirmwareDevice{}, err
		}

		return &UEFIFirmwareDevice{}, fmt.Errorf("OVMF secureboot code,vars missing, check: %s", OVMFPathbase)
	}

	found, err := insecureBoot4M.Exists()
	if err != nil {
		return &UEFIFirmwareDevice{}, err
	}
	if found {
		return &insecureBoot4M, nil
	}

	found, err = insecure.Exists()
	if found {
		return &insecure, nil
	}

	return &UEFIFirmwareDevice{}, fmt.Errorf("OVMF code,vars missing, check: %s", OVMFPathbase)
}
