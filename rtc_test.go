package qcli

import "testing"

var (
	rtcString = "-rtc base=utc,driftfix=slew,clock=host"
)

func TestAppendRTC(t *testing.T) {
	rtc := RTC{
		Base:     UTC,
		Clock:    Host,
		DriftFix: Slew,
	}

	testAppend(rtc, rtcString, t)
}

func TestBadRTC(t *testing.T) {
	c := &Config{}
	c.appendRTC()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}

	c = &Config{
		RTC: RTC{
			Clock: RTCClock("invalid"),
		},
	}
	c.appendRTC()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}

	c = &Config{
		RTC: RTC{
			Clock:    Host,
			DriftFix: RTCDriftFix("invalid"),
		},
	}
	c.appendRTC()
	if len(c.qemuParams) != 0 {
		t.Errorf("Expected empty qemuParams, found %s", c.qemuParams)
	}
}
