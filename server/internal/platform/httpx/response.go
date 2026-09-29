package httpx

import "github.com/gin-gonic/gin"

const RequestIDKey = "request_id"

type ErrorBody struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"request_id"`
	Details   map[string]any `json:"details"`
}

func OK(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

func Error(c *gin.Context, status int, code, message string) {
	requestID, _ := c.Get(RequestIDKey)
	c.AbortWithStatusJSON(status, gin.H{"error": ErrorBody{Code: code, Message: message, RequestID: stringValue(requestID), Details: map[string]any{}}})
}

func stringValue(value any) string {
	result, _ := value.(string)
	return result
}
