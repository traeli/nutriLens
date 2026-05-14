package service

import (
	"log"

	"nutrilens/internal/model"

	"gorm.io/gorm"
)

type ShareService struct {
	db    *gorm.DB
	score *ScoreService
}

func NewShareService(db *gorm.DB, scoreSvc *ScoreService) *ShareService {
	return &ShareService{db: db, score: scoreSvc}
}

// RecordShare logs a share event and awards score.
func (s *ShareService) RecordShare(userID uint, shareType string) error {
	record := model.ShareRecord{
		UserID:    userID,
		ShareType: shareType,
	}
	if err := s.db.Create(&record).Error; err != nil {
		return err
	}

	// Award share points
	return s.score.AddScore(userID, "share", 5, "分享海报")
}

// HandleInvite processes an invitation when a new user registers via a shared QR code.
// inviterID is the user who shared, inviteeID is the new user.
func (s *ShareService) HandleInvite(inviterID, inviteeID uint) error {
	if inviterID == 0 || inviteeID == 0 || inviterID == inviteeID {
		return nil
	}

	// Record invite relation
	invite := model.InviteRelation{
		InviterUserID: inviterID,
		InviteeUserID: inviteeID,
	}
	if err := s.db.Where("inviter_user_id = ? AND invitee_user_id = ?", inviterID, inviteeID).
		FirstOrCreate(&invite).Error; err != nil {
		log.Printf("[Share] failed to create invite relation: inviter=%d, invitee=%d, err=%v", inviterID, inviteeID, err)
		return err
	}

	// Set inviter_id on user
	s.db.Model(&model.User{}).Where("id = ?", inviteeID).Update("inviter_id", inviterID)

	// Award inviter score and create friend relation
	return s.score.OnInvite(inviterID, inviteeID)
}

// HandleInviteeFirstAnalyze awards the inviter when their invitee completes first analysis.
func (s *ShareService) HandleInviteeFirstAnalyze(inviteeID uint) {
	var user model.User
	if err := s.db.Select("inviter_id").First(&user, inviteeID).Error; err != nil || user.InviterID == 0 {
		return
	}

	// Check if already awarded
	var count int64
	s.db.Model(&model.ScoreLog{}).
		Where("user_id = ? AND action = ? AND description LIKE ?", user.InviterID, "invitee_analyze", "%"+string(rune(inviteeID))+"%").
		Count(&count)
	if count > 0 {
		return
	}

	if err := s.score.OnInviteeAnalyze(user.InviterID, inviteeID); err != nil {
		log.Printf("[Share] failed to award invitee first analyze: inviter=%d, invitee=%d, err=%v", user.InviterID, inviteeID, err)
	}
}
