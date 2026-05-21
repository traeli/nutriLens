package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"nutrilens/internal/cos"
	"nutrilens/internal/middleware"
	"nutrilens/internal/model"
	"nutrilens/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Svc        *Services
	Cfg        ConfigProvider
	COS        *cos.Client
	ScoreSvc   *service.ScoreService
	AchieveSvc *service.AchievementService
	ShareSvc   *service.ShareService
	NotifySvc  *service.NotifyService
	WebhookSvc *service.WebhookService
}

type Services struct {
	Auth   *service.WechatService
	Food   *service.FoodService
	Wheel  *service.WheelService
	Notify *service.NotifyService
}

type ConfigProvider interface {
	GetJWTSecret() string
	GetJWTExpire() int
	GetUploadDir() string
}

// ==================== Auth ====================

type WxLoginReq struct {
	Code      string `json:"code" binding:"required"`
	InviterID uint   `json:"inviter_id"`
}

func (h *Handler) WxLogin(c *gin.Context) {
	var req WxLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[WxLogin] bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	var openID string

	// Dev mock: 游客模式下 code 为 "the code is a mock one"
	if req.Code == "the code is a mock one" {
		openID = "mock_openid_dev"
		log.Println("[WxLogin] dev mock mode, using mock_openid_dev")
	} else {
		var err error
		openID, _, err = h.Svc.Auth.Code2Session(req.Code)
		if err != nil {
			log.Printf("[WxLogin] code2Session failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "wechat login failed: " + err.Error()})
			return
		}
	}

	db := h.Svc.Food.DB()
	var user model.User
	result := db.Where("open_id = ?", openID).First(&user)
	isNewUser := result.Error != nil

	if isNewUser {
		user = model.User{OpenID: openID}
		if err := db.Create(&user).Error; err != nil {
			log.Printf("[WxLogin] create user failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed"})
			return
		}

		// Handle invitation
		if req.InviterID > 0 && h.ShareSvc != nil {
			if err := h.ShareSvc.HandleInvite(req.InviterID, user.ID); err != nil {
				log.Printf("[WxLogin] handle invite failed: inviter=%d, err=%v", req.InviterID, err)
			}
		}
	}

	token, err := middleware.GenerateToken(user.ID, user.OpenID, h.Cfg.GetJWTSecret(), h.Cfg.GetJWTExpire())
	if err != nil {
		log.Printf("[WxLogin] generate token failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generate token failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":       token,
		"user_id":     user.ID,
		"has_profile": user.Height > 0 && user.Weight > 0,
	})
}

// ==================== User Profile ====================

func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	user := h.Svc.Food.GetUser(userID)
	if user.ID == 0 {
		log.Printf("[GetProfile] user not found: id=%d", userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

type UpdateProfileReq struct {
	Nickname  string  `json:"nickname"`
	AvatarURL string  `json:"avatar_url"`
	Height    float64 `json:"height" binding:"required"`
	Weight    float64 `json:"weight" binding:"required"`
	Age       int     `json:"age" binding:"required"`
	Gender    int     `json:"gender" binding:"required,oneof=1 2"`
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[UpdateProfile] bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Svc.Food.DB().Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"nickname":   req.Nickname,
		"avatar_url": req.AvatarURL,
		"height":     req.Height,
		"weight":     req.Weight,
		"age":        req.Age,
		"gender":     req.Gender,
	}).Error; err != nil {
		log.Printf("[UpdateProfile] db update failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

// ==================== Food Analysis ====================

func (h *Handler) AnalyzeImage(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		ImageKey string `json:"image_key" binding:"required"`
		MealType int    `json:"meal_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AnalyzeImage] bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "image_key is required"})
		return
	}

	if req.MealType == 0 {
		req.MealType = 1
	}

	// Build full COS URL from image_key
	imageURL := h.COS.ObjectURL(req.ImageKey)

	records, suggestion, err := h.Svc.Food.AnalyzeImage(c.Request.Context(), userID, req.MealType, imageURL)
	if err != nil {
		if err.Error() == "not_food" {
			c.JSON(http.StatusOK, gin.H{"is_food": false, "records": nil, "suggestion": ""})
			return
		}
		log.Printf("[AnalyzeImage] AI analysis failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed: " + err.Error()})
		return
	}

	stampText := service.ComputeStampText(records)

	c.JSON(http.StatusOK, gin.H{"is_food": true, "records": records, "suggestion": suggestion, "stamp_text": stampText})

	// Trigger score and achievement check asynchronously
	go func() {
		if h.ScoreSvc != nil {
			if err := h.ScoreSvc.OnAnalyze(userID); err != nil {
				log.Printf("[AnalyzeImage] score update failed: user_id=%d, err=%v", userID, err)
			}
		}
		if h.ShareSvc != nil {
			h.ShareSvc.HandleInviteeFirstAnalyze(userID)
		}
		if h.AchieveSvc != nil {
			h.AchieveSvc.CheckAndUnlock(userID)
		}
	}()
}

type AnalyzeTextReq struct {
	Description string `json:"description" binding:"required"`
	MealType    int    `json:"meal_type"`
}

func (h *Handler) AnalyzeText(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req AnalyzeTextReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AnalyzeText] bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MealType == 0 {
		req.MealType = 1
	}

	records, suggestion, err := h.Svc.Food.AnalyzeText(c.Request.Context(), userID, req.MealType, req.Description)
	if err != nil {
		if err.Error() == "not_food" {
			c.JSON(http.StatusOK, gin.H{"is_food": false, "records": nil, "suggestion": ""})
			return
		}
		log.Printf("[AnalyzeText] AI analysis failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed: " + err.Error()})
		return
	}

	stampText := service.ComputeStampText(records)

	c.JSON(http.StatusOK, gin.H{"is_food": true, "records": records, "suggestion": suggestion, "stamp_text": stampText})

	// Trigger score and achievement check asynchronously
	go func() {
		if h.ScoreSvc != nil {
			if err := h.ScoreSvc.OnAnalyze(userID); err != nil {
				log.Printf("[AnalyzeText] score update failed: user_id=%d, err=%v", userID, err)
			}
		}
		if h.ShareSvc != nil {
			h.ShareSvc.HandleInviteeFirstAnalyze(userID)
		}
		if h.AchieveSvc != nil {
			h.AchieveSvc.CheckAndUnlock(userID)
		}
	}()
}

func (h *Handler) ListFoodRecords(c *gin.Context) {
	userID := c.GetUint("user_id")
	date := c.Query("date")
	mealType, _ := strconv.Atoi(c.DefaultQuery("meal_type", "0"))

	records, err := h.Svc.Food.ListRecords(userID, date, mealType)
	if err != nil {
		log.Printf("[ListFoodRecords] query failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"records": records})
}

func (h *Handler) GetFoodRecord(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	record, err := h.Svc.Food.GetRecord(userID, id)
	if err != nil {
		log.Printf("[GetFoodRecord] not found: user_id=%d, id=%s, err=%v", userID, id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (h *Handler) DeleteFoodRecord(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	if err := h.Svc.Food.DeleteRecord(userID, id); err != nil {
		log.Printf("[DeleteFoodRecord] delete failed: user_id=%d, id=%s, err=%v", userID, id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) DailySummary(c *gin.Context) {
	userID := c.GetUint("user_id")
	date := c.Query("date")

	summary, err := h.Svc.Food.DailySummary(userID, date)
	if err != nil {
		log.Printf("[DailySummary] query failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) DailyAnalysis(c *gin.Context) {
	userID := c.GetUint("user_id")
	date := c.Query("date")

	analysis, err := h.Svc.Food.GetDailyAnalysis(c.Request.Context(), userID, date)
	if err != nil {
		log.Printf("[DailyAnalysis] failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, analysis)
}

func (h *Handler) MonthlySummary(c *gin.Context) {
	userID := c.GetUint("user_id")
	month := c.Query("month") // format: 2006-01

	summary, err := h.Svc.Food.MonthlySummary(userID, month)
	if err != nil {
		log.Printf("[MonthlySummary] query failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// ==================== Wheel ====================

func (h *Handler) ListDishes(c *gin.Context) {
	userID := c.GetUint("user_id")
	category := c.Query("category")

	dishes, err := h.Svc.Wheel.ListDishes(userID, category)
	if err != nil {
		log.Printf("[ListDishes] query failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"dishes": dishes})
}

func (h *Handler) SpinWheel(c *gin.Context) {
	userID := c.GetUint("user_id")
	category := c.DefaultQuery("category", "")

	dish, total, err := h.Svc.Wheel.Spin(userID, category)
	if err != nil {
		log.Printf("[SpinWheel] spin failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spin failed"})
		return
	}
	if dish == nil {
		log.Printf("[SpinWheel] no dishes available: user_id=%d", userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "no dishes available"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dish": dish, "total": total})

	// Award spin score asynchronously
	go func() {
		if h.ScoreSvc != nil {
			h.ScoreSvc.AddScore(userID, "spin", 2, "使用转盘")
		}
	}()
}

type CreateDishReq struct {
	Name     string  `json:"name" binding:"required"`
	Category string  `json:"category"`
	Calories float64 `json:"calories"`
	Weight   int     `json:"weight"`
	Recipe   string  `json:"recipe"`
}

func (h *Handler) CreateDish(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req CreateDishReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[CreateDish] bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Weight <= 0 {
		req.Weight = 50
	}
	if req.Category == "" {
		req.Category = "其他"
	}

	dish := &model.Dish{
		Name:     req.Name,
		Category: req.Category,
		Calories: req.Calories,
		Weight:   req.Weight,
		IsSystem: false,
		UserID:   userID,
		Recipe:   req.Recipe,
	}
	if err := h.Svc.Wheel.CreateDish(dish); err != nil {
		log.Printf("[CreateDish] create failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create dish failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"dish": dish})
}

func (h *Handler) CreateDishAI(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		Name   string `json:"name" binding:"required"`
		Weight int    `json:"weight"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[CreateDishAI] bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dish, err := h.Svc.Wheel.CreateDishWithAI(c.Request.Context(), userID, req.Name, req.Weight)
	if err != nil {
		if err.Error() == "not_food" {
			c.JSON(http.StatusOK, gin.H{"is_food": false})
			return
		}
		log.Printf("[CreateDishAI] failed: user_id=%d, name=%s, err=%v", userID, req.Name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_food": true, "dish": dish})
}

func (h *Handler) DeleteDish(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	if err := h.Svc.Wheel.DeleteDish(userID, id); err != nil {
		log.Printf("[DeleteDish] delete failed: user_id=%d, id=%s, err=%v", userID, id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ==================== Rank Visibility ====================

func (h *Handler) ToggleRank(c *gin.Context) {
	userID := c.GetUint("user_id")
	var user model.User
	if err := h.Svc.Food.DB().Select("id, show_on_rank").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	newVal := !user.ShowOnRank
	h.Svc.Food.DB().Model(&user).Update("show_on_rank", newVal)
	c.JSON(http.StatusOK, gin.H{"show_on_rank": newVal})
}

// ==================== Privacy ====================

type AgreePrivacyReq struct {
	AgreementType string `json:"agreement_type" binding:"required"`
	Version       string `json:"version"`
}

func (h *Handler) AgreePrivacy(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req AgreePrivacyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[AgreePrivacy] bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agreement := model.PrivacyAgreement{
		UserID:        userID,
		AgreementType: req.AgreementType,
		Version:       req.Version,
		AgreedAt:      time.Now(),
	}
	if err := h.Svc.Food.DB().Create(&agreement).Error; err != nil {
		log.Printf("[AgreePrivacy] db create failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save agreement failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "agreed"})
}

func (h *Handler) PrivacyStatus(c *gin.Context) {
	userID := c.GetUint("user_id")
	var count int64
	h.Svc.Food.DB().Model(&model.PrivacyAgreement{}).Where("user_id = ?", userID).Count(&count)
	c.JSON(http.StatusOK, gin.H{"agreed": count > 0})
}

// ==================== Upload ====================

type PresignUploadReq struct {
	BizType  string `json:"biz_type" binding:"required"`
	Filename string `json:"filename" binding:"required"`
}

func (h *Handler) PresignUpload(c *gin.Context) {
	var req PresignUploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[PresignUpload] bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uploadURL, objectKey, objectURL, err := h.COS.PresignPutURL(c.Request.Context(), cos.BizType(req.BizType), req.Filename, 30*time.Minute)
	if err != nil {
		log.Printf("[PresignUpload] presign failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_url": uploadURL,
		"object_key": objectKey,
		"object_url": objectURL,
	})
}

func (h *Handler) UploadImage(c *gin.Context) {
	// 5MB limit
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5<<20)

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image is required or too large (max 5MB)"})
		return
	}
	defer file.Close()

	bizType := cos.BizType(c.DefaultPostForm("biz_type", string(cos.BizFeedback)))
	result, err := h.COS.Upload(c.Request.Context(), bizType, header.Filename, file)
	if err != nil {
		log.Printf("[UploadImage] COS upload failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"object_key": result.ObjectKey,
		"object_url": result.ObjectURL,
	})
}

// ==================== Feedback ====================

func (h *Handler) CreateFeedback(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		Type        string          `json:"type" binding:"required"`
		Subject     string          `json:"subject" binding:"required"`
		Content     string          `json:"content" binding:"required"`
		Attachments model.JSONArray `json:"attachments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[CreateFeedback] bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	feedback := model.Feedback{
		UserID:      userID,
		Type:        req.Type,
		Subject:     req.Subject,
		Content:     req.Content,
		Attachments: req.Attachments,
	}
	if err := h.Svc.Food.DB().Create(&feedback).Error; err != nil {
		log.Printf("[CreateFeedback] create failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create feedback failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"feedback": feedback})
}

func (h *Handler) ListFeedback(c *gin.Context) {
	userID := c.GetUint("user_id")
	var list []model.Feedback
	if err := h.Svc.Food.DB().Where("user_id = ?", userID).Order("created_at DESC").Find(&list).Error; err != nil {
		log.Printf("[ListFeedback] query failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"feedbacks": list})
}

// ==================== Score / Rank ====================

func (h *Handler) GetMyScore(c *gin.Context) {
	userID := c.GetUint("user_id")
	if h.ScoreSvc == nil {
		c.JSON(http.StatusOK, gin.H{"total_score": 0, "week_score": 0, "streak_days": 0, "total_rank": 0, "week_rank": 0})
		return
	}
	c.JSON(http.StatusOK, h.ScoreSvc.GetMyScore(userID))
}

func (h *Handler) GetRank(c *gin.Context) {
	scope := c.DefaultQuery("scope", "week")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit > 100 {
		limit = 100
	}

	if h.ScoreSvc == nil {
		c.JSON(http.StatusOK, gin.H{"rank": []interface{}{}})
		return
	}

	if scope == "friend" {
		userID := c.GetUint("user_id")
		c.JSON(http.StatusOK, gin.H{"rank": h.ScoreSvc.GetFriendRank(userID, offset, limit)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rank": h.ScoreSvc.GetRank(scope, offset, limit)})
}

func (h *Handler) GetScoreLogs(c *gin.Context) {
	userID := c.GetUint("user_id")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if h.ScoreSvc == nil {
		c.JSON(http.StatusOK, gin.H{"logs": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": h.ScoreSvc.GetScoreLogs(userID, offset, limit)})
}

// ==================== Share ====================

func (h *Handler) RecordShare(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		ShareType string `json:"share_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.ShareType = "poster"
	}

	if h.ShareSvc == nil {
		c.JSON(http.StatusOK, gin.H{"message": "recorded"})
		return
	}

	if err := h.ShareSvc.RecordShare(userID, req.ShareType); err != nil {
		log.Printf("[RecordShare] failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "record share failed"})
		return
	}

	// Check achievements after share
	var newUnlocks []interface{}
	if h.AchieveSvc != nil {
		newUnlocks = make([]interface{}, 0)
		for _, a := range h.AchieveSvc.CheckAndUnlock(userID) {
			newUnlocks = append(newUnlocks, a)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "recorded", "new_achievements": newUnlocks})
}

// ==================== Share Poster ====================

func (h *Handler) GetSharePoster(c *gin.Context) {
	var share model.Share
	if err := h.Svc.Food.DB().Where("type = ?", "poster").First(&share).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no poster available"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"image_url": share.ImageURL})
}

// ==================== Achievement ====================

func (h *Handler) ListAchievements(c *gin.Context) {
	userID := c.GetUint("user_id")
	if h.AchieveSvc == nil {
		c.JSON(http.StatusOK, gin.H{"achievements": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"achievements": h.AchieveSvc.ListAchievements(userID)})
}

func (h *Handler) SetTitle(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		AchievementID uint `json:"achievement_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "achievement_id is required"})
		return
	}

	if h.AchieveSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "achievement service not available"})
		return
	}

	if err := h.AchieveSvc.SetTitle(userID, req.AchievementID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "title updated"})
}

func (h *Handler) CheckAchievements(c *gin.Context) {
	userID := c.GetUint("user_id")
	if h.AchieveSvc == nil {
		c.JSON(http.StatusOK, gin.H{"new_achievements": []interface{}{}})
		return
	}
	newUnlocks := h.AchieveSvc.CheckAndUnlock(userID)
	c.JSON(http.StatusOK, gin.H{"new_achievements": newUnlocks})
}

// ==================== Gitea Webhook ====================

func (h *Handler) GiteaWebhook(c *gin.Context) {
	// Read and parse Gitea push event body
	body, err := c.GetRawData()
	if err != nil {
		log.Printf("[GiteaWebhook] read body failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var giteaEvent struct {
		Before     string `json:"before"`
		After      string `json:"after"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(body, &giteaEvent); err != nil {
		log.Printf("[GiteaWebhook] parse body failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	repoName := giteaEvent.Repository.FullName
	if repoName == "" || h.WebhookSvc == nil {
		log.Printf("[GiteaWebhook] missing repo name or webhook service unavailable")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing repository info"})
		return
	}

	// Look up project by repo full name in database
	project, err := h.WebhookSvc.FindProjectByRepoName(repoName)
	if err != nil {
		log.Printf("[GiteaWebhook] project not found for repo=%s, err=%v", repoName, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "project not configured for " + repoName})
		return
	}

	log.Printf("[GiteaWebhook] matched project=%s before=%s after=%s", project.Name, giteaEvent.Before, giteaEvent.After)

	// Process webhook asynchronously
	go func() {
		if err := h.WebhookSvc.ProcessWebhook(project, giteaEvent.Before, giteaEvent.After); err != nil {
			log.Printf("[Webhook] process failed for project=%s: %v", project.Name, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "webhook received",
		"project": project.Name,
	})
}
