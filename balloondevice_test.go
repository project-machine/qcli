package qemu

import "testing"

func TestVirtioBalloonValid(t *testing.T) {
	balloon := BalloonDevice{
		ID: "",
	}

	if err := balloon.Valid(); err == nil {
		t.Fatalf("balloon should be not valid when ID is empty")
	}

	balloon.ID = "balloon0"
	if err := balloon.Valid(); err != nil {
		t.Fatalf("balloon should be valid: %s", err)
	}
}

func TestAppendVirtioBalloon(t *testing.T) {
	balloonDevice := BalloonDevice{
		ID:      "balloon",
		ROMFile: romfile,
	}

	var deviceString = "-device " + string(VirtioBalloon) + "-" + string(TransportPCI)
	deviceString += ",id=" + balloonDevice.ID + ",romfile=" + balloonDevice.ROMFile

	var OnDeflateOnOMM = ",deflate-on-oom=on"
	var OffDeflateOnOMM = ",deflate-on-oom=off"

	var OnDisableModern = ",disable-modern=true"
	var OffDisableModern = ",disable-modern=false"

	testAppend(balloonDevice, deviceString+OffDeflateOnOMM+OffDisableModern, t)

	balloonDevice.DeflateOnOOM = true
	testAppend(balloonDevice, deviceString+OnDeflateOnOMM+OffDisableModern, t)

	balloonDevice.DisableModern = true
	testAppend(balloonDevice, deviceString+OnDeflateOnOMM+OnDisableModern, t)

}
