package middleware

import (
	"net/http"
	"strings"

	platformauth "shijibu/internal/platform/auth"
	"shijibu/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

func Authenticate(tokens *platformauth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(value, "Bearer ") {
			httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录")
			return
		}
		userID, err := tokens.Parse(strings.TrimSpace(strings.TrimPrefix(value, "Bearer ")))
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "登录状态已失效")
			return
		}
		c.Set(UserIDKey, userID)
		c.Next()
	}
}

func UserID(c *gin.Context) uint {
	value, _ := c.Get(UserIDKey)
	userID, _ := value.(uint)
	return userID
}
