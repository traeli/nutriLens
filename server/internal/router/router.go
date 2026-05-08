package router

import (
	"nutrilens/internal/handler"
	"nutrilens/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, jwtSecret string, h *handler.Handler) {
	r.Use(middleware.CORS())

	api := r.Group("/api/v1")
	{
		// Public routes
		auth := api.Group("/auth")
		{
			auth.POST("/wx-login", h.WxLogin)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.JWTAuth(jwtSecret))
		{
			// User profile
			protected.GET("/user/profile", h.GetProfile)
			protected.PUT("/user/profile", h.UpdateProfile)

			// Food analysis
			protected.POST("/food/analyze/image", h.AnalyzeImage)
			protected.POST("/food/analyze/text", h.AnalyzeText)
			protected.GET("/food/records", h.ListFoodRecords)
			protected.GET("/food/records/:id", h.GetFoodRecord)
			protected.DELETE("/food/records/:id", h.DeleteFoodRecord)
			protected.GET("/food/daily-summary", h.DailySummary)
			protected.GET("/food/monthly-summary", h.MonthlySummary)

			// Wheel
			protected.GET("/wheel/dishes", h.ListDishes)
			protected.POST("/wheel/spin", h.SpinWheel)
			protected.POST("/wheel/dishes", h.CreateDish)
			protected.POST("/wheel/dishes/ai", h.CreateDishAI)
			protected.DELETE("/wheel/dishes/:id", h.DeleteDish)

			// Privacy
			protected.POST("/privacy/agree", h.AgreePrivacy)
			protected.GET("/privacy/status", h.PrivacyStatus)

			// Upload
			protected.POST("/upload/image", h.UploadImage)
		}
	}
}
