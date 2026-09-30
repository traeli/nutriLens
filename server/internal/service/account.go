package service

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"shijibu/internal/model"
	platformauth "shijibu/internal/platform/auth"

	"gorm.io/gorm"
)

type WeChatSessionProvider interface {
	Code2Session(ctx context.Context, code string) (string, error)
}

const CurrentAgreementVersion = "2.0"

func RequiredAgreementTypes() []string {
	return []string{"privacy", "user_service", "community"}
}

type AccountService struct {
	db         *gorm.DB
	wechat     WeChatSessionProvider
	tokens     *platformauth.TokenManager
	uploadRoot string
}

type LoginResult struct {
	Token      string     `json:"token"`
	UserID     uint       `json:"user_id"`
	HasProfile bool       `json:"has_profile"`
	Profile    model.User `json:"profile"`
}

func NewAccountService(db *gorm.DB, wechat WeChatSessionProvider, tokens *platformauth.TokenManager, uploadRoot ...string) *AccountService {
	root := ""
	if len(uploadRoot) > 0 {
		root = strings.TrimSpace(uploadRoot[0])
	}
	return &AccountService{db: db, wechat: wechat, tokens: tokens, uploadRoot: root}
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

func (s *AccountService) IsAccountActive(ctx context.Context, userID uint) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ? AND account_status = ? AND deleted_at IS NULL", userID, "active").Count(&count).Error
	return count == 1, err
}

// DeleteAccount permanently removes user-owned records and identifying account data.
// Files are removed after the database transaction commits so a filesystem error cannot
// leave the relational data only partially deleted.
func (s *AccountService) DeleteAccount(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidInput
	}
	storedObjects := make([]string, 0)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var activeUsers int64
		if err := tx.Model(&model.User{}).Where("id = ? AND account_status = ? AND deleted_at IS NULL", userID, "active").Count(&activeUsers).Error; err != nil {
			return err
		}
		if activeUsers != 1 {
			return ErrNotFound
		}

		var recordIDs, versionIDs, appealIDs, moderationTaskIDs []uint
		if err := tx.Table("visit_records").Where("user_id = ?", userID).Pluck("id", &recordIDs).Error; err != nil {
			return err
		}
		if len(recordIDs) > 0 {
			if err := tx.Table("visit_record_versions").Where("record_id IN ?", recordIDs).Pluck("id", &versionIDs).Error; err != nil {
				return err
			}
			if err := appendStoredObjects(tx, &storedObjects, "record_media", "object_key", "record_id IN ?", recordIDs); err != nil {
				return err
			}
			if err := appendStoredObjects(tx, &storedObjects, "record_media", "processed_object_key", "record_id IN ?", recordIDs); err != nil {
				return err
			}
			if err := appendStoredObjects(tx, &storedObjects, "record_evidences", "original_object_key", "record_id IN ?", recordIDs); err != nil {
				return err
			}
			if err := appendStoredObjects(tx, &storedObjects, "record_evidences", "masked_object_key", "record_id IN ?", recordIDs); err != nil {
				return err
			}
		}
		if err := appendStoredObjects(tx, &storedObjects, "nutrition_records", "image_object_key", "user_id = ?", userID); err != nil {
			return err
		}

		appeals := tx.Table("appeal_cases").Where("complainant_user_id = ?", userID)
		if len(recordIDs) > 0 {
			appeals = appeals.Or("record_id IN ?", recordIDs)
		}
		if err := appeals.Pluck("id", &appealIDs).Error; err != nil {
			return err
		}
		materials := tx.Table("case_materials").Where("submitted_by_type = ? AND submitted_by_id = ?", "user", userID)
		if len(appealIDs) > 0 {
			materials = materials.Or("case_id IN ?", appealIDs)
		}
		var materialKeys []string
		if err := materials.Pluck("object_key", &materialKeys).Error; err != nil {
			return err
		}
		storedObjects = append(storedObjects, materialKeys...)

		tasks := tx.Table("moderation_tasks").Where("1 = 0")
		if len(versionIDs) > 0 {
			tasks = tasks.Or("record_version_id IN ?", versionIDs)
		}
		if len(recordIDs) > 0 {
			tasks = tasks.Or("target_type = ? AND target_id IN ?", "visit_record", recordIDs)
		}
		if err := tasks.Pluck("id", &moderationTaskIDs).Error; err != nil {
			return err
		}

		if len(moderationTaskIDs) > 0 {
			if err := tx.Exec("DELETE FROM moderation_actions WHERE task_id IN ?", moderationTaskIDs).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM moderation_tasks WHERE id IN ?", moderationTaskIDs).Error; err != nil {
				return err
			}
		}
		if len(recordIDs) > 0 {
			if err := tx.Exec("DELETE FROM moderation_actions WHERE target_type IN ? AND target_id IN ?", []string{"record", "visit_record"}, recordIDs).Error; err != nil {
				return err
			}
		}
		if len(appealIDs) > 0 {
			if err := tx.Exec("DELETE FROM case_materials WHERE case_id IN ?", appealIDs).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM appeal_cases WHERE id IN ?", appealIDs).Error; err != nil {
				return err
			}
		}
		if err := tx.Exec("DELETE FROM case_materials WHERE submitted_by_type = ? AND submitted_by_id = ?", "user", userID).Error; err != nil {
			return err
		}
		if len(versionIDs) > 0 {
			if err := tx.Exec("DELETE FROM content_safety_checks WHERE target_type = ? AND target_id IN ?", "record_version", versionIDs).Error; err != nil {
				return err
			}
		}
		if len(recordIDs) > 0 {
			if err := tx.Exec("DELETE FROM content_reports WHERE target_type IN ? AND target_id IN ?", []string{"record", "visit_record"}, recordIDs).Error; err != nil {
				return err
			}
			for _, table := range []string{"helpful_votes", "outdated_signals", "record_media", "record_evidences"} {
				if err := tx.Exec("DELETE FROM "+table+" WHERE record_id IN ?", recordIDs).Error; err != nil {
					return err
				}
			}
			if err := tx.Exec("UPDATE visit_records SET current_version_id = NULL WHERE id IN ?", recordIDs).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM visit_record_versions WHERE record_id IN ?", recordIDs).Error; err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM visit_records WHERE id IN ?", recordIDs).Error; err != nil {
				return err
			}
		}

		if err := tx.Exec("UPDATE places SET created_by = NULL WHERE created_by = ?", userID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`UPDATE visit_records AS records
			SET helpful_count = GREATEST(0, records.helpful_count - votes.vote_count::integer)
			FROM (SELECT record_id, COUNT(*) AS vote_count FROM helpful_votes WHERE user_id = ? GROUP BY record_id) AS votes
			WHERE records.id = votes.record_id`, userID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`UPDATE visit_records AS records
			SET outdated_count = GREATEST(0, records.outdated_count - signals.signal_count::integer)
			FROM (SELECT record_id, COUNT(*) AS signal_count FROM outdated_signals WHERE user_id = ? GROUP BY record_id) AS signals
			WHERE records.id = signals.record_id`, userID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`UPDATE visit_records AS records
			SET report_count = GREATEST(0, records.report_count - reports.report_count::integer)
			FROM (
				SELECT target_id, COUNT(*) AS report_count FROM content_reports
				WHERE reporter_user_id = ? AND target_type IN ('record', 'visit_record') GROUP BY target_id
			) AS reports
			WHERE records.id = reports.target_id`, userID).Error; err != nil {
			return err
		}
		for _, deletion := range []struct {
			table  string
			column string
		}{
			{"privacy_agreements", "user_id"}, {"user_trust_profiles", "user_id"}, {"contribution_accounts", "user_id"},
			{"user_badges", "user_id"}, {"publisher_verifications", "user_id"}, {"place_favorites", "user_id"},
			{"helpful_votes", "user_id"}, {"outdated_signals", "user_id"}, {"route_favorites", "user_id"},
			{"route_journeys", "user_id"}, {"nutrition_records", "user_id"}, {"content_reports", "reporter_user_id"},
		} {
			if err := tx.Exec("DELETE FROM "+deletion.table+" WHERE "+deletion.column+" = ?", userID).Error; err != nil {
				return err
			}
		}
		return tx.Exec("DELETE FROM nutrilens_users WHERE id = ?", userID).Error
	})
	if err != nil {
		return err
	}
	s.removeStoredObjects(storedObjects)
	return nil
}

func appendStoredObjects(tx *gorm.DB, target *[]string, table, column, condition string, args ...any) error {
	var values []string
	if err := tx.Table(table).Where(condition, args...).Pluck(column, &values).Error; err != nil {
		return err
	}
	*target = append(*target, values...)
	return nil
}

func (s *AccountService) removeStoredObjects(objectKeys []string) {
	if s.uploadRoot == "" {
		return
	}
	root := filepath.Clean(s.uploadRoot)
	seen := make(map[string]struct{}, len(objectKeys))
	for _, objectKey := range objectKeys {
		objectKey = strings.TrimSpace(objectKey)
		if objectKey == "" {
			continue
		}
		target := filepath.Join(root, filepath.FromSlash(objectKey))
		relative, err := filepath.Rel(root, target)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			log.Printf("skip unsafe account file path %q", objectKey)
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			log.Printf("remove deleted account file %q: %v", objectKey, err)
		}
	}
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
	if version != CurrentAgreementVersion || (agreementType != "privacy" && agreementType != "user_service" && agreementType != "community") {
		return ErrInvalidInput
	}
	agreement := model.PrivacyAgreement{
		UserID: userID, AgreementType: agreementType, Version: version,
		IPHash: ipHash, ClientVersion: strings.TrimSpace(clientVersion), AgreedAt: time.Now(),
	}
	return s.db.Where("user_id = ? AND agreement_type = ? AND version = ?", userID, agreementType, version).
		FirstOrCreate(&agreement).Error
}
