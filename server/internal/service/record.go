package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"shijibu/internal/model"
	"shijibu/internal/platform/wechat"

	"gorm.io/gorm"
)

type TextSafetyProvider interface {
	CheckText(context.Context, string, string) (wechat.TextSafetyResult, error)
}
type RecordService struct {
	db     *gorm.DB
	safety TextSafetyProvider
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

func NewRecordService(db *gorm.DB, safety ...TextSafetyProvider) *RecordService {
	s := &RecordService{db: db}
	if len(safety) > 0 {
		s.safety = safety[0]
	}
	return s
}

func (s *RecordService) Create(userID uint, input RecordInput, requestKey string) (*RecordView, error) {
	if input.Visibility == "" {
		input.Visibility = "private"
	}
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
		created = model.VisitRecord{UserID: userID, PlaceID: input.PlaceID, Visibility: input.Visibility, PublishStatus: "draft", RiskLevel: "low"}
		if requestKey != "" {
			created.CreateRequestKey = &requestKey
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		version := versionFromInput(created.ID, userID, 1, input)
		if err := tx.Create(&version).Error; err != nil {
			return err
		}
		if err := replaceVersionTags(tx, version.ID, input.TagCodes); err != nil {
			return err
		}
		created.CurrentVersionID = &version.ID
		return tx.Model(&created).Update("current_version_id", version.ID).Error
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
	if input.Visibility == "" {
		input.Visibility = "private"
	}
	if err := validateRecordInput(input); err != nil {
		return nil, err
	}
	var record model.VisitRecord
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND user_id = ?", recordID, userID).First(&record).Error; err != nil {
			return mapNotFound(err)
		}
		if record.PublishStatus != "draft" && record.PublishStatus != "rejected" {
			return ErrConflict
		}
		placeID, err := resolveRecordPlace(tx, userID, input)
		if err != nil {
			return err
		}
		input.PlaceID = placeID
		var latest int
		if err := tx.Model(&model.VisitRecordVersion{}).Where("record_id = ?", record.ID).Select("COALESCE(MAX(version_no), 0)").Scan(&latest).Error; err != nil {
			return err
		}
		version := versionFromInput(record.ID, userID, latest+1, input)
		if err := tx.Create(&version).Error; err != nil {
			return err
		}
		if err := replaceVersionTags(tx, version.ID, input.TagCodes); err != nil {
			return err
		}
		record.PlaceID = input.PlaceID
		record.Visibility = input.Visibility
		record.PublishStatus = "draft"
		record.CurrentVersionID = &version.ID
		return tx.Model(&record).Updates(map[string]any{"place_id": input.PlaceID, "visibility": input.Visibility, "publish_status": "draft", "current_version_id": version.ID, "submitted_at": nil}).Error
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

func (s *RecordService) SubmitPublic(ctx context.Context, userID, recordID uint) (*RecordView, error) {
	view, err := s.GetOwned(userID, recordID)
	if err != nil {
		return nil, err
	}
	if view.Record.PublishStatus != "draft" && view.Record.PublishStatus != "rejected" {
		return nil, ErrConflict
	}
	if len([]rune(strings.TrimSpace(view.Version.Content))) < 20 {
		return nil, ErrInvalidInput
	}
	if len(view.Tags) == 0 {
		return nil, ErrInvalidInput
	}
	var flag model.FeatureFlag
	if err := s.db.First(&flag, "key = ?", "public_submission_enabled").Error; err == nil && !flag.Enabled {
		return nil, ErrDisabled
	}
	var verification model.PublisherVerification
	if err := s.db.Where("user_id = ? AND status = ?", userID, "verified").First(&verification).Error; err != nil {
		return nil, ErrForbidden
	}
	var agreementCount int64
	if err := s.db.Model(&model.PrivacyAgreement{}).Where("user_id = ? AND version = ? AND agreement_type IN ?", userID, "2.0", []string{"privacy", "user_service", "community"}).Count(&agreementCount).Error; err != nil {
		return nil, err
	}
	if agreementCount != 3 {
		return nil, ErrForbidden
	}
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil || user.AccountStatus != "active" {
		return nil, ErrForbidden
	}
	var trust model.UserTrustProfile
	if err := s.db.First(&trust, "user_id = ?", userID).Error; err != nil {
		return nil, ErrForbidden
	}
	var submittedToday int64
	dayStart := time.Now().Truncate(24 * time.Hour)
	if err := s.db.Model(&model.VisitRecord{}).Where("user_id = ? AND submitted_at >= ?", userID, dayStart).Count(&submittedToday).Error; err != nil {
		return nil, err
	}
	if submittedToday >= int64(trust.DailyPublishLimit) {
		return nil, ErrRateLimited
	}
	riskLabels := recordRiskLabels(view.Version.Content)
	provider, providerResult, rawReference := "local_rules", "passed", ""
	if s.safety != nil {
		result, safetyErr := s.safety.CheckText(ctx, user.OpenID, view.Version.Content)
		if safetyErr != nil {
			return nil, ErrUnavailable
		}
		provider, providerResult, rawReference = "wechat", result.Suggest, strconv.Itoa(result.RiskLabel)
		if result.Suggest != "pass" {
			riskLabels = append(riskLabels, "wechat_"+strconv.Itoa(result.RiskLabel))
		}
	}
	riskJSON, _ := json.Marshal(riskLabels)
	now := time.Now()
	err = s.db.Transaction(func(tx *gorm.DB) error {
		check := model.ContentSafetyCheck{TargetType: "record_version", TargetID: view.Version.ID, Provider: provider, CheckType: "text", Result: providerResult, RiskLabels: model.JSONDocument(riskJSON), RawResponseRef: rawReference, CheckedAt: now}
		if len(riskLabels) > 0 {
			check.Result = "review"
		}
		if err := tx.Create(&check).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.VisitRecordVersion{}).Where("id = ? AND record_id = ?", view.Version.ID, recordID).Update("visibility", "public").Error; err != nil {
			return err
		}
		if err := tx.Model(&model.VisitRecord{}).Where("id = ? AND user_id = ?", recordID, userID).Updates(map[string]any{
			"visibility": "public", "publish_status": "pending_review", "submitted_at": now,
		}).Error; err != nil {
			return err
		}
		priority := len(riskLabels) * 10
		dueAt := now.Add(24 * time.Hour)
		return tx.Create(&model.ModerationTask{TaskType: "record_publish", TargetType: "visit_record", TargetID: recordID, RecordVersionID: &view.Version.ID, Priority: priority, RiskLabels: model.JSONDocument(riskJSON), Status: "pending", DueAt: &dueAt}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetOwned(userID, recordID)
}

func recordRiskLabels(content string) []string {
	content = strings.ToLower(content)
	rules := map[string][]string{
		"serious_accusation": {"中毒", "诈骗", "违法", "违禁品", "黑店"},
		"personal_attack":    {"骗子", "垃圾", "人渣", "去死"},
		"privacy":            {"手机号", "电话", "微信号", "身份证"},
	}
	labels := make([]string, 0)
	for label, words := range rules {
		for _, word := range words {
			if strings.Contains(content, word) {
				labels = append(labels, label)
				break
			}
		}
	}
	return labels
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
		Select("places.id AS place_id, places.name AS place_name, places.city_code, places.business_area, places.longitude, places.latitude, COUNT(visit_records.id) AS record_count, MAX(visit_record_versions.visit_date)::text AS last_visit_at").
		Joins("JOIN places ON places.id = visit_records.place_id").
		Joins("JOIN visit_record_versions ON visit_record_versions.id = visit_records.current_version_id").
		Where("visit_records.user_id = ? AND visit_records.deleted_at IS NULL", userID).
		Group("places.id, places.name, places.city_code, places.business_area, places.longitude, places.latitude").
		Order("MAX(visit_record_versions.visit_date) DESC").Scan(&points).Error
	return points, err
}

func (s *RecordService) loadView(record model.VisitRecord) (*RecordView, error) {
	if record.CurrentVersionID == nil {
		return nil, ErrNotFound
	}
	view := &RecordView{Record: record}
	if err := s.db.First(&view.Version, *record.CurrentVersionID).Error; err != nil {
		return nil, mapNotFound(err)
	}
	if err := s.db.First(&view.Place, record.PlaceID).Error; err != nil {
		return nil, mapNotFound(err)
	}
	if err := s.db.Table("tags").Joins("JOIN visit_record_tag_links l ON l.tag_id = tags.id").Where("l.record_version_id = ?", view.Version.ID).Order("tags.sort_order, tags.id").Find(&view.Tags).Error; err != nil {
		return nil, err
	}
	if err := s.db.Where("record_version_id = ?", view.Version.ID).Order("sort_order, id").Find(&view.Media).Error; err != nil {
		return nil, err
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

// findPlaceByName matches a place by city + case-insensitive name, preferring
// operator-activated rows over the user's own pending submissions.
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

// SubmitPlace stores a user-proposed place with status=pending; it becomes
// searchable only after an operator activates it. Submitting a name that
// already exists in the city returns the existing row without duplicating.
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

func versionFromInput(recordID, userID uint, versionNo int, input RecordInput) model.VisitRecordVersion {
	visibility := input.Visibility
	if visibility == "" {
		visibility = "private"
	}
	return model.VisitRecordVersion{RecordID: recordID, VersionNo: versionNo, EditorUserID: userID, VisitDate: input.VisitDate, ConsumerType: strings.TrimSpace(input.ConsumerType), Conclusion: input.Conclusion, PriceMin: input.PriceMin, PriceMax: input.PriceMax, AverageCost: input.AverageCost, WaitMinutes: input.WaitMinutes, MealPeriod: strings.TrimSpace(input.MealPeriod), Dishes: input.Dishes, Content: strings.TrimSpace(input.Content), Visibility: visibility, ChangeSummary: strings.TrimSpace(input.ChangeSummary)}
}

func replaceVersionTags(tx *gorm.DB, versionID uint, codes []string) error {
	if len(codes) == 0 {
		return nil
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
		return err
	}
	if len(tags) != len(clean) {
		return ErrInvalidInput
	}
	links := make([]model.VisitRecordTagLink, 0, len(tags))
	for _, tag := range tags {
		links = append(links, model.VisitRecordTagLink{RecordVersionID: versionID, TagID: tag.ID})
	}
	return tx.Create(&links).Error
}

func mapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
