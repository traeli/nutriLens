package router

import (
	"net/http"
	"path/filepath"

	"shijibu/internal/handler"
	"shijibu/internal/middleware"
	platformauth "shijibu/internal/platform/auth"

	"github.com/gin-gonic/gin"
)

func New(h *handler.Handler, tokens *platformauth.TokenManager, allowedOrigins []string, uploadDir ...string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.RequestID(), middleware.CORS(allowedOrigins))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "nutrilens-api"})
	})
	if len(uploadDir) > 0 && uploadDir[0] != "" {
		r.Static("/uploads", filepath.Join(uploadDir[0], "public"))
	}

	v1 := r.Group("/api/v1")
	v1.POST("/auth/wx-login", h.Login)
	v1.GET("/cities", h.Cities)
	v1.GET("/home/summary", h.HomeSummary)
	v1.GET("/routes/nearby", h.NearbyRoute)
	v1.GET("/routes/:id", h.GetRoute)
	v1.GET("/cities/:code/poster-theme", h.CityPosterTheme)
	v1.GET("/places/search", h.SearchPlaces)
	v1.GET("/places/:id", h.GetPlace)
	v1.GET("/places/:id/experiences", h.PlaceExperiences)
	v1.GET("/experiences", h.Experiences)
	v1.GET("/experiences/:id", h.GetExperience)
	v1.GET("/tags", h.Tags)
	v1.GET("/agreements/current", h.CurrentAgreements)

	authorized := v1.Group("")
	if h.Accounts != nil {
		authorized.Use(middleware.Authenticate(tokens, h.Accounts))
	} else {
		authorized.Use(middleware.Authenticate(tokens))
	}
	authorized.GET("/user/profile", h.GetProfile)
	authorized.PUT("/user/profile", h.UpdateProfile)
	authorized.DELETE("/user/account", h.DeleteAccount)
	authorized.POST("/agreements/accept", h.AcceptAgreement)
	authorized.POST("/records", h.CreateRecord)
	authorized.POST("/places", h.SubmitPlace)
	authorized.GET("/records/:id", h.GetRecord)
	authorized.PATCH("/records/:id", h.UpdateRecord)
	authorized.DELETE("/records/:id", h.DeleteRecord)
	authorized.POST("/records/:id/submit-public", h.SubmitRecord)
	authorized.GET("/records/:id/review-status", h.RecordReviewStatus)
	authorized.GET("/me/records", h.MyRecords)
	authorized.GET("/me/footprints/summary", h.FootprintSummary)
	authorized.GET("/me/footprints/map", h.FootprintMap)
	authorized.GET("/me/contribution", h.Contribution)
	authorized.GET("/me/badges", h.Badges)
	authorized.POST("/speech/transcribe", h.TranscribeSpeech)
	authorized.POST("/publisher-verification/phone", h.VerifyPublisherPhone)
	authorized.GET("/publisher-verification/status", h.PublisherVerificationStatus)
	authorized.POST("/records/:id/media", h.UploadRecordMedia)
	authorized.DELETE("/records/:id/media/:media_id", h.DeleteRecordMedia)
	authorized.POST("/records/:id/evidences", h.UploadEvidence)
	authorized.POST("/places/:id/favorite", h.FavoritePlace)
	authorized.DELETE("/places/:id/favorite", h.UnfavoritePlace)
	authorized.GET("/me/favorite-places", h.FavoritePlaces)
	authorized.POST("/experiences/:id/helpful", h.Helpful)
	authorized.DELETE("/experiences/:id/helpful", h.Unhelpful)
	authorized.POST("/experiences/:id/outdated", h.MarkOutdated)
	authorized.POST("/reports", h.CreateReport)
	authorized.GET("/me/reports", h.MyReports)
	authorized.POST("/records/:id/appeals", h.CreateAppeal)
	authorized.GET("/me/appeals", h.MyAppeals)
	authorized.POST("/routes/:id/favorite", h.FavoriteRoute)
	authorized.DELETE("/routes/:id/favorite", h.UnfavoriteRoute)
	authorized.POST("/routes/:id/start", h.StartRoute)
	authorized.PATCH("/route-journeys/:id", h.UpdateJourney)
	authorized.GET("/me/route-journeys", h.MyJourneys)
	authorized.POST("/nutrition/records", h.CreateNutritionRecord)
	authorized.GET("/nutrition/records", h.NutritionRecords)
	authorized.DELETE("/nutrition/records/:id", h.DeleteNutritionRecord)
	authorized.GET("/nutrition/summary", h.NutritionSummary)
	return r
}
