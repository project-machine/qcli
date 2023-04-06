package qcli

import (
	"strings"
	"testing"
)

var (
	smbFile           = "-smbios file=foo"
	smbType0Bios      = "-smbios type=0,vendor=Vendor,version=Version,date=Date,release=1.0,uefi=on"
	smbType1System    = "-smbios type=1,manufacturer=Manufacturer,product=Product,version=Version,serial=Serial,uuid=UUID,sku=SKU,family=Family"
	smbType2Baseboard = "-smbios type=2,manufacturer=Manufacturer,product=Product,version=Version,serial=Serial,asset=Asset,location=Location"
	smbType3Chassis   = "-smbios type=3,manufacturer=Manufacturer,version=Version,serial=Serial,asset=Asset,sku=SKU"
	smbType4Processor = "-smbios type=4,sock_pfx=SocketPrefix,manufacturer=Manufacturer,version=Version,serial=Serial,asset=Asset,part=Part"
	smbType17Memory   = "-smbios type=17,loc_pfx=LocationPrefix,bank=Bank,manufacturer=Manufacturer,serial=Serial,asset=Asset,part=Part,speed=3600"
)

var bios = SMTableBIOS{
	Vendor:  "Vendor",
	Version: "Version",
	Date:    "Date",
	Release: "1.0",
	UEFI:    "on",
}
var system = SMTableSystem{
	Manufacturer: "Manufacturer",
	Product:      "Product",
	Version:      "Version",
	Serial:       "Serial",
	UUID:         "UUID",
	SKU:          "SKU",
	Family:       "Family",
}
var baseboard = SMTableBaseboard{
	Manufacturer: "Manufacturer",
	Product:      "Product",
	Version:      "Version",
	Serial:       "Serial",
	Asset:        "Asset",
	Location:     "Location",
}
var chassis = SMTableChassis{
	Manufacturer: "Manufacturer",
	Version:      "Version",
	Serial:       "Serial",
	Asset:        "Asset",
	SKU:          "SKU",
}
var processor = SMTableProcessor{
	SocketPrefix: "SocketPrefix",
	Manufacturer: "Manufacturer",
	Version:      "Version",
	Serial:       "Serial",
	Asset:        "Asset",
	Part:         "Part",
}
var memory = SMTableMemory{
	LocationPrefix: "LocationPrefix",
	Bank:           "Bank",
	Manufacturer:   "Manufacturer",
	Serial:         "Serial",
	Asset:          "Asset",
	Part:           "Part",
	Speed:          "3600",
}
var smbFull = SMBIOSInfo{
	BIOS:       bios,
	System:     system,
	Baseboard:  baseboard,
	Chassis:    chassis,
	Processors: []SMTableProcessor{processor},
	Memory:     []SMTableMemory{memory},
}

func TestAppendSMBIOSFile(t *testing.T) {
	smb := SMBIOSInfo{
		File: "foo",
	}
	testAppend(smb, smbFile, t)
}

func TestAppendSMBIOSType0BIOS(t *testing.T) {
	smb := SMBIOSInfo{
		BIOS: bios,
	}
	testAppend(smb, smbType0Bios, t)
}

func TestAppendSMBIOSType1System(t *testing.T) {
	smb := SMBIOSInfo{
		System: system,
	}
	testAppend(smb, smbType1System, t)
}

func TestAppendSMBIOSType2Baseboard(t *testing.T) {
	smb := SMBIOSInfo{
		Baseboard: baseboard,
	}
	testAppend(smb, smbType2Baseboard, t)
}

func TestAppendSMBIOSType3Chassis(t *testing.T) {
	smb := SMBIOSInfo{
		Chassis: chassis,
	}
	testAppend(smb, smbType3Chassis, t)
}

func TestAppendSMBIOSType4Processor(t *testing.T) {
	smb := SMBIOSInfo{
		Processors: []SMTableProcessor{processor},
	}
	testAppend(smb, smbType4Processor, t)
}

func TestAppendSMBIOSType17Memory(t *testing.T) {
	smb := SMBIOSInfo{
		Memory: []SMTableMemory{memory},
	}
	testAppend(smb, smbType17Memory, t)
}

func TestAppendSMBIOSFUll(t *testing.T) {
	tables := []string{smbType0Bios, smbType1System, smbType2Baseboard, smbType3Chassis, smbType4Processor, smbType17Memory}
	smbFullStr := strings.Join(tables, " ")
	testAppend(smbFull, smbFullStr, t)
}
