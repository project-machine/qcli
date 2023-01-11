package qcli

import (
	"fmt"
)

type UEFIFirmwareDevice struct {
	Code string `yaml:"uefi-code"`
	Vars string `yaml:"uefi-vars"`
}

const (
	UEFIVarsFileName = "uefi-nvram.fd"
	SecCodePath      = "/usr/share/OVMF/OVMF_CODE.secboot.fd"
	UbuntuSecVars    = "/usr/share/OVMF/OVMF_VARS.ms.fd"
	CentosSecVars    = "/usr/share/OVMF/OVMF_VARS.secboot.fd"
	UnSecCodePath    = "/usr/share/OVMF/OVMF_CODE.fd"
	UnSecVarsPath    = "/usr/share/OVMF/OVMF_VARS.fd"
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

// NewSystemUEFIFirmwareDevice looks at the local system to collect expected
// OVMF firmware files, callers will need to make a copy of the of the Vars
// template file before using it in a running VM.
func NewSystemUEFIFirmwareDevice(useSecureBoot bool) (*UEFIFirmwareDevice, error) {
	var uefiDev *UEFIFirmwareDevice

	if useSecureBoot {
		uefiDev.Code = SecCodePath
		if PathExists(UbuntuSecVars) {
			uefiDev.Vars = UbuntuSecVars
		} else if PathExists(CentosSecVars) {
			uefiDev.Vars = CentosSecVars
		} else {
			return uefiDev, fmt.Errorf("secureboot requested, but no secureboot OVMF variables found")
		}
	} else {
		if PathExists(UnSecCodePath) {
			uefiDev.Code = UnSecCodePath
		} else {
			uefiDev.Code = SecCodePath
		}
		if PathExists(UnSecVarsPath) {
			uefiDev.Vars = UnSecVarsPath
		} else {
			return uefiDev, fmt.Errorf("OMVF variables template missing: %s", UnSecVarsPath)
		}
	}

	return uefiDev, nil
}
