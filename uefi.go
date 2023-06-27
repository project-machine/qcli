package qcli

import (
	"fmt"
)

type UEFIFirmwareDevice struct {
	Code string `yaml:"uefi-code"`
	Vars string `yaml:"uefi-vars"`
}

const (
	UEFIVarsFileName = "uefi_nvram.fd"
	SecCodePath      = "/usr/share/OVMF/OVMF_CODE.secboot.fd"
	UbuntuSecVars    = "/usr/share/OVMF/OVMF_VARS.ms.fd"
	CentosSecVars    = "/usr/share/OVMF/OVMF_VARS.secboot.fd"
	UnSecCodePath    = "/usr/share/OVMF/OVMF_CODE.fd"
	UnSecVarsPath    = "/usr/share/OVMF/OVMF_VARS.fd"
	SecCodePathAarch64      = "/usr/share/AAVMF/AAVMF_CODE.ms.fd"
	UbuntuSecVarsAarch64    = "/usr/share/AAVMF/AAVMF_VARS.ms.fd"
	UnSecCodePathAarch64    = "/usr/share/AAVMF/AAVMF_CODE.fd"
	UnSecVarsPathAarch64    = "/usr/share/AAVMF/AAVMF_VARS.fd"
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
		qemuParams = append(qemuParams, "-drive", "if=pflash,format=raw,readonly=on,file="+u.Code)
	}
	if u.Vars != "" {
		qemuParams = append(qemuParams, "-drive", "if=pflash,format=raw,file="+u.Vars)
	}

	return qemuParams
}

//Helper function to find paths (either code or vars firmware) for new UEFIFirmwareDevice
func selectPath(paths []string) string {
    for _, path := range paths {
        if PathExists(path) {
            return path
        }
    }
    return ""
}

// NewSystemUEFIFirmwareDevice looks at the local system to collect expected
// OVMF firmware files, callers will need to make a copy of the of the Vars
// template file before using it in a running VM.
func NewSystemUEFIFirmwareDevice(useSecureBoot bool) (*UEFIFirmwareDevice, error) {
	uefiDev := UEFIFirmwareDevice{}
	//can add in more paths as necessary
	var SecCodePaths = []string{SecCodePath, SecCodePathAarch64}
	var SecVarPaths = []string{UbuntuSecVarsAarch64, CentosSecVars, UbuntuSecVarsAarch64}
	var UnSecCodePaths = []string{UnSecCodePath, UnSecCodePathAarch64}
	var UnSecVarPaths = []string{UnSecVarsPath, UnSecVarsPathAarch64}

	if (useSecureBoot) {
		secCode := selectPath(SecCodePaths)
		if secCode == "" {
			return &uefiDev, fmt.Errorf("Secureboot requested, but no secureboot Code file found")
		}
		uefiDev.Code = secCode
		secVars := selectPath(SecVarPaths)
		if secVars == "" {
			return &uefiDev, fmt.Errorf("Secureboot requested, but no secureboot Vars file found")
		}
		uefiDev.Vars = secVars
	} else {
		codePath := selectPath(UnSecCodePaths)
		if codePath == "" {
			codePath = selectPath(SecCodePaths)
		}
		if codePath == "" {
			return &uefiDev, fmt.Errorf("Failed to find UEFI code firmware")
		}
		uefiDev.Code = codePath
		unsecVars := selectPath(UnSecVarPaths)
		if unsecVars == "" {
			return &uefiDev, fmt.Errorf("Secureboot requested, but no secureboot Vars file found")
		}
		uefiDev.Vars = unsecVars
	}
	return &uefiDev, nil
}
