package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBindRecordInputAcceptsQuickRecordPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{
		"place_id":42,
		"visit_date":"2026-09-16",
		"consumer_type":"self",
		"conclusion":"hot",
		"average_cost":null,
		"wait_minutes":null,
		"dishes":[],
		"content":"",
		"visibility":"private",
		"tag_codes":[]
	}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/records", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = request

	input, ok := bindRecordInput(context)
	if !ok {
		t.Fatalf("bindRecordInput() rejected valid quick-record payload: status=%d body=%s", response.Code, response.Body.String())
	}
	if input.PlaceID != 42 || input.Conclusion != "hot" {
		t.Fatalf("bindRecordInput() = %#v", input)
	}
}
