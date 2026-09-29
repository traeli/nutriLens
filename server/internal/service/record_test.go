package service

import (
	"testing"
	"time"
)

func TestValidateRecordInputAcceptsSelectedPlaceAndHotConclusion(t *testing.T) {
	input := RecordInput{
		PlaceID:    42,
		VisitDate:  time.Now(),
		Conclusion: "hot",
		Visibility: "private",
	}
	if err := validateRecordInput(input); err != nil {
		t.Fatalf("validateRecordInput() error = %v", err)
	}
}

func TestValidateRecordInputRejectsTypedPlaceWithoutMapSelection(t *testing.T) {
	input := RecordInput{
		PlaceName: "巷口面馆", CityCode: "shanghai", VisitDate: time.Now(),
		Conclusion: "recommend", Visibility: "private",
	}
	if err := validateRecordInput(input); err == nil {
		t.Fatal("validateRecordInput() error = nil, want selected place validation error")
	}
}

func TestValidCoordinate(t *testing.T) {
	if !validCoordinate(121.4288, 31.2282) {
		t.Fatal("validCoordinate() rejected a valid Shanghai coordinate")
	}
	if validCoordinate(0, 0) || validCoordinate(181, 31) {
		t.Fatal("validCoordinate() accepted an invalid coordinate")
	}
}

func TestValidateRecordInputRejectsMissingPlaceIdentity(t *testing.T) {
	input := RecordInput{VisitDate: time.Now(), Conclusion: "recommend", Visibility: "private"}
	if err := validateRecordInput(input); err == nil {
		t.Fatal("validateRecordInput() error = nil, want validation error")
	}
}
