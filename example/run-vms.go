package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"test-govmm/pkg/qemu"
	"time"

	log "github.com/sirupsen/logrus"
)

type qmpTestLogger struct{}

func (l qmpTestLogger) V(level int32) bool {
	return true
}

func (l qmpTestLogger) Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

func (l qmpTestLogger) Warningf(format string, v ...interface{}) {
	l.Infof(format, v...)
}

func (l qmpTestLogger) Errorf(format string, v ...interface{}) {
	l.Errorf(format, v...)
}

func getFullConfig() qemu.Config {
	// return qemu.Config{}

	config := qemu.Config{
		Machine: qemu.Machine{
			Type:         qemu.MachineTypePC35,
			Acceleration: qemu.MachineAccelerationKVM,
			SMM:          "on",
		},
		CPUModel:      "qemu64",
		CPUModelFlags: []string{"+x2apic"},
		Memory: qemu.Memory{
			Size: "4096M",
		},
		RngDevices: []qemu.RngDevice{
			qemu.RngDevice{
				Driver:    qemu.VirtioRng,
				ID:        "rng0",
				Bus:       "pcie.0",
				Addr:      "3",
				Transport: qemu.TransportPCI,
				Filename:  qemu.RngDevUrandom,
			},
		},
		BlkDevices: []qemu.BlockDevice{
			qemu.BlockDevice{
				Driver:    qemu.PFlash,
				ID:        "pflash0",
				File:      "/usr/share/OVMF/OVMF_CODE.fd",
				Format:    qemu.RAW,
				Interface: qemu.PFlashInterface,
				ReadOnly:  true,
				DriveOnly: true,
			},
			qemu.BlockDevice{
				Driver:    qemu.PFlash,
				ID:        "pflash1",
				File:      "uefi_nvram.fd",
				Format:    qemu.RAW,
				Interface: qemu.PFlashInterface,
				DriveOnly: true,
			},
			qemu.BlockDevice{
				Driver:        qemu.VirtioBlock,
				ID:            "drive0",
				File:          "boot.qcow2",
				AIO:           qemu.Threads,
				Format:        qemu.QCOW2,
				Interface:     qemu.NoInterface,
				DisableModern: true,
				Serial:        "ssd-boot",
				BlockSize:     512,
				RotationRate:  1,
				Cache:         qemu.CacheModeUnsafe,
				Discard:       qemu.DiscardUnmap,
				DetectZeroes:  qemu.DetectZeroesUnmap,
				BootIndex:     0,
			},
		},
		NetDevices: []qemu.NetDevice{
			qemu.NetDevice{
				Driver:     qemu.VirtioNet,
				Type:       qemu.USER,
				ID:         "user0",
				MACAddress: "01:02:de:ad:be:ef",
				Bus:        "pcie.0",
				User: qemu.NetDeviceUser{
					IPV4: true,
					HostForward: qemu.PortRule{
						Protocol: "tcp",
						Host:     qemu.Port{Port: 22222},
						Guest:    qemu.Port{Port: 22},
					},
				},
			},
		},
		LegacySerialDevices: []qemu.LegacySerialDevice{
			qemu.LegacySerialDevice{
				MonMux: true,
			},
		},
		PCIeRootPortDevices: []qemu.PCIeRootPortDevice{
			qemu.PCIeRootPortDevice{
				ID:            "root-port.0x6.0",
				Bus:           "pcie.0",
				Chassis:       "0x0",
				Slot:          "0x00",
				Port:          "0x0",
				Addr:          "0x4",
				Multifunction: true,
			},
			qemu.PCIeRootPortDevice{
				ID:            "root-port.0x6.1",
				Bus:           "pcie.0",
				Chassis:       "0x1",
				Slot:          "0x00",
				Port:          "0x1",
				Addr:          "0x6.0x1",
				Multifunction: false,
			},
		},
		GlobalParams: []string{
			"ICH9-LPC.disable_s3=1",
			"driver=cfi.pflash01,property=secure,value=on",
		},
		Knobs: qemu.Knobs{
			NoGraphic:     true,
			NoHPET:        true,
			Snapshot:      true,
			HugePages:     true,
			MemPrealloc:   true,
			FileBackedMem: true,
			MemShared:     true,
		},
		SMP: qemu.SMP{
			CPUs: 4,
		},
	}

	return config
}

func getRngConfig() qemu.Config {
	config := qemu.Config{
		Machine: qemu.Machine{
			Type:         qemu.MachineTypePC35,
			Acceleration: qemu.MachineAccelerationKVM,
			SMM:          "on",
		},
		CPUModel:      "qemu64",
		CPUModelFlags: []string{"+x2apic"},
		Memory: qemu.Memory{
			Size: "4096",
		},
		RngDevices: []qemu.RngDevice{
			qemu.RngDevice{
				Driver:    qemu.VirtioRng,
				ID:        "rng0",
				Bus:       "pcie.0",
				Addr:      "3",
				Transport: qemu.TransportPCI,
				Filename:  qemu.RngDevUrandom,
			},
		},
		GlobalParams: []string{
			"ICH9-LPC.disable_s3=1",
			"driver=cfi.pflash01,property=secure,value=on",
		},
		Knobs: qemu.Knobs{
			NoGraphic:     true,
			NoHPET:        true,
			Snapshot:      true,
			HugePages:     true,
			MemPrealloc:   true,
			FileBackedMem: true,
			MemShared:     true,
		},
		SMP: qemu.SMP{
			CPUs: 4,
		},
	}

	return config
}

func writeConfig() {
	config := getRngConfig()

	content, err := qemu.MarshalConfig(config)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Writing machine config to machine.yaml...")
	ioutil.WriteFile("machine.yaml", content, 0644)
	fmt.Printf("done\n")
}

func readConfig(confFile string) (*qemu.Config, error) {
	log.Infof("Reading machine config from %s...", confFile)
	content, err := ioutil.ReadFile(confFile)
	if err != nil {
		panic(err)
	}

	log.Infof("Loading config into qemu.Config...")
	config, err := qemu.UnmarshalConfig(content)
	if err != nil {
		panic(err)
	}

	/*
		fmt.Printf("done\n")
		fmt.Printf("config:\n%+v\n", config)

		fmt.Printf("Generating VM command line...")
		logger := qmpTestLogger{}
		params, err := qemu.ConfigureParams(config, logger)
		if err != nil {
			panic(err)
		}
		fmt.Printf("done\n")
		fmt.Printf("\ncommand:\n%s\n", strings.Join(params, " "))
		fmt.Printf("done\n")
	*/

	return config, err
}

type VM struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Config *qemu.Config
	Cmd    *exec.Cmd
}

func newVM(ctx context.Context, vmConfig *qemu.Config) (*VM, error) {
	vmCtx, cancelFn := context.WithCancel(ctx)

	params, err := qemu.ConfigureParams(vmConfig, nil)
	if err != nil {
		return &VM{}, err
	}

	return &VM{
		Ctx:    vmCtx,
		Cancel: cancelFn,
		Config: vmConfig,
		Cmd:    exec.Command(vmConfig.Path, params...),
	}, nil
}

func runVM(vm *VM) error {
	log.Infof("VM:%s starting QEMU process", vm.Config.Name)
	err := vm.Cmd.Start()
	if err != nil {
		return err
	}

	log.Infof("VM:%s waiting for QEMU process to exit...", vm.Config.Name)
	err = vm.Cmd.Wait()
	if err != nil {
		return err
	}
	log.Infof("VM:%s QEMU process exited", vm.Config.Name)

	return nil
}

func BackgroundVM(config *qemu.Config, timeout time.Duration) error {
	var wg sync.WaitGroup
	disconnectedCh := make(chan struct{})
	doneCh := make(chan struct{})
	vmName := config.Name

	log.Infof("VM:%s starting in background", vmName)

	if config.Path == "" {
		config.Path = "qemu-system-x86_64"
	}

	// daemonize this VM
	// config.Knobs.Daemonize = true

	ctx := context.Background()
	vm, err := newVM(ctx, config)

	wg.Add(1)
	go func(v *VM) {
		runVM(v)
		wg.Done()
		close(doneCh)
	}(vm)

	// Set up our options.  We don't want any logging or to receive any events.
	cfg := qemu.QMPConfig{
		Logger: qmpTestLogger{},
	}

	// FIXME: sort out wait for socket
	// Start monitoring the qemu instance.  This functon will block until we have
	// connect to the QMP socket and received the welcome message.
	time.Sleep(2 * time.Second) // some delay on start up...

	qmpSocketFile := config.QMPSockets[0].Name
	log.Infof("VM:%s connecting to QMP socket %s", vmName, qmpSocketFile)
	q, qver, err := qemu.QMPStart(context.Background(), qmpSocketFile, cfg, disconnectedCh)
	if err != nil {
		return fmt.Errorf("Failed to connect to qmp socket: %s", err.Error())
	}
	log.Infof("VM:%s QMP:%v QMPVersion:%v", vmName, q, qver)

	// This has to be the first command executed in a QMP session.
	err = q.ExecuteQMPCapabilities(context.Background())
	if err != nil {
		return err
	}

	log.Infof("VM:%s querying CPUInfo via QMP...", vmName)
	cpuInfo, err := q.ExecQueryCpus(context.TODO())
	if err != nil {
		return err
	}
	log.Infof("VM:%s has %d CPUs", vmName, len(cpuInfo))

	log.Infof("VM:%s querying VM Status via QMP...", vmName)
	status, err := q.ExecuteQueryStatus(context.TODO())
	if err != nil {
		return err
	}
	log.Infof("VM:%s Status:%s Running:%v", vmName, status.Status, status.Running)

	log.Infof("VM:%s Waiting 20 seconds for boot to prompt...", vmName)
	time.Sleep(time.Second * 20)

	// Let's try to shutdown the VM.  If it hasn't shutdown in 10 seconds we'll
	// send a quit message.
	log.Infof("VM:%s trying graceful shutdown via system_powerdown (%s timeout before cancelling)..", vmName, timeout.String())
	err = q.ExecuteSystemPowerdown(ctx)
	if err != nil {
		log.Errorf("VM:%s error:%s", vmName, err.Error())
	}

	select {
	case <-doneCh:
		log.Infof("VM:%s has exited without cancel", vmName)
	case <-time.After(timeout):
		log.Warnf("VM:%s timed out, killing via cancel context...", vmName)
		vm.Cancel()
	}
	// should be no-op but we need to make sure it's gone
	wg.Wait()
	<-disconnectedCh

	log.Infof("VM:%s background vm func returns", vmName)
	return nil
}

func StartVM(config *qemu.Config) error {
	ctx := config.Ctx
	if ctx == nil {
		ctx = context.TODO()
	}

	if config.Path == "" {
		config.Path = "qemu-system-x86_64"
	}

	params, err := qemu.ConfigureParams(config, qmpTestLogger{})
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, config.Path, params...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	var wg sync.WaitGroup
	vm1 := "machine-vm1.yaml"
	vm2 := "machine-vm2.yaml"

	//if len(os.Args) > 1 {
	//	confFile = os.Args[1]
	//}

	log.Infof("Reading config for VM1")
	vm1Config, err := readConfig(vm1)
	if err != nil {
		panic(err)
	}

	log.Infof("Reading config for VM2")
	vm2Config, err := readConfig(vm2)
	if err != nil {
		panic(err)
	}

	wg.Add(1)
	go func(cfg *qemu.Config, wg *sync.WaitGroup) {
		log.Infof("VM:%s starting in background...", cfg.Name)
		err = BackgroundVM(cfg, 10*time.Second)
		if err != nil {
			panic(err)
		}
		log.Infof("VM:%s done", cfg.Name)
		wg.Done()
	}(vm1Config, &wg)

	wg.Add(1)
	go func(cfg *qemu.Config, wg *sync.WaitGroup) {
		log.Infof("VM:%s starting in background...", cfg.Name)
		err = BackgroundVM(cfg, 100*time.Millisecond)
		if err != nil {
			panic(err)
		}
		log.Infof("VM:%s done", cfg.Name)
		wg.Done()
	}(vm2Config, &wg)

	log.Infof("Waiting for background VMs to complete...")
	wg.Wait()
	log.Infof("All done, exiting...")
}
