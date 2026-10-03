package service

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	model "shijibu/internal/model/pgsql"
	platformauth "shijibu/internal/platform/auth"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WeChatSessionProvider interface {
	Code2Session(ctx context.Context, code string) (string, error)
}

const CurrentAgreementVersion = "2.0"

func RequiredAgreementTypes() []string {
	return []string{"privacy", "user_service", "community"}
}

type AccountService struct {
	db          *gorm.DB
	wechat      WeChatSessionProvider
	tokens      *platformauth.SessionManager
	uploadRoot  string
	emailCodes  EmailCodeStore
	emailSender EmailSender
	emailSecret []byte
}

type LoginResult struct {
	platformauth.TokenPair
	UserID     uint       `json:"user_id"`
	HasProfile bool       `json:"has_profile"`
	Profile    model.User `json:"profile"`
}

func NewAccountService(db *gorm.DB, wechat WeChatSessionProvider, tokens *platformauth.SessionManager, uploadRoot ...string) *AccountService {
	root := ""
	if len(uploadRoot) > 0 {
		root = strings.TrimSpace(uploadRoot[0])
	}
	return &AccountService{db: db, wechat: wechat, tokens: tokens, uploadRoot: root}
}

// Login 用微信临时 code 确认身份，为首次登录的用户建档，并签发已保存到 Redis 的双 Token。
// 任一步骤失败都不会返回登录成功；用户创建与附属账户初始化目前是独立的数据库操作。
func (s *AccountService) Login(ctx context.Context, code string) (*LoginResult, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrInvalidInput
	}
	openID, err := s.wechat.Code2Session(ctx, code) // 向微信验证 code，获取用户 OpenID。
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
	if err := s.ensureAccountState(user.ID); err != nil { // 补齐信用档案和贡献账户，不重置已有数据。
		return nil, err
	}
	token, err := s.tokens.Issue(ctx, user.ID) // 生成双 Token 并保存 Redis 会话，失败则不返回登录成功。
	if err != nil {
		return nil, err
	}
	return &LoginResult{TokenPair: *token, UserID: user.ID, HasProfile: strings.TrimSpace(user.Nickname) != "", Profile: user}, nil
}

// ensureAccountState 补齐用户的信用档案和贡献账户，不创建用户本身。
// 已有档案直接读取，不重置默认值；缺失时才创建。两份附属数据处于同一事务，失败则回滚。
// 该事务不包含 Login 中的用户创建。
func (s *AccountService) ensureAccountState(userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error { return ensureAccountStateTx(tx, userID) }) // 两份附属数据一起提交。
}

// ensureAccountStateTx 复用调用方事务补齐附属账户，冲突时保留原有数据。
func ensureAccountStateTx(tx *gorm.DB, userID uint) error {
	trust := model.UserTrustProfile{UserID: userID, TrustLevel: "new", DailyPublishLimit: 1, RiskFlags: model.JSONDocument("{}"), UpdatedAt: time.Now()}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&trust).Error; err != nil {
		return err
	}
	contribution := model.ContributionAccount{UserID: userID, UpdatedAt: time.Now()}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&contribution).Error
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

// Refresh rechecks account state before rotating either credential.
func (s *AccountService) Refresh(ctx context.Context, value string) (*platformauth.TokenPair, error) {
	userID, err := s.tokens.ValidateRefresh(ctx, value)
	if err != nil {
		return nil, err
	}
	active, err := s.IsAccountActive(ctx, userID)
	if err != nil {
		return nil, platformauth.ErrStoreUnavailable
	}
	if !active {
		return nil, platformauth.ErrUnauthorized
	}
	return s.tokens.Refresh(ctx, value)
}
