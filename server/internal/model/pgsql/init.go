// Package pgsql 负责 PostgreSQL 模型和应用数据库生命周期。
package pgsql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"shijibu/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var defaultDB *gorm.DB

// Init 打开 PostgreSQL、检查连接、迁移表结构并生成种子数据；是否调用由应用入口决定。
func Init() error {
	if defaultDB != nil {
		return nil
	}
	cfg := config.Current().Database
	db, err := open(cfg)
	if err != nil {
		return fmt.Errorf("initialize database connection: %w", err)
	}
	if err := selfCheck(context.Background(), db, cfg.CheckTimeout); err != nil {
		_ = closeDB(db)
		return fmt.Errorf("database self-check: %w", err)
	}
	if err := db.AutoMigrate(autoMigrateModels()...); err != nil {
		_ = closeDB(db)
		return fmt.Errorf("database migration: %w", err)
	}
	if err := migrateLegacyAccountStatuses(db); err != nil {
		_ = closeDB(db)
		return fmt.Errorf("account status compatibility migration: %w", err)
	}
	if err := applyTableComments(db); err != nil {
		_ = closeDB(db)
		return fmt.Errorf("database table comments: %w", err)
	}
	if err := migrateLegacyRestaurantReviews(db); err != nil {
		_ = closeDB(db)
		return fmt.Errorf("database compatibility migration: %w", err)
	}
	if err := migrateLegacyVisitRecords(db); err != nil {
		_ = closeDB(db)
		return fmt.Errorf("visit record consolidation migration: %w", err)
	}
	if err := Seed(db); err != nil {
		_ = closeDB(db)
		return fmt.Errorf("database seed: %w", err)
	}
	defaultDB = db
	return nil
}

// migrateLegacyAccountStatuses 将旧版本允许登录的空账号状态统一为 active，
// 避免登录成功后又被认证中间件判定为不可用。已软删除账号不参与回填。
func migrateLegacyAccountStatuses(tx *gorm.DB) error {
	return tx.Exec(`UPDATE nutrilens_users
		SET account_status = 'active', updated_at = NOW()
		WHERE deleted_at IS NULL AND BTRIM(COALESCE(account_status, '')) = ''`).Error
}

// autoMigrateModels 集中声明由应用管理的持久化模型，便于迁移和结构测试共用同一份清单。
func autoMigrateModels() []any {
	return []any{
		&User{}, &PrivacyAgreement{},
		&City{}, &CityPosterTheme{}, &Place{}, &Tag{},
		&CityRoute{}, &CityRouteStop{},
		&VisitRecord{},
		&RestaurantReview{}, &RestaurantReviewVersion{}, &RestaurantReviewTagLink{},
		&UserTrustProfile{}, &ContributionAccount{}, &Badge{}, &UserBadge{},
		&PublisherVerification{}, &ContentSafetyCheck{}, &ModerationTask{}, &FeatureFlag{},
		&PlaceFavorite{}, &HelpfulVote{}, &OutdatedSignal{},
		&ContentReport{}, &AppealCase{},
		&RouteFavorite{}, &RouteJourney{}, &NutritionRecord{},
	}
}

// applyTableComments 为 AutoMigrate 管理的表补充中文用途说明；重复执行时 PostgreSQL 会更新已有注释。
func applyTableComments(tx *gorm.DB) error {
	statements := []string{
		`COMMENT ON TABLE nutrilens_users IS '用户账号'`,
		`COMMENT ON TABLE privacy_agreements IS '用户隐私协议确认记录'`,
		`COMMENT ON TABLE cities IS '运营城市'`,
		`COMMENT ON TABLE city_poster_themes IS '城市饮食记忆海报主题'`,
		`COMMENT ON TABLE places IS '餐饮地点'`,
		`COMMENT ON TABLE tags IS '内容标签'`,
		`COMMENT ON TABLE city_routes IS '城市餐饮路线'`,
		`COMMENT ON TABLE city_route_stops IS '城市餐饮路线站点'`,
		`COMMENT ON TABLE visit_records IS '用户私人到店足迹'`,
		`COMMENT ON TABLE restaurant_reviews IS '用户公开餐厅评论'`,
		`COMMENT ON TABLE restaurant_review_versions IS '餐厅评论版本'`,
		`COMMENT ON TABLE restaurant_review_tag_links IS '餐厅评论版本与标签关联'`,
		`COMMENT ON TABLE user_trust_profiles IS '用户可信度档案'`,
		`COMMENT ON TABLE contribution_accounts IS '用户贡献账户'`,
		`COMMENT ON TABLE badges IS '贡献徽章定义'`,
		`COMMENT ON TABLE user_badges IS '用户已获得徽章'`,
		`COMMENT ON TABLE publisher_verifications IS '内容发布者认证记录'`,
		`COMMENT ON TABLE content_safety_checks IS '内容安全检查记录'`,
		`COMMENT ON TABLE moderation_tasks IS '人工审核任务'`,
		`COMMENT ON TABLE feature_flags IS '功能开关'`,
		`COMMENT ON TABLE place_favorites IS '用户收藏的餐饮地点'`,
		`COMMENT ON TABLE helpful_votes IS '评论有帮助投票'`,
		`COMMENT ON TABLE outdated_signals IS '评论过时反馈'`,
		`COMMENT ON TABLE content_reports IS '内容举报'`,
		`COMMENT ON TABLE appeal_cases IS '审核申诉'`,
		`COMMENT ON TABLE route_favorites IS '用户收藏的餐饮路线'`,
		`COMMENT ON TABLE route_journeys IS '用户路线打卡进度'`,
		`COMMENT ON TABLE nutrition_records IS '用户私人饮食记录'`,
	}
	for _, statement := range statements {
		if err := tx.Exec(statement).Error; err != nil {
			return fmt.Errorf("apply table comments: %w", err)
		}
	}
	return nil
}

// migrateLegacyRestaurantReviews 将旧公开到店记录迁入评论域，再把来源记录恢复为私人足迹。
func migrateLegacyRestaurantReviews(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("visit_record_versions") {
		return nil
	}
	statements := []string{
		`INSERT INTO restaurant_reviews (id, user_id, place_id, visit_record_id, status, risk_level, submitted_at, published_at, helpful_count, outdated_count, report_count, created_at, updated_at, deleted_at)
		 SELECT records.id, records.user_id, records.place_id, records.id,
		 CASE records.publish_status WHEN 'pending_review' THEN 'pending' WHEN 'published' THEN 'published' ELSE 'rejected' END,
		 records.risk_level, COALESCE(records.submitted_at, records.created_at), records.published_at,
		 records.helpful_count, records.outdated_count, records.report_count, records.created_at, records.updated_at, records.deleted_at
		 FROM visit_records AS records
		 WHERE records.publish_status IN ('pending_review', 'published', 'rejected', 'hidden', 'reported')
		 ON CONFLICT (id) DO NOTHING`,
		`INSERT INTO restaurant_review_versions (id, review_id, version_no, editor_user_id, visit_date, conclusion, average_cost, wait_minutes, meal_period, dishes, content, change_summary, created_at)
		 SELECT versions.id, reviews.id, versions.version_no, versions.editor_user_id, versions.visit_date,
		 versions.conclusion, versions.average_cost, versions.wait_minutes, versions.meal_period,
		 versions.dishes, versions.content, versions.change_summary, versions.created_at
		 FROM restaurant_reviews AS reviews
		 JOIN visit_records AS records ON records.id = reviews.visit_record_id
		 JOIN visit_record_versions AS versions ON versions.id = records.current_version_id
		 ON CONFLICT (id) DO NOTHING`,
		`UPDATE restaurant_reviews AS reviews SET current_version_id = versions.id
		 FROM restaurant_review_versions AS versions
		 WHERE versions.review_id = reviews.id AND reviews.current_version_id IS NULL`,
		`UPDATE restaurant_reviews SET status = 'rejected' WHERE status NOT IN ('pending', 'published', 'rejected')`,
		`INSERT INTO restaurant_review_tag_links(review_version_id, tag_id)
		 SELECT links.record_version_id, links.tag_id FROM visit_record_tag_links AS links
		 JOIN restaurant_review_versions AS versions ON versions.id = links.record_version_id
		 ON CONFLICT DO NOTHING`,
		`UPDATE moderation_tasks AS tasks
		 SET target_type = 'restaurant_review', target_id = reviews.id, review_version_id = reviews.current_version_id
		 FROM restaurant_reviews AS reviews
		 WHERE tasks.target_type = 'visit_record' AND tasks.target_id = reviews.visit_record_id`,
		`UPDATE visit_record_versions AS versions SET visibility = 'private'
		 FROM restaurant_reviews AS reviews
		 WHERE versions.record_id = reviews.visit_record_id AND versions.visibility <> 'private'`,
		`UPDATE visit_records AS records
		 SET visibility = 'private', publish_status = 'draft', risk_level = 'low', submitted_at = NULL, published_at = NULL
		 FROM restaurant_reviews AS reviews
		 WHERE records.id = reviews.visit_record_id`,
		`UPDATE restaurant_reviews AS reviews
		 SET media = COALESCE((
			 SELECT jsonb_agg(to_jsonb(media) ORDER BY media.sort_order, media.id)
			 FROM record_media AS media WHERE media.record_id = reviews.visit_record_id
		 ), '[]'::jsonb)
		 WHERE reviews.media = '[]'::jsonb`,
		`SELECT setval(
			pg_get_serial_sequence('restaurant_review_versions', 'id'),
			GREATEST(COALESCE((SELECT MAX(id) FROM restaurant_review_versions), 0), 1),
			EXISTS (SELECT 1 FROM restaurant_review_versions)
		)`,
	}
	for _, statement := range statements {
		if err := tx.Exec(statement).Error; err != nil {
			return fmt.Errorf("migrate legacy restaurant reviews: %w", err)
		}
	}
	return nil
}

// migrateLegacyVisitRecords 把旧版本表、标签关系、媒体和凭证一次性回填到足迹主表。
// 迁移是幂等的；旧表暂时保留用于回滚，但业务代码不再读写它们。
func migrateLegacyVisitRecords(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("visit_record_versions") {
		return nil
	}
	statement := `UPDATE visit_records AS records
		SET version_no = versions.version_no,
			visit_date = versions.visit_date,
			consumer_type = versions.consumer_type,
			conclusion = versions.conclusion,
			price_min = versions.price_min,
			price_max = versions.price_max,
			average_cost = versions.average_cost,
			wait_minutes = versions.wait_minutes,
			meal_period = versions.meal_period,
			dishes = versions.dishes,
			content = versions.content,
			change_summary = versions.change_summary,
			tag_ids = COALESCE((
				SELECT jsonb_agg(links.tag_id ORDER BY tags.sort_order, links.tag_id)
				FROM visit_record_tag_links AS links
				JOIN tags ON tags.id = links.tag_id
				WHERE links.record_version_id = versions.id
			), '[]'::jsonb),
			media = COALESCE((
				SELECT jsonb_agg(to_jsonb(media) ORDER BY media.sort_order, media.id)
				FROM record_media AS media WHERE media.record_id = records.id
			), '[]'::jsonb),
			evidences = COALESCE((
				SELECT jsonb_agg(to_jsonb(evidence) ORDER BY evidence.id)
				FROM record_evidences AS evidence WHERE evidence.record_id = records.id
			), '[]'::jsonb)
		FROM visit_record_versions AS versions
		WHERE versions.id = records.current_version_id
			AND records.visit_date IS NULL`
	if err := tx.Exec(statement).Error; err != nil {
		return fmt.Errorf("consolidate legacy visit records: %w", err)
	}
	return nil
}

// DB 返回由 Init 创建的数据库连接。
func DB() *gorm.DB {
	if defaultDB == nil {
		panic("database is not initialized")
	}
	return defaultDB
}

// Close 释放已经初始化的 PostgreSQL 连接池。
func Close() error {
	if defaultDB == nil {
		return nil
	}
	db := defaultDB
	defaultDB = nil
	return closeDB(db)
}

func open(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, errors.New("open PostgreSQL connection")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.New("get PostgreSQL connection pool")
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	return db, nil
}

func selfCheck(ctx context.Context, db *gorm.DB, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func closeDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
