package service

import (
	"testing"

	model "shijibu/internal/model/pgsql"
)

func TestNormalizeNearbyRoutePlan(t *testing.T) {
	tests := []struct {
		name       string
		options    NearbyRouteOptions
		wantStops  int
		wantRadius float64
		wantScene  string
	}{
		{name: "two hours", options: NearbyRouteOptions{Budget: "2h"}, wantStops: 1, wantRadius: 2500},
		{name: "half day family", options: NearbyRouteOptions{Budget: "half", Scene: "family"}, wantStops: 2, wantRadius: 3500, wantScene: "family"},
		{name: "full day", options: NearbyRouteOptions{Budget: "day"}, wantStops: 3, wantRadius: 12000},
		{name: "legacy limit", options: NearbyRouteOptions{Limit: 3}, wantStops: 3, wantRadius: 5000},
		{name: "unknown values", options: NearbyRouteOptions{Budget: "forever", Scene: "unknown"}, wantStops: 2, wantRadius: 5000},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := normalizeNearbyRoutePlan(test.options)
			if plan.targetStops != test.wantStops || plan.radiusMeters != test.wantRadius || plan.scene != test.wantScene {
				t.Fatalf("plan = %#v, want stops=%d radius=%v scene=%q", plan, test.wantStops, test.wantRadius, test.wantScene)
			}
		})
	}
}

func TestNearbyCandidateScoreUsesSceneSignals(t *testing.T) {
	lowCost, highCost := 45.0, 180.0
	lowWait, highWait := 8.0, 55.0
	cheapFast := nearbyCandidate{
		place: model.Place{Category: "restaurant"}, stop: RouteStopView{DistanceMeters: 800},
		signal: nearbyPlaceSignal{ReviewCount: 2, AverageCost: &lowCost, AverageWait: &lowWait},
	}
	expensiveSlow := nearbyCandidate{
		place: model.Place{Category: "restaurant"}, stop: RouteStopView{DistanceMeters: 800},
		signal: nearbyPlaceSignal{ReviewCount: 2, AverageCost: &highCost, AverageWait: &highWait},
	}
	for _, scene := range []string{"budget_friendly", "no_queue", "family", "elderly"} {
		plan := normalizeNearbyRoutePlan(NearbyRouteOptions{Budget: "half", Scene: scene})
		if nearbyCandidateScore(cheapFast, plan) <= nearbyCandidateScore(expensiveSlow, plan) {
			t.Fatalf("scene %q did not prefer the matching cost/wait signals", scene)
		}
	}
}

func TestSelectNearbyCandidatesRespectsBudgetAndDiversity(t *testing.T) {
	candidates := []nearbyCandidate{
		{place: model.Place{ID: 1, Category: "restaurant", BusinessArea: "A"}, stop: RouteStopView{DistanceMeters: 500}, score: 40},
		{place: model.Place{ID: 2, Category: "restaurant", BusinessArea: "A"}, stop: RouteStopView{DistanceMeters: 600}, score: 39},
		{place: model.Place{ID: 3, Category: "drink", BusinessArea: "B"}, stop: RouteStopView{DistanceMeters: 700}, score: 35},
		{place: model.Place{ID: 4, Category: "stall", BusinessArea: "C"}, stop: RouteStopView{DistanceMeters: 800}, score: 30},
		{place: model.Place{ID: 5, Category: "night_market", BusinessArea: "D"}, stop: RouteStopView{DistanceMeters: 900}, score: 25},
		{place: model.Place{ID: 6, Category: "restaurant", BusinessArea: "E"}, stop: RouteStopView{DistanceMeters: 1000}, score: 20},
	}
	short := selectNearbyCandidates(candidates, normalizeNearbyRoutePlan(NearbyRouteOptions{Budget: "2h"}))
	fullDay := selectNearbyCandidates(candidates, normalizeNearbyRoutePlan(NearbyRouteOptions{Budget: "day"}))
	if len(short) != 1 || len(fullDay) != 3 {
		t.Fatalf("selected stops: short=%d full-day=%d", len(short), len(fullDay))
	}
	if fullDay[0].place.ID == fullDay[1].place.ID || fullDay[1].place.Category == "restaurant" {
		t.Fatalf("full-day route did not apply category diversity: %#v", fullDay)
	}
}
