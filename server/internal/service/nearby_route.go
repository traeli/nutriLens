package service

import (
	"errors"
	"fmt"
	"math"
	"strings"

	model "shijibu/internal/model/pgsql"

	"gorm.io/gorm"
)

// NearbyRouteOptions 是实时路线的用户显式输入。Budget 和 Scene 都是可选项，
// 未传时保留旧客户端的 3～4 站行为。
type NearbyRouteOptions struct {
	CityCode  string
	Longitude float64
	Latitude  float64
	Limit     int
	Budget    string
	Scene     string
}

type nearbyRoutePlan struct {
	budget       string
	budgetLabel  string
	scene        string
	sceneLabel   string
	targetStops  int
	radiusMeters float64
}

type nearbyPlaceSignal struct {
	PlaceID        uint
	ReviewCount    int64
	AverageCost    *float64
	AverageWait    *float64
	LateNightCount int64
}

type nearbyCandidate struct {
	place  model.Place
	stop   RouteStopView
	signal nearbyPlaceSignal
	score  float64
}

var nearbySceneLabels = map[string]string{
	"first_visit": "第一次来", "family": "亲子同行", "rainy_day": "雨天照吃",
	"late_night": "深夜到访", "solo": "一个人吃", "local_explore": "本地人探店",
	"couple": "两人约会", "friends": "朋友聚餐", "elderly": "长辈同行",
	"budget_friendly": "省心实惠", "no_queue": "不想排队",
}

func normalizeNearbyRoutePlan(options NearbyRouteOptions) nearbyRoutePlan {
	plan := nearbyRoutePlan{budget: "half", budgetLabel: "半天", targetStops: 2, radiusMeters: 5000}
	switch strings.TrimSpace(options.Budget) {
	case "2h":
		plan.budget, plan.budgetLabel, plan.targetStops, plan.radiusMeters = "2h", "2 小时", 1, 2500
	case "day":
		plan.budget, plan.budgetLabel, plan.targetStops, plan.radiusMeters = "day", "一整天", 3, 12000
	case "half":
		// 使用默认值。
	default:
		if options.Limit >= 3 && options.Limit <= 4 {
			plan.targetStops = options.Limit
		}
	}
	if label, ok := nearbySceneLabels[strings.TrimSpace(options.Scene)]; ok {
		plan.scene, plan.sceneLabel = strings.TrimSpace(options.Scene), label
	}
	switch plan.scene {
	case "family", "rainy_day", "elderly":
		plan.radiusMeters *= 0.7
	}
	return plan
}

// NearbyRoute 根据本次定位、可用时间和场景偏好即时选点，不读取预设 city_routes。
func (s *DiscoveryService) NearbyRoute(options NearbyRouteOptions) (*RouteDetailView, error) {
	options.CityCode = strings.TrimSpace(options.CityCode)
	if options.CityCode == "" || !validPublicCoordinate(options.Longitude, options.Latitude) {
		return nil, ErrInvalidInput
	}
	plan := normalizeNearbyRoutePlan(options)

	var city model.City
	if err := s.db.Where("code = ? AND enabled = ?", options.CityCode, true).First(&city).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var places []model.Place
	if err := s.db.Where("city_code = ? AND status = ? AND longitude <> 0 AND latitude <> 0", options.CityCode, "active").Find(&places).Error; err != nil {
		return nil, err
	}
	if len(places) == 0 {
		return emptyNearbyRoute(options.CityCode), nil
	}

	signals, err := s.nearbyPlaceSignals(placeIDs(places))
	if err != nil {
		return nil, err
	}
	candidates := buildNearbyCandidates(places, signals, options, plan)
	selected := selectNearbyCandidates(candidates, plan)
	stops, routeDistance := orderNearbyStops(selected, options.Longitude, options.Latitude)

	tag := "实时路线"
	if plan.sceneLabel != "" {
		tag = plan.sceneLabel
	}
	metaParts := []string{plan.budgetLabel, fmt.Sprintf("%d 站", len(stops))}
	if routeDistance > 0 {
		metaParts = append(metaParts, formatRouteDistance(routeDistance))
	}
	return &RouteDetailView{
		CityCode: options.CityCode,
		Title:    fmt.Sprintf("从你附近开始的 %d 站吃喝路线", len(stops)),
		Tag:      tag,
		Meta:     strings.Join(metaParts, " · "),
		Footnote: "路线按当前位置与本次筛选实时生成；营业状态、交通方式与路况请以地图 App 为准",
		Stops:    stops,
		Source:   "nearby",
	}, nil
}

func emptyNearbyRoute(cityCode string) *RouteDetailView {
	return &RouteDetailView{CityCode: cityCode, Title: "附近吃喝路线", Meta: "附近暂时没有可用地点", Stops: []RouteStopView{}, Source: "nearby"}
}

func (s *DiscoveryService) nearbyPlaceSignals(ids []uint) (map[uint]nearbyPlaceSignal, error) {
	var rows []nearbyPlaceSignal
	err := s.db.Table("restaurant_reviews AS reviews").
		Select(`reviews.place_id, COUNT(reviews.id) AS review_count,
			AVG(versions.average_cost) AS average_cost, AVG(versions.wait_minutes) AS average_wait,
			COUNT(reviews.id) FILTER (WHERE versions.meal_period = 'late_night') AS late_night_count`).
		Joins("JOIN restaurant_review_versions AS versions ON versions.id = reviews.current_version_id").
		Where("reviews.status = ? AND reviews.deleted_at IS NULL", "published").
		Where("reviews.place_id IN ?", ids).
		Group("reviews.place_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]nearbyPlaceSignal, len(rows))
	for _, row := range rows {
		result[row.PlaceID] = row
	}
	return result, nil
}

func buildNearbyCandidates(places []model.Place, signals map[uint]nearbyPlaceSignal, options NearbyRouteOptions, plan nearbyRoutePlan) []nearbyCandidate {
	candidates := make([]nearbyCandidate, 0, len(places))
	for _, place := range places {
		distance := haversineMeters(options.Longitude, options.Latitude, place.Longitude, place.Latitude)
		signal := signals[place.ID]
		candidate := nearbyCandidate{
			place: place, signal: signal,
			stop: RouteStopView{PlaceID: place.ID, Name: place.Name, Address: place.Address, Longitude: place.Longitude, Latitude: place.Latitude,
				DistanceMeters: int(math.Round(distance)), ExperienceCount: signal.ReviewCount, IconKey: nearbyIconKey(place.Category)},
		}
		candidate.score = nearbyCandidateScore(candidate, plan)
		candidates = append(candidates, candidate)
	}
	return candidates
}

func nearbyCandidateScore(candidate nearbyCandidate, plan nearbyRoutePlan) float64 {
	distanceWeight := 30.0
	if plan.scene == "rainy_day" || plan.scene == "family" || plan.scene == "elderly" {
		distanceWeight = 45
	}
	score := math.Log1p(float64(candidate.signal.ReviewCount))*16 - float64(candidate.stop.DistanceMeters)/plan.radiusMeters*distanceWeight
	lowCost := candidate.signal.AverageCost != nil && *candidate.signal.AverageCost <= 80
	lowWait := candidate.signal.AverageWait != nil && *candidate.signal.AverageWait <= 20
	switch plan.scene {
	case "first_visit":
		score += math.Log1p(float64(candidate.signal.ReviewCount)) * 10
	case "family", "elderly":
		if lowWait {
			score += 18
		}
		if candidate.place.Category == "restaurant" {
			score += 10
		}
	case "rainy_day":
		if candidate.place.Category == "restaurant" || candidate.place.Category == "drink" {
			score += 12
		}
	case "late_night":
		score += float64(candidate.signal.LateNightCount) * 16
		if candidate.place.Category == "night_market" {
			score += 18
		}
	case "solo":
		if lowCost {
			score += 14
		}
		if lowWait {
			score += 12
		}
		if candidate.place.Category == "drink" || candidate.place.Category == "stall" {
			score += 10
		}
	case "local_explore":
		if candidate.signal.ReviewCount > 0 && candidate.signal.ReviewCount <= 3 {
			score += 14
		}
	case "couple":
		if candidate.place.Category == "restaurant" || candidate.place.Category == "drink" {
			score += 14
		}
	case "friends":
		if candidate.place.Category == "restaurant" || candidate.place.Category == "night_market" {
			score += 14
		}
	case "budget_friendly":
		if lowCost {
			score += 24
		}
	case "no_queue":
		if lowWait {
			score += 24
		}
	}
	return score
}

func selectNearbyCandidates(candidates []nearbyCandidate, plan nearbyRoutePlan) []nearbyCandidate {
	withinRadius := make([]nearbyCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if float64(candidate.stop.DistanceMeters) <= plan.radiusMeters {
			withinRadius = append(withinRadius, candidate)
		}
	}
	if len(withinRadius) < plan.targetStops {
		withinRadius = candidates
	}
	remaining := append([]nearbyCandidate(nil), withinRadius...)
	selected := make([]nearbyCandidate, 0, plan.targetStops)
	usedCategories := make(map[string]bool)
	usedAreas := make(map[string]bool)
	for len(remaining) > 0 && len(selected) < plan.targetStops {
		bestIndex, bestScore := 0, math.Inf(-1)
		for index, candidate := range remaining {
			score := candidate.score
			if !usedCategories[candidate.place.Category] {
				score += 8
			}
			area := candidate.place.BusinessArea
			if area == "" {
				area = candidate.place.District
			}
			if area != "" && !usedAreas[area] {
				score += 4
			}
			if score > bestScore {
				bestIndex, bestScore = index, score
			}
		}
		chosen := remaining[bestIndex]
		selected = append(selected, chosen)
		usedCategories[chosen.place.Category] = true
		area := chosen.place.BusinessArea
		if area == "" {
			area = chosen.place.District
		}
		usedAreas[area] = true
		remaining = append(remaining[:bestIndex], remaining[bestIndex+1:]...)
	}
	return selected
}

func orderNearbyStops(candidates []nearbyCandidate, longitude, latitude float64) ([]RouteStopView, float64) {
	remaining := append([]nearbyCandidate(nil), candidates...)
	stops := make([]RouteStopView, 0, len(candidates))
	var total float64
	for len(remaining) > 0 {
		bestIndex, bestDistance := 0, math.Inf(1)
		for index, candidate := range remaining {
			distance := haversineMeters(longitude, latitude, candidate.place.Longitude, candidate.place.Latitude)
			if distance < bestDistance {
				bestIndex, bestDistance = index, distance
			}
		}
		chosen := remaining[bestIndex]
		total += bestDistance
		chosen.stop.Index = len(stops) + 1
		stops = append(stops, chosen.stop)
		longitude, latitude = chosen.place.Longitude, chosen.place.Latitude
		remaining = append(remaining[:bestIndex], remaining[bestIndex+1:]...)
	}
	return stops, total
}

func formatRouteDistance(meters float64) string {
	if meters >= 1000 {
		return fmt.Sprintf("约 %.1f km", meters/1000)
	}
	return fmt.Sprintf("约 %d m", int(math.Round(meters)))
}
