package pgsql

import (
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// legacyCityRouteStop 保留本次变更前的生产表结构，用于验证已有数据可以由
// AutoMigrate 原地升级，而不是只验证空库建表。
type legacyCityRouteStop struct {
	ID        uint   `gorm:"primaryKey"`
	RouteID   uint   `gorm:"not null;index"`
	PlaceID   uint   `gorm:"not null;index"`
	SortOrder int    `gorm:"not null;default:0"`
	Note      string `gorm:"size:240;not null;default:''"`
}

func (legacyCityRouteStop) TableName() string { return "city_route_stops" }

// TestProductionSchemaUpgrade 仅在显式提供隔离测试库时运行。数据库名必须固定，
// 防止误把破坏性升级演练指向开发库或正式库。
func TestProductionSchemaUpgrade(t *testing.T) {
	dsn := os.Getenv("NUTRILENS_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("set NUTRILENS_MIGRATION_TEST_DSN to run PostgreSQL upgrade smoke test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open migration test database: %v", err)
	}
	var databaseName string
	if err := db.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil {
		t.Fatalf("read migration test database name: %v", err)
	}
	if databaseName != "nutrilens_migration_test" {
		t.Fatalf("refuse destructive migration test against database %q", databaseName)
	}

	if err := db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public").Error; err != nil {
		t.Fatalf("reset migration test schema: %v", err)
	}
	legacyModels := []any{
		&User{}, &PrivacyAgreement{},
		&City{}, &CityPosterTheme{}, &Place{}, &Tag{},
		&CityRoute{}, &legacyCityRouteStop{},
		&VisitRecord{},
		&RestaurantReview{}, &RestaurantReviewVersion{}, &RestaurantReviewTagLink{},
		&UserTrustProfile{}, &ContributionAccount{}, &Badge{}, &UserBadge{},
		&PublisherVerification{}, &ContentSafetyCheck{}, &ModerationTask{}, &FeatureFlag{},
		&PlaceFavorite{}, &HelpfulVote{}, &OutdatedSignal{},
		&ContentReport{}, &AppealCase{},
		&RouteFavorite{}, &RouteJourney{}, &NutritionRecord{},
	}
	if err := db.AutoMigrate(legacyModels...); err != nil {
		t.Fatalf("create legacy production schema: %v", err)
	}
	// 复现旧生产表：主键存在，但缺少自增序列和默认值。AutoMigrate 不会补回该默认值。
	if err := db.Exec(`ALTER TABLE restaurant_review_versions ALTER COLUMN id DROP DEFAULT;
		DROP SEQUENCE IF EXISTS restaurant_review_versions_id_seq`).Error; err != nil {
		t.Fatalf("simulate legacy review version primary key: %v", err)
	}
	if err := seedLegacyDiscoveryData(db); err != nil {
		t.Fatalf("seed legacy production data: %v", err)
	}

	if err := db.AutoMigrate(autoMigrateModels()...); err != nil {
		t.Fatalf("upgrade production schema: %v", err)
	}
	if err := migrateLegacyRestaurantReviews(db); err != nil {
		t.Fatalf("migrate legacy restaurant reviews: %v", err)
	}
	if err := ensureRestaurantReviewVersionSequence(db); err != nil {
		t.Fatalf("repair restaurant review version sequence: %v", err)
	}
	if err := migratePublishedReviewsToDiscoveryPosts(db); err != nil {
		t.Fatalf("migrate published reviews to discovery posts: %v", err)
	}
	if err := applyTableComments(db); err != nil {
		t.Fatalf("apply upgraded schema comments: %v", err)
	}
	if err := Seed(db); err != nil {
		t.Fatalf("seed upgraded production schema: %v", err)
	}
	if err := Seed(db); err != nil {
		t.Fatalf("repeat upgraded production seed: %v", err)
	}

	for _, column := range []string{"recommended_food", "food_category_code", "icon_key"} {
		if !db.Migrator().HasColumn(&CityRouteStop{}, column) {
			t.Errorf("upgraded city_route_stops missing column %s", column)
		}
	}
	if !db.Migrator().HasTable(&DiscoveryScenario{}) {
		t.Fatal("upgraded schema missing discovery_scenarios")
	}
	var scenarioCount int64
	if err := db.Model(&DiscoveryScenario{}).Count(&scenarioCount).Error; err != nil {
		t.Fatalf("count discovery scenarios: %v", err)
	}
	if scenarioCount != 11 {
		t.Fatalf("discovery scenario count after repeated seed = %d, want 11", scenarioCount)
	}

	version := RestaurantReviewVersion{
		ReviewID: 999999, VersionNo: 1, EditorUserID: 1, VisitDate: time.Now(),
		Conclusion: "recommend", Content: "sequence migration verification", CreatedAt: time.Now(),
	}
	if err := db.Create(&version).Error; err != nil {
		t.Fatalf("insert review version with generated ID: %v", err)
	}
	if version.ID == 0 {
		t.Fatal("generated review version ID is zero")
	}

	var stop CityRouteStop
	if err := db.Joins("JOIN city_routes ON city_routes.id = city_route_stops.route_id").
		Where("city_routes.title = ? AND city_route_stops.sort_order = ?", "曹家渡的四站早餐", 1).
		First(&stop).Error; err != nil {
		t.Fatalf("load upgraded legacy route stop: %v", err)
	}
	if stop.RecommendedFood != "鲜肉生煎" || stop.FoodCategoryCode != "shengjian" || stop.IconKey != "shengjian" {
		t.Fatalf("legacy route stop metadata was not backfilled: %#v", stop)
	}
}

func seedLegacyDiscoveryData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := seedCityCatalog(tx); err != nil {
			return err
		}
		if err := seedPlaces(tx); err != nil {
			return err
		}
		if err := createSeed(tx, "poster themes", seedPosterThemes()); err != nil {
			return err
		}
		if err := createSeed(tx, "experience tags", seedTags()); err != nil {
			return err
		}
		if err := createSeed(tx, "city routes", seedRoutes()); err != nil {
			return err
		}
		if err := createSeed(tx, "feature flags", seedFeatureFlags()); err != nil {
			return err
		}
		var route CityRoute
		if err := tx.Where("city_code = ? AND title = ?", "310000", "曹家渡的四站早餐").First(&route).Error; err != nil {
			return err
		}
		var place Place
		if err := tx.Where("city_code = ? AND name = ?", "310000", "老周生煎").First(&place).Error; err != nil {
			return err
		}
		return tx.Create(&legacyCityRouteStop{RouteID: route.ID, PlaceID: place.ID, SortOrder: 1, Note: "生产环境已有说明"}).Error
	})
}
