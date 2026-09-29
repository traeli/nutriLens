package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"shijibu/internal/model"

	"gorm.io/gorm"
)

type DiscoveryService struct{ db *gorm.DB }

type ExperienceView struct {
	Record           model.VisitRecord        `json:"record"`
	Version          model.VisitRecordVersion `json:"version"`
	Place            model.Place              `json:"place"`
	Author           AuthorSummary            `json:"author"`
	Tags             []model.Tag              `json:"tags"`
	Media            []model.RecordMedia      `json:"media"`
	EvidenceVerified bool                     `json:"evidence_verified"`
}

type AuthorSummary struct {
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

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
	Index           int     `json:"index"`
	PlaceID         uint    `json:"place_id,omitempty"`
	Name            string  `json:"name"`
	Address         string  `json:"address,omitempty"`
	Longitude       float64 `json:"longitude,omitempty"`
	Latitude        float64 `json:"latitude,omitempty"`
	DistanceMeters  int     `json:"distance_meters,omitempty"`
	ExperienceCount int64   `json:"experience_count,omitempty"`
	Note            string  `json:"note,omitempty"`
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
	City             model.City         `json:"city"`
	Cities           []model.City       `json:"cities"`
	RouteGuides      []RouteGuideView   `json:"route_guides"`
	FeaturedRoute    *FeaturedRouteView `json:"featured_route"`
	DailyPicks       []DailyPlaceView   `json:"daily_picks"`
	CommunityRecords []HomeCardView     `json:"community_records"`
	Stats            HomeStatsView      `json:"stats"`
}

func NewDiscoveryService(db *gorm.DB) *DiscoveryService { return &DiscoveryService{db: db} }

func (s *DiscoveryService) Cities() ([]model.City, error) {
	var cities []model.City
	err := s.db.Where("enabled = ?", true).
		Order("is_default DESC, sort_order, code").
		Find(&cities).Error
	return cities, err
}

// HomeSummary aggregates every data block the mini program home page renders:
// city switcher, route swiper, featured route card, daily picks, community
// records and lightweight city stats, keyed by the required city code.
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
			featured = &FeaturedRouteView{
				ID: route.ID, Title: route.Title, Subtitle: route.Meta,
				Stops: parseRouteStops(route.Stops), Footnote: route.Footnote,
			}
		}
	}

	page, err := s.listExperiences(cityCode, 0, "", 30)
	if err != nil {
		return nil, err
	}
	views := page.Items
	dailyPicks := make([]DailyPlaceView, 0, 2)
	communityRecords := make([]HomeCardView, 0, 3)
	seenPlaces := make(map[uint]bool)
	for _, view := range views {
		if len(communityRecords) < 3 {
			communityRecords = append(communityRecords, s.homeCard(view, true))
		}
		if len(dailyPicks) < 2 && !seenPlaces[view.Place.ID] {
			pick, pickErr := s.dailyPlace(view)
			if pickErr != nil {
				return nil, pickErr
			}
			dailyPicks = append(dailyPicks, pick)
			seenPlaces[view.Place.ID] = true
		}
		if len(dailyPicks) == 2 && len(communityRecords) == 3 {
			break
		}
	}

	stats, err := s.homeStats(cityCode)
	if err != nil {
		return nil, err
	}

	return &HomeSummaryView{
		City: city, Cities: cities, RouteGuides: routeGuides, FeaturedRoute: featured,
		DailyPicks: dailyPicks, CommunityRecords: communityRecords, Stats: stats,
	}, nil
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

	var stops []RouteStopView
	err := s.db.Table("city_route_stops AS route_stops").
		Select("route_stops.sort_order AS index, route_stops.place_id, places.name, places.address, places.longitude, places.latitude, route_stops.note").
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
	return &RouteDetailView{
		ID: route.ID, CityCode: route.CityCode, Title: route.Title, Tag: route.Tag,
		Meta: route.Meta, Image: route.Image, Footnote: route.Footnote, Stops: stops,
	}, nil
}

func (s *DiscoveryService) NearbyRoute(cityCode string, longitude, latitude float64, limit int) (*RouteDetailView, error) {
	cityCode = strings.TrimSpace(cityCode)
	if cityCode == "" || !validPublicCoordinate(longitude, latitude) {
		return nil, ErrInvalidInput
	}
	if limit < 3 {
		limit = 4
	}
	if limit > 4 {
		limit = 4
	}

	var city model.City
	if err := s.db.Where("code = ? AND enabled = ?", cityCode, true).First(&city).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var places []model.Place
	if err := s.db.Where("city_code = ? AND status = ? AND longitude <> 0 AND latitude <> 0", cityCode, "active").Find(&places).Error; err != nil {
		return nil, err
	}
	if len(places) == 0 {
		return &RouteDetailView{CityCode: cityCode, Title: "附近吃喝路线", Meta: "附近暂时没有可用地点", Stops: []RouteStopView{}, Source: "nearby"}, nil
	}
	type placeExperienceCount struct {
		PlaceID uint
		Count   int64
	}
	var countRows []placeExperienceCount
	if err := s.db.Model(&model.VisitRecord{}).
		Select("place_id, COUNT(*) AS count").
		Where("publish_status = ? AND visibility = ? AND deleted_at IS NULL", "published", "public").
		Where("place_id IN ?", placeIDs(places)).
		Group("place_id").Scan(&countRows).Error; err != nil {
		return nil, err
	}
	experienceCounts := make(map[uint]int64, len(countRows))
	for _, row := range countRows {
		experienceCounts[row.PlaceID] = row.Count
	}
	stops := make([]RouteStopView, 0, len(places))
	for _, place := range places {
		stops = append(stops, RouteStopView{
			PlaceID: place.ID, Name: place.Name, Address: place.Address,
			Longitude: place.Longitude, Latitude: place.Latitude,
			DistanceMeters:  int(math.Round(haversineMeters(longitude, latitude, place.Longitude, place.Latitude))),
			ExperienceCount: experienceCounts[place.ID],
		})
	}
	reviewed := make([]RouteStopView, 0, len(stops))
	for _, stop := range stops {
		if stop.ExperienceCount > 0 {
			reviewed = append(reviewed, stop)
		}
	}
	if len(reviewed) >= 3 {
		stops = reviewed
	}
	sort.SliceStable(stops, func(i, j int) bool { return stops[i].DistanceMeters < stops[j].DistanceMeters })
	if len(stops) > limit {
		stops = stops[:limit]
	}
	for index := range stops {
		stops[index].Index = index + 1
	}
	return &RouteDetailView{
		CityCode: cityCode, Title: fmt.Sprintf("从你附近开始的 %d 站吃喝路线", len(stops)),
		Tag: "附近路线", Meta: "按直线距离由近到远排列", Footnote: "实际交通方式与路况请以地图 App 为准",
		Stops: stops, Source: "nearby",
	}, nil
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
		stops = append(stops, RouteStopView{Index: len(stops) + 1, Name: name})
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
		labels = [2]string{"到店", "真实记录"}
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
	err := s.db.Table("visit_records").
		Select("COUNT(visit_records.id) AS experience_count, COUNT(visit_records.id) FILTER (WHERE versions.conclusion = 'recommend') AS recommend_count, AVG(versions.average_cost) AS average_cost").
		Joins("JOIN visit_record_versions AS versions ON versions.id = visit_records.current_version_id").
		Where("visit_records.place_id = ? AND visit_records.publish_status = ? AND visit_records.visibility = ? AND visit_records.deleted_at IS NULL", view.Place.ID, "published", "public").
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
	if err := s.db.Model(&model.VisitRecord{}).
		Joins("JOIN places ON places.id = visit_records.place_id").
		Where("places.city_code = ? AND visit_records.publish_status = ? AND visit_records.visibility = ? AND visit_records.deleted_at IS NULL", cityCode, "published", "public").
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
	db := s.db.Table("places").Select("places.*, COUNT(visit_records.id) AS experience_count").
		Joins("LEFT JOIN visit_records ON visit_records.place_id = places.id AND visit_records.publish_status = 'published' AND visit_records.visibility = 'public' AND visit_records.deleted_at IS NULL").
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
		order = "COUNT(visit_records.id) DESC, places.id DESC"
		db = db.Having("COUNT(visit_records.id) > 0")
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
	var stats struct {
		ExperienceCount int64
		RecommendCount  int64
		AverageCost     *float64
	}
	if err := s.db.Table("visit_records").
		Select("COUNT(visit_records.id) AS experience_count, COUNT(visit_records.id) FILTER (WHERE versions.conclusion = 'recommend') AS recommend_count, AVG(versions.average_cost) AS average_cost").
		Joins("JOIN visit_record_versions AS versions ON versions.id = visit_records.current_version_id").
		Where("visit_records.place_id = ? AND visit_records.publish_status = ? AND visit_records.visibility = ? AND visit_records.deleted_at IS NULL", id, "published", "public").
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	detail.ExperienceCount = stats.ExperienceCount
	detail.RecommendCount = stats.RecommendCount
	detail.AverageCost = stats.AverageCost
	var media model.RecordMedia
	if err := s.db.Table("record_media AS media").
		Select("media.*").
		Joins("JOIN visit_records AS records ON records.id = media.record_id").
		Where("records.place_id = ? AND records.publish_status = ? AND records.visibility = ? AND records.deleted_at IS NULL AND media.safety_status = ? AND media.desensitize_status = ?", id, "published", "public", "passed", "completed").
		Order("records.id DESC, media.sort_order, media.id").First(&media).Error; err == nil {
		detail.ImageURL = media.PublicURL
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
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
	db := s.db.Model(&model.VisitRecord{}).
		Where("visit_records.publish_status = ? AND visit_records.visibility = ?", "published", "public")
	if placeID > 0 {
		db = db.Where("visit_records.place_id = ?", placeID)
	}
	if cityCode = strings.TrimSpace(cityCode); cityCode != "" {
		db = db.Joins("JOIN places ON places.id = visit_records.place_id").Where("places.city_code = ?", cityCode)
	}
	if cursorID := parseCursor(cursor); cursorID > 0 {
		db = db.Where("visit_records.id < ?", cursorID)
	}
	var records []model.VisitRecord
	if err := db.Order("visit_records.id DESC").Limit(limit + 1).Find(&records).Error; err != nil {
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
	var record model.VisitRecord
	if err := s.db.Where("id = ? AND publish_status = ? AND visibility = ?", id, "published", "public").First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	views, err := s.loadExperienceViews([]model.VisitRecord{record})
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

func (s *DiscoveryService) loadExperienceViews(records []model.VisitRecord) ([]ExperienceView, error) {
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
		var user model.User
		if err := s.db.Select("nickname", "avatar_url").First(&user, record.UserID).Error; err == nil {
			view.Author = AuthorSummary{Nickname: user.Nickname, AvatarURL: user.AvatarURL}
		}
		if err := s.db.Where("record_version_id = ? AND safety_status = ? AND desensitize_status = ?", view.Version.ID, "passed", "completed").Order("sort_order, id").Find(&view.Media).Error; err != nil {
			return nil, err
		}
		if err := s.db.Table("tags").Joins("JOIN visit_record_tag_links l ON l.tag_id = tags.id").Where("l.record_version_id = ?", view.Version.ID).Order("tags.sort_order").Find(&view.Tags).Error; err != nil {
			return nil, err
		}
		var verifiedCount int64
		if err := s.db.Model(&model.RecordEvidence{}).Where("record_id = ? AND verify_status = ?", record.ID, "verified").Count(&verifiedCount).Error; err != nil {
			return nil, err
		}
		view.EvidenceVerified = verifiedCount > 0
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
