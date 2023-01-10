package qemu

import "testing"

func TestAppendFwcfg(t *testing.T) {
	fwcfgString := "-fw_cfg name=opt/com.mycompany/blob,file=./my_blob.bin"
	fwcfg := FwCfg{
		Name: "opt/com.mycompany/blob",
		File: "./my_blob.bin",
	}
	testAppend(fwcfg, fwcfgString, t)

	fwcfgString = "-fw_cfg name=opt/com.mycompany/blob,string=foo"
	fwcfg = FwCfg{
		Name: "opt/com.mycompany/blob",
		Str:  "foo",
	}
	testAppend(fwcfg, fwcfgString, t)
}

func TestBadFwcfg(t *testing.T) {
	c := &Config{}
	c.appendFwCfg(nil)
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}

	c = &Config{
		FwCfg: []FwCfg{
			{
				Name: "name=opt/com.mycompany/blob",
				File: "./my_blob.bin",
				Str:  "foo",
			},
		},
	}
	c.appendFwCfg(nil)
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}
