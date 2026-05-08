package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"nutrilens/internal/middleware"
	"nutrilens/internal/model"
	"nutrilens/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Svc *Services
	Cfg ConfigProvider
}

type Services struct {
	Auth  *service.WechatService
	Food  *service.FoodService
	Wheel *service.WheelService
}

type ConfigProvider interface {
	GetJWTSecret() string
	GetJWTExpire() int
	GetUploadDir() string
}

// ==================== Auth ====================

type WxLoginReq struct {
	Code string `json:"code" binding:"required"`
}

func (h *Handler) WxLogin(c *gin.Context) {
	var req WxLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[WxLogin] bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	openID, _, err := h.Svc.Auth.Code2Session(req.Code)
	if err != nil {
		log.Printf("[WxLogin] code2Session failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "wechat login failed: " + err.Error()})
		return
	}

	db := h.Svc.Food.DB()
	var user model.User
	result := db.Where("open_id = ?", openID).First(&user)
	if result.Error != nil {
		user = model.User{OpenID: openID}
		if err := db.Create(&user).Error; err != nil {
			log.Printf("[WxLogin] create user failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed"})
			return
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
	Nickname string  `json:"nickname"`
	Height   float64 `json:"height" binding:"required"`
	Weight   float64 `json:"weight" binding:"required"`
	Age      int     `json:"age" binding:"required"`
	Gender   int     `json:"gender" binding:"required,oneof=1 2"`
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
		"nickname": req.Nickname,
		"height":   req.Height,
		"weight":   req.Weight,
		"age":      req.Age,
		"gender":   req.Gender,
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

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		log.Printf("[AnalyzeImage] missing image: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "image is required"})
		return
	}
	defer file.Close()

	uploadDir := h.Cfg.GetUploadDir()
	filename := fmt.Sprintf("%d_%d_%s", userID, time.Now().UnixMilli(), header.Filename)
	filePath := filepath.Join(uploadDir, filename)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("[AnalyzeImage] mkdir failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create upload dir failed"})
		return
	}
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("[AnalyzeImage] create file failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save image failed"})
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	fileBytes, _ := os.ReadFile(filePath)
	mealType, _ := strconv.Atoi(c.DefaultPostForm("meal_type", "1"))

	record, err := h.Svc.Food.AnalyzeImage(c.Request.Context(), userID, mealType, fileBytes, "/uploads/"+filename)
	if err != nil {
		if err.Error() == "not_food" {
			c.JSON(http.StatusOK, gin.H{"is_food": false})
			return
		}
		log.Printf("[AnalyzeImage] AI analysis failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI analysis failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_food": true, "record": record})
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

	c.JSON(http.StatusOK, gin.H{"is_food": true, "records": records, "suggestion": suggestion})
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

func (h *Handler) UploadImage(c *gin.Context) {
	userID := c.GetUint("user_id")
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		log.Printf("[UploadImage] missing image: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "image is required"})
		return
	}
	defer file.Close()

	uploadDir := h.Cfg.GetUploadDir()
	filename := fmt.Sprintf("%d_%d_%s", userID, time.Now().UnixMilli(), filepath.Base(header.Filename))
	filePath := filepath.Join(uploadDir, filename)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("[UploadImage] mkdir failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create dir failed"})
		return
	}
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("[UploadImage] create file failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	c.JSON(http.StatusOK, gin.H{
		"url":      "/uploads/" + filename,
		"filename": filename,
	})
}
