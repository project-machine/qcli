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
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// 32 slots available, place auto-created devices
// starting at the end and decrement as needed
// to avoid impacting
const PCISlotMax int = 31

// slot 0, 1, and 2 are always taken
const PCISlotOffset = 3

type PCIBus [PCISlotMax]bool

func (bus *PCIBus) SetSlot(slot int) error {
	if slot > PCISlotMax {
		return fmt.Errorf("Slot %d must be < %d", slot, PCISlotMax)
	}
	bus[slot] = true
	log.Debugf("PCIBus: allocated slot %s", fmt.Sprintf("0x%02x", slot))
	return nil
}

func (bus *PCIBus) GetSlot(busAddr string) int {
	// see if supplised busAddr string is set, if so use that
	if busAddr != "" {
		slot, _ := parseBusAddrString(busAddr)
		if slot > 0 {
			status := bus[slot]
			if !status {
				if err := bus.SetSlot(slot); err != nil {
					log.Debugf("Could not set PCI Bus slot: %v", err)
				}
				return slot
			}
		}
	}
	// should we error or allocate an open slot?

	// start from the top end of PCI range and descend to PCI offset to avoid
	// using typically assigned pci slots
	for slot := PCISlotMax - 1; slot > PCISlotOffset; slot-- {
		status := bus[slot]
		if !status {
			if err := bus.SetSlot(slot); err != nil {
				log.Debugf("Could not set PCI Bus slot: %v", err)
			}
			return slot
		}
	}
	log.Errorf("PCIBus(%v) No PCI slots remaining", bus)
	return -1
}

func parseBusAddrString(addr string) (int, error) {
	addrString := addr

	// someone tossed a pcie.0/1  or something
	if strings.Contains(addr, "/") {
		toks := strings.Split(addr, "/")
		addrString = toks[1]
	}

	// if someone already makes it hex (0x12)
	addrString = strings.Replace(addrString, "0x", "", -1)
	addrString = strings.Replace(addrString, "0X", "", -1)

	addrInt, err := strconv.Atoi(addr)
	if err != nil {
		return -1, err
	}

	return addrInt, nil
}

// PCIeRootPortDevice represents a memory balloon device.
type PCIeRootPortDevice struct {
	ID string `yaml:"id"` // format: rp{n}, n>=0

	Bus     string `yaml:"bus"`     // default is pcie.0
	Chassis string `yaml:"chassis"` // (slot, chassis) pair is mandatory and must be unique for each pcie-root-port, >=0, default is 0x00
	Slot    string `yaml:"slot"`    // >=0, default is 0x00
	Port    string `yaml:"port"`    // specify which port of the PCIeRootBus (pcie.0 bus) to use.

	Multifunction bool   `yaml:"multifunction"` // true => "on", false => "off", default is off
	Addr          string `yaml:"addr"`          // >=0, default is 0x00

	// The PCIE-PCI bridge can be hot-plugged only into pcie-root-port that has 'bus-reserve' property value to
	// provide secondary bus for the hot-plugged bridge.
	BusReserve    string `yaml:"bus-reserve"`
	Pref64Reserve string `yaml:"pref64-reserve"` // reserve prefetched MMIO aperture, 64-bit
	Pref32Reserve string `yaml:"pref32-reserve"` // reserve prefetched MMIO aperture, 32-bit
	MemReserve    string `yaml:"memory-reserve"` // reserve non-prefetched MMIO aperture, 32-bit *only*
	IOReserve     string `yaml:"io-reserve"`     // IO reservation

	ROMFile string `yaml:"rom-file"` // ROMFile specifies the ROM file being used for this device.

	// Transport is the virtio transport for this device.
	Transport VirtioTransport `yaml:"transport"`
}

// QemuParams returns the qemu parameters built out of the PCIeRootPortDevice.
func (b PCIeRootPortDevice) QemuParams(config *Config) []string {
	var qemuParams []string
	var deviceParams []string
	driver := PCIeRootPort

	deviceParams = append(deviceParams, fmt.Sprintf("%s,id=%s", driver, b.ID))

	if b.Bus == "" {
		b.Bus = "pcie.0"
	}
	deviceParams = append(deviceParams, fmt.Sprintf("bus=%s", b.Bus))

	if b.Chassis == "" {
		b.Chassis = "0x00"
	}
	deviceParams = append(deviceParams, fmt.Sprintf("chassis=%s", b.Chassis))

	if b.Slot == "" {
		b.Slot = "0x00"
	}
	deviceParams = append(deviceParams, fmt.Sprintf("slot=%s", b.Slot))

	if b.Port != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("port=%s", b.Port))
	}

	if b.Addr == "" {
		b.Addr = "0x00"
	}
	deviceParams = append(deviceParams, fmt.Sprintf("addr=%s", b.Addr))

	if b.Multifunction {
		deviceParams = append(deviceParams, "multifunction=on")
	} else {
		// don't emit multifuction=off for sub-function devices
		if !strings.Contains(b.Addr, ".") {
			deviceParams = append(deviceParams, "multifunction=off")
		}
	}

	if b.BusReserve != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("bus-reserve=%s", b.BusReserve))
	}

	if b.Pref64Reserve != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("pref64-reserve=%s", b.Pref64Reserve))
	}

	if b.Pref32Reserve != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("pref32-reserve=%s", b.Pref32Reserve))
	}

	if b.MemReserve != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("mem-reserve=%s", b.MemReserve))
	}

	if b.IOReserve != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("io-reserve=%s", b.IOReserve))
	}

	if b.Transport.isVirtioPCI(config) && b.ROMFile != "" {
		deviceParams = append(deviceParams, fmt.Sprintf("romfile=%s", b.ROMFile))
	}

	qemuParams = append(qemuParams, "-device")
	qemuParams = append(qemuParams, strings.Join(deviceParams, ","))
	return qemuParams
}

// Valid returns true if the PCIeRootPortDevice structure is valid and complete.
func (b PCIeRootPortDevice) Valid() error {
	// the "pref32-reserve" and "pref64-reserve" hints are mutually exclusive.
	if b.Pref64Reserve != "" && b.Pref32Reserve != "" {
		return fmt.Errorf("PCIeRootPortDevice Pref64Reserve and Pref32Reserve are mutually exclusive")
	}

	if b.ID == "" {
		return fmt.Errorf("PCIeRootPortDevice has empty ID field")
	}

	return nil
}

func NewPCIeRootMultifunctionPortRange(idPrefix, bus, baseAddr string, numPorts int) ([]Device, error) {
	devices := []Device{}

	if idPrefix == "" {
		return devices, fmt.Errorf("Empty idPrefix provided")
	}

	if baseAddr == "" {
		return devices, fmt.Errorf("Empty baseAddr provided")
	}

	if numPorts < 1 {
		return devices, fmt.Errorf("numPorts must be greater than 0")
	}

	for p := 0; p < numPorts; p++ {
		rootPortID := fmt.Sprintf("%s.%s.%d", idPrefix, baseAddr, p)
		port := fmt.Sprintf("0x%x", p)
		chassis := fmt.Sprintf("0x%x", p)
		addr := fmt.Sprintf("%s.0x%x", baseAddr, p)

		pcieRootPort := PCIeRootPortDevice{
			ID:      rootPortID,
			Port:    port,
			Chassis: chassis,
			Addr:    addr,
			Bus:     bus,
		}

		if p == 0 {
			pcieRootPort.Multifunction = true
		}

		if err := pcieRootPort.Valid(); err != nil {
			return devices, fmt.Errorf("Error generating PCIeRootPortDevice: %+v", pcieRootPort)
		}
		devices = append(devices, pcieRootPort)
	}

	return devices, nil
}
