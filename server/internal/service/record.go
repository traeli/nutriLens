package service

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	model "shijibu/internal/model/pgsql"
)

type RecordService struct {
	db *gorm.DB
}

type RecordInput struct {
	PlaceID       uint
	PlaceName     string
	CityCode      string
	VisitDate     time.Time
	ConsumerType  string
	Conclusion    string
	PriceMin      *float64
	PriceMax      *float64
	AverageCost   *float64
	WaitMinutes   *int
	MealPeriod    string
	Dishes        model.JSONDocument
	Content       string
	Visibility    string
	TagCodes      []string
	ChangeSummary string
}

type RecordView struct {
	Record  model.VisitRecord        `json:"record"`
	Version model.VisitRecordVersion `json:"version"`
	Place   model.Place              `json:"place"`
	Tags    []model.Tag              `json:"tags"`
	Media   []model.RecordMedia      `json:"media"`
}

type FootprintSummary struct {
	PlaceCount        int64 `json:"place_count"`
	BusinessAreaCount int64 `json:"business_area_count"`
	MonthRecordCount  int64 `json:"month_record_count"`
	TotalRecordCount  int64 `json:"total_record_count"`
}

type FootprintPoint struct {
	PlaceID      uint    `json:"place_id"`
	PlaceName    string  `json:"place_name"`
	CityCode     string  `json:"city_code"`
	BusinessArea string  `json:"business_area"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	RecordCount  int64   `json:"record_count"`
	LastVisitAt  string  `json:"last_visit_at"`
}

func NewRecordService(db *gorm.DB) *RecordService {
	return &RecordService{db: db}
}

func (s *RecordService) Create(userID uint, input RecordInput, requestKey string) (*RecordView, error) {
	// 到店足迹始终是私人数据；公开状态只属于由足迹创建的餐厅评论。
	input.Visibility = "private"
	if err := validateRecordInput(input); err != nil {
		return nil, err
	}
	requestKey = strings.TrimSpace(requestKey)
	if requestKey != "" {
		var existing model.VisitRecord
		err := s.db.Where("user_id = ? AND create_request_key = ?", userID, requestKey).First(&existing).Error
		if err == nil {
			return s.loadView(existing)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	var created model.VisitRecord
	err := s.db.Transaction(func(tx *gorm.DB) error {
		placeID, err := resolveRecordPlace(tx, userID, input)
		if err != nil {
			return err
		}
		input.PlaceID = placeID
		tagIDs, err := resolveTagIDs(tx, input.TagCodes)
		if err != nil {
			return err
		}
		created = recordFromInput(userID, input, tagIDs)
		if requestKey != "" {
			created.CreateRequestKey = &requestKey
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.loadView(created)
}

func (s *RecordService) GetOwned(userID, recordID uint) (*RecordView, error) {
	var record model.VisitRecord
	if err := s.db.Where("id = ? AND user_id = ?", recordID, userID).First(&record).Error; err != nil {
		return nil, mapNotFound(err)
	}
	return s.loadView(record)
}

func (s *RecordService) Update(userID, recordID uint, input RecordInput) (*RecordView, error) {
	input.Visibility = "private"
	if err := validateRecordInput(input); err != nil {
		return nil, err
	}
	var record model.VisitRecord
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND user_id = ?", recordID, userID).First(&record).Error; err != nil {
			return mapNotFound(err)
		}
		placeID, err := resolveRecordPlace(tx, userID, input)
		if err != nil {
			return err
		}
		input.PlaceID = placeID
		tagIDs, err := resolveTagIDs(tx, input.TagCodes)
		if err != nil {
			return err
		}
		record.PlaceID = input.PlaceID
		record.Visibility = input.Visibility
		record.PublishStatus = "draft"
		record.VersionNo++
		applyRecordInput(&record, input, tagIDs)
		return tx.Save(&record).Error
	})
	if err != nil {
		return nil, err
	}
	return s.loadView(record)
}

func (s *RecordService) Delete(userID, recordID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", recordID, userID).Delete(&model.VisitRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *RecordService) ListMine(userID uint, status, cityCode, cursor string, limit int) (Page[RecordView], error) {
	limit = normalizeLimit(limit)
	db := s.db.Model(&model.VisitRecord{}).Where("visit_records.user_id = ?", userID)
	if status = strings.TrimSpace(status); status != "" {
		db = db.Where("visit_records.publish_status = ?", status)
	}
	if cityCode = strings.TrimSpace(cityCode); cityCode != "" {
		db = db.Joins("JOIN places ON places.id = visit_records.place_id").Where("places.city_code = ?", cityCode)
	}
	if cursorID := parseCursor(cursor); cursorID > 0 {
		db = db.Where("visit_records.id < ?", cursorID)
	}
	var records []model.VisitRecord
	if err := db.Order("visit_records.id DESC").Limit(limit + 1).Find(&records).Error; err != nil {
		return Page[RecordView]{}, err
	}
	hasMore := len(records) > limit
	if hasMore {
		records = records[:limit]
	}
	views := make([]RecordView, 0, len(records))
	for _, record := range records {
		view, err := s.loadView(record)
		if err != nil {
			return Page[RecordView]{}, err
		}
		views = append(views, *view)
	}
	page := Page[RecordView]{Items: views, HasMore: hasMore}
	if hasMore && len(views) > 0 {
		page.NextCursor = strconv.FormatUint(uint64(views[len(views)-1].Record.ID), 10)
	}
	return page, nil
}

func (s *RecordService) FootprintSummary(userID uint) (*FootprintSummary, error) {
	var result FootprintSummary
	base := s.db.Model(&model.VisitRecord{}).Where("user_id = ?", userID)
	if err := base.Count(&result.TotalRecordCount).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.VisitRecord{}).Where("user_id = ?", userID).Distinct("place_id").Count(&result.PlaceCount).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.VisitRecord{}).Joins("JOIN places ON places.id = visit_records.place_id").Where("visit_records.user_id = ? AND places.business_area <> ''", userID).Distinct("places.business_area").Count(&result.BusinessAreaCount).Error; err != nil {
		return nil, err
	}
	monthStart := time.Now().In(time.Local)
	monthStart = time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, monthStart.Location())
	if err := s.db.Model(&model.VisitRecord{}).Where("user_id = ? AND created_at >= ?", userID, monthStart).Count(&result.MonthRecordCount).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *RecordService) FootprintMap(userID uint) ([]FootprintPoint, error) {
	var points []FootprintPoint
	err := s.db.Table("visit_records").
		Select("places.id AS place_id, places.name AS place_name, places.city_code, places.business_area, places.longitude, places.latitude, COUNT(visit_records.id) AS record_count, MAX(visit_records.visit_date)::text AS last_visit_at").
		Joins("JOIN places ON places.id = visit_records.place_id").
		Where("visit_records.user_id = ? AND visit_records.deleted_at IS NULL", userID).
		Group("places.id, places.name, places.city_code, places.business_area, places.longitude, places.latitude").
		Order("MAX(visit_records.visit_date) DESC").Scan(&points).Error
	return points, err
}

func (s *RecordService) loadView(record model.VisitRecord) (*RecordView, error) {
	versionID := record.ID
	record.CurrentVersionID = &versionID
	view := &RecordView{Record: record, Version: versionFromRecord(record), Media: []model.RecordMedia(record.Media)}
	if err := s.db.First(&view.Place, record.PlaceID).Error; err != nil {
		return nil, mapNotFound(err)
	}
	if len(record.TagIDs) > 0 {
		if err := s.db.Where("id IN ?", []uint(record.TagIDs)).Order("sort_order, id").Find(&view.Tags).Error; err != nil {
			return nil, err
		}
	}
	return view, nil
}

func validateRecordInput(input RecordInput) error {
	if input.PlaceID == 0 || input.VisitDate.IsZero() {
		return ErrInvalidInput
	}
	if input.Conclusion != "recommend" && input.Conclusion != "neutral" && input.Conclusion != "caution" && input.Conclusion != "hot" {
		return ErrInvalidInput
	}
	if input.Visibility != "private" && input.Visibility != "public" {
		return ErrInvalidInput
	}
	if len([]rune(input.Content)) > 500 || len(input.TagCodes) > 12 {
		return ErrInvalidInput
	}
	if input.WaitMinutes != nil && *input.WaitMinutes < 0 {
		return ErrInvalidInput
	}
	return nil
}

func resolveRecordPlace(tx *gorm.DB, userID uint, input RecordInput) (uint, error) {
	if input.PlaceID == 0 {
		return 0, ErrInvalidInput
	}
	var place model.Place
	if err := tx.Where("id = ? AND (status = ? OR created_by = ?)", input.PlaceID, "active", userID).First(&place).Error; err != nil {
		return 0, mapNotFound(err)
	}
	return place.ID, nil
}

// findPlaceByName 按城市和不区分大小写的名称匹配地点，优先返回运营已启用的地点，
// 其次返回当前用户待审核的地点。
func findPlaceByName(tx *gorm.DB, userID uint, name, cityCode string) (*model.Place, bool, error) {
	var place model.Place
	err := tx.Where("city_code = ? AND LOWER(name) = LOWER(?) AND (status = ? OR created_by = ?)", cityCode, name, "active", userID).
		Order("CASE WHEN status = 'active' THEN 0 ELSE 1 END, id").First(&place).Error
	if err == nil {
		return &place, true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	return nil, false, err
}

type PlaceSubmissionInput struct {
	Name         string
	CityCode     string
	Category     string
	District     string
	BusinessArea string
	Address      string
	Longitude    float64
	Latitude     float64
	POIProvider  string
}

var placeCategories = map[string]bool{"restaurant": true, "stall": true, "night_market": true, "drink": true}

// SubmitPlace 以待审核状态保存用户提交的地点，运营启用后才进入搜索结果；
// 同一城市存在同名地点时直接返回已有数据，避免重复创建。
func (s *RecordService) SubmitPlace(userID uint, input PlaceSubmissionInput) (*model.Place, error) {
	name := strings.TrimSpace(input.Name)
	cityCode := strings.TrimSpace(input.CityCode)
	address := strings.TrimSpace(input.Address)
	if name == "" || cityCode == "" || address == "" || len([]rune(name)) > 160 || len(cityCode) > 16 || len([]rune(address)) > 300 || !validCoordinate(input.Longitude, input.Latitude) {
		return nil, ErrInvalidInput
	}
	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = "restaurant"
	}
	if !placeCategories[category] {
		return nil, ErrInvalidInput
	}

	var place *model.Place
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var city model.City
		if err := tx.Where("code = ? AND enabled = ?", cityCode, true).First(&city).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvalidInput
			}
			return err
		}
		existing, found, err := findNearbyPlaceByName(tx, userID, name, cityCode, input.Longitude, input.Latitude)
		if err != nil {
			return err
		}
		if found {
			place = existing
			return nil
		}
		created := model.Place{
			Name: name, Category: category, CityCode: cityCode,
			District: strings.TrimSpace(input.District), BusinessArea: strings.TrimSpace(input.BusinessArea),
			Address: address, Longitude: input.Longitude, Latitude: input.Latitude,
			Status: "pending", CreatedBy: &userID,
		}
		if provider := strings.TrimSpace(input.POIProvider); provider != "" {
			created.POIProvider = &provider
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		place = &created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return place, nil
}

func validCoordinate(longitude, latitude float64) bool {
	return longitude >= -180 && longitude <= 180 && latitude >= -90 && latitude <= 90 && (longitude != 0 || latitude != 0)
}

func findNearbyPlaceByName(tx *gorm.DB, userID uint, name, cityCode string, longitude, latitude float64) (*model.Place, bool, error) {
	var places []model.Place
	if err := tx.Where("city_code = ? AND LOWER(name) = LOWER(?) AND (status = ? OR created_by = ?)", cityCode, name, "active", userID).
		Order("CASE WHEN status = 'active' THEN 0 ELSE 1 END, id").Find(&places).Error; err != nil {
		return nil, false, err
	}
	for index := range places {
		if distanceMeters(latitude, longitude, places[index].Latitude, places[index].Longitude) <= 80 {
			return &places[index], true, nil
		}
	}
	return nil, false, nil
}

func distanceMeters(latitudeA, longitudeA, latitudeB, longitudeB float64) float64 {
	const earthRadius = 6371000.0
	toRadians := func(value float64) float64 { return value * math.Pi / 180 }
	latA, latB := toRadians(latitudeA), toRadians(latitudeB)
	deltaLat := toRadians(latitudeB - latitudeA)
	deltaLon := toRadians(longitudeB - longitudeA)
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) + math.Cos(latA)*math.Cos(latB)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func recordFromInput(userID uint, input RecordInput, tagIDs []uint) model.VisitRecord {
	record := model.VisitRecord{
		UserID: userID, PlaceID: input.PlaceID, Visibility: "private", PublishStatus: "draft",
		RiskLevel: "low", VersionNo: 1, Media: model.JSONList[model.RecordMedia]{}, Evidences: model.JSONList[model.RecordEvidence]{},
	}
	applyRecordInput(&record, input, tagIDs)
	return record
}

func applyRecordInput(record *model.VisitRecord, input RecordInput, tagIDs []uint) {
	record.PlaceID = input.PlaceID
	record.VisitDate = input.VisitDate
	record.ConsumerType = strings.TrimSpace(input.ConsumerType)
	record.Conclusion = input.Conclusion
	record.PriceMin = input.PriceMin
	record.PriceMax = input.PriceMax
	record.AverageCost = input.AverageCost
	record.WaitMinutes = input.WaitMinutes
	record.MealPeriod = strings.TrimSpace(input.MealPeriod)
	record.Dishes = input.Dishes
	record.Content = strings.TrimSpace(input.Content)
	record.ChangeSummary = strings.TrimSpace(input.ChangeSummary)
	record.TagIDs = model.JSONList[uint](tagIDs)
}

func versionFromRecord(record model.VisitRecord) model.VisitRecordVersion {
	return model.VisitRecordVersion{
		ID: record.ID, RecordID: record.ID, VersionNo: record.VersionNo, EditorUserID: record.UserID,
		VisitDate: record.VisitDate, ConsumerType: record.ConsumerType, Conclusion: record.Conclusion,
		PriceMin: record.PriceMin, PriceMax: record.PriceMax, AverageCost: record.AverageCost,
		WaitMinutes: record.WaitMinutes, MealPeriod: record.MealPeriod, Dishes: record.Dishes,
		Content: record.Content, Visibility: "private", ChangeSummary: record.ChangeSummary, CreatedAt: record.UpdatedAt,
	}
}

func resolveTagIDs(tx *gorm.DB, codes []string) ([]uint, error) {
	if len(codes) == 0 {
		return []uint{}, nil
	}
	clean := make([]string, 0, len(codes))
	seen := make(map[string]bool)
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code != "" && !seen[code] {
			seen[code] = true
			clean = append(clean, code)
		}
	}
	var tags []model.Tag
	if err := tx.Where("code IN ? AND enabled = ?", clean, true).Find(&tags).Error; err != nil {
		return nil, err
	}
	if len(tags) != len(clean) {
		return nil, ErrInvalidInput
	}
	tagIDs := make([]uint, 0, len(tags))
	for _, tag := range tags {
		tagIDs = append(tagIDs, tag.ID)
	}
	return tagIDs, nil
}

func mapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
