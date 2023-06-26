// +build aarch64 arm64

package qcli

import "testing"

func TestAppendMachineAarch64Virt(t *testing.T){
	machineString := "-machine virt,accel=kvm"
	machine := Machine {
		Type:			MachineTypeVirtAarch64,
		Acceleration:	MachineAccelerationKVM,
	}
	testAppend(machine, machineString, t)
}
