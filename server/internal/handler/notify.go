package handler

import (
	"log"
	"net/http"

	"nutrilens/internal/service"

	"github.com/gin-gonic/gin"
)

// WechatVerify handles WeChat server token verification (GET request).
// WeChat sends signature, timestamp, nonce, echostr — we verify and return echostr.
func (h *Handler) WechatVerify(c *gin.Context) {
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")

	if signature == "" || timestamp == "" || nonce == "" || echostr == "" {
		c.String(400, "missing parameters")
		return
	}

	if !h.Svc.Auth.VerifySignature(timestamp, nonce, signature) {
		c.String(403, "signature verification failed")
		return
	}

	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(echostr))
}

func (h *Handler) GetNotifySettings(c *gin.Context) {
	userID := c.GetUint("user_id")

	setting, err := h.NotifySvc.GetOrCreateSettings(userID)
	if err != nil {
		log.Printf("[GetNotifySettings] failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get settings failed"})
		return
	}
	c.JSON(http.StatusOK, setting)
}

func (h *Handler) UpdateNotifySettings(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req service.UpdateNotifySettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[UpdateNotifySettings] bad request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.NotifySvc.UpdateSettings(userID, req); err != nil {
		log.Printf("[UpdateNotifySettings] failed: user_id=%d, err=%v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "settings updated"})
}
