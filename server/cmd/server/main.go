package main

import (
	"context"
	"fmt"
	"log"
	"nutrilens/config"
	"nutrilens/internal/cache"
	"nutrilens/internal/handler"
	"nutrilens/internal/model"
	"nutrilens/internal/router"
	"nutrilens/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type appConfig struct {
	*config.Config
}

func (a *appConfig) GetJWTSecret() string { return a.JWT.Secret }
func (a *appConfig) GetJWTExpire() int    { return a.JWT.Expire }
func (a *appConfig) GetUploadDir() string { return a.Upload.Dir }

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Database
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	initDB(db)

	// Redis
	redisClient, err := cache.NewClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to create redis client: %v", err)
	}
	if err := redisClient.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to connect redis: %v", err)
	}
	fmt.Println("Redis connected")

	rateLimiter := cache.NewRateLimiter(redisClient, "ai_rate")

	// User tag getter — queries DB for user's tag
	getTag := func(userID uint) string {
		var user model.User
		if err := db.Select("tag").Where("id = ?", userID).First(&user).Error; err != nil {
			return ""
		}
		return user.Tag
	}

	// Services
	wechatSvc := service.NewWechatService(cfg.WeChat.AppID, cfg.WeChat.AppSecret)
	deepseekSvc := service.NewDeepSeekService(cfg.DeepSeek.APIKey, cfg.DeepSeek.BaseURL)
	foodSvc := service.NewFoodService(db, deepseekSvc)
	wheelSvc := service.NewWheelService(db, deepseekSvc)

	// Handler
	appCfg := &appConfig{cfg}
	h := &handler.Handler{
		Svc: &handler.Services{
			Auth:  wechatSvc,
			Food:  foodSvc,
			Wheel: wheelSvc,
		},
		Cfg: appCfg,
	}

	// Gin
	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()
	r.Static("/uploads", cfg.Upload.Dir)
	router.Setup(r, cfg.JWT.Secret, h, rateLimiter, getTag, cfg.RateLimit.Daily)

	addr := cfg.Server.Port
	fmt.Printf("Server starting on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initDB runs DDL auto-migration via GORM and seeds initial data.
func initDB(db *gorm.DB) {
	fmt.Println("Running database migration...")

	if err := db.AutoMigrate(
		&model.User{},
		&model.FoodRecord{},
		&model.Dish{},
		&model.PrivacyAgreement{},
		&model.AICallLog{},
	); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	fmt.Println("Database migration completed.")
	seedDishes(db)
}

func seedDishes(db *gorm.DB) {
	var count int64
	db.Model(&model.Dish{}).Where("is_system = ?", true).Count(&count)
	if count > 0 {
		return
	}

	dishes := []model.Dish{
		// 中式
		{Name: "宫保鸡丁", Category: "中式", Calories: 350, Weight: 80, IsSystem: true},
		{Name: "红烧肉", Category: "中式", Calories: 520, Weight: 60, IsSystem: true},
		{Name: "番茄炒蛋", Category: "中式", Calories: 200, Weight: 90, IsSystem: true},
		{Name: "麻婆豆腐", Category: "中式", Calories: 280, Weight: 75, IsSystem: true},
		{Name: "糖醋排骨", Category: "中式", Calories: 450, Weight: 65, IsSystem: true},
		{Name: "清炒时蔬", Category: "中式", Calories: 120, Weight: 95, IsSystem: true},
		{Name: "鱼香肉丝", Category: "中式", Calories: 320, Weight: 70, IsSystem: true},
		{Name: "蛋炒饭", Category: "中式", Calories: 400, Weight: 85, IsSystem: true},
		{Name: "水煮鱼", Category: "中式", Calories: 380, Weight: 55, IsSystem: true},
		{Name: "白切鸡", Category: "中式", Calories: 250, Weight: 75, IsSystem: true},
		// 西式
		{Name: "牛排", Category: "西式", Calories: 480, Weight: 70, IsSystem: true},
		{Name: "意大利面", Category: "西式", Calories: 420, Weight: 75, IsSystem: true},
		{Name: "凯撒沙拉", Category: "西式", Calories: 180, Weight: 90, IsSystem: true},
		{Name: "汉堡", Category: "西式", Calories: 550, Weight: 60, IsSystem: true},
		{Name: "三明治", Category: "西式", Calories: 320, Weight: 80, IsSystem: true},
		// 轻食
		{Name: "鸡胸肉沙拉", Category: "轻食", Calories: 250, Weight: 90, IsSystem: true},
		{Name: "藜麦蔬菜碗", Category: "轻食", Calories: 300, Weight: 85, IsSystem: true},
		{Name: "牛油果吐司", Category: "轻食", Calories: 280, Weight: 80, IsSystem: true},
		{Name: "酸奶水果杯", Category: "轻食", Calories: 180, Weight: 90, IsSystem: true},
		{Name: "全麦三明治", Category: "轻食", Calories: 300, Weight: 85, IsSystem: true},
		// 面食
		{Name: "兰州拉面", Category: "面食", Calories: 450, Weight: 75, IsSystem: true},
		{Name: "重庆小面", Category: "面食", Calories: 380, Weight: 70, IsSystem: true},
		{Name: "炸酱面", Category: "面食", Calories: 400, Weight: 80, IsSystem: true},
		{Name: "馄饨", Category: "面食", Calories: 300, Weight: 85, IsSystem: true},
		// 粤式
		{Name: "虾饺", Category: "粤式", Calories: 200, Weight: 80, IsSystem: true},
		{Name: "肠粉", Category: "粤式", Calories: 180, Weight: 85, IsSystem: true},
		{Name: "煲仔饭", Category: "粤式", Calories: 480, Weight: 65, IsSystem: true},
		{Name: "叉烧饭", Category: "粤式", Calories: 520, Weight: 60, IsSystem: true},
	}
	db.Create(&dishes)
	fmt.Printf("Seeded %d system dishes\n", len(dishes))
}
