package handler

import (
	"net/http"
	"strconv"

	"shijibu/internal/middleware"
	"shijibu/internal/platform/httpx"
	"shijibu/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Cities(c *gin.Context) {
	items, err := h.Discovery.Cities()
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"items": items})
}

func (h *Handler) CityPosterTheme(c *gin.Context) {
	result, err := h.Discovery.CityPosterTheme(c.Param("code"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, result)
}

func (h *Handler) HomeSummary(c *gin.Context) {
	result, err := h.Discovery.HomeSummary(c.Query("city_code"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, result)
}

func (h *Handler) SearchPlaces(c *gin.Context) {
	page, err := h.Discovery.SearchPlaces(c.Query("city_code"), c.Query("q"), c.Query("cursor"), queryLimit(c), c.Query("recommended") == "1")
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, page)
}

func (h *Handler) GetPlace(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	place, err := h.Discovery.GetPlace(id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, place)
}

type placeSubmissionRequest struct {
	Name         string  `json:"name" binding:"required"`
	CityCode     string  `json:"city_code" binding:"required"`
	Category     string  `json:"category"`
	District     string  `json:"district"`
	BusinessArea string  `json:"business_area"`
	Address      string  `json:"address" binding:"required"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	POIProvider  string  `json:"poi_provider"`
}

func (h *Handler) SubmitPlace(c *gin.Context) {
	var request placeSubmissionRequest
	if c.ShouldBindJSON(&request) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "地点信息格式不正确")
		return
	}
	place, err := h.Records.SubmitPlace(middleware.UserID(c), service.PlaceSubmissionInput{
		Name: request.Name, CityCode: request.CityCode, Category: request.Category,
		District: request.District, BusinessArea: request.BusinessArea, Address: request.Address,
		Longitude: request.Longitude, Latitude: request.Latitude, POIProvider: request.POIProvider,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, place)
}

func (h *Handler) GetRoute(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	result, err := h.Discovery.GetRoute(id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, result)
}

func (h *Handler) NearbyRoute(c *gin.Context) {
	longitude, longitudeErr := strconv.ParseFloat(c.Query("longitude"), 64)
	latitude, latitudeErr := strconv.ParseFloat(c.Query("latitude"), 64)
	if longitudeErr != nil || latitudeErr != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "当前位置坐标不正确")
		return
	}
	result, err := h.Discovery.NearbyRoute(c.Query("city_code"), longitude, latitude, queryLimit(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, result)
}

func (h *Handler) Experiences(c *gin.Context) {
	page, err := h.Discovery.ListExperiences(c.Query("city_code"), c.Query("cursor"), queryLimit(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, page)
}

func (h *Handler) GetExperience(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	view, err := h.Discovery.GetExperience(id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, view)
}

func (h *Handler) PlaceExperiences(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	page, err := h.Discovery.ListPlaceExperiences(id, c.Query("cursor"), queryLimit(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, page)
}

func (h *Handler) Tags(c *gin.Context) {
	items, err := h.Discovery.Tags()
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"items": items})
}

func queryLimit(c *gin.Context) int {
	value, _ := strconv.Atoi(c.Query("limit"))
	return value
}
