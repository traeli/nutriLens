package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"shijibu/internal/model"
	platformauth "shijibu/internal/platform/auth"

	"gorm.io/gorm"
)

type WeChatSessionProvider interface {
	Code2Session(ctx context.Context, code string) (string, error)
}

type AccountService struct {
	db     *gorm.DB
	wechat WeChatSessionProvider
	tokens *platformauth.TokenManager
}

type LoginResult struct {
	Token      string     `json:"token"`
	UserID     uint       `json:"user_id"`
	HasProfile bool       `json:"has_profile"`
	Profile    model.User `json:"profile"`
}

func NewAccountService(db *gorm.DB, wechat WeChatSessionProvider, tokens *platformauth.TokenManager) *AccountService {
	return &AccountService{db: db, wechat: wechat, tokens: tokens}
}

func (s *AccountService) Login(ctx context.Context, code string) (*LoginResult, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrInvalidInput
	}
	openID, err := s.wechat.Code2Session(ctx, code)
	if err != nil {
		return nil, err
	}
	var user model.User
	err = s.db.WithContext(ctx).Where("open_id = ?", openID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = model.User{OpenID: openID, AccountStatus: "active"}
		if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if user.AccountStatus != "" && user.AccountStatus != "active" {
		return nil, ErrForbidden
	}
	if err := s.ensureAccountState(user.ID); err != nil {
		return nil, err
	}
	token, err := s.tokens.Issue(user.ID)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token, UserID: user.ID, HasProfile: strings.TrimSpace(user.Nickname) != "", Profile: user}, nil
}

func (s *AccountService) ensureAccountState(userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		trust := model.UserTrustProfile{UserID: userID, TrustLevel: "new", DailyPublishLimit: 1, RiskFlags: model.JSONDocument("{}"), UpdatedAt: time.Now()}
		if err := tx.Where("user_id = ?", userID).FirstOrCreate(&trust).Error; err != nil {
			return err
		}
		contribution := model.ContributionAccount{UserID: userID, UpdatedAt: time.Now()}
		return tx.Where("user_id = ?", userID).FirstOrCreate(&contribution).Error
	})
}

func (s *AccountService) Get(userID uint) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *AccountService) Update(userID uint, nickname, avatarURL string) (*model.User, error) {
	nickname = strings.TrimSpace(nickname)
	avatarURL = strings.TrimSpace(avatarURL)
	if len([]rune(nickname)) > 64 || len(avatarURL) > 512 {
		return nil, ErrInvalidInput
	}
	result := s.db.Model(&model.User{}).Where("id = ? AND account_status = ?", userID, "active").Updates(map[string]interface{}{
		"nickname": nickname, "avatar_url": avatarURL, "updated_at": time.Now(),
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return s.Get(userID)
}

func (s *AccountService) AcceptAgreement(userID uint, agreementType, version, ipHash, clientVersion string) error {
	agreementType = strings.TrimSpace(agreementType)
	version = strings.TrimSpace(version)
	if userID == 0 || agreementType == "" || version == "" {
		return ErrInvalidInput
	}
	if agreementType != "privacy" && agreementType != "user_service" && agreementType != "community" {
		return ErrInvalidInput
	}
	agreement := model.PrivacyAgreement{
		UserID: userID, AgreementType: agreementType, Version: version,
		IPHash: ipHash, ClientVersion: strings.TrimSpace(clientVersion), AgreedAt: time.Now(),
	}
	return s.db.Where("user_id = ? AND agreement_type = ? AND version = ?", userID, agreementType, version).
		FirstOrCreate(&agreement).Error
}
