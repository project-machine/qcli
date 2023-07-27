package qcli

import "testing"

func TestAppendMachine(t *testing.T) {
	machineString := "-machine pc-lite,accel=kvm,kernel_irqchip=on,nvdimm=on"
	machine := Machine{
		Type:          "pc-lite",
		Acceleration:  MachineAccelerationKVM,
		KernelIRQChip: "on",
		NVDIMM:        "on",
	}
	testAppend(machine, machineString, t)

	machineString = "-machine pc-lite,accel=kvm,kernel_irqchip=on,nvdimm=on,gic-version=host,usb=off"
	machine = Machine{
		Type:          "pc-lite",
		Acceleration:  MachineAccelerationKVM,
		KernelIRQChip: "on",
		NVDIMM:        "on",
		Options:       "gic-version=host,usb=off",
	}
	testAppend(machine, machineString, t)

	machineString = "-machine microvm,accel=kvm,pic=off,pit=off"
	machine = Machine{
		Type:         "microvm",
		Acceleration: MachineAccelerationKVM,
		Options:      "pic=off,pit=off",
	}
	testAppend(machine, machineString, t)

	machineString = "-machine q35,accel=kvm,smm=on"
	machine = Machine{
		Type:         MachineTypePC35,
		Acceleration: MachineAccelerationKVM,
		SMM:          "on",
	}
	testAppend(machine, machineString, t)
}

func TestAppendEmptyMachine(t *testing.T) {
	machine := Machine{}

	testAppend(machine, "", t)
}

func TestBadMachine(t *testing.T) {
	c := &Config{}
	c.appendMachine()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestAppendMachineAarch64Virt(t *testing.T) {
	machineString := "-machine virt,accel=kvm"
	machine := Machine{
		Type:         MachineTypeVirt,
		Acceleration: MachineAccelerationKVM,
	}
	testAppend(machine, machineString, t)
}
