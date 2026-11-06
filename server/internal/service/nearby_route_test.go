package service

import (
	mathrand "math/rand"
	"strconv"
	"strings"
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
		{name: "two hours", options: NearbyRouteOptions{Budget: "2h"}, wantStops: 3, wantRadius: 2500},
		{name: "half day family", options: NearbyRouteOptions{Budget: "half", Scene: "family"}, wantStops: 4, wantRadius: 3500, wantScene: "family"},
		{name: "full day", options: NearbyRouteOptions{Budget: "day"}, wantStops: 6, wantRadius: 12000},
		{name: "legacy limit", options: NearbyRouteOptions{Limit: 3}, wantStops: 3, wantRadius: 5000},
		{name: "unknown values", options: NearbyRouteOptions{Budget: "forever", Scene: "unknown"}, wantStops: 4, wantRadius: 5000},
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

func TestNearbyCandidateScoreFallsBackToCategoryWithoutReviews(t *testing.T) {
	stall := nearbyCandidate{place: model.Place{Category: "stall"}, stop: RouteStopView{DistanceMeters: 800}}
	restaurant := nearbyCandidate{place: model.Place{Category: "restaurant"}, stop: RouteStopView{DistanceMeters: 800}}
	for _, scene := range []string{"budget_friendly", "no_queue", "local_explore"} {
		plan := normalizeNearbyRoutePlan(NearbyRouteOptions{Budget: "half", Scene: scene})
		if nearbyCandidateScore(stall, plan) <= nearbyCandidateScore(restaurant, plan) {
			t.Fatalf("scene %q did not use category fallback", scene)
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
	short := selectNearbyCandidates(candidates, normalizeNearbyRoutePlan(NearbyRouteOptions{Budget: "2h"}), mathrand.New(mathrand.NewSource(1)))
	fullDay := selectNearbyCandidates(candidates, normalizeNearbyRoutePlan(NearbyRouteOptions{Budget: "day"}), mathrand.New(mathrand.NewSource(1)))
	if len(short) != 3 || len(fullDay) != 6 {
		t.Fatalf("selected stops: short=%d full-day=%d", len(short), len(fullDay))
	}
	seen := make(map[uint]bool)
	for _, candidate := range fullDay {
		if seen[candidate.place.ID] {
			t.Fatalf("full-day route selected duplicate place %d", candidate.place.ID)
		}
		seen[candidate.place.ID] = true
	}
}

func TestSelectNearbyCandidatesVariesWithinQualityPool(t *testing.T) {
	plan := nearbyRoutePlan{targetStops: 2, radiusMeters: 5000}
	candidates := make([]nearbyCandidate, 0, 10)
	for index := 0; index < 9; index++ {
		candidates = append(candidates, nearbyCandidate{
			place: model.Place{ID: uint(index + 1), Category: "restaurant"},
			stop:  RouteStopView{DistanceMeters: 500 + index*10},
			score: 100 - float64(index),
		})
	}
	// 极低分地点应被排除在 targetStops*3 的优质候选池之外。
	candidates = append(candidates, nearbyCandidate{place: model.Place{ID: 99}, stop: RouteStopView{DistanceMeters: 600}, score: -1000})

	routes := make(map[string]bool)
	for seed := int64(1); seed <= 20; seed++ {
		selected := selectNearbyCandidates(candidates, plan, mathrand.New(mathrand.NewSource(seed)))
		ids := make([]string, 0, len(selected))
		for _, candidate := range selected {
			if candidate.place.ID == 99 {
				t.Fatal("low-quality candidate escaped the weighted candidate pool")
			}
			ids = append(ids, strconv.FormatUint(uint64(candidate.place.ID), 10))
		}
		routes[strings.Join(ids, ",")] = true
	}
	if len(routes) < 2 {
		t.Fatalf("weighted selection did not vary across requests: %#v", routes)
	}
}
