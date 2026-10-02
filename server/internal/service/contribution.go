package service

import (
	"errors"

	model "shijibu/internal/model/pgsql"

	"gorm.io/gorm"
)

type ContributionService struct{ db *gorm.DB }

type BadgeView struct {
	model.Badge
	UnlockedAt string `json:"unlocked_at"`
}

type ContributionView struct {
	Account model.ContributionAccount `json:"account"`
	Trust   model.UserTrustProfile    `json:"trust"`
}

func NewContributionService(db *gorm.DB) *ContributionService {
	return &ContributionService{db: db}
}

func (s *ContributionService) Get(userID uint) (*ContributionView, error) {
	var account model.ContributionAccount
	err := s.db.Where("user_id = ?", userID).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		account = model.ContributionAccount{UserID: userID}
	} else if err != nil {
		return nil, err
	}
	var trust model.UserTrustProfile
	err = s.db.Where("user_id = ?", userID).First(&trust).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		trust = model.UserTrustProfile{UserID: userID, TrustLevel: "new", DailyPublishLimit: 1}
	} else if err != nil {
		return nil, err
	}
	return &ContributionView{Account: account, Trust: trust}, nil
}

func (s *ContributionService) Badges(userID uint) ([]BadgeView, error) {
	var rows []struct {
		model.Badge
		UnlockedAt string `gorm:"column:unlocked_at"`
	}
	err := s.db.Table("badges").Select("badges.*, user_badges.unlocked_at::text AS unlocked_at").
		Joins("LEFT JOIN user_badges ON user_badges.badge_id = badges.id AND user_badges.user_id = ?", userID).
		Where("badges.enabled = ?", true).
		Order("user_badges.unlocked_at DESC NULLS LAST, badges.id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]BadgeView, 0, len(rows))
	for _, row := range rows {
		result = append(result, BadgeView{Badge: row.Badge, UnlockedAt: row.UnlockedAt})
	}
	return result, nil
}
