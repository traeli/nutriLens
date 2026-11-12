package router

import (
	"net/http"
	"path/filepath"

	"shijibu/internal/handler"
	"shijibu/internal/middleware"

	"github.com/gin-gonic/gin"
)

func New(h *handler.Handler, tokens middleware.TokenAuthenticator, allowedOrigins []string, uploadDir ...string) *gin.Engine {
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
	v1.POST("/auth/login", h.Login)

	v1.POST("/auth/refresh", h.RefreshToken)

	// Public discovery endpoints contain only published or editorial content. Keep
	// them outside the authenticated group so a visitor can understand and browse
	// the product before deciding whether to sign in.
	v1.GET("/cities", h.Cities)
	v1.GET("/home/summary", h.HomeSummary)
	v1.GET("/routes/nearby", h.NearbyRoute)
	v1.GET("/routes/:id", h.GetRoute)
	v1.GET("/cities/:code/poster-theme", h.CityPosterTheme)
	v1.GET("/places/search", h.SearchPlaces)
	v1.GET("/places/:id", h.GetPlace)
	v1.GET("/discovery/posts", h.DiscoveryPosts)
	v1.GET("/discovery/posts/:id", h.GetDiscoveryPost)
	v1.GET("/places/:id/discovery-posts", h.PlaceDiscoveryPosts)
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
	authorized.POST("/places/:id/favorite", h.FavoritePlace)
	authorized.DELETE("/places/:id/favorite", h.UnfavoritePlace)
	authorized.GET("/me/favorite-places", h.FavoritePlaces)
	authorized.POST("/discovery/posts/:id/like", h.LikeDiscoveryPost)
	authorized.DELETE("/discovery/posts/:id/like", h.UnlikeDiscoveryPost)
	authorized.GET("/me/discovery-post-likes", h.MyDiscoveryPostLikes)
	authorized.POST("/reports", h.CreateReport)
	authorized.GET("/me/reports", h.MyReports)
	authorized.POST("/routes/:id/favorite", h.FavoriteRoute)
	authorized.DELETE("/routes/:id/favorite", h.UnfavoriteRoute)
	authorized.POST("/routes/:id/start", h.StartRoute)
	authorized.PATCH("/route-journeys/:id", h.UpdateJourney)
	authorized.GET("/me/route-journeys", h.MyJourneys)
	return r
}
