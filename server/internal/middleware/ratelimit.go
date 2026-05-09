package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"nutrilens/internal/cache"

	"github.com/gin-gonic/gin"
)

// UserTagGetter returns the tag string for a given user ID.
// This keeps the middleware decoupled from the database layer.
type UserTagGetter func(userID uint) string

// RateLimit returns a Gin middleware that enforces daily AI call limits.
// Users with tag "vip" bypass the limit entirely.
func RateLimit(limiter *cache.RateLimiter, getTag UserTagGetter, limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// VIP users bypass rate limiting
		tag := getTag(userID)
		if tag == "vip" {
			c.Next()
			return
		}

		identifier := strconv.FormatUint(uint64(userID), 10)
		count, remaining, err := limiter.Allow(c.Request.Context(), identifier, limit)
		if err != nil {
			// Redis error — fail open to avoid blocking all users
			fmt.Printf("[RateLimit] redis error: user_id=%d, err=%v\n", userID, err)
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-RateLimit-Used", strconv.FormatInt(count, 10))

		if count > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":     "今日AI分析次数已达上限",
				"limit":     limit,
				"used":      count,
				"remaining": 0,
			})
			return
		}

		c.Next()
	}
}
