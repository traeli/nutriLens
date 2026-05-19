package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"nutrilens/internal/model"

	"gorm.io/gorm"
)

// AIProviderService dispatches AI calls to the configured provider per task type.
// It reads model config from ai_models and prompt templates from ai_prompts (both managed via Directus).
type AIProviderService struct {
	db *gorm.DB
}

func NewAIProviderService(db *gorm.DB) *AIProviderService {
	return &AIProviderService{db: db}
}

// GetActiveModel returns the enabled model config for a given task type.
func (s *AIProviderService) GetActiveModel(taskType string) (*model.AIModel, error) {
	var m model.AIModel
	if err := s.db.Where("task_type = ? AND enabled = ?", taskType, true).First(&m).Error; err != nil {
		return nil, fmt.Errorf("no enabled AI model for task type '%s'", taskType)
	}
	return &m, nil
}

// GetActivePrompt returns the enabled prompt config for a given task type.
func (s *AIProviderService) GetActivePrompt(taskType string) (*model.AIPrompt, error) {
	var p model.AIPrompt
	if err := s.db.Where("task_type = ? AND enabled = ?", taskType, true).First(&p).Error; err != nil {
		return nil, fmt.Errorf("no enabled AI prompt for task type '%s'", taskType)
	}
	return &p, nil
}

// renderPrompt replaces {{variable}} placeholders in the template.
func renderPrompt(tmpl string, vars map[string]string) string {
	result := tmpl
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}

// callAI makes a chat completions API call to the configured provider.
func (s *AIProviderService) callAI(ctx context.Context, m *model.AIModel, payload map[string]interface{}) (string, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", m.BaseURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s request failed: %w", m.Provider, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// Check for API error response
	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	if json.Unmarshal(respBody, &errResp) == nil && errResp.Error.Message != "" {
		return "", fmt.Errorf("%s API error: %s", m.Provider, errResp.Error.Message)
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("parse %s response failed: %s", m.Provider, string(respBody))
	}
	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no response from %s, body: %s", m.Provider, string(respBody))
	}

	return extractJSON(apiResp.Choices[0].Message.Content), nil
}

// IdentifyImageFood uses the vision model to only identify food items and quantities from an image.
// It returns an IdentifyResult with food names and estimated weights, without nutritional analysis.
func (s *AIProviderService) IdentifyImageFood(ctx context.Context, imageURL string) (*IdentifyResult, error) {
	m, err := s.GetActiveModel("image_identify")
	if err != nil {
		return nil, err
	}
	p, err := s.GetActivePrompt("image_identify")
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"model": m.ModelName,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": imageURL,
						},
					},
					{
						"type": "text",
						"text": p.UserPromptTemplate,
					},
				},
			},
		},
	}

	content, err := s.callAI(ctx, m, payload)
	if err != nil {
		return nil, err
	}

	var result IdentifyResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("parse identify response failed: %w, body: %s", err, content)
	}
	return &result, nil
}

// AnalyzeImage analyzes a food image using the configured image model (e.g., 智谱 GLM-4V).
func (s *AIProviderService) AnalyzeImage(ctx context.Context, imageURL string, user model.User) (*AnalyzeResult, error) {
	m, err := s.GetActiveModel("image")
	if err != nil {
		return nil, err
	}
	p, err := s.GetActivePrompt("image")
	if err != nil {
		return nil, err
	}

	userInfo := fmt.Sprintf("用户信息: 身高%.0fcm, 体重%.0fkg, 年龄%d岁, 性别%s",
		user.Height, user.Weight, user.Age, genderStr(user.Gender))

	userPrompt := renderPrompt(p.UserPromptTemplate, map[string]string{
		"user_info": userInfo,
	})

	payload := map[string]interface{}{
		"model": m.ModelName,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": imageURL,
						},
					},
					{
						"type": "text",
						"text": userPrompt,
					},
				},
			},
		},
	}

	content, err := s.callAI(ctx, m, payload)
	if err != nil {
		return nil, err
	}

	var result AnalyzeResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return &AnalyzeResult{
			FoodName:   "未知食物",
			Calories:   0,
			Suggestion: content,
		}, nil
	}
	return &result, nil
}

// AnalyzeByText analyzes text using the configured text model (e.g., DeepSeek).
func (s *AIProviderService) AnalyzeByText(ctx context.Context, description string, user model.User) (*AnalyzeResult, error) {
	m, err := s.GetActiveModel("text")
	if err != nil {
		return nil, err
	}
	p, err := s.GetActivePrompt("text")
	if err != nil {
		return nil, err
	}

	userInfo := fmt.Sprintf("身高%.0fcm, 体重%.0fkg, 年龄%d岁, 性别%s",
		user.Height, user.Weight, user.Age, genderStr(user.Gender))

	userPrompt := renderPrompt(p.UserPromptTemplate, map[string]string{
		"description": description,
		"user_info":   userInfo,
	})

	messages := []map[string]interface{}{}
	if p.SystemPrompt != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": p.SystemPrompt,
		})
	}
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userPrompt,
	})

	payload := map[string]interface{}{
		"model":    m.ModelName,
		"messages": messages,
	}

	content, err := s.callAI(ctx, m, payload)
	if err != nil {
		return nil, err
	}

	var result AnalyzeResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return &AnalyzeResult{
			FoodName:   "未知食物",
			Calories:   0,
			Suggestion: content,
		}, nil
	}
	return &result, nil
}

// AnalyzeDish analyzes a dish name using the configured dish model.
func (s *AIProviderService) AnalyzeDish(ctx context.Context, dishName string) (*DishAnalysisResult, error) {
	m, err := s.GetActiveModel("dish")
	if err != nil {
		return nil, err
	}
	p, err := s.GetActivePrompt("dish")
	if err != nil {
		return nil, err
	}

	userPrompt := renderPrompt(p.UserPromptTemplate, map[string]string{
		"dish_name": dishName,
	})

	messages := []map[string]interface{}{}
	if p.SystemPrompt != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": p.SystemPrompt,
		})
	}
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userPrompt,
	})

	payload := map[string]interface{}{
		"model":    m.ModelName,
		"messages": messages,
	}

	content, err := s.callAI(ctx, m, payload)
	if err != nil {
		return nil, err
	}

	var result DishAnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("parse dish JSON failed: %w", err)
	}
	return &result, nil
}

// RecommendMeal generates a personalized meal recommendation for push notifications.
func (s *AIProviderService) RecommendMeal(ctx context.Context, userInfo, todayFoods, nutrientSummary, preferredDishes, mealTypeName string) (string, error) {
	m, err := s.GetActiveModel("meal_recommend")
	if err != nil {
		return "", err
	}
	p, err := s.GetActivePrompt("meal_recommend")
	if err != nil {
		return "", err
	}

	userPrompt := renderPrompt(p.UserPromptTemplate, map[string]string{
		"user_info":        userInfo,
		"today_foods":      todayFoods,
		"nutrient_summary": nutrientSummary,
		"preferred_dishes": preferredDishes,
		"meal_type_name":   mealTypeName,
	})

	messages := []map[string]interface{}{}
	if p.SystemPrompt != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": p.SystemPrompt,
		})
	}
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userPrompt,
	})

	payload := map[string]interface{}{
		"model":    m.ModelName,
		"messages": messages,
	}

	content, err := s.callAI(ctx, m, payload)
	if err != nil {
		return "", err
	}
	return content, nil
}

// AnalyzeDailySummary generates a daily health suggestion based on the user's food intake.
func (s *AIProviderService) AnalyzeDailySummary(ctx context.Context, foodList string, nutrientSummary string, user model.User) (string, error) {
	m, err := s.GetActiveModel("daily_analysis")
	if err != nil {
		return "", err
	}
	p, err := s.GetActivePrompt("daily_analysis")
	if err != nil {
		return "", err
	}

	userInfo := fmt.Sprintf("身高%.0fcm, 体重%.0fkg, 年龄%d岁, 性别%s",
		user.Height, user.Weight, user.Age, genderStr(user.Gender))

	userPrompt := renderPrompt(p.UserPromptTemplate, map[string]string{
		"food_list":        foodList,
		"nutrient_summary": nutrientSummary,
		"user_info":        userInfo,
	})

	messages := []map[string]interface{}{}
	if p.SystemPrompt != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": p.SystemPrompt,
		})
	}
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userPrompt,
	})

	payload := map[string]interface{}{
		"model":    m.ModelName,
		"messages": messages,
	}

	content, err := s.callAI(ctx, m, payload)
	if err != nil {
		return "", err
	}

	return content, nil
}
