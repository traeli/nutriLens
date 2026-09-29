package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"shijibu/internal/middleware"
	"shijibu/internal/platform/httpx"
)

func (h *Handler) UploadRecordMedia(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "请选择图片")
		return
	}
	item, err := h.Media.AddRecordMedia(middleware.UserID(c), id, file, c.PostForm("media_type"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, item)
}

func (h *Handler) UploadEvidence(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "请选择凭证图片")
		return
	}
	item, err := h.Media.AddEvidence(middleware.UserID(c), id, file, c.PostForm("evidence_type"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusCreated, item)
}

func (h *Handler) DeleteRecordMedia(c *gin.Context) {
	recordID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var uri struct {
		MediaID uint `uri:"media_id" binding:"required"`
	}
	if c.ShouldBindUri(&uri) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "图片 ID 不正确")
		return
	}
	if err := h.Media.DeleteRecordMedia(middleware.UserID(c), recordID, uri.MediaID); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
