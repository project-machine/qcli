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
	SecCodePathAarch64      = "/usr/share/AAVMF/AAVM_CODE.ms.fd"
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

//Helper function to check and set CODE path for new UEFIFirmwareDevice
func (uefiDev *UEFIFirmwareDevice) checkAndSetCodePaths(codePaths []string) (bool){
	for _, cp := range codePaths {
		if PathExists(cp) {
			uefiDev.Code = cp
			return true
		}
	}
	return false
}
//Helper function to check and set VARS path for new UEFIFirmwareDevice
func (uefiDev *UEFIFirmwareDevice) checkAndSetVarPaths(varPaths []string) (bool){
	for _, vp := range varPaths {
		if PathExists(vp) {
			uefiDev.Vars = vp
			return true
		}
	}
	return false
}

// NewSystemUEFIFirmwareDevice looks at the local system to collect expected
// OVMF firmware files, callers will need to make a copy of the of the Vars
// template file before using it in a running VM.
func NewSystemUEFIFirmwareDevice(useSecureBoot bool) (*UEFIFirmwareDevice, error) {
	uefiDev := UEFIFirmwareDevice{}
	//can add in more paths as necessary
	var SecCodePaths = []string{SecCodePath, CentosSecVars, SecCodePathAarch64}
	var SecVarPaths = []string{UbuntuSecVarsAarch64, UbuntuSecVarsAarch64}
	var UnSecCodePaths = []string{UnSecCodePath, UnSecCodePathAarch64}
	var UnSecVarPaths = []string{UnSecVarsPath, UnSecVarsPathAarch64}
	var setCode bool
	var setVars bool

	if (useSecureBoot) {
		setCode = uefiDev.checkAndSetCodePaths(SecCodePaths)
		if (!setCode) {
			return &uefiDev, fmt.Errorf("Secureboot requested, but no secureboot CODE firmware file found")
		}
		setVars = uefiDev.checkAndSetVarPaths(SecVarPaths)
		if (!setVars) {
			return &uefiDev, fmt.Errorf("Secureboot requested, but no secureboot VARS firmware file found")
		}
	} else {
		setCode = uefiDev.checkAndSetCodePaths(UnSecCodePaths)
		if (!setCode) {
			setCode = uefiDev.checkAndSetCodePaths(SecCodePaths)
		}
		if (!setCode) {
			return &uefiDev, fmt.Errorf("Failed to find UEFI code firmware")
		}
		setVars = uefiDev.checkAndSetVarPaths(UnSecVarPaths)
		if (!setVars) {
			return &uefiDev, fmt.Errorf("Failed to find UEFI Vars firmware")
		}
	}
	return &uefiDev, nil
}
