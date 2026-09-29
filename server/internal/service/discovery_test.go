package service

import "testing"

func TestHaversineMeters(t *testing.T) {
	distance := haversineMeters(121.4737, 31.2304, 121.4837, 31.2304)
	if distance < 900 || distance > 1100 {
		t.Fatalf("haversineMeters() = %.2f, want about 950 meters", distance)
	}
}

func TestValidPublicCoordinate(t *testing.T) {
	if !validPublicCoordinate(121.4737, 31.2304) {
		t.Fatal("validPublicCoordinate() rejected a valid coordinate")
	}
	if validPublicCoordinate(0, 0) || validPublicCoordinate(181, 31) {
		t.Fatal("validPublicCoordinate() accepted an invalid coordinate")
	}
}
