package service

import (
	"reflect"
	"testing"
	"time"

	model "shijibu/internal/model/pgsql"
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

func TestRecordFromInputKeepsVisitDataInAggregate(t *testing.T) {
	input := RecordInput{
		PlaceID: 42, VisitDate: time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC),
		ConsumerType: " dine_in ", Conclusion: "recommend", MealPeriod: " dinner ",
		Dishes: model.JSONDocument(`[{"name":"面"}]`), Content: " 很好吃 ", Visibility: "private",
	}
	record := recordFromInput(9, input, []uint{2, 5})
	if record.UserID != 9 || record.PlaceID != 42 || record.VersionNo != 1 {
		t.Fatalf("record identity = %#v", record)
	}
	if record.ConsumerType != "dine_in" || record.MealPeriod != "dinner" || record.Content != "很好吃" {
		t.Fatalf("record normalized fields = %#v", record)
	}
	if !reflect.DeepEqual([]uint(record.TagIDs), []uint{2, 5}) {
		t.Fatalf("record tag ids = %#v", record.TagIDs)
	}
	if record.Media == nil || record.Evidences == nil {
		t.Fatal("record JSON arrays must be initialized")
	}
}
