package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"time"

	"shijibu/internal/model"
	"shijibu/internal/platform/wechat"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PhoneProvider interface {
	GetPhoneNumber(context.Context, string) (wechat.PhoneResult, error)
}

type VerificationService struct {
	db     *gorm.DB
	phones PhoneProvider
	key    []byte
}

func NewVerificationService(db *gorm.DB, phones PhoneProvider, encodedKey string) *VerificationService {
	key, _ := base64.StdEncoding.DecodeString(encodedKey)
	return &VerificationService{db: db, phones: phones, key: key}
}

func (s *VerificationService) Status(userID uint) (*model.PublisherVerification, error) {
	var item model.PublisherVerification
	if err := s.db.Select("user_id", "verification_channel", "status", "verified_at", "created_at", "updated_at").First(&item, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &model.PublisherVerification{UserID: userID, Status: "unverified"}, nil
		}
		return nil, err
	}
	return &item, nil
}

func (s *VerificationService) VerifyPhone(ctx context.Context, userID uint, code string) (*model.PublisherVerification, error) {
	if len(s.key) != 32 || s.phones == nil {
		return nil, ErrUnavailable
	}
	phone, err := s.phones.GetPhoneNumber(ctx, code)
	if err != nil {
		return nil, ErrUnavailable
	}
	encrypted, err := encryptValue(s.key, phone.PhoneNumber)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(phone.CountryCode + ":" + phone.PurePhoneNumber))
	now := time.Now()
	item := &model.PublisherVerification{UserID: userID, PhoneEncrypted: encrypted, PhoneHash: hex.EncodeToString(hash[:]), VerificationChannel: "wechat_phone", Status: "verified", VerifiedAt: now, CreatedAt: now, UpdatedAt: now}
	err = s.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}}, DoUpdates: clause.AssignmentColumns([]string{"phone_encrypted", "phone_hash", "status", "verified_at", "updated_at"})}).Create(item).Error
	if err != nil {
		return nil, err
	}
	item.PhoneEncrypted, item.PhoneHash = "", ""
	return item, nil
}

func encryptValue(key []byte, value string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(value), nil)), nil
}
