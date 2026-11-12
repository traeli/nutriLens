package service

import (
	"testing"

	model "shijibu/internal/model/pgsql"
)

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

func TestMergeScenarioFiltersPrefersCityOverride(t *testing.T) {
	rows := []model.DiscoveryScenario{
		{ID: 1, Code: "solo", Label: "一个人吃", ConditionType: "scene", ConditionValue: "solo", SortOrder: 50},
		{ID: 2, CityCode: "310000", Code: "solo", Label: "独自探店", ConditionType: "scene", ConditionValue: "solo_shanghai", SortOrder: 10},
		{ID: 3, Code: "family", Label: "亲子同行", ConditionType: "scene", ConditionValue: "family", SortOrder: 20},
	}

	got := mergeScenarioFilters(rows, "310000")
	if len(got) != 2 {
		t.Fatalf("mergeScenarioFilters() length = %d, want 2", len(got))
	}
	if got[0].Code != "solo" || got[0].Label != "独自探店" || got[0].Query.Value != "solo_shanghai" {
		t.Fatalf("mergeScenarioFilters() city override = %#v", got[0])
	}
	if got[1].Code != "family" {
		t.Fatalf("mergeScenarioFilters() order = %#v", got)
	}
}

func TestNearbyIconKey(t *testing.T) {
	if got := nearbyIconKey("drink"); got != "coffee" {
		t.Fatalf("nearbyIconKey(drink) = %q, want coffee", got)
	}
	if got := nearbyIconKey("restaurant"); got != "bowl" {
		t.Fatalf("nearbyIconKey(restaurant) = %q, want bowl", got)
	}
}

func TestPublicEditorialAuthorIsPlatformOwned(t *testing.T) {
	if editorialAuthorName != "饭友记编辑" {
		t.Fatalf("editorialAuthorName = %q", editorialAuthorName)
	}
}
