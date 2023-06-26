package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"qcli"
	"sync"
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

func getFullConfig() qcli.Config {
	// return qcli.Config{}

	config := qcli.Config{
		Machine: qcli.Machine{
			Type:         qcli.MachineTypePC35,
			Acceleration: qcli.MachineAccelerationKVM,
			SMM:          "on",
		},
		CPUModel:      "qemu64",
		CPUModelFlags: []string{"+x2apic"},
		Memory: qcli.Memory{
			Size: "4096M",
		},
		RngDevices: []qcli.RngDevice{
			qcli.RngDevice{
				Driver:    qcli.VirtioRng,
				ID:        "rng0",
				Bus:       "pcie.0",
				Addr:      "3",
				Transport: qcli.TransportPCI,
				Filename:  qcli.RngDevUrandom,
			},
		},
		BlkDevices: []qcli.BlockDevice{
			qcli.BlockDevice{
				Driver:    qcli.PFlash,
				ID:        "pflash0",
				File:      "/usr/share/OVMF/OVMF_CODE.fd",
				Format:    qcli.RAW,
				Interface: qcli.PFlashInterface,
				ReadOnly:  true,
				DriveOnly: true,
			},
			qcli.BlockDevice{
				Driver:    qcli.PFlash,
				ID:        "pflash1",
				File:      "uefi-nvram.fd",
				Format:    qcli.RAW,
				Interface: qcli.PFlashInterface,
				DriveOnly: true,
			},
			qcli.BlockDevice{
				Driver:        qcli.VirtioBlock,
				ID:            "drive0",
				File:          "boot.qcow2",
				AIO:           qcli.Threads,
				Format:        qcli.QCOW2,
				Interface:     qcli.NoInterface,
				DisableModern: true,
				Serial:        "ssd-boot",
				BlockSize:     512,
				RotationRate:  1,
				Cache:         qcli.CacheModeUnsafe,
				Discard:       qcli.DiscardUnmap,
				DetectZeroes:  qcli.DetectZeroesUnmap,
				BootIndex:     "0",
			},
		},
		NetDevices: []qcli.NetDevice{
			qcli.NetDevice{
				Driver:     qcli.VirtioNet,
				Type:       qcli.USER,
				ID:         "user0",
				MACAddress: "01:02:de:ad:be:ef",
				Bus:        "pcie.0",
				User: qcli.NetDeviceUser{
					IPV4: true,
					HostForward: qcli.PortRule{
						Protocol: "tcp",
						Host:     qcli.Port{Port: 22222},
						Guest:    qcli.Port{Port: 22},
					},
				},
			},
		},
		LegacySerialDevices: []qcli.LegacySerialDevice{
			qcli.LegacySerialDevice{
				MonMux: true,
			},
		},
		PCIeRootPortDevices: []qcli.PCIeRootPortDevice{
			qcli.PCIeRootPortDevice{
				ID:            "root-port.0x6.0",
				Bus:           "pcie.0",
				Chassis:       "0x0",
				Slot:          "0x00",
				Port:          "0x0",
				Addr:          "0x4",
				Multifunction: true,
			},
			qcli.PCIeRootPortDevice{
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
		Knobs: qcli.Knobs{
			NoGraphic:     true,
			NoHPET:        true,
			Snapshot:      true,
			HugePages:     true,
			MemPrealloc:   true,
			FileBackedMem: true,
			MemShared:     true,
		},
		SMP: qcli.SMP{
			CPUs: 4,
		},
	}

	return config
}

func getRngConfig() qcli.Config {
	config := qcli.Config{
		Machine: qcli.Machine{
			Type:         qcli.MachineTypePC35,
			Acceleration: qcli.MachineAccelerationKVM,
			SMM:          "on",
		},
		CPUModel:      "qemu64",
		CPUModelFlags: []string{"+x2apic"},
		Memory: qcli.Memory{
			Size: "4096",
		},
		RngDevices: []qcli.RngDevice{
			qcli.RngDevice{
				Driver:    qcli.VirtioRng,
				ID:        "rng0",
				Bus:       "pcie.0",
				Addr:      "3",
				Transport: qcli.TransportPCI,
				Filename:  qcli.RngDevUrandom,
			},
		},
		GlobalParams: []string{
			"ICH9-LPC.disable_s3=1",
			"driver=cfi.pflash01,property=secure,value=on",
		},
		Knobs: qcli.Knobs{
			NoGraphic:     true,
			NoHPET:        true,
			Snapshot:      true,
			HugePages:     true,
			MemPrealloc:   true,
			FileBackedMem: true,
			MemShared:     true,
		},
		SMP: qcli.SMP{
			CPUs: 4,
		},
	}

	return config
}

func writeConfig() {
	config := getRngConfig()

	content, err := qcli.MarshalConfig(config)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Writing machine config to machine.yaml...")
	ioutil.WriteFile("machine.yaml", content, 0644)
	fmt.Printf("done\n")
}

func readConfig(confFile string) (*qcli.Config, error) {
	log.Infof("Reading machine config from %s...", confFile)
	content, err := ioutil.ReadFile(confFile)
	if err != nil {
		panic(err)
	}

	log.Infof("Loading config into qcli.Config...")
	config, err := qcli.UnmarshalConfig(content)
	if err != nil {
		panic(err)
	}

	/*
		fmt.Printf("done\n")
		fmt.Printf("config:\n%+v\n", config)

		fmt.Printf("Generating VM command line...")
		logger := qmpTestLogger{}
		params, err := qcli.ConfigureParams(config, logger)
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
	Config *qcli.Config
	Cmd    *exec.Cmd
	QMP    *qcli.QMP
}

func newVM(ctx context.Context, vmConfig *qcli.Config) (*VM, error) {
	vmCtx, cancelFn := context.WithCancel(ctx)

	params, err := qcli.ConfigureParams(vmConfig, nil)
	if err != nil {
		return &VM{}, err
	}
	log.Infof("Cmd: %s", params)

	return &VM{
		Ctx:    vmCtx,
		Cancel: cancelFn,
		Config: vmConfig,
		Cmd:    exec.Command(vmConfig.Path, params...),
	}, nil
}

func runVM(vm *VM) error {
	log.Infof("VM:%s starting QEMU process", vm.Config.Name)
	var stderr bytes.Buffer

	vm.Cmd.Stderr = &stderr
	err := vm.Cmd.Start()
	if err != nil {
		log.Errorf("VM:%s failed with: %s", stderr.String())
		return err
	}

	log.Infof("VM:%s waiting for QEMU process to exit...", vm.Config.Name)
	err = vm.Cmd.Wait()
	if err != nil {
		log.Errorf("VM:%s wait failed with: %s", stderr.String())
		return err
	}
	log.Infof("VM:%s QEMU process exited", vm.Config.Name)

	return nil
}

func BackgroundVM(config *qcli.Config, timeout time.Duration) error {
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
	cfg := qcli.QMPConfig{
		Logger: qmpTestLogger{},
	}

	// FIXME: sort out wait for socket
	// Start monitoring the qemu instance.  This functon will block until we have
	// connect to the QMP socket and received the welcome message.
	time.Sleep(2 * time.Second) // some delay on start up...

	qmpSocketFile := config.QMPSockets[0].Name
	log.Infof("VM:%s connecting to QMP socket %s", vmName, qmpSocketFile)
	q, qver, err := qcli.QMPStart(context.Background(), qmpSocketFile, cfg, disconnectedCh)
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

func StartVM(config *qcli.Config) error {
	ctx := config.Ctx
	if ctx == nil {
		ctx = context.TODO()
	}

	if config.Path == "" {
		config.Path = "qemu-system-x86_64"
	}

	params, err := qcli.ConfigureParams(config, qmpTestLogger{})
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

	vmConfigs := []qcli.Config{}

	//if len(os.Args) > 1 {
	//	confFile = os.Args[1]
	//}

	log.Infof("Reading config for VM1")
	vm1Config, err := readConfig(vm1)
	if err != nil {
		panic(err)
	}
	vmConfigs = append(vmConfigs, vm1Config)

	log.Infof("Reading config for VM2")
	vm2Config, err := readConfig(vm2)
	if err != nil {
		panic(err)
	}
	vmConfigs = append(vmConfigs, vm1Config)

	wg.Add(1)
	go func(cfg *qcli.Config, wg *sync.WaitGroup) {
		log.Infof("VM:%s starting in background...", cfg.Name)
		err = BackgroundVM(cfg, 10*time.Second)
		if err != nil {
			panic(err)
		}
		log.Infof("VM:%s done", cfg.Name)
		wg.Done()
	}(vm1Config, &wg)

	wg.Add(1)
	go func(cfg *qcli.Config, wg *sync.WaitGroup) {
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
