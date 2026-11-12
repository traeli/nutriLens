package service

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"

	model "shijibu/internal/model/pgsql"

	"gorm.io/gorm"
)

type DiscoveryService struct{ db *gorm.DB }

type ExperienceView struct {
	Record           model.RestaurantReview        `json:"record"`
	Version          model.RestaurantReviewVersion `json:"version"`
	Place            model.Place                   `json:"place"`
	Author           AuthorSummary                 `json:"author"`
	Tags             []model.Tag                   `json:"tags"`
	Media            []model.RecordMedia           `json:"media"`
	EvidenceVerified bool                          `json:"evidence_verified"`
}

type AuthorSummary struct {
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

const editorialAuthorName = "饭友记编辑"

type PlaceSearchView struct {
	model.Place
	ExperienceCount int64 `json:"experience_count"`
}

type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

type CityPosterThemeView struct {
	CityCode        string   `json:"city_code"`
	Version         int      `json:"version"`
	Title           string   `json:"title"`
	Subtitle        string   `json:"subtitle"`
	BackgroundImage string   `json:"background_image"`
	AccentColor     string   `json:"accent_color"`
	SecondaryColor  string   `json:"secondary_color"`
	Motifs          []string `json:"motifs"`
}

type RouteGuideView struct {
	ID      uint   `json:"id"`
	Tag     string `json:"tag"`
	Title   string `json:"title"`
	Meta    string `json:"meta"`
	Keyword string `json:"keyword"`
	Image   string `json:"image"`
}

type RouteStopView struct {
	Index            int     `json:"index"`
	PlaceID          uint    `json:"place_id,omitempty"`
	Name             string  `json:"name"`
	Address          string  `json:"address,omitempty"`
	Longitude        float64 `json:"longitude,omitempty"`
	Latitude         float64 `json:"latitude,omitempty"`
	DistanceMeters   int     `json:"distance_meters,omitempty"`
	ExperienceCount  int64   `json:"experience_count,omitempty"`
	RecommendedFood  string  `json:"recommended_food,omitempty"`
	FoodCategoryCode string  `json:"food_category_code,omitempty"`
	ThemeCode        string  `json:"theme_code,omitempty"`
	ThemeIconURL     string  `json:"theme_icon_url,omitempty"`
	IconKey          string  `json:"icon_key,omitempty"`
	Note             string  `json:"note,omitempty"`
}

type ScenarioQueryView struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ScenarioFilterView struct {
	Code      string            `json:"code"`
	Label     string            `json:"label"`
	SortOrder int               `json:"sort_order"`
	Query     ScenarioQueryView `json:"query"`
}

type FeaturedRouteView struct {
	ID       uint            `json:"id"`
	Title    string          `json:"title"`
	Subtitle string          `json:"subtitle"`
	Stops    []RouteStopView `json:"stops"`
	Footnote string          `json:"footnote"`
}

type RouteDetailView struct {
	ID       uint            `json:"id,omitempty"`
	CityCode string          `json:"city_code"`
	Title    string          `json:"title"`
	Tag      string          `json:"tag"`
	Meta     string          `json:"meta"`
	Image    string          `json:"image"`
	Footnote string          `json:"footnote"`
	Stops    []RouteStopView `json:"stops"`
	Source   string          `json:"source,omitempty"`
}

type HomeCardAuthor struct {
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

type HomeCardView struct {
	ID              uint            `json:"id"`
	RecordID        uint            `json:"record_id"`
	PlaceID         uint            `json:"place_id"`
	PlaceName       string          `json:"place_name"`
	Area            string          `json:"area,omitempty"`
	Excerpt         string          `json:"excerpt,omitempty"`
	VisitedAt       string          `json:"visited_at"`
	ImageURL        string          `json:"image_url"`
	Conclusion      string          `json:"conclusion,omitempty"`
	ConclusionLabel string          `json:"conclusion_label,omitempty"`
	Verdict         string          `json:"verdict,omitempty"`
	Author          *HomeCardAuthor `json:"author,omitempty"`
}

type DailyPlaceView struct {
	PlaceID         uint     `json:"place_id"`
	Name            string   `json:"name"`
	Category        string   `json:"category"`
	Area            string   `json:"area,omitempty"`
	Address         string   `json:"address"`
	ImageURL        string   `json:"image_url,omitempty"`
	ExperienceCount int64    `json:"experience_count"`
	RecommendCount  int64    `json:"recommend_count"`
	AverageCost     *float64 `json:"average_cost,omitempty"`
	LatestVisitAt   string   `json:"latest_visit_at"`
}

type PlaceDetailView struct {
	model.Place
	ImageURL        string   `json:"image_url,omitempty"`
	ExperienceCount int64    `json:"experience_count"`
	RecommendCount  int64    `json:"recommend_count"`
	AverageCost     *float64 `json:"average_cost,omitempty"`
}

type HomeStatsView struct {
	ExperienceCount int64 `json:"experience_count"`
	PlaceCount      int64 `json:"place_count"`
}

type HomeSummaryView struct {
	City            model.City           `json:"city"`
	Cities          []model.City         `json:"cities"`
	ScenarioFilters []ScenarioFilterView `json:"scenario_filters"`
	RouteGuides     []RouteGuideView     `json:"route_guides"`
	FeaturedRoute   *FeaturedRouteView   `json:"featured_route"`
	DailyPicks      []DailyPlaceView     `json:"daily_picks"`
	DiscoveryPosts  []HomeCardView       `json:"discovery_posts"`
	Stats           HomeStatsView        `json:"stats"`
}

func NewDiscoveryService(db *gorm.DB) *DiscoveryService { return &DiscoveryService{db: db} }

func (s *DiscoveryService) Cities() ([]model.City, error) {
	var cities []model.City
	err := s.db.Where("enabled = ?", true).
		Order("is_default DESC, sort_order, code").
		Find(&cities).Error
	return cities, err
}

// HomeSummary 按必填城市编码聚合小程序首页需要的城市切换、路线轮播、精选路线、
// 每日推荐、编辑食记和城市概览数据。
func (s *DiscoveryService) HomeSummary(cityCode string) (*HomeSummaryView, error) {
	cityCode = strings.TrimSpace(cityCode)
	if cityCode == "" {
		return nil, ErrInvalidInput
	}

	var city model.City
	if err := s.db.Where("code = ? AND enabled = ?", cityCode, true).First(&city).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	cities, err := s.Cities()
	if err != nil {
		return nil, err
	}

	routes, err := s.cityRoutes(cityCode)
	if err != nil {
		return nil, err
	}

	routeGuides := make([]RouteGuideView, 0, len(routes))
	var featured *FeaturedRouteView
	for _, route := range routes {
		routeGuides = append(routeGuides, RouteGuideView{
			ID: route.ID, Tag: route.Tag, Title: route.Title,
			Meta: route.Meta, Keyword: route.Keyword, Image: route.Image,
		})
		if featured == nil && route.IsFeatured {
			stops, stopErr := s.routeStops(route)
			if stopErr != nil {
				return nil, stopErr
			}
			featured = &FeaturedRouteView{
				ID: route.ID, Title: route.Title, Subtitle: route.Meta,
				Stops: stops, Footnote: route.Footnote,
			}
		}
	}
	scenarioFilters, err := s.scenarioFilters(cityCode)
	if err != nil {
		return nil, err
	}

	page, err := s.ListDiscoveryPosts(cityCode, "", 30, 0)
	if err != nil {
		return nil, err
	}
	views := page.Items
	dailyPicks := make([]DailyPlaceView, 0, 2)
	editorialNotes := make([]HomeCardView, 0, 3)
	seenPlaces := make(map[uint]bool)
	for _, view := range views {
		if len(editorialNotes) < 3 {
			editorialNotes = append(editorialNotes, s.discoveryHomeCard(view))
		}
		if len(dailyPicks) < 2 && !seenPlaces[view.Place.ID] {
			dailyPicks = append(dailyPicks, s.discoveryDailyPlace(view))
			seenPlaces[view.Place.ID] = true
		}
		if len(dailyPicks) == 2 && len(editorialNotes) == 3 {
			break
		}
	}

	stats, err := s.homeStats(cityCode)
	if err != nil {
		return nil, err
	}

	return &HomeSummaryView{
		City: city, Cities: cities, ScenarioFilters: scenarioFilters, RouteGuides: routeGuides, FeaturedRoute: featured,
		DailyPicks: dailyPicks, DiscoveryPosts: editorialNotes, Stats: stats,
	}, nil
}

func (s *DiscoveryService) discoveryHomeCard(view DiscoveryPostView) HomeCardView {
	area := view.Place.BusinessArea
	if area == "" {
		area = view.Place.District
	}
	imageURL := ""
	if len(view.Images) > 0 {
		imageURL = view.Images[0].ImageURL
	}
	publishedAt := view.Post.CreatedAt
	if view.Post.PublishedAt != nil {
		publishedAt = *view.Post.PublishedAt
	}
	return HomeCardView{
		ID: view.Post.ID, RecordID: view.Post.ID, PlaceID: view.Place.ID, PlaceName: view.Place.Name,
		Area: area, Excerpt: truncateRunes(view.Post.Content, 28), VisitedAt: publishedAt.Format("2006-01-02"),
		ImageURL: imageURL, Conclusion: "editorial", ConclusionLabel: "编辑发现", Verdict: "信息整理",
		Author: &HomeCardAuthor{Nickname: view.Post.AuthorName},
	}
}

func (s *DiscoveryService) discoveryDailyPlace(view DiscoveryPostView) DailyPlaceView {
	area := view.Place.BusinessArea
	if area == "" {
		area = view.Place.District
	}
	publishedAt := view.Post.CreatedAt
	if view.Post.PublishedAt != nil {
		publishedAt = *view.Post.PublishedAt
	}
	pick := DailyPlaceView{PlaceID: view.Place.ID, Name: view.Place.Name, Category: view.Place.Category, Area: area,
		Address: view.Place.Address, ExperienceCount: 1, AverageCost: view.Place.AverageCost, LatestVisitAt: publishedAt.Format("2006-01-02")}
	if len(view.Images) > 0 {
		pick.ImageURL = view.Images[0].ImageURL
	}
	return pick
}

func (s *DiscoveryService) scenarioFilters(cityCode string) ([]ScenarioFilterView, error) {
	var rows []model.DiscoveryScenario
	err := s.db.Where("enabled = ? AND city_code IN ?", true, []string{"", cityCode}).
		Order("sort_order, id").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return mergeScenarioFilters(rows, cityCode), nil
}

func mergeScenarioFilters(rows []model.DiscoveryScenario, cityCode string) []ScenarioFilterView {
	selected := make(map[string]model.DiscoveryScenario, len(rows))
	for _, row := range rows {
		current, exists := selected[row.Code]
		if !exists || (row.CityCode == cityCode && current.CityCode != cityCode) {
			selected[row.Code] = row
		}
	}
	ordered := make([]model.DiscoveryScenario, 0, len(selected))
	for _, row := range selected {
		ordered = append(ordered, row)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].SortOrder == ordered[j].SortOrder {
			return ordered[i].ID < ordered[j].ID
		}
		return ordered[i].SortOrder < ordered[j].SortOrder
	})
	views := make([]ScenarioFilterView, 0, len(ordered))
	for _, row := range ordered {
		views = append(views, ScenarioFilterView{
			Code: row.Code, Label: row.Label, SortOrder: row.SortOrder,
			Query: ScenarioQueryView{Type: row.ConditionType, Value: row.ConditionValue},
		})
	}
	return views
}

func (s *DiscoveryService) cityRoutes(cityCode string) ([]model.CityRoute, error) {
	var routes []model.CityRoute
	err := s.db.Where("city_code = ? AND enabled = ?", cityCode, true).
		Order("sort_order, id").Find(&routes).Error
	return routes, err
}

func (s *DiscoveryService) GetRoute(id uint) (*RouteDetailView, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}
	var route model.CityRoute
	if err := s.db.Table("city_routes").
		Select("city_routes.*").
		Joins("JOIN cities ON cities.code = city_routes.city_code AND cities.enabled = TRUE").
		Where("city_routes.id = ? AND city_routes.enabled = TRUE", id).
		First(&route).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	stops, err := s.routeStops(route)
	if err != nil {
		return nil, err
	}
	return &RouteDetailView{
		ID: route.ID, CityCode: route.CityCode, Title: route.Title, Tag: route.Tag,
		Meta: route.Meta, Image: route.Image, Footnote: route.Footnote, Stops: stops,
	}, nil
}

func (s *DiscoveryService) routeStops(route model.CityRoute) ([]RouteStopView, error) {
	var stops []RouteStopView
	err := s.db.Table("city_route_stops AS route_stops").
		Select("route_stops.sort_order AS index, route_stops.place_id, places.name, places.address, places.longitude, places.latitude, places.theme_code, route_stops.recommended_food, route_stops.food_category_code, route_stops.icon_key, route_stops.note").
		Joins("JOIN places ON places.id = route_stops.place_id AND places.status = 'active'").
		Where("route_stops.route_id = ?", route.ID).
		Order("route_stops.sort_order, route_stops.id").
		Scan(&stops).Error
	if err != nil {
		return nil, err
	}
	if len(stops) == 0 {
		stops = parseRouteStops(route.Stops)
	}
	for index := range stops {
		stops[index].ThemeCode = normalizePlaceThemeCode(stops[index].ThemeCode, "")
		stops[index].IconKey = themeIconKey(stops[index].ThemeCode)
	}
	themeIcons, err := s.placeThemeIconURLs()
	if err != nil {
		return nil, err
	}
	applyThemeIconURLs(stops, themeIcons)
	return stops, nil
}

func nearbyIconKey(category string) string {
	if category == "drink" {
		return "coffee"
	}
	return "bowl"
}

var placeThemeCodes = map[string]bool{
	"dumpling_bun": true, "noodles": true, "rice_meal": true, "local_dishes": true,
	"hotpot": true, "bbq_skewers": true, "street_snack": true, "breakfast": true,
	"coffee_tea": true, "dessert": true, "bakery": true, "seafood": true,
	"vegetarian": true, "bar_drinks": true, "general": true,
}

func normalizePlaceThemeCode(themeCode, category string) string {
	themeCode = strings.TrimSpace(themeCode)
	if placeThemeCodes[themeCode] {
		return themeCode
	}
	switch strings.TrimSpace(category) {
	case "drink":
		return "coffee_tea"
	case "stall", "night_market":
		return "street_snack"
	default:
		return "general"
	}
}

// themeIconKey 返回稳定语义键；具体图形和颜色由小程序控制。
func themeIconKey(themeCode string) string {
	return normalizePlaceThemeCode(themeCode, "")
}

func (s *DiscoveryService) placeThemeIconURLs() (map[string]string, error) {
	var themes []model.PlaceTheme
	if err := s.db.Where("enabled = ?", true).Find(&themes).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string, len(themes))
	for _, theme := range themes {
		result[theme.Code] = strings.TrimSpace(theme.IconURL)
	}
	return result, nil
}

func applyThemeIconURLs(stops []RouteStopView, iconURLs map[string]string) {
	for index := range stops {
		stops[index].ThemeIconURL = iconURLs[stops[index].ThemeCode]
	}
}

func placeIDs(places []model.Place) []uint {
	ids := make([]uint, 0, len(places))
	for _, place := range places {
		ids = append(ids, place.ID)
	}
	return ids
}

func validPublicCoordinate(longitude, latitude float64) bool {
	return longitude >= -180 && longitude <= 180 && latitude >= -90 && latitude <= 90 && (longitude != 0 || latitude != 0)
}

func haversineMeters(longitude1, latitude1, longitude2, latitude2 float64) float64 {
	const earthRadius = 6371000.0
	toRadians := func(value float64) float64 { return value * math.Pi / 180 }
	lat1, lat2 := toRadians(latitude1), toRadians(latitude2)
	deltaLat := toRadians(latitude2 - latitude1)
	deltaLng := toRadians(longitude2 - longitude1)
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func parseRouteStops(raw model.JSONDocument) []RouteStopView {
	var names []string
	if err := json.Unmarshal(raw, &names); err != nil {
		return []RouteStopView{}
	}
	stops := make([]RouteStopView, 0, len(names))
	for _, name := range names {
		if name = strings.TrimSpace(name); name == "" {
			continue
		}
		stops = append(stops, RouteStopView{Index: len(stops) + 1, Name: name, IconKey: "bowl"})
	}
	return stops
}

var homeConclusionLabels = map[string][2]string{
	"recommend": {"推荐", "值得一试"},
	"neutral":   {"一般", "体验中性"},
	"caution":   {"谨慎", "留意细节"},
}

func (s *DiscoveryService) homeCard(view ExperienceView, withAuthor bool) HomeCardView {
	labels, ok := homeConclusionLabels[view.Version.Conclusion]
	if !ok {
		labels = [2]string{"编辑食记", "编辑推荐"}
	}
	area := view.Place.BusinessArea
	if area == "" {
		area = view.Place.District
	}
	imageURL := ""
	if len(view.Media) > 0 {
		imageURL = view.Media[0].PublicURL
	}
	card := HomeCardView{
		ID: view.Record.ID, RecordID: view.Record.ID, PlaceID: view.Place.ID, PlaceName: view.Place.Name, Area: area,
		Excerpt:   truncateRunes(view.Version.Content, 28),
		VisitedAt: view.Version.VisitDate.Format("2006-01-02"),
		ImageURL:  imageURL, Conclusion: view.Version.Conclusion,
		ConclusionLabel: labels[0], Verdict: labels[1],
	}
	if withAuthor {
		card.Author = &HomeCardAuthor{Nickname: view.Author.Nickname, AvatarURL: view.Author.AvatarURL}
	}
	return card
}

func (s *DiscoveryService) dailyPlace(view ExperienceView) (DailyPlaceView, error) {
	area := view.Place.BusinessArea
	if area == "" {
		area = view.Place.District
	}
	pick := DailyPlaceView{
		PlaceID: view.Place.ID, Name: view.Place.Name, Category: view.Place.Category,
		Area: area, Address: view.Place.Address, LatestVisitAt: view.Version.VisitDate.Format("2006-01-02"),
	}
	if len(view.Media) > 0 {
		pick.ImageURL = view.Media[0].PublicURL
	}
	var stats struct {
		ExperienceCount int64
		RecommendCount  int64
		AverageCost     *float64
	}
	err := s.db.Table("restaurant_reviews AS reviews").
		Select("COUNT(reviews.id) AS experience_count, COUNT(reviews.id) FILTER (WHERE versions.conclusion = 'recommend') AS recommend_count, AVG(versions.average_cost) AS average_cost").
		Joins("JOIN restaurant_review_versions AS versions ON versions.id = reviews.current_version_id").
		Where("reviews.place_id = ? AND reviews.status = ? AND reviews.deleted_at IS NULL", view.Place.ID, "published").
		Scan(&stats).Error
	if err != nil {
		return DailyPlaceView{}, err
	}
	pick.ExperienceCount = stats.ExperienceCount
	pick.RecommendCount = stats.RecommendCount
	pick.AverageCost = stats.AverageCost
	return pick, nil
}

func truncateRunes(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit]) + "…"
}

func (s *DiscoveryService) homeStats(cityCode string) (HomeStatsView, error) {
	var stats HomeStatsView
	if err := s.db.Model(&model.DiscoveryPost{}).
		Joins("JOIN places ON places.id = discovery_posts.place_id").
		Where("places.city_code = ? AND discovery_posts.status = ?", cityCode, "published").
		Count(&stats.ExperienceCount).Error; err != nil {
		return stats, err
	}
	err := s.db.Model(&model.Place{}).
		Where("city_code = ? AND status = ?", cityCode, "active").
		Count(&stats.PlaceCount).Error
	return stats, err
}

func (s *DiscoveryService) CityPosterTheme(cityCode string) (*CityPosterThemeView, error) {
	cityCode = strings.TrimSpace(cityCode)
	if cityCode == "" {
		return nil, ErrInvalidInput
	}

	var row model.CityPosterTheme
	if err := s.db.Table("city_poster_themes").
		Select("city_poster_themes.*").
		Joins("JOIN cities ON cities.code = city_poster_themes.city_code AND cities.enabled = TRUE").
		Where("city_poster_themes.city_code = ?", cityCode).
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var motifs []string
	if err := json.Unmarshal(row.Motifs, &motifs); err != nil {
		return nil, err
	}
	return &CityPosterThemeView{
		CityCode: row.CityCode, Version: row.Version, Title: row.Title, Subtitle: row.Subtitle,
		BackgroundImage: row.BackgroundImage, AccentColor: row.AccentColor,
		SecondaryColor: row.SecondaryColor, Motifs: motifs,
	}, nil
}

func (s *DiscoveryService) SearchPlaces(cityCode, query, cursor string, limit int, recommended bool) (Page[PlaceSearchView], error) {
	limit = normalizeLimit(limit)
	db := s.db.Table("places").Select("places.*, COUNT(posts.id) AS experience_count").
		Joins("LEFT JOIN discovery_posts AS posts ON posts.place_id = places.id AND posts.status = 'published'").
		Where("places.status = ?", "active")
	if cityCode = strings.TrimSpace(cityCode); cityCode != "" {
		db = db.Where("places.city_code = ?", cityCode)
	}
	if query = strings.TrimSpace(query); query != "" {
		like := "%" + query + "%"
		db = db.Where("places.name ILIKE ? OR places.business_area ILIKE ? OR places.district ILIKE ?", like, like, like)
	}
	if cursorID := parseCursor(cursor); cursorID > 0 {
		db = db.Where("places.id < ?", cursorID)
	}
	var rows []PlaceSearchView
	order := "places.id DESC"
	if recommended {
		order = "COUNT(posts.id) DESC, places.id DESC"
		db = db.Having("COUNT(posts.id) > 0")
	}
	if err := db.Group("places.id").Order(order).Limit(limit + 1).Scan(&rows).Error; err != nil {
		return Page[PlaceSearchView]{}, err
	}
	return makePage(rows, limit, func(item PlaceSearchView) uint { return item.ID }), nil
}

func (s *DiscoveryService) GetPlace(id uint) (*PlaceDetailView, error) {
	var place model.Place
	if err := s.db.Where("id = ? AND status = ?", id, "active").First(&place).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	detail := &PlaceDetailView{Place: place}
	var postCount int64
	if err := s.db.Model(&model.DiscoveryPost{}).
		Where("place_id = ? AND status = ?", id, "published").Count(&postCount).Error; err != nil {
		return nil, err
	}
	detail.ExperienceCount = postCount
	detail.AverageCost = place.AverageCost
	var image model.DiscoveryPostImage
	if err := s.db.Table("discovery_post_images AS images").
		Joins("JOIN discovery_posts AS posts ON posts.id = images.post_id AND posts.status = 'published'").
		Where("posts.place_id = ?", id).Order("posts.published_at DESC, images.sort_order, images.id").First(&image).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if image.ID > 0 {
		detail.ImageURL = image.ImageURL
	}
	return detail, nil
}

func (s *DiscoveryService) ListExperiences(cityCode, cursor string, limit int) (Page[ExperienceView], error) {
	return s.listExperiences(cityCode, 0, cursor, limit)
}

func (s *DiscoveryService) ListPlaceExperiences(placeID uint, cursor string, limit int) (Page[ExperienceView], error) {
	if placeID == 0 {
		return Page[ExperienceView]{}, ErrInvalidInput
	}
	return s.listExperiences("", placeID, cursor, limit)
}

func (s *DiscoveryService) listExperiences(cityCode string, placeID uint, cursor string, limit int) (Page[ExperienceView], error) {
	limit = normalizeLimit(limit)
	db := s.db.Model(&model.RestaurantReview{}).
		Where("restaurant_reviews.status = ?", "published")
	if placeID > 0 {
		db = db.Where("restaurant_reviews.place_id = ?", placeID)
	}
	if cityCode = strings.TrimSpace(cityCode); cityCode != "" {
		db = db.Joins("JOIN places ON places.id = restaurant_reviews.place_id").Where("places.city_code = ?", cityCode)
	}
	if cursorID := parseCursor(cursor); cursorID > 0 {
		db = db.Where("restaurant_reviews.id < ?", cursorID)
	}
	var records []model.RestaurantReview
	if err := db.Order("restaurant_reviews.id DESC").Limit(limit + 1).Find(&records).Error; err != nil {
		return Page[ExperienceView]{}, err
	}
	hasMore := len(records) > limit
	if hasMore {
		records = records[:limit]
	}
	views, err := s.loadExperienceViews(records)
	if err != nil {
		return Page[ExperienceView]{}, err
	}
	page := Page[ExperienceView]{Items: views}
	page.HasMore = hasMore
	if hasMore && len(records) > 0 {
		page.NextCursor = strconv.FormatUint(uint64(records[len(records)-1].ID), 10)
	}
	return page, nil
}

func (s *DiscoveryService) GetExperience(id uint) (*ExperienceView, error) {
	var record model.RestaurantReview
	if err := s.db.Where("id = ? AND status = ?", id, "published").First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	views, err := s.loadExperienceViews([]model.RestaurantReview{record})
	if err != nil {
		return nil, err
	}
	return &views[0], nil
}

func (s *DiscoveryService) Tags() ([]model.Tag, error) {
	var tags []model.Tag
	err := s.db.Where("enabled = ?", true).Order("sort_order, id").Find(&tags).Error
	return tags, err
}

func (s *DiscoveryService) loadExperienceViews(records []model.RestaurantReview) ([]ExperienceView, error) {
	views := make([]ExperienceView, 0, len(records))
	for _, record := range records {
		view := ExperienceView{Record: record}
		if record.CurrentVersionID == nil {
			continue
		}
		if err := s.db.First(&view.Version, *record.CurrentVersionID).Error; err != nil {
			return nil, err
		}
		if err := s.db.First(&view.Place, record.PlaceID).Error; err != nil {
			return nil, err
		}
		view.Author = AuthorSummary{Nickname: editorialAuthorName}
		for _, media := range record.Media {
			if media.SafetyStatus == "passed" && media.DesensitizeStatus == "completed" {
				view.Media = append(view.Media, media)
			}
		}
		if err := s.db.Table("tags").Joins("JOIN restaurant_review_tag_links links ON links.tag_id = tags.id").Where("links.review_version_id = ?", view.Version.ID).Order("tags.sort_order").Find(&view.Tags).Error; err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, nil
}

func normalizeLimit(value int) int {
	if value <= 0 {
		return 20
	}
	if value > 50 {
		return 50
	}
	return value
}

func parseCursor(value string) uint {
	id, _ := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return uint(id)
}

func makePage[T any](rows []T, limit int, id func(T) uint) Page[T] {
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	page := Page[T]{Items: rows}
	page.HasMore = hasMore
	if hasMore && len(rows) > 0 {
		page.NextCursor = strconv.FormatUint(uint64(id(rows[len(rows)-1])), 10)
	}
	return page
}
