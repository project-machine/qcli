package qcli

import "testing"

func TestAppendMachineAarch64Virt(t *testing.T){
	machineString := "-machine virt,accel=kvm"
	machine := Machine {
		Type:			MachineTypeVirt,
		Acceleration:	MachineAccelerationKVM,
	}
	testAppend(machine, machineString, t)
}
