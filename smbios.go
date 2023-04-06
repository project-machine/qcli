package qcli

import (
	"fmt"
	"strings"
)

/*
smbios:
  file:
  bios:
  system:
  baseboard:
  chassis:
  processors:
  - processor
  - processor
  memory:
  - memory
  - memory
*/

type SMBIOSInfo struct {
	File       string             `yaml:"file,omitempty"`       // -smbios file
	BIOS       SMTableBIOS        `yaml:"bios,omitempty"`       // -smbios type=0
	System     SMTableSystem      `yaml:"system,omitempty"`     // -smbios type=1
	Baseboard  SMTableBaseboard   `yaml:"baseboard,omitempty"`  // -smbios type=2
	Chassis    SMTableChassis     `yaml:"chassis,omitempty"`    // -smbios type=3
	Processors []SMTableProcessor `yaml:"processors,omitempty"` // -smbios type=4
	Memory     []SMTableMemory    `yaml:"memory,omitempty"`     // -smbios type=17
}

const SMTableBIOSType = 0

type SMTableBIOS struct {
	Vendor  string `yaml:"vendor,omitempty"`
	Version string `yaml:"version,omitempty"`
	Date    string `yaml:"date,omitempty"`
	Release string `yaml:"release,omitempty"`
	UEFI    string `yaml:"uefi,omitempty"`
}

func (table SMTableBIOS) Valid() error {
	if table.Release != "" {
		var major int
		var minor int
		_, err := fmt.Sscanf(table.Release, "%d.%d", &major, &minor)
		if err != nil {
			return fmt.Errorf("SMTableBIOS Type=0 Release field is not in <digit>.<digit> format")
		}
	}
	if table.UEFI != "" {
		val := strings.ToLower(table.UEFI)
		if val != "on" && val != "off" {
			return fmt.Errorf("SMTableBIOS Type=0 UEFI field is not 'on' or 'off': %s", table.UEFI)
		}
	}
	return nil
}

func (table SMTableBIOS) QemuParams(config *Config) []string {
	var qemuParams []string
	var tableParams []string
	typeParam := fmt.Sprintf("type=%d", SMTableBIOSType)

	if table.Vendor != "" {
		tableParams = append(tableParams, "vendor="+table.Vendor)
	}
	if table.Version != "" {
		tableParams = append(tableParams, "version="+table.Version)
	}
	if table.Date != "" {
		tableParams = append(tableParams, "date="+table.Date)
	}
	if table.Release != "" {
		tableParams = append(tableParams, "release="+table.Release)
	}
	if table.UEFI != "" {
		tableParams = append(tableParams, "uefi="+table.UEFI)
	}
	if len(tableParams) > 0 {
		qemuParams = append(qemuParams, "-smbios")
		tableParams = append([]string{typeParam}, tableParams...)
		qemuParams = append(qemuParams, strings.Join(tableParams, ","))
	}
	// fmt.Printf("SMTableBIOS: table=%v tparams=%v qparams=%v\n", table, tableParams, qemuParams)
	return qemuParams
}

const SMTableSystemType = 1

type SMTableSystem struct {
	Manufacturer string `yaml:"manufacturer,omitempty"`
	Product      string `yaml:"product,omitempty"`
	Version      string `yaml:"version,omitempty"`
	Serial       string `yaml:"serial,omitempty"`
	UUID         string `yaml:"uuid,omitempty"`
	SKU          string `yaml:"sku,omitempty"`
	Family       string `yaml:"family,omitempty"`
}

func (table SMTableSystem) Valid() error {
	// no format requirements
	return nil
}

func (table SMTableSystem) QemuParams(config *Config) []string {
	var qemuParams []string
	var tableParams []string
	typeParam := fmt.Sprintf("type=%d", SMTableSystemType)

	if table.Manufacturer != "" {
		tableParams = append(tableParams, "manufacturer="+table.Manufacturer)
	}
	if table.Product != "" {
		tableParams = append(tableParams, "product="+table.Product)
	}
	if table.Version != "" {
		tableParams = append(tableParams, "version="+table.Version)
	}
	if table.Serial != "" {
		tableParams = append(tableParams, "serial="+table.Serial)
	}
	if table.UUID != "" {
		tableParams = append(tableParams, "uuid="+table.UUID)
	}
	if table.SKU != "" {
		tableParams = append(tableParams, "sku="+table.SKU)
	}
	if table.Family != "" {
		tableParams = append(tableParams, "family="+table.Family)
	}

	if len(tableParams) > 0 {
		qemuParams = append(qemuParams, "-smbios")
		tableParams = append([]string{typeParam}, tableParams...)
		qemuParams = append(qemuParams, strings.Join(tableParams, ","))
	}
	return qemuParams
}

const SMTableBaseboardType = 2

type SMTableBaseboard struct {
	Manufacturer string `yaml:"manufacturer,omitempty"`
	Product      string `yaml:"product,omitempty"`
	Version      string `yaml:"version,omitempty"`
	Serial       string `yaml:"serial,omitempty"`
	Asset        string `yaml:"asset,omitempty"`
	Location     string `yaml:"location,omitempty"`
}

func (table SMTableBaseboard) Valid() error {
	// no format requirements
	return nil
}

func (table SMTableBaseboard) QemuParams(config *Config) []string {
	var qemuParams []string
	var tableParams []string
	typeParam := fmt.Sprintf("type=%d", SMTableBaseboardType)

	if table.Manufacturer != "" {
		tableParams = append(tableParams, "manufacturer="+table.Manufacturer)
	}
	if table.Product != "" {
		tableParams = append(tableParams, "product="+table.Product)
	}
	if table.Version != "" {
		tableParams = append(tableParams, "version="+table.Version)
	}
	if table.Serial != "" {
		tableParams = append(tableParams, "serial="+table.Serial)
	}
	if table.Asset != "" {
		tableParams = append(tableParams, "asset="+table.Asset)
	}
	if table.Location != "" {
		tableParams = append(tableParams, "location="+table.Location)
	}

	if len(tableParams) > 0 {
		qemuParams = append(qemuParams, "-smbios")
		tableParams = append([]string{typeParam}, tableParams...)
		qemuParams = append(qemuParams, strings.Join(tableParams, ","))
	}
	return qemuParams
}

const SMTableChassisType = 3

type SMTableChassis struct {
	Manufacturer string `yaml:"manufacturer,omitempty"`
	Version      string `yaml:"version,omitempty"`
	Serial       string `yaml:"serial,omitempty"`
	Asset        string `yaml:"asset,omitempty"`
	SKU          string `yaml:"sku,omitempty"`
}

func (table SMTableChassis) Valid() error {
	// no format requirements
	return nil
}

func (table SMTableChassis) QemuParams(config *Config) []string {
	var qemuParams []string
	var tableParams []string
	typeParam := fmt.Sprintf("type=%d", SMTableChassisType)

	if table.Manufacturer != "" {
		tableParams = append(tableParams, "manufacturer="+table.Manufacturer)
	}
	if table.Version != "" {
		tableParams = append(tableParams, "version="+table.Version)
	}
	if table.Serial != "" {
		tableParams = append(tableParams, "serial="+table.Serial)
	}
	if table.Asset != "" {
		tableParams = append(tableParams, "asset="+table.Asset)
	}
	if table.SKU != "" {
		tableParams = append(tableParams, "sku="+table.SKU)
	}
	if len(tableParams) > 0 {
		qemuParams = append(qemuParams, "-smbios")
		tableParams = append([]string{typeParam}, tableParams...)
		qemuParams = append(qemuParams, strings.Join(tableParams, ","))
	}
	return qemuParams
}

const SMTableProcessorType = 4

type SMTableProcessor struct {
	SocketPrefix string `yaml:"socket-prefix,omitempty"`
	Manufacturer string `yaml:"manufacturer,omitempty"`
	Version      string `yaml:"version,omitempty"`
	Serial       string `yaml:"serial,omitempty"`
	Asset        string `yaml:"asset,omitempty"`
	Part         string `yaml:"part,omitempty"`
}

func (table SMTableProcessor) Valid() error {
	// no format requirements
	return nil
}

func (table SMTableProcessor) QemuParams(config *Config) []string {
	var qemuParams []string
	var tableParams []string
	typeParam := fmt.Sprintf("type=%d", SMTableProcessorType)

	if table.SocketPrefix != "" {
		tableParams = append(tableParams, "sock_pfx="+table.SocketPrefix)
	}
	if table.Manufacturer != "" {
		tableParams = append(tableParams, "manufacturer="+table.Manufacturer)
	}
	if table.Version != "" {
		tableParams = append(tableParams, "version="+table.Version)
	}
	if table.Serial != "" {
		tableParams = append(tableParams, "serial="+table.Serial)
	}
	if table.Asset != "" {
		tableParams = append(tableParams, "asset="+table.Asset)
	}
	if table.Part != "" {
		tableParams = append(tableParams, "part="+table.Part)
	}

	if len(tableParams) > 0 {
		qemuParams = append(qemuParams, "-smbios")
		tableParams = append([]string{typeParam}, tableParams...)
		qemuParams = append(qemuParams, strings.Join(tableParams, ","))
	}
	return qemuParams
}

const SMTableMemoryType = 17

type SMTableMemory struct {
	LocationPrefix string `yaml:"location-prefix,omitempty"`
	Bank           string `yaml:"bank,omitempty"`
	Manufacturer   string `yaml:"manufacturer,omitempty"`
	Serial         string `yaml:"serial,omitempty"`
	Asset          string `yaml:"asset,omitempty"`
	Part           string `yaml:"part,omitempty"`
	Speed          string `yaml:"speed,omitempty"`
}

func (table SMTableMemory) Valid() error {
	if table.Speed != "" {
		var speed int
		_, err := fmt.Sscanf(table.Speed, "%d", &speed)
		if err != nil {
			return fmt.Errorf("SMTableMemory Type=17 Speed field must be a number, found: %s", table.Speed)
		}
	}
	return nil
}

func (table SMTableMemory) QemuParams(config *Config) []string {
	var qemuParams []string
	var tableParams []string
	typeParam := fmt.Sprintf("type=%d", SMTableMemoryType)

	if table.LocationPrefix != "" {
		tableParams = append(tableParams, "loc_pfx="+table.LocationPrefix)
	}
	if table.Bank != "" {
		tableParams = append(tableParams, "bank="+table.Bank)
	}
	if table.Manufacturer != "" {
		tableParams = append(tableParams, "manufacturer="+table.Manufacturer)
	}
	if table.Serial != "" {
		tableParams = append(tableParams, "serial="+table.Serial)
	}
	if table.Asset != "" {
		tableParams = append(tableParams, "asset="+table.Asset)
	}
	if table.Part != "" {
		tableParams = append(tableParams, "part="+table.Part)
	}
	if table.Speed != "" {
		tableParams = append(tableParams, "speed="+table.Speed)
	}
	if len(tableParams) > 0 {
		qemuParams = append(qemuParams, "-smbios")
		tableParams = append([]string{typeParam}, tableParams...)
		qemuParams = append(qemuParams, strings.Join(tableParams, ","))
	}
	return qemuParams
}

/*
type SMBIOSInfo struct {
	File       string             `yaml:"file,omitempty"`       // -smbios file
	BIOS       SMTableBIOS        `yaml:"bios,omitempty"`       // -smbios type=0
	System     SMTableSystem      `yaml:"system,omitempty"`     // -smbios type=1
	Baseboard  SMTableBaseboard   `yaml:"baseboard,omitempty"`  // -smbios type=2
	Chassis    SMTableChassis     `yaml:"chassis,omitempty"`    // -smbios type=3
	Processors []SMTableProcessor `yaml:"processors,omitempty"` // -smbios type=4
	Memory     []SMTableMemory    `yaml:"memory,omitempty"`     // -smbios type=17
}
*/

// Valid returns true if the SMBIOSInfo structure is valid and complete.
func (smb SMBIOSInfo) Valid() error {
	if err := smb.BIOS.Valid(); err != nil {
		return err
	}
	if err := smb.System.Valid(); err != nil {
		return err
	}
	if err := smb.Baseboard.Valid(); err != nil {
		return err
	}
	if err := smb.Chassis.Valid(); err != nil {
		return err
	}
	for _, proc := range smb.Processors {
		if err := proc.Valid(); err != nil {
			return err
		}
	}
	for _, mem := range smb.Memory {
		if err := mem.Valid(); err != nil {
			return err
		}
	}
	return nil
}

// QemuParams returns the qemu parameters built out of the SMBIOSInfo object
func (smb SMBIOSInfo) QemuParams(config *Config) []string {
	var qemuParams []string

	if smb.File != "" {
		qemuParams = append(qemuParams, "-smbios", "file="+smb.File)
	}
	qemuParams = append(qemuParams, smb.BIOS.QemuParams(config)...)
	qemuParams = append(qemuParams, smb.System.QemuParams(config)...)
	qemuParams = append(qemuParams, smb.Baseboard.QemuParams(config)...)
	qemuParams = append(qemuParams, smb.Chassis.QemuParams(config)...)
	for _, proc := range smb.Processors {
		qemuParams = append(qemuParams, proc.QemuParams(config)...)
	}
	for _, mem := range smb.Memory {
		qemuParams = append(qemuParams, mem.QemuParams(config)...)
	}

	return qemuParams
}

func (config *Config) appendSMBIOSInfo() error {
	//fmt.Printf("config called appendSMBIOSInfo()\n")
	if err := config.SMBIOS.Valid(); err != nil {
		return err
	}
	config.qemuParams = append(config.qemuParams, config.SMBIOS.QemuParams(config)...)
	return nil
}
