/*
// Copyright contributors to the Virtual Machine Manager for Go project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
*/

package qcli

import (
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

const agentUUID = "4cb19522-1e18-439a-883a-f9b2a3a95f5e"
const volumeUUID = "67d86208-b46c-4465-9018-e14187d4010"

const DevNo = "fe.1.1234"

func testAppend(structure interface{}, expected string, t *testing.T) {
	var config Config
	testConfigAppend(&config, structure, expected, t)
}

func testConfig(config *Config, expected string, t *testing.T) {
	params, err := ConfigureParams(config, nil)
	if err != nil {
		t.Fatalf("Failed to append parameters: %s", err.Error())
	}
	result := strings.Join(params, " ")
	if expected != result {
		t.Fatalf("Failed to append parameters\nexpected[%s]\n!=\n   found[%s]", expected, result)
	}
}

func testConfigAppend(config *Config, structure interface{}, expected string, t *testing.T) {

	switch s := structure.(type) {
	case Machine:
		config.Machine = s
		config.appendMachine()
	case FwCfg:
		config.FwCfg = []FwCfg{s}
		config.appendFwCfg(nil)

	case Device:
		config.devices = []Device{s}
		err := config.appendDevices()
		if err != nil {
			t.Fatalf("Failed to append Device '%v', error: %s", s, err)
		}
	case Object:
		objParams := s.QemuParams(config)
		config.qemuParams = append(config.qemuParams, objParams...)

	case TPMDevice:
		config.TPM = s
		config.appendTPM()

	case SMBIOSInfo:
		config.SMBIOS = s
		if err := config.appendSMBIOSInfo(); err != nil {
			t.Fatalf("Failed ot append SMBIOS '%v', error: %s", s, err)
		}

	case Knobs:
		config.Knobs = s
		config.appendKnobs()

	case Kernel:
		config.Kernel = s
		config.appendKernel()

	case Memory:
		config.Memory = s
		config.appendMemory()

	case SMP:
		config.SMP = s
		if err := config.appendCPUs(); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

	case QMPSocket:
		config.QMPSockets = []QMPSocket{s}
		config.appendQMPSockets()

	case []QMPSocket:
		config.QMPSockets = s
		config.appendQMPSockets()

	case RTC:
		config.RTC = s
		config.appendRTC()

	case IOThread:
		config.IOThreads = []IOThread{s}
		config.appendIOThreads()
	case Incoming:
		config.Incoming = s
		config.appendIncoming()
	}

	result := strings.Join(config.qemuParams, " ")
	if expected != result {
		t.Fatalf("Failed to append parameters\nexpected[%s]\n!=\n   found[%s]", expected, result)
	}
}

func TestAppendKnobsAllTrue(t *testing.T) {
	var knobsString = "-no-user-config -nodefaults -nographic --no-reboot -daemonize -overcommit mem-lock=on -S -no-hpet -snapshot"
	knobs := Knobs{
		NoUserConfig:  true,
		NoDefaults:    true,
		NoGraphic:     true,
		NoReboot:      true,
		Daemonize:     true,
		MemPrealloc:   true,
		FileBackedMem: true,
		MemShared:     true,
		Mlock:         true,
		Stopped:       true,
		NoHPET:        true,
		Snapshot:      true,
	}

	testAppend(knobs, knobsString, t)
}

func TestAppendKnobsAllFalse(t *testing.T) {
	var knobsString = ""
	knobs := Knobs{
		NoUserConfig:  false,
		NoDefaults:    false,
		NoGraphic:     false,
		NoReboot:      false,
		MemPrealloc:   false,
		FileBackedMem: false,
		MemShared:     false,
		Mlock:         false,
		Stopped:       false,
		NoHPET:        false,
		Snapshot:      false,
	}

	testAppend(knobs, knobsString, t)
}

func TestAppendMemoryHugePages(t *testing.T) {
	conf := &Config{
		Memory: Memory{
			Size:   "1G",
			Slots:  8,
			MaxMem: "3G",
			Path:   "foobar",
		},
	}
	memString := "-m 1G,slots=8,maxmem=3G"
	testConfigAppend(conf, conf.Memory, memString, t)

	knobs := Knobs{
		HugePages:     true,
		MemPrealloc:   true,
		FileBackedMem: true,
		MemShared:     true,
	}
	objMemString := "-object memory-backend-file,id=dimm1,size=1G,mem-path=/dev/hugepages,share=on,prealloc=on"
	numaMemString := "-numa node,memdev=dimm1"
	memBackendString := "-machine memory-backend=dimm1"

	knobsString := objMemString + " "
	if isDimmSupported(nil) {
		knobsString += numaMemString
	} else {
		knobsString += memBackendString
	}

	testConfigAppend(conf, knobs, memString+" "+knobsString, t)
}

func TestAppendMemoryMemPrealloc(t *testing.T) {
	conf := &Config{
		Memory: Memory{
			Size:   "1G",
			Slots:  8,
			MaxMem: "3G",
			Path:   "foobar",
		},
	}
	memString := "-m 1G,slots=8,maxmem=3G"
	testConfigAppend(conf, conf.Memory, memString, t)

	knobs := Knobs{
		MemPrealloc: true,
		MemShared:   true,
	}
	objMemString := "-object memory-backend-ram,id=dimm1,size=1G,share=on,prealloc=on"
	numaMemString := "-numa node,memdev=dimm1"
	memBackendString := "-machine memory-backend=dimm1"

	knobsString := objMemString + " "
	if isDimmSupported(nil) {
		knobsString += numaMemString
	} else {
		knobsString += memBackendString
	}

	testConfigAppend(conf, knobs, memString+" "+knobsString, t)
}

func TestAppendMemoryMemShared(t *testing.T) {
	conf := &Config{
		Memory: Memory{
			Size:   "1G",
			Slots:  8,
			MaxMem: "3G",
			Path:   "foobar",
		},
	}
	memString := "-m 1G,slots=8,maxmem=3G"
	testConfigAppend(conf, conf.Memory, memString, t)

	knobs := Knobs{
		FileBackedMem: true,
		MemShared:     true,
	}
	objMemString := "-object memory-backend-file,id=dimm1,size=1G,mem-path=foobar,share=on"
	numaMemString := "-numa node,memdev=dimm1"
	memBackendString := "-machine memory-backend=dimm1"

	knobsString := objMemString + " "
	if isDimmSupported(nil) {
		knobsString += numaMemString
	} else {
		knobsString += memBackendString
	}

	testConfigAppend(conf, knobs, memString+" "+knobsString, t)
}

func TestAppendMemoryFileBackedMem(t *testing.T) {
	conf := &Config{
		Memory: Memory{
			Size:   "1G",
			Slots:  8,
			MaxMem: "3G",
			Path:   "foobar",
		},
	}
	memString := "-m 1G,slots=8,maxmem=3G"
	testConfigAppend(conf, conf.Memory, memString, t)

	knobs := Knobs{
		FileBackedMem: true,
		MemShared:     false,
	}
	objMemString := "-object memory-backend-file,id=dimm1,size=1G,mem-path=foobar"
	numaMemString := "-numa node,memdev=dimm1"
	memBackendString := "-machine memory-backend=dimm1"

	knobsString := objMemString + " "
	if isDimmSupported(nil) {
		knobsString += numaMemString
	} else {
		knobsString += memBackendString
	}

	testConfigAppend(conf, knobs, memString+" "+knobsString, t)
}

func TestAppendMemoryFileBackedMemPrealloc(t *testing.T) {
	conf := &Config{
		Memory: Memory{
			Size:   "1G",
			Slots:  8,
			MaxMem: "3G",
			Path:   "foobar",
		},
	}
	memString := "-m 1G,slots=8,maxmem=3G"
	testConfigAppend(conf, conf.Memory, memString, t)

	knobs := Knobs{
		FileBackedMem: true,
		MemShared:     true,
		MemPrealloc:   true,
	}
	objMemString := "-object memory-backend-file,id=dimm1,size=1G,mem-path=foobar,share=on,prealloc=on"
	numaMemString := "-numa node,memdev=dimm1"
	memBackendString := "-machine memory-backend=dimm1"

	knobsString := objMemString + " "
	if isDimmSupported(nil) {
		knobsString += numaMemString
	} else {
		knobsString += memBackendString
	}

	testConfigAppend(conf, knobs, memString+" "+knobsString, t)
}

func TestNoRebootKnob(t *testing.T) {
	conf := &Config{}

	knobs := Knobs{
		NoReboot: true,
	}
	knobsString := "--no-reboot"

	testConfigAppend(conf, knobs, knobsString, t)
}

var kernelString = "-kernel /opt/vmlinux.container -initrd /opt/initrd.container -append root=/dev/pmem0p1 rootflags=dax,data=ordered,errors=remount-ro rw rootfstype=ext4 tsc=reliable"

func TestAppendKernel(t *testing.T) {
	kernel := Kernel{
		Path:       "/opt/vmlinux.container",
		InitrdPath: "/opt/initrd.container",
		Params:     "root=/dev/pmem0p1 rootflags=dax,data=ordered,errors=remount-ro rw rootfstype=ext4 tsc=reliable",
	}

	testAppend(kernel, kernelString, t)
}

var memoryString = "-m 2G,slots=2,maxmem=3G"

func TestAppendMemory(t *testing.T) {
	memory := Memory{
		Size:   "2G",
		Slots:  2,
		MaxMem: "3G",
		Path:   "",
	}

	testAppend(memory, memoryString, t)
}

var cpusString = "-smp 2,cores=1,threads=2,sockets=2,maxcpus=6"

func TestAppendCPUs(t *testing.T) {
	smp := SMP{
		CPUs:    2,
		Sockets: 2,
		Cores:   1,
		Threads: 2,
		MaxCPUs: 6,
	}

	testAppend(smp, cpusString, t)
}

func TestFailToAppendCPUs(t *testing.T) {
	config := Config{
		SMP: SMP{
			CPUs:    2,
			Sockets: 2,
			Cores:   1,
			Threads: 2,
			MaxCPUs: 1,
		},
	}

	if err := config.appendCPUs(); err == nil {
		t.Fatalf("Expected appendCPUs to fail")
	}
}

var pidfile = "/run/vc/vm/iamsandboxid/pidfile"
var logfile = "/run/vc/vm/iamsandboxid/logfile"
var qemuString = "-name cc-qemu -cpu host -uuid " + agentUUID + " -pidfile " + pidfile + " -D " + logfile

func TestAppendStrings(t *testing.T) {
	config := Config{
		Path:     "qemu",
		Name:     "cc-qemu",
		UUID:     agentUUID,
		CPUModel: "host",
		PidFile:  pidfile,
		LogFile:  logfile,
	}

	config.appendName()
	config.appendCPUModel()
	config.appendUUID()
	config.appendPidFile()
	config.appendLogFile()

	result := strings.Join(config.qemuParams, " ")
	if result != qemuString {
		t.Fatalf("Failed to append parameters [%s] != [%s]", result, qemuString)
	}
}

var ioThreadString = "-object iothread,id=iothread1"

func TestAppendIOThread(t *testing.T) {
	ioThread := IOThread{
		ID: "iothread1",
	}

	testAppend(ioThread, ioThreadString, t)
}

var incomingStringFD = "-S -incoming fd:3"

func TestAppendIncomingFD(t *testing.T) {
	source := Incoming{
		MigrationType: MigrationFD,
		FD:            os.Stdout,
	}

	testAppend(source, incomingStringFD, t)
}

var incomingStringExec = "-S -incoming exec:test migration cmd"

func TestAppendIncomingExec(t *testing.T) {
	source := Incoming{
		MigrationType: MigrationExec,
		Exec:          "test migration cmd",
	}

	testAppend(source, incomingStringExec, t)
}

var incomingStringDefer = "-S -incoming defer"

func TestAppendIncomingDefer(t *testing.T) {
	source := Incoming{
		MigrationType: MigrationDefer,
	}

	testAppend(source, incomingStringDefer, t)
}

func TestBadName(t *testing.T) {
	c := &Config{}
	c.appendName()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestBadCPUModel(t *testing.T) {
	c := &Config{}
	c.appendCPUModel()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestValidCPUModelAndCPUModelFlags(t *testing.T) {
	c := &Config{
		CPUModel:      "host",
		CPUModelFlags: []string{"+flag1", "-flag2"},
	}
	c.appendCPUModel()
	expected := []string{"-cpu", "host,+flag1,-flag2"}
	ok := reflect.DeepEqual(expected, c.qemuParams)
	if !ok {
		t.Errorf("Expected %v, found %v", expected, c.qemuParams)
	}
}

func TestBadDevices(t *testing.T) {
	c := &Config{}
	c.appendDevices()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}

	c = &Config{
		devices: []Device{
			FSDevice{},
			FSDevice{
				ID:       "id0",
				MountTag: "tag",
			},
			CharDevice{},
			CharDevice{
				ID: "id1",
			},
			NetDevice{},
			NetDevice{
				ID:   "id1",
				Type: IPVTAP,
				Tap: NetDeviceTap{
					IFName: "if",
				},
			},
			SerialDevice{},
			SerialDevice{
				ID: "id0",
			},
			BlockDevice{},
			BlockDevice{
				Driver: "drv",
				ID:     "id1",
			},
			VhostUserDevice{},
			VhostUserDevice{
				CharDevID: "devid",
			},
			VhostUserDevice{
				CharDevID:  "devid",
				SocketPath: "/var/run/sock",
			},
			VhostUserDevice{
				CharDevID:     "devid",
				SocketPath:    "/var/run/sock",
				VhostUserType: VhostUserNet,
			},
			VhostUserDevice{
				CharDevID:     "devid",
				SocketPath:    "/var/run/sock",
				VhostUserType: VhostUserSCSI,
			},
		},
	}

	c.appendDevices()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestBadGlobalParams(t *testing.T) {
	c := &Config{}
	c.appendGlobalParams()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestValidGlobalParams(t *testing.T) {
	c := &Config{GlobalParams: []string{"param1", "param2"}}
	expected := []string{"-global", "param1", "-global", "param2"}
	c.appendGlobalParams()
	ok := reflect.DeepEqual(expected, c.qemuParams)
	if !ok {
		t.Errorf("Expected %v, found %v", expected, c.qemuParams)
	}
}

func TestBadPFlash(t *testing.T) {
	c := &Config{}
	c.appendPFlashParam()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestValidPFlash(t *testing.T) {
	c := &Config{}
	c.PFlash = []string{"flash0", "flash1"}
	c.appendPFlashParam()
	expected := []string{"-pflash", "flash0", "-pflash", "flash1"}
	ok := reflect.DeepEqual(expected, c.qemuParams)
	if !ok {
		t.Errorf("Expected %v, found %v", expected, c.qemuParams)
	}
}

func TestBadSeccompSandbox(t *testing.T) {
	c := &Config{}
	c.appendSeccompSandbox()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestValidSeccompSandbox(t *testing.T) {
	c := &Config{}
	c.SeccompSandbox = string("on,obsolete=deny")
	c.appendSeccompSandbox()
	expected := []string{"-sandbox", "on,obsolete=deny"}
	ok := reflect.DeepEqual(expected, c.qemuParams)
	if !ok {
		t.Errorf("Expected %v, found %v", expected, c.qemuParams)
	}
}

func TestBadVGA(t *testing.T) {
	c := &Config{}
	c.appendVGA()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestBadKernel(t *testing.T) {
	c := &Config{}
	c.appendKernel()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestBadMemoryKnobs(t *testing.T) {
	c := &Config{}
	c.appendMemoryKnobs()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}

	c = &Config{
		Knobs: Knobs{
			HugePages: true,
		},
	}
	c.appendMemoryKnobs()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}

	c = &Config{
		Knobs: Knobs{
			MemShared: true,
		},
	}
	c.appendMemoryKnobs()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}

	c = &Config{
		Knobs: Knobs{
			MemPrealloc: true,
		},
	}
	c.appendMemoryKnobs()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestBadBios(t *testing.T) {
	c := &Config{}
	c.appendBios()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestBadIOThreads(t *testing.T) {
	c := &Config{}
	c.appendIOThreads()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}

	c = &Config{
		IOThreads: []IOThread{{ID: ""}},
	}
	c.appendIOThreads()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestBadIncoming(t *testing.T) {
	c := &Config{}
	c.appendIncoming()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}

func TestBadCPUs(t *testing.T) {
	c := &Config{}
	if err := c.appendCPUs(); err != nil {
		t.Fatalf("No error expected got %v", err)
	}
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}

	c = &Config{
		SMP: SMP{
			MaxCPUs: 1,
			CPUs:    2,
		},
	}
	if c.appendCPUs() == nil {
		t.Errorf("Error expected")
	}
}

var (
	fullUefiVM           = "-machine q35,accel=kvm,smm=on -cpu qemu64,+x2apic -m 4096 -device pcie-root-port,id=root-port.0x4.0,bus=pcie.0,chassis=0x0,slot=0x00,port=0x0,addr=0x5,multifunction=on -device pcie-root-port,id=root-port.0x4.1,bus=pcie.0,chassis=0x1,slot=0x00,port=0x1,addr=0x5.0x1 -object rng-random,id=rng0,filename=/dev/urandom -device virtio-rng-pci,rng=rng0,bus=pcie.0,addr=0x03 -drive file=boot.qcow2,id=drive0,if=none,format=qcow2,aio=threads,cache=unsafe,discard=unmap,detect-zeroes=unmap -device virtio-blk-pci,drive=drive0,serial=ssd-boot,bootindex=0,disable-modern=true,addr=0x04,bus=pcie.0,logical_block_size=512,physical_block_size=512,scsi=off,config-wce=off -netdev user,id=user0,ipv4=on,hostfwd=tcp::22222-:22 -device virtio-net-pci,netdev=user0,mac=01:02:de:ad:be:ef,bus=pcie.0,disable-modern=false -chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off -chardev socket,id=monitor0,path=/tmp/monitor.sock,server=on,wait=off -serial chardev:serial0 -monitor chardev:monitor0 -drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE.fd -drive if=pflash,format=raw,file=uefi_nvram.fd -global ICH9-LPC.disable_s3=1 -global driver=cfi.pflash01,property=secure,value=on -object memory-backend-file,id=dimm1,size=4096,mem-path=/dev/hugepages,share=on,prealloc=on -numa node,memdev=dimm1 -nographic -no-hpet -snapshot -smp 4"
	fullBiosVM           = "-machine q35,accel=kvm,smm=on -cpu qemu64,+x2apic -m 4096 -device pcie-root-port,id=root-port.0x4.0,bus=pcie.0,chassis=0x0,slot=0x00,port=0x0,addr=0x5,multifunction=on -device pcie-root-port,id=root-port.0x4.1,bus=pcie.0,chassis=0x1,slot=0x00,port=0x1,addr=0x5.0x1 -object rng-random,id=rng0,filename=/dev/urandom -device virtio-rng-pci,rng=rng0,bus=pcie.0,addr=0x03 -drive file=boot.qcow2,id=drive0,if=none,format=qcow2,aio=threads,cache=unsafe,discard=unmap,detect-zeroes=unmap -device virtio-blk-pci,drive=drive0,serial=ssd-boot,bootindex=0,disable-modern=true,addr=0x04,bus=pcie.0,logical_block_size=512,physical_block_size=512,scsi=off,config-wce=off -netdev user,id=user0,ipv4=on,hostfwd=tcp::22222-:22 -device virtio-net-pci,netdev=user0,mac=01:02:de:ad:be:ef,bus=pcie.0,disable-modern=false -chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off -chardev socket,id=monitor0,path=/tmp/monitor.sock,server=on,wait=off -serial chardev:serial0 -monitor chardev:monitor0 -global ICH9-LPC.disable_s3=1 -global driver=cfi.pflash01,property=secure,value=on -object memory-backend-file,id=dimm1,size=4096,mem-path=/dev/hugepages,share=on,prealloc=on -numa node,memdev=dimm1 -nographic -no-hpet -snapshot -smp 4"
	fullUefiVMSpice      = "-machine q35,accel=kvm,smm=on -cpu qemu64,+x2apic -m 4096 -spice port=5901,addr=127.0.0.1 -device virtio-serial-pci -device virtserialport,chardev=spicechannel0,name=com.redhat.spice.0 -chardev spicevmc,id=spicechannel0,name=vdagent -device pcie-root-port,id=root-port.0x4.0,bus=pcie.0,chassis=0x0,slot=0x00,port=0x0,addr=0x5,multifunction=on -device pcie-root-port,id=root-port.0x4.1,bus=pcie.0,chassis=0x1,slot=0x00,port=0x1,addr=0x5.0x1 -object rng-random,id=rng0,filename=/dev/urandom -device virtio-rng-pci,rng=rng0,bus=pcie.0,addr=0x03 -drive file=boot.qcow2,id=drive0,if=none,format=qcow2,aio=threads,cache=unsafe,discard=unmap,detect-zeroes=unmap -device virtio-blk-pci,drive=drive0,serial=ssd-boot,bootindex=0,disable-modern=true,addr=0x04,bus=pcie.0,logical_block_size=512,physical_block_size=512,scsi=off,config-wce=off -netdev user,id=user0,ipv4=on,hostfwd=tcp::22222-:22 -device virtio-net-pci,netdev=user0,mac=01:02:de:ad:be:ef,bus=pcie.0,disable-modern=false -chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off -chardev socket,id=monitor0,path=/tmp/monitor.sock,server=on,wait=off -serial chardev:serial0 -monitor chardev:monitor0 -drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE.fd -drive if=pflash,format=raw,file=uefi_nvram.fd -global ICH9-LPC.disable_s3=1 -global driver=cfi.pflash01,property=secure,value=on -object memory-backend-file,id=dimm1,size=4096,mem-path=/dev/hugepages,share=on,prealloc=on -numa node,memdev=dimm1 -nographic -no-hpet -snapshot -smp 4"
	fullUefiVMTPM        = "-machine q35,accel=kvm,smm=on -cpu qemu64,+x2apic -m 4096 -chardev socket,id=chrtpm0,path=tpm.socket -tpmdev emulator,id=tpm0,chardev=chrtpm0 -device tpm-tis,tpmdev=tpm0 -device pcie-root-port,id=root-port.0x4.0,bus=pcie.0,chassis=0x0,slot=0x00,port=0x0,addr=0x5,multifunction=on -device pcie-root-port,id=root-port.0x4.1,bus=pcie.0,chassis=0x1,slot=0x00,port=0x1,addr=0x5.0x1 -object rng-random,id=rng0,filename=/dev/urandom -device virtio-rng-pci,rng=rng0,bus=pcie.0,addr=0x03 -drive file=boot.qcow2,id=drive0,if=none,format=qcow2,aio=threads,cache=unsafe,discard=unmap,detect-zeroes=unmap -device virtio-blk-pci,drive=drive0,serial=ssd-boot,bootindex=0,disable-modern=true,addr=0x04,bus=pcie.0,logical_block_size=512,physical_block_size=512,scsi=off,config-wce=off -netdev user,id=user0,ipv4=on,hostfwd=tcp::22222-:22 -device virtio-net-pci,netdev=user0,mac=01:02:de:ad:be:ef,bus=pcie.0,disable-modern=false -chardev socket,id=serial0,path=/tmp/console.sock,server=on,wait=off -chardev socket,id=monitor0,path=/tmp/monitor.sock,server=on,wait=off -serial chardev:serial0 -monitor chardev:monitor0 -drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE.fd -drive if=pflash,format=raw,file=uefi_nvram.fd -global ICH9-LPC.disable_s3=1 -global driver=cfi.pflash01,property=secure,value=on -object memory-backend-file,id=dimm1,size=4096,mem-path=/dev/hugepages,share=on,prealloc=on -numa node,memdev=dimm1 -nographic -no-hpet -snapshot -smp 4"
	fullUefiAarch64VM    = "-machine virt,accel=kvm -cpu host -m 1G -drive file=udisk.img,id=hd0,if=none,format=qcow2 -device virtio-blk-pci,drive=hd0,serial=hd0,disable-modern=false,addr=0x1e,bus=pcie.0,scsi=off,config-wce=off -drive file=ubuntu-22.04.2-live-server-arm64.iso,id=cdrom0,if=none,format=raw,media=cdrom,readonly=on -device virtio-blk-pci,drive=cdrom0,serial=cdrom0,bootindex=0,disable-modern=false,addr=0x1d,bus=pcie.0,scsi=off,config-wce=off -drive if=pflash,format=raw,readonly=on,file=/usr/share/AAVMF/AAVMF_CODE.ms.fd -drive if=pflash,format=raw,file=uefi_nvram.fd -object memory-backend-ram,id=dimm1,size=1G -numa node,memdev=dimm1 -nographic"
	fullUefiAarch64VMTPM = "-machine virt,accel=kvm -cpu host -m 1G -chardev socket,id=chrtpm0,path=tpm.socket -tpmdev emulator,id=tpm0,chardev=chrtpm0 -device tpm-tis-device,tpmdev=tpm0 -drive file=udisk.img,id=hd0,if=none,format=qcow2 -device virtio-blk-pci,drive=hd0,serial=hd0,disable-modern=false,addr=0x1e,bus=pcie.0,scsi=off,config-wce=off -drive file=ubuntu-22.04.2-live-server-arm64.iso,id=cdrom0,if=none,format=raw,media=cdrom,readonly=on -device virtio-blk-pci,drive=cdrom0,serial=cdrom0,bootindex=0,disable-modern=false,addr=0x1d,bus=pcie.0,scsi=off,config-wce=off -drive if=pflash,format=raw,readonly=on,file=/usr/share/AAVMF/AAVMF_CODE.ms.fd -drive if=pflash,format=raw,file=uefi_nvram.fd -object memory-backend-ram,id=dimm1,size=1G -numa node,memdev=dimm1 -nographic"
)

func fullVMConfig() *Config {
	c := &Config{
		Machine: Machine{
			Type:         MachineTypePC35,
			Acceleration: MachineAccelerationKVM,
			SMM:          "on",
		},
		CPUModel:      "qemu64",
		CPUModelFlags: []string{"+x2apic"},
		Memory: Memory{
			Size: "4096",
		},
		RngDevices: []RngDevice{
			RngDevice{
				Driver:    VirtioRng,
				ID:        "rng0",
				Bus:       "pcie.0",
				Transport: TransportPCI,
				Filename:  RngDevUrandom,
				Addr:      "3",
			},
		},
		BlkDevices: []BlockDevice{
			BlockDevice{
				Driver:        VirtioBlock,
				ID:            "drive0",
				File:          "boot.qcow2",
				BusAddr:       "4",
				AIO:           Threads,
				Format:        QCOW2,
				Interface:     NoInterface,
				DisableModern: true,
				Serial:        "ssd-boot",
				BlockSize:     512,
				Cache:         CacheModeUnsafe,
				Discard:       DiscardUnmap,
				DetectZeroes:  DetectZeroesUnmap,
				BootIndex:     "0",
			},
		},
		NetDevices: []NetDevice{
			NetDevice{
				Driver:     VirtioNet,
				Type:       USER,
				ID:         "user0",
				MACAddress: "01:02:de:ad:be:ef",
				Bus:        "pcie.0",
				User: NetDeviceUser{
					IPV4: true,
					HostForward: []PortRule{
						PortRule{
							Protocol: "tcp",
							Host:     Port{Port: 22222},
							Guest:    Port{Port: 22},
						},
					},
				},
			},
		},
		CharDevices: []CharDevice{
			CharDevice{
				Driver:  LegacySerial,
				Backend: Socket,
				ID:      "serial0",
				Path:    "/tmp/console.sock",
			},
			CharDevice{
				Driver:  LegacySerial,
				Backend: Socket,
				ID:      "monitor0",
				Path:    "/tmp/monitor.sock",
			},
		},
		LegacySerialDevices: []LegacySerialDevice{
			LegacySerialDevice{
				ChardevID: "serial0",
			},
		},
		MonitorDevices: []MonitorDevice{
			MonitorDevice{
				ChardevID: "monitor0",
			},
		},
		PCIeRootPortDevices: []PCIeRootPortDevice{
			PCIeRootPortDevice{
				ID:            "root-port.0x4.0",
				Bus:           "pcie.0",
				Chassis:       "0x0",
				Slot:          "0x00",
				Port:          "0x0",
				Addr:          "0x5",
				Multifunction: true,
			},
			PCIeRootPortDevice{
				ID:            "root-port.0x4.1",
				Bus:           "pcie.0",
				Chassis:       "0x1",
				Slot:          "0x00",
				Port:          "0x1",
				Addr:          "0x5.0x1",
				Multifunction: false,
			},
		},
		GlobalParams: []string{
			"ICH9-LPC.disable_s3=1",
			"driver=cfi.pflash01,property=secure,value=on",
		},
		Knobs: Knobs{
			NoGraphic:     true,
			NoHPET:        true,
			Snapshot:      true,
			HugePages:     true,
			MemPrealloc:   true,
			FileBackedMem: true,
			MemShared:     true,
		},
		SMP: SMP{
			CPUs: 4,
		},
	}
	return c
}

func fullVMConfigAarch64() *Config {
	c := &Config{
		Machine: Machine{
			Type:         MachineTypeVirt,
			Acceleration: MachineAccelerationKVM,
		},
		CPUModel: "host",
		Memory: Memory{
			Size: "1G",
		},
		BlkDevices: []BlockDevice{
			BlockDevice{
				Driver:    VirtioBlock,
				ID:        "hd0",
				File:      "udisk.img",
				Format:    QCOW2,
				Interface: NoInterface,
			},
			BlockDevice{
				Driver:    VirtioBlock,
				Interface: NoInterface,
				ID:        "cdrom0",
				File:      "ubuntu-22.04.2-live-server-arm64.iso",
				Format:    RAW,
				ReadOnly:  true,
				Media:     "cdrom",
				BootIndex: "0",
			},
		},
		Knobs: Knobs{
			NoGraphic: true,
		},
	}
	return c
}

func TestFullUEFIMachineCommand(t *testing.T) {
	var c *Config
	u := UEFIFirmwareDevice{Code: "", Vars: "uefi_nvram.fd"}
	expected := ""
	switch runtime.GOARCH {
	case "aarch64", "arm64":
		c = fullVMConfigAarch64()
		u.Code = "/usr/share/AAVMF/AAVMF_CODE.ms.fd"
		expected = fullUefiAarch64VM
	case "x86_64", "amd64":
		c = fullVMConfig()
		u.Code = "/usr/share/OVMF/OVMF_CODE.fd"
		expected = fullUefiVM
	}

	c.UEFIFirmwareDevices = append(c.UEFIFirmwareDevices, u)

	qemuParams, err := ConfigureParams(c, nil)
	if err != nil {
		t.Fatalf("Failed to Configure parameters, error: %s", err.Error())
	}
	result := strings.Join(qemuParams, " ")
	if expected != result {
		t.Fatalf("Failed to append parameters\nexpected[%s]\n!=\nfound    [%s]", expected, result)
	}
}

func TestFullUEFISpiceMachineCommand(t *testing.T) {
	c := fullVMConfig()

	u := UEFIFirmwareDevice{
		Code: "/usr/share/OVMF/OVMF_CODE.fd",
		Vars: "uefi_nvram.fd",
	}
	c.UEFIFirmwareDevices = append(c.UEFIFirmwareDevices, u)

	c.SpiceDevice = SpiceDevice{Port: "5901"}

	expected := fullUefiVMSpice
	qemuParams, err := ConfigureParams(c, nil)
	if err != nil {
		t.Fatalf("Failed to Configure parameters, error: %s", err.Error())
	}
	result := strings.Join(qemuParams, " ")
	if expected != result {
		t.Fatalf("Failed to append parameters\nexpected[%s]\n!=\nfound    [%s]", expected, result)
	}
}

func TestFullUEFITPMMachineCommand(t *testing.T) {
	var c *Config
	u := UEFIFirmwareDevice{Code: "", Vars: "uefi_nvram.fd"}
	expected := ""
	switch runtime.GOARCH {
	case "aarch64", "arm64":
		c = fullVMConfigAarch64()
		u.Code = "/usr/share/AAVMF/AAVMF_CODE.ms.fd"
		expected = fullUefiAarch64VMTPM
	case "x86_64", "amd64":
		c = fullVMConfig()
		u.Code = "/usr/share/OVMF/OVMF_CODE.fd"
		expected = fullUefiVMTPM
	}

	c.UEFIFirmwareDevices = append(c.UEFIFirmwareDevices, u)

	c.TPM = TPMDevice{
		ID:     "tpm0",
		Driver: TPMTISDevice,
		Path:   "tpm.socket",
		Type:   TPMEmulatorDevice,
	}

	qemuParams, err := ConfigureParams(c, nil)
	if err != nil {
		t.Fatalf("Failed to Configure parameters, error: %s", err.Error())
	}
	result := strings.Join(qemuParams, " ")
	if expected != result {
		t.Fatalf("Failed to append parameters\nexpected[%s]\n!=\nfound    [%s]", expected, result)
	}
}

func TestFullBiosMachineCommand(t *testing.T) {
	c := fullVMConfig()

	expected := fullBiosVM
	qemuParams, err := ConfigureParams(c, nil)
	if err != nil {
		t.Fatalf("Failed to Configure parameters, error: %s", err.Error())
	}
	result := strings.Join(qemuParams, " ")
	if expected != result {
		t.Fatalf("Failed to append parameters\nexpected[%s]\n!=\nfound    [%s]", expected, result)
	}
}

func TestGetSocketsPath(t *testing.T) {
	serial := "/tmp/serial.sock"
	monitor := "/tmp/monitor.sock"
	qmp := "/tmp/qmp.sock"
	expected := []string{serial, monitor, qmp}

	c := &Config{
		LegacySerialDevices: []LegacySerialDevice{
			LegacySerialDevice{
				Backend: Socket,
				Path:    serial,
			},
		},
		MonitorDevices: []MonitorDevice{
			MonitorDevice{
				Backend: Socket,
				Path:    monitor,
			},
		},
		QMPSockets: []QMPSocket{
			QMPSocket{
				Type: Unix,
				Name: qmp,
			},
		},
	}

	sockets, err := GetSocketPaths(c)
	if err != nil {
		t.Fatalf("Failed to get sockets from config: %s", err)
	}

	// sort them
	sort.Strings(expected)
	sort.Strings(sockets)

	ok := reflect.DeepEqual(expected, sockets)
	if !ok {
		t.Errorf("Expected %v, found %v", expected, sockets)
	}
}
