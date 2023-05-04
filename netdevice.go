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

// Package qemu provides methods and types for launching and managing QEMU
// instances.  Instances can be launched with the LaunchQemu function and
// managed thereafter via QMPStart and the QMP object that this function
// returns.  To manage a qemu instance after it has been launched you need
// to pass the -qmp option during launch requesting the qemu instance to create
// a QMP unix domain manageent socket, e.g.,
// -qmp unix:/tmp/qmp-socket,server,nowait.  For more information see the
// example below.

package qcli

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// NetDeviceType is a qemu networking device type.
type NetDeviceType string

const (
	// USER is SLIRP user-space networking device type
	USER NetDeviceType = "user"

	// MCASTSOCKET is a socket networking device type
	MCASTSOCKET NetDeviceType = "mcastsocket"

	// TAP is a TAP networking device type.
	TAP NetDeviceType = "tap"

	// MACVTAP is a macvtap networking device type.
	MACVTAP NetDeviceType = "macvtap"

	// IPVTAP is a ipvtap virtual networking device type.
	IPVTAP NetDeviceType = "ipvtap"

	// VETHTAP is a veth-tap virtual networking device type.
	VETHTAP NetDeviceType = "vethtap"

	// VFIO is a direct assigned PCI device or PCI VF
	VFIO NetDeviceType = "VFIO"

	// VHOSTUSER is a vhost-user port (socket)
	VHOSTUSER NetDeviceType = "vhostuser"
)

// QemuNetdevParam converts to the QEMU -netdev parameter notation
func (n NetDeviceType) QemuNetdevParam(netdev *NetDevice, config *Config) string {
	if netdev.Transport == "" {
		netdev.Transport = netdev.Transport.defaultTransport(config)
	}

	switch n {
	case USER:
		return "user"
	case MCASTSOCKET:
		return "socket"
	case TAP, MACVTAP, IPVTAP, VETHTAP:
		return "tap" // -netdev tap,<props> -device virtio-net-pci
	case VFIO:
		if netdev.Transport == TransportMMIO {
			log.Fatal("vfio devices are not support with the MMIO transport")
		}
		return "" // -device vfio-pci (no netdev)
	case VHOSTUSER:
		if netdev.Transport == TransportCCW {
			log.Fatal("vhost-user devices are not supported on IBM Z")
		}
		return "vhost-user" // -netdev vhost-user,<props> (no device)
	default:
		return ""

	}
}

// QemuDeviceParam converts to the QEMU -device parameter notation
func (n NetDeviceType) QemuDeviceParam(netdev *NetDevice, config *Config) DeviceDriver {
	if netdev.Transport == "" {
		netdev.Transport = netdev.Transport.defaultTransport(config)
	}

	if netdev.Driver != VirtioNet {
		return netdev.Driver
	}

	// Handle virtio-net transport
	var device string

	switch n {
	case MCASTSOCKET:
		device = "virtio-net"
	case USER:
		device = "virtio-net"
	case TAP:
		device = "virtio-net"
	case MACVTAP:
		device = "virtio-net"
	case IPVTAP:
		device = "virtio-net"
	case VETHTAP:
		device = "virtio-net" // -netdev type=tap -device virtio-net-pci
	case VFIO:
		if netdev.Transport == TransportMMIO {
			log.Fatal("vfio devices are not support with the MMIO transport")
		}
		device = "vfio" // -device vfio-pci (no netdev)
	case VHOSTUSER:
		if netdev.Transport == TransportCCW {
			log.Fatal("vhost-user devices are not supported on IBM Z")
		}
		return "" // -netdev type=vhost-user (no device)
	default:
		return ""
	}

	switch netdev.Transport {
	case TransportPCI:
		return DeviceDriver(device + "-pci")
	case TransportCCW:
		return DeviceDriver(device + "-ccw")
	case TransportMMIO:
		return DeviceDriver(device + "-device")
	default:
		return ""
	}
}

// -netdev tap,ifname=,downscript=,script=
type NetDeviceTap struct {
	// IfName is the interface name,
	IFName string `yaml:"ifname"`

	// DownScript is the tap interface deconfiguration script.
	DownScript string `yaml:"downscript-file"`

	// Script is the tap interface configuration script.
	Script string `yaml:"upscript-file"`
}

type Port struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type PortRule struct {
	Protocol string `yaml:"protocol"`
	Host     Port   `yaml:"host-port"`
	Guest    Port   `yaml:"guest-port"`
}

/*
func (p *PortRule) UnmarshalYAML(unmarshal func(interface{}) error) error {
	DefaultPortProtocol := "tcp"
	DefaultPortHostAddress := ""
	DefaultPortGuestAddress := ""
	var ruleVal map[string]string
	var err error

	if err = unmarshal(&ruleVal); err != nil {
		return err
	}

	for hostVal, guestVal := range ruleVal {
		hostToks := strings.Split(hostVal, ":")
		if len(hostToks) == 3 {
			p.Protocol = hostToks[0]
			p.Host.Address = hostToks[1]
			p.Host.Port, err = strconv.Atoi(hostToks[2])
			if err != nil {
				return err
			}
		} else if len(hostToks) == 2 {
			p.Protocol = DefaultPortProtocol
			p.Host.Address = hostToks[0]
			p.Host.Port, err = strconv.Atoi(hostToks[1])
			if err != nil {
				return err
			}
		} else {
			p.Protocol = DefaultPortProtocol
			p.Host.Address = DefaultPortHostAddress
			p.Host.Port, err = strconv.Atoi(hostToks[0])
			if err != nil {
				return err
			}
		}
		guestToks := strings.Split(guestVal, ":")
		if len(guestToks) == 2 {
			p.Guest.Address = guestToks[0]
			p.Guest.Port, err = strconv.Atoi(guestToks[1])
			if err != nil {
				return err
			}
		} else {
			p.Guest.Address = DefaultPortGuestAddress
			p.Guest.Port, err = strconv.Atoi(guestToks[0])
			if err != nil {
				return err
			}
		}
		break
	}
	if p.Protocol != "tcp" && p.Protocol != "udp" {
		return fmt.Errorf("Invalid PortRule.Protocol value: %s . Must be 'tcp' or 'udp'", p.Protocol)
	}
	return nil
}
*/

func (p PortRule) String() string {
	return fmt.Sprintf("%s:%s:%d-%s:%d", p.Protocol,
		p.Host.Address, p.Host.Port, p.Guest.Address, p.Guest.Port)
}

const EmptyPortRule = "::0-:0"

// -netdev user,
type NetDeviceUser struct {
	IPV4        bool       `yaml:"ipv4-enable"`
	IPV4NetAddr string     `yaml:"ipv4-network-address"`
	HostForward []PortRule `yaml:"host-port-rules"`
}

// -netdev socket,listen=
type NetDeviceMcastSocket struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
}

// -netdev socket,mcast=
// -netdev socket,udp=

// NetDevice represents a guest networking device
type NetDevice struct {
	// Type is the netdev type (e.g. tap).
	Type NetDeviceType `yaml:"type"`

	// Driver is the qemu device driver
	Driver DeviceDriver `yaml:"driver"`

	// ID is the netdevice identifier.
	ID string `yaml:"id"`

	// Bus is the bus path name of a PCI device.
	Bus string `yaml:"bus"`

	// Addr is the address offset of a PCI device.
	Addr string `yaml:"address"`

	// FDs represents the list of already existing file descriptors to be used.
	// This is mostly useful for mq support.
	FDs      []*os.File
	VhostFDs []*os.File

	// VHost enables virtio device emulation from the host kernel instead of from qemu.
	VHost bool `yaml:"vhost-enable"`

	// MACAddress is the networking device interface MAC address.
	MACAddress string `yaml:"macaddress"`

	// DisableModern prevents qemu from relying on fast MMIO.
	DisableModern bool `yaml:"disable-modern"`

	// ROMFile specifies the ROM file being used for this device.
	ROMFile string `yaml:"rom-file"`

	// DevNo identifies the ccw devices for s390x architecture
	DevNo string `yaml:"ccw-dev-no"`

	// Transport is the virtio transport for this device.
	Transport VirtioTransport `yaml:"transport"`

	// -netdev tap,.*
	Tap NetDeviceTap `yaml:"tap-device"`

	// -netdev user,.*
	User NetDeviceUser `yaml:"user-device"`

	// -netdev socket,mcast=
	McastSocket NetDeviceMcastSocket `yaml:"mcast-socket"`

	// bootindex
	BootIndex string `yaml:"bootindex"`
}

// VirtioNetTransport is a map of the virtio-net device name that corresponds
// to each transport.
var VirtioNetTransport = map[VirtioTransport]string{
	TransportPCI:  "virtio-net-pci",
	TransportCCW:  "virtio-net-ccw",
	TransportMMIO: "virtio-net-device",
}

// Valid returns true if the NetDevice structure is valid and complete.
func (netdev NetDevice) Valid() error {
	if netdev.ID == "" {
		return fmt.Errorf("NetDevice has empty ID field")
	}

	if netdev.Type == "" {
		return fmt.Errorf("NetDevice has empty Type field")
	}

	switch netdev.Type {
	case USER, MCASTSOCKET, TAP, MACVTAP:
		break
	default:
		return fmt.Errorf("NetDevice has Unknown Type value: %s", netdev.Type)
	}

	if netdev.Type == TAP && netdev.Tap.IFName == "" {
		return fmt.Errorf("Netdevice Type=TAP has empty IFName field")
	}

	if netdev.Type == MCASTSOCKET {
		if netdev.McastSocket.Address == "" {
			return fmt.Errorf("Netdevice Type=MCASTSOCKET has empty Address field")
		}
		if netdev.McastSocket.Port == "" {
			return fmt.Errorf("Netdevice Type=MCASTSOCKET has empty Port field")
		}
	}

	return nil
}

// mqParameter returns the parameters for multi-queue driver. If the driver is a PCI device then the
// vector flag is required. If the driver is a CCW type than the vector flag is not implemented and only
// multi-queue option mq needs to be activated. See comment in libvirt code at
// https://github.com/libvirt/libvirt/blob/6e7e965dcd3d885739129b1454ce19e819b54c25/src/qemu/qemu_command.c#L3633
func (netdev NetDevice) mqParameter(config *Config) string {
	p := []string{"mq=on"}

	if netdev.Transport.isVirtioPCI(config) {
		// https://www.linux-kvm.org/page/Multiqueue
		// -netdev tap,vhost=on,queues=N
		// enable mq and specify msix vectors in qemu cmdline
		// (2N+2 vectors, N for tx queues, N for rx queues, 1 for config, and one for possible control vq)
		// -device virtio-net-pci,mq=on,vectors=2N+2...
		// enable mq in guest by 'ethtool -L eth0 combined $queue_num'
		// Clearlinux automatically sets up the queues properly
		// The agent implementation should do this to ensure that it is
		// always set
		vectors := len(netdev.FDs)*2 + 2
		p = append(p, fmt.Sprintf("vectors=%d", vectors))
	}

	return strings.Join(p, ",")
}

// QemuDeviceParams returns the -device parameters for this network device
func (netdev NetDevice) QemuDeviceParams(config *Config) []string {
	var deviceParams []string

	driver := netdev.Type.QemuDeviceParam(&netdev, config)
	if driver == "" {
		return nil
	}

	deviceParams = append(deviceParams, fmt.Sprintf("%s", driver))
	deviceParams = append(deviceParams, fmt.Sprintf("netdev=%s", netdev.ID))
	deviceParams = append(deviceParams, fmt.Sprintf("mac=%s", netdev.MACAddress))

	if netdev.Bus != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("bus=%s", netdev.Bus))
	}

	if netdev.Addr != "" {
		addr, err := strconv.Atoi(netdev.Addr)
		if err == nil && addr >= 0 {
			deviceParams = append(deviceParams, fmt.Sprintf("addr=0x%02x", addr))
		}
	}

	if netdev.BootIndex != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("bootindex=%s", netdev.BootIndex))
	}

	if strings.HasPrefix(string(driver), "virtio") {
		if s := netdev.Transport.disableModern(config, netdev.DisableModern); s != "" {
			deviceParams = append(deviceParams, s)
		}
	}

	if len(netdev.FDs) > 0 {
		// Note: We are appending to the device params here
		deviceParams = append(deviceParams, netdev.mqParameter(config))
	}

	if netdev.Transport.isVirtioPCI(config) && netdev.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", netdev.ROMFile))
	}

	if netdev.Transport.isVirtioCCW(config) {
		if config.Knobs.IOMMUPlatform {
			deviceParams = append(deviceParams, "iommu_platform=on")
		}
		deviceParams = append(deviceParams, fmt.Sprintf("devno=%s", netdev.DevNo))
	}

	return deviceParams
}

// QemuNetdevParams returns the -netdev parameters for this network device
func (netdev NetDevice) QemuNetdevParams(config *Config) []string {
	var netdevParams []string

	netdevType := netdev.Type.QemuNetdevParam(&netdev, config)
	if netdevType == "" {
		return nil
	}

	netdevParams = append(netdevParams, netdevType)
	netdevParams = append(netdevParams, fmt.Sprintf("id=%s", netdev.ID))

	if netdev.VHost {
		netdevParams = append(netdevParams, "vhost=on")
		if len(netdev.VhostFDs) > 0 {
			var fdParams []string
			qemuFDs := config.appendFDs(netdev.VhostFDs)
			for _, fd := range qemuFDs {
				fdParams = append(fdParams, fmt.Sprintf("%d", fd))
			}
			netdevParams = append(netdevParams, fmt.Sprintf("vhostfds=%s", strings.Join(fdParams, ":")))
		}
	}

	switch netdev.Type {
	case TAP:
		if len(netdev.FDs) > 0 {
			var fdParams []string

			qemuFDs := config.appendFDs(netdev.FDs)
			for _, fd := range qemuFDs {
				fdParams = append(fdParams, fmt.Sprintf("%d", fd))
			}

			netdevParams = append(netdevParams, fmt.Sprintf("fds=%s", strings.Join(fdParams, ":")))

		} else {
			netdevParams = append(netdevParams, fmt.Sprintf("ifname=%s", netdev.Tap.IFName))
			if netdev.Tap.DownScript != "" {
				netdevParams = append(netdevParams, fmt.Sprintf("downscript=%s", netdev.Tap.DownScript))
			}
			if netdev.Tap.Script != "" {
				netdevParams = append(netdevParams, fmt.Sprintf("script=%s", netdev.Tap.Script))
			}
		}
	case USER:
		if netdev.User.IPV4 {
			netdevParams = append(netdevParams, "ipv4=on")
		} else {
			netdevParams = append(netdevParams, "ipv4=off")
		}

		for _, rule := range netdev.User.HostForward {
			hostfwd := rule.String()
			if hostfwd != EmptyPortRule {
				netdevParams = append(netdevParams, fmt.Sprintf("hostfwd=%s", hostfwd))
			}
		}

		if netdev.User.IPV4NetAddr != "" {
			netdevParams = append(netdevParams, fmt.Sprintf("net=%s", netdev.User.IPV4NetAddr))
		}
	case MCASTSOCKET:
		var mcastParam string

		mcastParam = fmt.Sprintf("mcast=%s:%s", netdev.McastSocket.Address, netdev.McastSocket.Port)
		netdevParams = append(netdevParams, mcastParam)
	}

	return netdevParams
}

// QemuParams returns the qemu parameters built out of this network device.
func (netdev NetDevice) QemuParams(config *Config) []string {
	var netdevParams []string
	var deviceParams []string
	var qemuParams []string

	// Macvtap can only be connected via fds
	if (netdev.Type == MACVTAP) && (len(netdev.FDs) == 0) {
		return nil // implicit error
	}

	if netdev.Type.QemuNetdevParam(&netdev, config) != "" {
		netdevParams = netdev.QemuNetdevParams(config)
		if netdevParams != nil {
			qemuParams = append(qemuParams, "-netdev")
			qemuParams = append(qemuParams, strings.Join(netdevParams, ","))
		}
	}

	if netdev.Type.QemuDeviceParam(&netdev, config) != "" {
		deviceParams = netdev.QemuDeviceParams(config)
		if deviceParams != nil {
			qemuParams = append(qemuParams, "-device")
			qemuParams = append(qemuParams, strings.Join(deviceParams, ","))
		}
	}

	return qemuParams
}
