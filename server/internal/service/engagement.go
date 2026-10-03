package service

import (
	"encoding/json"
	"strings"
	"time"

	model "shijibu/internal/model/pgsql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EngagementService struct{ db *gorm.DB }

func NewEngagementService(db *gorm.DB) *EngagementService { return &EngagementService{db: db} }

func (s *EngagementService) FavoritePlace(userID, placeID uint) error {
	if userID == 0 || placeID == 0 || s.db.First(&model.Place{}, placeID).Error != nil {
		return ErrNotFound
	}
	return s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.PlaceFavorite{UserID: userID, PlaceID: placeID}).Error
}

func (s *EngagementService) UnfavoritePlace(userID, placeID uint) error {
	return s.db.Where("user_id = ? AND place_id = ?", userID, placeID).Delete(&model.PlaceFavorite{}).Error
}

func (s *EngagementService) FavoritePlaces(userID uint) ([]model.Place, error) {
	var items []model.Place
	err := s.db.Table("places").Joins("JOIN place_favorites f ON f.place_id = places.id").
		Where("f.user_id = ?", userID).Order("f.created_at DESC").Find(&items).Error
	return items, err
}

func (s *EngagementService) Helpful(userID, recordID uint) error {
	var record model.RestaurantReview
	if err := s.db.Where("id = ? AND status = ?", recordID, "published").First(&record).Error; err != nil {
		return mapNotFound(err)
	}
	if record.UserID == userID {
		return ErrForbidden
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.HelpfulVote{UserID: userID, RecordID: recordID})
		if result.Error != nil || result.RowsAffected == 0 {
			return result.Error
		}
		return tx.Model(&model.RestaurantReview{}).Where("id = ?", recordID).UpdateColumn("helpful_count", gorm.Expr("helpful_count + 1")).Error
	})
}

func (s *EngagementService) Unhelpful(userID, recordID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND record_id = ?", userID, recordID).Delete(&model.HelpfulVote{})
		if result.Error != nil || result.RowsAffected == 0 {
			return result.Error
		}
		return tx.Model(&model.RestaurantReview{}).Where("id = ?", recordID).
			UpdateColumn("helpful_count", gorm.Expr("GREATEST(helpful_count - 1, 0)")).Error
	})
}

func (s *EngagementService) MarkOutdated(userID, recordID uint, reason string) error {
	if len([]rune(reason)) > 500 {
		return ErrInvalidInput
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var record model.RestaurantReview
		if err := tx.Where("id = ? AND status = ?", recordID, "published").First(&record).Error; err != nil {
			return mapNotFound(err)
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.OutdatedSignal{UserID: userID, RecordID: recordID, Reason: strings.TrimSpace(reason), Status: "pending"})
		if result.Error != nil || result.RowsAffected == 0 {
			return result.Error
		}
		return tx.Model(&record).UpdateColumn("outdated_count", gorm.Expr("outdated_count + 1")).Error
	})
}

var reportReasons = map[string]bool{
	"false_experience": true, "defamation": true, "privacy_portrait": true,
	"ad": true, "malicious_competitor": true, "outdated": true, "other": true,
}

func (s *EngagementService) CreateReport(userID uint, targetType string, targetID uint, reason, description string) (*model.ContentReport, error) {
	targetType, reason, description = strings.TrimSpace(targetType), strings.TrimSpace(reason), strings.TrimSpace(description)
	if userID == 0 || targetID == 0 || (targetType != "record" && targetType != "review") || !reportReasons[reason] || len([]rune(description)) > 1000 {
		return nil, ErrInvalidInput
	}
	report := &model.ContentReport{ReporterUserID: userID, TargetType: "restaurant_review", TargetID: targetID, ReasonCode: reason, Description: description, Status: "pending"}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var record model.RestaurantReview
		if err := tx.Where("id = ? AND status = ?", targetID, "published").First(&record).Error; err != nil {
			return mapNotFound(err)
		}
		var recent int64
		if err := tx.Model(&model.ContentReport{}).Where("reporter_user_id = ? AND created_at >= ?", userID, time.Now().Add(-24*time.Hour)).Count(&recent).Error; err != nil {
			return err
		}
		if recent >= 10 {
			return ErrRateLimited
		}
		if err := tx.Create(report).Error; err != nil {
			return err
		}
		return tx.Model(&record).UpdateColumn("report_count", gorm.Expr("report_count + 1")).Error
	})
	return report, err
}

func (s *EngagementService) Reports(userID uint) ([]model.ContentReport, error) {
	var items []model.ContentReport
	err := s.db.Where("reporter_user_id = ?", userID).Order("id DESC").Limit(100).Find(&items).Error
	return items, err
}

func (s *EngagementService) CreateAppeal(userID, recordID uint, text string) (*model.AppealCase, error) {
	text = strings.TrimSpace(text)
	if recordID == 0 || len([]rune(text)) < 10 || len([]rune(text)) > 2000 {
		return nil, ErrInvalidInput
	}
	var record model.RestaurantReview
	if err := s.db.Where("id = ? AND user_id = ?", recordID, userID).First(&record).Error; err != nil {
		return nil, mapNotFound(err)
	}
	if record.PublishStatus != "rejected" && record.PublishStatus != "hidden" && record.PublishStatus != "deleted" {
		return nil, ErrConflict
	}
	item := &model.AppealCase{CaseType: "user_appeal", RecordID: &recordID, ComplainantType: "user", ComplainantUserID: &userID, ClaimText: text, Status: "pending"}
	return item, s.db.Create(item).Error
}

func (s *EngagementService) Appeals(userID uint) ([]model.AppealCase, error) {
	var items []model.AppealCase
	err := s.db.Where("complainant_user_id = ?", userID).Order("id DESC").Limit(100).Find(&items).Error
	return items, err
}

func (s *EngagementService) FavoriteRoute(userID, routeID uint) error {
	if err := s.db.Where("id = ? AND enabled = ?", routeID, true).First(&model.CityRoute{}).Error; err != nil {
		return mapNotFound(err)
	}
	return s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.RouteFavorite{UserID: userID, RouteID: routeID}).Error
}

func (s *EngagementService) UnfavoriteRoute(userID, routeID uint) error {
	return s.db.Where("user_id = ? AND route_id = ?", userID, routeID).Delete(&model.RouteFavorite{}).Error
}

func (s *EngagementService) StartRoute(userID, routeID uint) (*model.RouteJourney, error) {
	if err := s.db.Where("id = ? AND enabled = ?", routeID, true).First(&model.CityRoute{}).Error; err != nil {
		return nil, mapNotFound(err)
	}
	journey := &model.RouteJourney{UserID: userID, RouteID: routeID, Status: "started", CompletedStopIDs: model.JSONDocument("[]"), StartedAt: time.Now(), UpdatedAt: time.Now()}
	return journey, s.db.Create(journey).Error
}

func (s *EngagementService) UpdateJourney(userID, journeyID uint, stopIDs []uint, complete bool) (*model.RouteJourney, error) {
	data, _ := json.Marshal(stopIDs)
	updates := map[string]any{"completed_stop_ids": model.JSONDocument(data), "updated_at": time.Now()}
	if complete {
		now := time.Now()
		updates["status"], updates["completed_at"] = "completed", &now
	}
	result := s.db.Model(&model.RouteJourney{}).Where("id = ? AND user_id = ?", journeyID, userID).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	var journey model.RouteJourney
	return &journey, s.db.Where("id = ? AND user_id = ?", journeyID, userID).First(&journey).Error
}

func (s *EngagementService) MyJourneys(userID uint) ([]model.RouteJourney, error) {
	var items []model.RouteJourney
	err := s.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&items).Error
	return items, err
}
