package middleware

import (
	"context"
	"net/http"
	"strings"

	platformauth "shijibu/internal/platform/auth"
	"shijibu/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

type AccountStatusChecker interface {
	IsAccountActive(context.Context, uint) (bool, error)
}

func Authenticate(tokens *platformauth.TokenManager, statusCheckers ...AccountStatusChecker) gin.HandlerFunc {
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
		if len(statusCheckers) > 0 && statusCheckers[0] != nil {
			active, err := statusCheckers[0].IsAccountActive(c.Request.Context(), userID)
			if err != nil {
				httpx.Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "服务暂时不可用")
				return
			}
			if !active {
				httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "账号已注销或不可用")
				return
			}
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
