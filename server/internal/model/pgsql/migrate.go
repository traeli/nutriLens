package pgsql

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrate 根据当前业务模型补齐表结构；新增模型时需要在这里注册。
// 不导入运营数据，也不删除旧表或旧字段。使用事务避免失败后留下部分结构。
func AutoMigrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 原 SQL 的单字段 UNIQUE 在模型中也需声明为 unique，而非仅声明 uniqueIndex。
		// 否则 GORM 会尝试按自己的命名删除旧约束，与 PostgreSQL 原约束名称不符。
		if err := tx.AutoMigrate(
			&User{}, &PrivacyAgreement{},
			&City{}, &CityPosterTheme{}, &Place{}, &Tag{},
			&CityRoute{}, &CityRouteStop{},
			&VisitRecord{}, &VisitRecordVersion{}, &VisitRecordTagLink{},
			&RecordMedia{}, &RecordEvidence{},
			&UserTrustProfile{}, &ContributionAccount{}, &Badge{}, &UserBadge{},
			&PublisherVerification{}, &ContentSafetyCheck{}, &ModerationTask{}, &FeatureFlag{},
			&PlaceFavorite{}, &HelpfulVote{}, &OutdatedSignal{},
			&ContentReport{}, &AppealCase{},
			&RouteFavorite{}, &RouteJourney{}, &NutritionRecord{},
		); err != nil {
			return fmt.Errorf("migrate business models: %w", err)
		}

		// 模型未声明的复合/部分唯一索引仍被业务去重和 SQL 种子导入依赖。
		// 沿用迁移脚本的名称，兼容已经通过 SQL 初始化的数据库。
		indexes := []string{
			`CREATE UNIQUE INDEX IF NOT EXISTS privacy_agreements_user_id_agreement_type_version_key ON privacy_agreements(user_id, agreement_type, version)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_cities_single_default ON cities(is_default) WHERE is_default = TRUE`,
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_places_provider_poi ON places(poi_provider, poi_id) WHERE poi_provider IS NOT NULL AND poi_id IS NOT NULL`,
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_city_routes_city_title ON city_routes(city_code, title)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS city_route_stops_route_id_sort_order_key ON city_route_stops(route_id, sort_order)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_visit_records_idempotency ON visit_records(user_id, create_request_key) WHERE create_request_key IS NOT NULL`,
			`CREATE UNIQUE INDEX IF NOT EXISTS visit_record_versions_record_id_version_no_key ON visit_record_versions(record_id, version_no)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_moderation_open_record ON moderation_tasks(target_type, target_id) WHERE status IN ('pending', 'assigned')`,
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_outdated_pending_unique ON outdated_signals(user_id, record_id) WHERE status = 'pending'`,
		}
		for _, statement := range indexes {
			if err := tx.Exec(statement).Error; err != nil {
				return fmt.Errorf("migrate business unique indexes: %w", err)
			}
		}
		return nil
	})
}
