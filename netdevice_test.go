package qcli

import (
	"io/ioutil"
	"os"
	"testing"
)

var (
	deviceNetworkPCIString         = "-netdev tap,id=tap0,vhost=on,ifname=ceth0,downscript=no,script=no -device virtio-net-pci,netdev=tap0,mac=01:02:de:ad:be:ef,bus=/pci-bus/pcie.0,addr=0xff,disable-modern=true,romfile=efi-virtio.rom"
	deviceNetworkPCIStringLowAddr  = "-netdev tap,id=tap0,vhost=on,ifname=ceth0,downscript=no,script=no -device virtio-net-pci,netdev=tap0,mac=01:02:de:ad:be:ef,bus=/pci-bus/pcie.0,addr=0x03,disable-modern=true,romfile=efi-virtio.rom"
	deviceNetworkPCIStringMq       = "-netdev tap,id=tap0,vhost=on,fds=3:4 -device virtio-net-pci,netdev=tap0,mac=01:02:de:ad:be:ef,bus=/pci-bus/pcie.0,addr=0xff,disable-modern=true,mq=on,vectors=6,romfile=efi-virtio.rom"
	deviceNetworkString            = "-netdev tap,id=tap0,vhost=on,ifname=ceth0,downscript=no,script=no -device virtio-net-pci,netdev=tap0,mac=01:02:de:ad:be:ef,disable-modern=true,romfile=efi-virtio.rom"
	deviceNetworkUserString        = "-netdev user,id=user0,ipv4=on,net=10.0.2.15/24 -device e1000,netdev=user0,mac=01:02:de:ad:be:ef,romfile="
	deviceNetworkUserHostFwdString = "-netdev user,id=user0,ipv4=on,hostfwd=tcp::22222-:22,hostfwd=tcp::8080-:80 -device virtio-net-pci,netdev=user0,mac=01:02:de:ad:be:ef,disable-modern=false"
	deviceNetworkMcastSocketString = "-netdev socket,id=sock0,mcast=230.0.0.1:1234 -device virtio-net-pci,netdev=sock0,mac=01:02:de:ad:be:ef,disable-modern=true"
	deviceNetworkTapMqString       = "-netdev tap,id=tap0,vhost=on,fds=3:4 -device virtio-net-pci,netdev=tap0,mac=01:02:de:ad:be:ef,disable-modern=true,mq=on,vectors=6,romfile=efi-virtio.rom"
)

func TestAppendDeviceNetworkTap(t *testing.T) {
	netdev := NetDevice{
		Driver:        VirtioNet,
		Type:          TAP,
		ID:            "tap0",
		VHost:         true,
		MACAddress:    "01:02:de:ad:be:ef",
		DisableModern: true,
		ROMFile:       "efi-virtio.rom",
		Tap: NetDeviceTap{
			IFName:     "ceth0",
			Script:     "no",
			DownScript: "no",
		},
	}

	if netdev.Transport.isVirtioCCW(nil) {
		netdev.DevNo = DevNo
	}

	testAppend(netdev, deviceNetworkString, t)
}

func TestAppendDeviceNetworkUser(t *testing.T) {
	netdev := NetDevice{
		Driver:     E1000,
		Type:       USER,
		ID:         "user0",
		MACAddress: "01:02:de:ad:be:ef",
		ROMFile:    DisabledNetDeviceROMFile,
		User: NetDeviceUser{
			IPV4:        true,
			IPV4NetAddr: "10.0.2.15/24",
		},
	}

	testAppend(netdev, deviceNetworkUserString, t)
}

func TestAppendDeviceNetworkUserHostForward(t *testing.T) {
	netdev := NetDevice{
		Driver:        VirtioNet,
		Type:          USER,
		ID:            "user0",
		DisableModern: false,
		MACAddress:    "01:02:de:ad:be:ef",
		User: NetDeviceUser{
			IPV4: true,
			HostForward: []PortRule{
				PortRule{
					Protocol: "tcp",
					Host:     Port{Port: 22222},
					Guest:    Port{Port: 22},
				},
				PortRule{
					Protocol: "tcp",
					Host:     Port{Port: 8080},
					Guest:    Port{Port: 80},
				},
			},
		},
	}

	testAppend(netdev, deviceNetworkUserHostFwdString, t)
}

func TestAppendDeviceNetworkMcastSocket(t *testing.T) {
	netdev := NetDevice{
		Driver:        VirtioNet,
		Type:          MCASTSOCKET,
		ID:            "sock0",
		DisableModern: true,
		MACAddress:    "01:02:de:ad:be:ef",
		McastSocket: NetDeviceMcastSocket{
			Address: "230.0.0.1",
			Port:    "1234",
		},
	}

	testAppend(netdev, deviceNetworkMcastSocketString, t)
}

func TestAppendDeviceNetworkTapMq(t *testing.T) {
	foo, _ := ioutil.TempFile(os.TempDir(), "govmm-qemu-test")
	bar, _ := ioutil.TempFile(os.TempDir(), "govmm-qemu-test")

	defer func() {
		_ = foo.Close()
		_ = bar.Close()
		_ = os.Remove(foo.Name())
		_ = os.Remove(bar.Name())
	}()

	netdev := NetDevice{
		Driver:        VirtioNet,
		Type:          TAP,
		ID:            "tap0",
		FDs:           []*os.File{foo, bar},
		VHost:         true,
		MACAddress:    "01:02:de:ad:be:ef",
		DisableModern: true,
		ROMFile:       "efi-virtio.rom",
		Tap: NetDeviceTap{
			IFName:     "ceth0",
			Script:     "no",
			DownScript: "no",
		},
	}
	if netdev.Transport.isVirtioCCW(nil) {
		netdev.DevNo = DevNo
	}

	testAppend(netdev, deviceNetworkTapMqString, t)
}

func TestAppendDeviceNetworkPCI(t *testing.T) {

	netdev := NetDevice{
		Driver:        VirtioNet,
		Type:          TAP,
		ID:            "tap0",
		Bus:           "/pci-bus/pcie.0",
		Addr:          "255",
		VHost:         true,
		MACAddress:    "01:02:de:ad:be:ef",
		DisableModern: true,
		ROMFile:       romfile,
		Tap: NetDeviceTap{
			IFName:     "ceth0",
			Script:     "no",
			DownScript: "no",
		},
	}

	if !netdev.Transport.isVirtioPCI(nil) {
		t.Skip("Test valid only for PCI devices")
	}

	testAppend(netdev, deviceNetworkPCIString, t)
}

func TestAppendDeviceNetworkPCILowAddr(t *testing.T) {

	netdev := NetDevice{
		Driver:        VirtioNet,
		Type:          TAP,
		ID:            "tap0",
		Bus:           "/pci-bus/pcie.0",
		Addr:          "3",
		VHost:         true,
		MACAddress:    "01:02:de:ad:be:ef",
		DisableModern: true,
		ROMFile:       romfile,
		Tap: NetDeviceTap{
			IFName:     "ceth0",
			Script:     "no",
			DownScript: "no",
		},
	}

	if !netdev.Transport.isVirtioPCI(nil) {
		t.Skip("Test valid only for PCI devices")
	}

	testAppend(netdev, deviceNetworkPCIStringLowAddr, t)
}

func TestAppendDeviceNetworkPCIMq(t *testing.T) {
	foo, _ := ioutil.TempFile(os.TempDir(), "govmm-qemu-test")
	bar, _ := ioutil.TempFile(os.TempDir(), "govmm-qemu-test")

	defer func() {
		_ = foo.Close()
		_ = bar.Close()
		_ = os.Remove(foo.Name())
		_ = os.Remove(bar.Name())
	}()

	netdev := NetDevice{
		Driver:        VirtioNet,
		Type:          TAP,
		ID:            "tap0",
		Bus:           "/pci-bus/pcie.0",
		Addr:          "255",
		FDs:           []*os.File{foo, bar},
		VHost:         true,
		MACAddress:    "01:02:de:ad:be:ef",
		DisableModern: true,
		ROMFile:       romfile,
		Tap: NetDeviceTap{
			IFName:     "ceth0",
			Script:     "no",
			DownScript: "no",
		},
	}

	if !netdev.Transport.isVirtioPCI(nil) {
		t.Skip("Test valid only for PCI devices")
	}

	testAppend(netdev, deviceNetworkPCIStringMq, t)
}
