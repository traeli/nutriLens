package main

import (
	"context"
	"fmt"
	"log"
	"nutrilens/config"
	"nutrilens/internal/cache"
	cosPkg "nutrilens/internal/cos"
	"nutrilens/internal/handler"
	"nutrilens/internal/model"
	"nutrilens/internal/router"
	"nutrilens/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var beijingLoc = time.FixedZone("CST", 8*3600)

type appConfig struct {
	*config.Config
}

func (a *appConfig) GetJWTSecret() string            { return a.JWT.Secret }
func (a *appConfig) GetJWTExpire() int               { return a.JWT.Expire }
func (a *appConfig) GetUploadDir() string            { return a.Upload.Dir }
func (a *appConfig) GetTemplateID() string           { return a.Notify.TemplateID }
func (a *appConfig) GetDefaultBreakfastTime() string { return a.Notify.BreakfastTime }
func (a *appConfig) GetDefaultLunchTime() string     { return a.Notify.LunchTime }
func (a *appConfig) GetDefaultDinnerTime() string    { return a.Notify.DinnerTime }

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
	initDB(db, cfg)

	// Redis
	redisClient, err := cache.NewClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to create redis client: %v", err)
	}
	if err := redisClient.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to connect redis: %v", err)
	}
	fmt.Println("Redis connected")

	// COS
	cosClient, err := cosPkg.NewClient(cfg.COS)
	if err != nil {
		log.Fatalf("Failed to create COS client: %v", err)
	}
	fmt.Println("COS connected")

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
	wechatSvc := service.NewWechatService(cfg.WeChat.AppID, cfg.WeChat.AppSecret, cfg.WeChat.Token, redisClient)
	aiProvider := service.NewAIProviderService(db)
	foodSvc := service.NewFoodService(db, aiProvider, cosClient)
	wheelSvc := service.NewWheelService(db, aiProvider)
	scoreSvc := service.NewScoreService(db)
	achieveSvc := service.NewAchievementService(db, scoreSvc)
	shareSvc := service.NewShareService(db, scoreSvc)
	appCfg := &appConfig{cfg}
	notifySvc := service.NewNotifyService(db, wechatSvc, aiProvider, appCfg)
	webhookSvc := service.NewWebhookService(cfg.DeepSeek.APIKey, cfg.DeepSeek.BaseURL, db)

	// Handler
	h := &handler.Handler{
		Svc: &handler.Services{
			Auth:   wechatSvc,
			Food:   foodSvc,
			Wheel:  wheelSvc,
			Notify: notifySvc,
		},
		Cfg:        appCfg,
		COS:        cosClient,
		ScoreSvc:   scoreSvc,
		AchieveSvc: achieveSvc,
		ShareSvc:   shareSvc,
		NotifySvc:  notifySvc,
		WebhookSvc: webhookSvc,
	}

	// Gin
	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()
	router.Setup(r, cfg.JWT.Secret, h, rateLimiter, getTag, cfg.RateLimit.Daily)

	// Start notification scheduler
	checkInterval := cfg.Notify.CheckInterval
	if checkInterval == "" {
		checkInterval = "@every 1m"
	}
	c := cron.New(cron.WithLocation(beijingLoc))
	c.AddFunc(checkInterval, func() {
		if err := notifySvc.CheckAndNotify(context.Background()); err != nil {
			log.Printf("[NotifyScheduler] check error: %v", err)
		}
	})
	c.Start()
	fmt.Println("Notification scheduler started")

	addr := cfg.Server.Port
	fmt.Printf("Server starting on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	c.Stop()
}

// initDB runs DDL auto-migration via GORM and seeds initial data.
func initDB(db *gorm.DB, cfg *config.Config) {
	fmt.Println("Running database migration...")

	if err := db.AutoMigrate(
		&model.User{},
		&model.FoodRecord{},
		&model.Dish{},
		&model.PrivacyAgreement{},
		&model.AICallLog{},
		&model.Feedback{},
		&model.AIModel{},
		&model.AIPrompt{},
		&model.UserDailyAnalysis{},
		&model.UserScore{},
		&model.ScoreLog{},
		&model.NotifySetting{},
		&model.NotifyLog{},
		&model.MealRecommendation{},
		&model.FriendRelation{},
		&model.Achievement{},
		&model.UserAchievement{},
		&model.ShareRecord{},
		&model.InviteRelation{},
		&model.Share{},
		&model.WebhookProject{},
	); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	fmt.Println("Database migration completed.")
	seedDishes(db)
	seedAIModels(db, cfg.DeepSeek.APIKey)
	seedAIPrompts(db)
	seedAchievements(db)
	seedShare(db)
	seedWebhookProjects(db)
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

func seedAIModels(db *gorm.DB, deepseekKey string) {
	type seedEntry struct {
		taskType string
		model    model.AIModel
	}
	seeds := []seedEntry{
		{
			taskType: "text",
			model: model.AIModel{
				Name:      "DeepSeek Chat (文本)",
				Provider:  "deepseek",
				ModelName: "deepseek-chat",
				APIKey:    deepseekKey,
				BaseURL:   "https://api.deepseek.com/v1/chat/completions",
				TaskType:  "text",
				Enabled:   true,
			},
		},
		{
			taskType: "dish",
			model: model.AIModel{
				Name:      "DeepSeek Chat (菜品)",
				Provider:  "deepseek",
				ModelName: "deepseek-chat",
				APIKey:    deepseekKey,
				BaseURL:   "https://api.deepseek.com/v1/chat/completions",
				TaskType:  "dish",
				Enabled:   true,
			},
		},
		{
			taskType: "image",
			model: model.AIModel{
				Name:      "智谱 GLM-4V (图片)",
				Provider:  "zhipu",
				ModelName: "glm-4v-flash",
				APIKey:    "",
				BaseURL:   "https://open.bigmodel.cn/api/paas/v4/chat/completions",
				TaskType:  "image",
				Enabled:   false,
			},
		},
		{
			taskType: "image_identify",
			model: model.AIModel{
				Name:      "智谱 GLM-4V (图片识别)",
				Provider:  "zhipu",
				ModelName: "glm-4v-flash",
				APIKey:    "",
				BaseURL:   "https://open.bigmodel.cn/api/paas/v4/chat/completions",
				TaskType:  "image_identify",
				Enabled:   true,
			},
		},
		{
			taskType: "daily_analysis",
			model: model.AIModel{
				Name:      "DeepSeek Chat (每日分析)",
				Provider:  "deepseek",
				ModelName: "deepseek-chat",
				APIKey:    deepseekKey,
				BaseURL:   "https://api.deepseek.com/v1/chat/completions",
				TaskType:  "daily_analysis",
				Enabled:   true,
			},
		},
		{
			taskType: "meal_recommend",
			model: model.AIModel{
				Name:      "DeepSeek Chat (餐食推荐)",
				Provider:  "deepseek",
				ModelName: "deepseek-chat",
				APIKey:    deepseekKey,
				BaseURL:   "https://api.deepseek.com/v1/chat/completions",
				TaskType:  "meal_recommend",
				Enabled:   true,
			},
		},
	}

	inserted := 0
	for _, s := range seeds {
		var count int64
		db.Model(&model.AIModel{}).Where("task_type = ?", s.taskType).Count(&count)
		if count > 0 {
			continue
		}
		if err := db.Create(&s.model).Error; err != nil {
			log.Printf("Failed to seed AI model for task_type=%s: %v", s.taskType, err)
		} else {
			inserted++
		}
	}
	if inserted > 0 {
		fmt.Printf("Seeded %d new AI models\n", inserted)
	}
}

func seedAIPrompts(db *gorm.DB) {
	type seedEntry struct {
		taskType string
		prompt   model.AIPrompt
	}
	seeds := []seedEntry{
		{
			taskType: "image",
			prompt: model.AIPrompt{
				Title:        "图片食物分析",
				TaskType:     "image",
				SystemPrompt: "",
				UserPromptTemplate: `请分析这张图片中的内容，返回JSON格式(不要其他内容):
{
  "is_food": true或false(图片中是否包含食物),
  "food_name": "食物名称",
  "unit": "kg",
  "unit_amount": 估算重量(基于图片中食物的视觉比例，参考成人正常食量，单位为kg的数字，如0.15表示150g),
  "calories": 估算总卡路里(kcal,数字，基于估算重量计算),
  "nutrients": {"protein": 蛋白质g, "carbs": 碳水g, "fat": 脂肪g, "fiber": 膳食纤维g, "sugar": 糖分g, "vitamin_c": 维生素C_mg},
  "suggestion": "根据{{user_info}}给出的个性化饮食建议，特别关注糖分摄入对血糖的影响"
}

注意:
- is_food字段必填。如果图片中没有食物(如风景、人物、动物等)，设为false，其他字段填默认值即可
- 如果is_food为true，所有营养数据必须基于unit_amount(估算重量)来计算
- 糖分(sugar)字段必填，单位为克
- 维生素C(vitamin_c)如果食物含有则填写，单位为mg，不含则为0
- unit固定为"kg"`,
				Enabled: true,
			},
		},
		{
			taskType: "image_identify",
			prompt: model.AIPrompt{
				Title:        "图片食物识别",
				TaskType:     "image_identify",
				SystemPrompt: "",
				UserPromptTemplate: `请识别这张图片中的食物，返回JSON格式(不要其他内容):
{
  "is_food": true或false(图片中是否包含食物),
  "foods": [
    {"name": "食物名称", "unit": "kg", "unit_amount": 估算重量(基于图片中食物的视觉比例，参考成人正常食量，单位为kg的数字，如0.15表示150g)},
    {"name": "食物名称", "unit": "kg", "unit_amount": 估算重量}
  ]
}

注意:
- is_food字段必填。如果图片中没有食物，设为false，foods为空数组
- 如果图片中有多种食物，每种都要列出
- unit固定为"kg"
- unit_amount基于图片中食物的视觉比例估算，参考成人正常食量
- 只需要识别食物名称和估算重量，不需要分析营养成分`,
				Enabled: true,
			},
		},
		{
			taskType: "text",
			prompt: model.AIPrompt{
				Title:        "文本食物分析",
				TaskType:     "text",
				SystemPrompt: "你是一个专业的营养师AI助手。用户会告诉你他们吃了什么，你需要将每种食物分开分析营养成分和卡路里，并根据用户的身体信息给出个性化的饮食建议。请始终以JSON格式回复。",
				UserPromptTemplate: `我吃了: {{description}}

我的信息: {{user_info}}

请返回JSON格式(不要其他内容):
{
  "is_food": true或false(用户描述的内容是否是食物),
  "foods": [
    {"name": "食物名称", "unit": "kg", "unit_amount": 估算重量(参考该用户身体信息的正常食量，单位为kg的数字，如0.1表示100g), "calories": 卡路里数字, "nutrients": {"protein": 蛋白质g, "carbs": 碳水g, "fat": 脂肪g, "fiber": 膳食纤维g, "sugar": 糖分g, "vitamin_c": 维生素C_mg}},
    {"name": "食物名称", "unit": "kg", "unit_amount": 估算重量, "calories": 卡路里数字, "nutrients": {"protein": 蛋白质g, "carbs": 碳水g, "fat": 脂肪g, "fiber": 膳食纤维g, "sugar": 糖分g, "vitamin_c": 维生素C_mg}}
  ],
  "suggestion": "根据用户信息给出的个性化饮食建议，特别关注糖分摄入对血糖的影响"
}

重要规则:
- is_food字段必填。如果用户描述的不是食物(如"我吃了个手机"、"今天天气不错"等)，设为false，foods为空数组
- unit固定为"kg"
- 如果用户没有明确说食物量，根据用户的身高体重年龄估算一个成人正常食用量作为unit_amount
- 所有营养数据必须基于unit_amount(估算重量)来计算
- 糖分(sugar)字段必填，单位为克，这对糖尿病患者非常重要
- 维生素C(vitamin_c)如果食物含有则填写，单位为mg，不含则为0`,
				Enabled: true,
			},
		},
		{
			taskType: "dish",
			prompt: model.AIPrompt{
				Title:        "菜品营养分析",
				TaskType:     "dish",
				SystemPrompt: "你是一个专业的营养师AI助手。根据菜品名称，分析其营养成分。请始终以JSON格式回复。",
				UserPromptTemplate: `菜品名称: {{dish_name}}

请返回JSON格式(不要其他内容):
{
  "is_food": true或false(这个名称是否是一种食物),
  "name": "{{dish_name}}",
  "category": "分类(中式/西式/日式/韩式/其他)",
  "unit": "kg",
  "unit_amount": 成人正常一份的重量(kg数字,如0.2表示200g),
  "calories": 一份的卡路里(kcal数字),
  "nutrients": {"protein": 蛋白质g, "carbs": 碳水g, "fat": 脂肪g, "fiber": 膳食纤维g, "sugar": 糖分g, "vitamin_c": 维生素C_mg}
}

注意:
- is_food字段必填。如果名称不是食物，设为false
- unit固定为"kg"
- 营养数据基于unit_amount(成人正常一份)计算
- sugar字段必填，这对糖尿病患者很重要`,
				Enabled: true,
			},
		},
		{
			taskType: "daily_analysis",
			prompt: model.AIPrompt{
				Title:        "每日健康分析",
				TaskType:     "daily_analysis",
				SystemPrompt: "你是一个专业的健康营养顾问AI。请严格按照要求的JSON格式回复，不要添加任何其他内容。",
				UserPromptTemplate: `用户今天的饮食记录:
{{food_list}}

今日营养摄入汇总:
{{nutrient_summary}}

用户身体信息: {{user_info}}

请返回JSON格式(不要其他内容):
{
  "nutrition_eval": "营养评估：今日营养摄入是否均衡，哪些过多或不足，30字以内",
  "diet_advice": "饮食建议：剩余时间建议吃什么注意什么，50字以内",
  "exercise_advice": "运动建议：根据今日卡路里建议的运动，40字以内"
}

注意:
- 三个字段都是必填
- 每个字段请严格控制在指定字数以内，语言简洁
- 如果糖分摄入过多，在diet_advice中重点提醒
- 运动建议要具体(如: 快走30分钟、慢跑20分钟等)
- 只返回JSON，不要markdown代码块或其他内容`,
				Enabled: true,
			},
		},
		{
			taskType: "meal_recommend",
			prompt: model.AIPrompt{
				Title:        "餐食推荐推送",
				TaskType:     "meal_recommend",
				SystemPrompt: "你是一个专业的营养师AI助手。根据用户的个人信息、今日已摄入的食物和营养素、用户的菜品偏好，为指定餐次推荐合适的菜品。回复必须严格遵循JSON格式。",
				UserPromptTemplate: `用户信息: {{user_info}}

今日已摄入食物:
{{today_foods}}

今日营养素汇总:
{{nutrient_summary}}

用户偏好菜品(喜爱度1-100):
{{preferred_dishes}}

目标餐次: {{meal_type_name}}

请为{{meal_type_name}}推荐2-3道菜品，结合用户偏好和今日营养缺口。返回JSON格式:
{
  "recommendations": [
    {"name": "菜品名称", "reason": "推荐理由(30字内)", "calories": 估算卡路里数字}
  ],
  "tip": "一句话饮食小贴士(40字内)"
}

注意:
- 优先从用户偏好菜品中选择，若没有合适偏好则推荐一般健康菜品
- 推荐理由要结合用户今天的营养素摄入情况
- 如果用户今天还没吃东西，按早餐标准推荐(清淡、有营养)
- 只返回JSON，不要其他内容`,
				Enabled: true,
			},
		},
	}

	inserted := 0
	for _, s := range seeds {
		var count int64
		db.Model(&model.AIPrompt{}).Where("task_type = ?", s.taskType).Count(&count)
		if count > 0 {
			continue
		}
		if err := db.Create(&s.prompt).Error; err != nil {
			log.Printf("Failed to seed AI prompt for task_type=%s: %v", s.taskType, err)
		} else {
			inserted++
		}
	}
	if inserted > 0 {
		fmt.Printf("Seeded %d new AI prompts\n", inserted)
	}
}

func seedAchievements(db *gorm.DB) {
	var count int64
	db.Model(&model.Achievement{}).Count(&count)
	if count > 0 {
		return
	}

	achievements := []model.Achievement{
		{Code: "newbie", Name: "萌新食客", Icon: "seedling", Description: "注册并完成首次食物识别", Rarity: "common", ConditionType: "first_analyze", ConditionValue: 1, RewardType: "", RewardValue: 0},
		{Code: "wheel_regular", Name: "转盘常客", Icon: "slot", Description: "使用转盘30次", Rarity: "common", ConditionType: "spin", ConditionValue: 30, RewardType: "", RewardValue: 0},
		{Code: "calorie_detective", Name: "热量侦探", Icon: "fire", Description: "累计识别50次食物", Rarity: "rare", ConditionType: "total_analyze", ConditionValue: 50, RewardType: "quota", RewardValue: 5},
		{Code: "week_streak", Name: "坚持一周", Icon: "calendar", Description: "连续打卡7天", Rarity: "rare", ConditionType: "streak", ConditionValue: 7, RewardType: "quota", RewardValue: 5},
		{Code: "fitness_star", Name: "健身达人", Icon: "dumbbell", Description: "连续7天热量在推荐范围内", Rarity: "rare", ConditionType: "healthy_streak", ConditionValue: 7, RewardType: "poster_template", RewardValue: 1},
		{Code: "three_meals_10", Name: "三餐不落", Icon: "utensils", Description: "完成10次三餐全记录", Rarity: "rare", ConditionType: "three_meals", ConditionValue: 10, RewardType: "quota", RewardValue: 10},
		{Code: "foodie", Name: "美食鉴赏家", Icon: "chef", Description: "创建20道自定义菜品", Rarity: "rare", ConditionType: "dish", ConditionValue: 20, RewardType: "quota", RewardValue: 5},
		{Code: "share_master", Name: "分享达人", Icon: "share", Description: "分享海报20次", Rarity: "rare", ConditionType: "share", ConditionValue: 20, RewardType: "poster_template", RewardValue: 2},
		{Code: "photographer", Name: "热量摄影师", Icon: "camera", Description: "累计识别200次食物", Rarity: "epic", ConditionType: "total_analyze", ConditionValue: 200, RewardType: "quota", RewardValue: 20},
		{Code: "month_streak", Name: "坚持一月", Icon: "calendar_check", Description: "连续打卡30天", Rarity: "epic", ConditionType: "streak", ConditionValue: 30, RewardType: "quota", RewardValue: 20},
		{Code: "popular", Name: "人气食客", Icon: "users", Description: "邀请10位好友注册", Rarity: "epic", ConditionType: "invite", ConditionValue: 10, RewardType: "quota", RewardValue: 30},
		{Code: "nutrition_master", Name: "营养大师", Icon: "bullseye", Description: "连续30天热量在推荐范围内", Rarity: "legendary", ConditionType: "healthy_streak", ConditionValue: 30, RewardType: "vip", RewardValue: 0},
		{Code: "nutrition_god", Name: "营养之神", Icon: "crown", Description: "累计识别1000次食物", Rarity: "legendary", ConditionType: "total_analyze", ConditionValue: 1000, RewardType: "vip", RewardValue: 0},
		{Code: "perfect_attendance", Name: "全勤之星", Icon: "trophy", Description: "连续打卡100天", Rarity: "legendary", ConditionType: "streak", ConditionValue: 100, RewardType: "vip", RewardValue: 0},
		{Code: "dragon", Name: "美食龙", Icon: "dragon", Description: "解锁所有普通+稀有称号", Rarity: "legendary", ConditionType: "all_common_rare", ConditionValue: 8, RewardType: "vip", RewardValue: 0},
	}

	if err := db.Create(&achievements).Error; err != nil {
		log.Printf("Failed to seed achievements: %v", err)
	} else {
		fmt.Printf("Seeded %d achievements\n", len(achievements))
	}
}

func seedShare(db *gorm.DB) {
	var count int64
	db.Model(&model.Share{}).Count(&count)
	if count > 0 {
		return
	}
	share := model.Share{
		ImageURL: "https://nutrilens-1429775067.cos.ap-guangzhou.myqcloud.com/share/ee7a24a6-784f-4b2f-abe8-d919276ae850.png",
		Type:     "poster",
	}
	if err := db.Create(&share).Error; err != nil {
		log.Printf("Failed to seed share: %v", err)
	} else {
		fmt.Println("Seeded share poster")
	}
}

func seedWebhookProjects(db *gorm.DB) {
	var count int64
	db.Model(&model.WebhookProject{}).Count(&count)
	if count > 0 {
		return
	}

	project := model.WebhookProject{
		Name:             "DRH APP服务",
		RepoName:         "SmartWearables/apiService",
		GitURL:           "http://120.76.141.107:33701/SmartWearables/apiService.git",
		GitToken:         "b8595975da7bd154bdb2d19c74a0b4a7fd918980",
		FeishuWebhookURL: "https://open.feishu.cn/open-apis/bot/v2/hook/167cd842-e3cf-42f7-b5ac-9774e5507060",
		FeishuKeyword:    "drhhh",
		Enabled:          true,
	}
	if err := db.Create(&project).Error; err != nil {
		log.Printf("Failed to seed webhook project: %v", err)
	} else {
		fmt.Println("Seeded webhook project")
	}
}
