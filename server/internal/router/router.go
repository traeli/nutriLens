package router

import (
	"nutrilens/internal/cache"
	"nutrilens/internal/handler"
	"nutrilens/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, jwtSecret string, h *handler.Handler, limiter *cache.RateLimiter, getTag middleware.UserTagGetter, dailyLimit int64) {
	r.Use(middleware.CORS())

	// WeChat server verification (no auth)
	r.GET("/", h.WechatVerify)
	r.GET("/wx/callback", h.WechatVerify)

	// Gitea webhook (no auth)
	r.POST("/webhook/gitea", h.GiteaWebhook)

	api := r.Group("/api/v1")
	{
		// WeChat server verification (no auth, used by public nginx proxy)
		api.GET("/wx/callback", h.WechatVerify)

		// Share poster (public)
		api.GET("/share/poster", h.GetSharePoster)

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
			protected.PUT("/user/rank-visibility", h.ToggleRank)

			// Food (non-AI)
			protected.GET("/food/records", h.ListFoodRecords)
			protected.GET("/food/records/:id", h.GetFoodRecord)
			protected.DELETE("/food/records/:id", h.DeleteFoodRecord)
			protected.GET("/food/daily-summary", h.DailySummary)
			protected.GET("/food/monthly-summary", h.MonthlySummary)
			protected.GET("/food/daily-analysis", h.DailyAnalysis)

			// Food AI analysis (rate limited)
			aiFood := protected.Group("")
			aiFood.Use(middleware.RateLimit(limiter, getTag, dailyLimit))
			{
				aiFood.POST("/food/analyze/image", h.AnalyzeImage)
				aiFood.POST("/food/analyze/text", h.AnalyzeText)
			}

			// Wheel
			protected.GET("/wheel/dishes", h.ListDishes)
			protected.POST("/wheel/spin", h.SpinWheel)
			protected.POST("/wheel/dishes", h.CreateDish)

			// Wheel AI dish creation (rate limited)
			aiWheel := protected.Group("")
			aiWheel.Use(middleware.RateLimit(limiter, getTag, dailyLimit))
			{
				aiWheel.POST("/wheel/dishes/ai", h.CreateDishAI)
			}

			protected.DELETE("/wheel/dishes/:id", h.DeleteDish)

			// Privacy
			protected.POST("/privacy/agree", h.AgreePrivacy)
			protected.GET("/privacy/status", h.PrivacyStatus)

			// Upload
			protected.POST("/upload/presign", h.PresignUpload)
			protected.POST("/upload/image", h.UploadImage)

			// Feedback
			protected.POST("/feedback", h.CreateFeedback)
			protected.GET("/feedback", h.ListFeedback)

			// Score / Rank
			protected.GET("/score/my", h.GetMyScore)
			protected.GET("/score/rank", h.GetRank)
			protected.GET("/score/log", h.GetScoreLogs)

			// Share
			protected.POST("/share/record", h.RecordShare)

			// Achievement
			protected.GET("/achievement/list", h.ListAchievements)
			protected.PUT("/achievement/title", h.SetTitle)
			protected.POST("/achievement/check", h.CheckAchievements)

			// Notify
			protected.GET("/notify/settings", h.GetNotifySettings)
			protected.PUT("/notify/settings", h.UpdateNotifySettings)
		}
	}
}
