package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"nutrilens/internal/model"
)

type FoodItem struct {
	Name       string        `json:"name"`
	Unit       string        `json:"unit"`
	UnitAmount float64       `json:"unit_amount"`
	Calories   float64       `json:"calories"`
	Nutrients  model.JSONMap `json:"nutrients"`
}

type AnalyzeResult struct {
	IsFood     bool          `json:"is_food"`
	FoodName   string        `json:"food_name"`
	Unit       string        `json:"unit"`
	UnitAmount float64       `json:"unit_amount"`
	Calories   float64       `json:"calories"`
	Nutrients  model.JSONMap `json:"nutrients"`
	Suggestion string        `json:"suggestion"`
	Foods      []FoodItem    `json:"foods"`
}

type DishAnalysisResult struct {
	IsFood     bool          `json:"is_food"`
	Name       string        `json:"name"`
	Category   string        `json:"category"`
	Calories   float64       `json:"calories"`
	Unit       string        `json:"unit"`
	UnitAmount float64       `json:"unit_amount"`
	Nutrients  model.JSONMap `json:"nutrients"`
}

type DeepSeekService struct {
	APIKey  string
	BaseURL string
}

func NewDeepSeekService(apiKey, baseURL string) *DeepSeekService {
	return &DeepSeekService{APIKey: apiKey, BaseURL: baseURL}
}

// AnalyzeByImage sends image to DeepSeek VL for food analysis.
func (s *DeepSeekService) AnalyzeByImage(ctx context.Context, imageBytes []byte, user model.User) (*AnalyzeResult, error) {
	b64Image := base64.StdEncoding.EncodeToString(imageBytes)
	userInfo := fmt.Sprintf("用户信息: 身高%.0fcm, 体重%.0fkg, 年龄%d岁, 性别%s",
		user.Height, user.Weight, user.Age, genderStr(user.Gender))

	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": "data:image/jpeg;base64," + b64Image,
						},
					},
					{
						"type": "text",
						"text": fmt.Sprintf(`请分析这张图片中的内容，返回JSON格式(不要其他内容):
{
  "is_food": true或false(图片中是否包含食物),
	  "food_name": "食物名称",
  "unit": "kg",
  "unit_amount": 估算重量(基于图片中食物的视觉比例，参考成人正常食量，单位为kg的数字，如0.15表示150g),
  "calories": 估算总卡路里(kcal,数字，基于估算重量计算),
  "nutrients": {"protein": 蛋白质g, "carbs": 碳水g, "fat": 脂肪g, "fiber": 膳食纤维g, "sugar": 糖分g, "vitamin_c": 维生素C_mg},
  "suggestion": "根据%s给出的个性化饮食建议，特别关注糖分摄入对血糖的影响"
}

注意:
- is_food字段必填。如果图片中没有食物(如风景、人物、动物等)，设为false，其他字段填默认值即可
	- 如果is_food为true，所有营养数据必须基于unit_amount(估算重量)来计算
- 糖分(sugar)字段必填，单位为克
- 维生素C(vitamin_c)如果食物含有则填写，单位为mg，不含则为0
- unit固定为"kg"`, userInfo),
					},
				},
			},
		},
	}

	return s.callAPI(ctx, payload)
}

// AnalyzeByText sends text description to DeepSeek Chat for food analysis.
func (s *DeepSeekService) AnalyzeByText(ctx context.Context, description string, user model.User) (*AnalyzeResult, error) {
	userInfo := fmt.Sprintf("身高%.0fcm, 体重%.0fkg, 年龄%d岁, 性别%s",
		user.Height, user.Weight, user.Age, genderStr(user.Gender))

	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]interface{}{
			{
				"role": "system",
				"content": "你是一个专业的营养师AI助手。用户会告诉你他们吃了什么，你需要将每种食物分开分析营养成分和卡路里，并根据用户的身体信息给出个性化的饮食建议。请始终以JSON格式回复。",
			},
			{
				"role": "user",
				"content": fmt.Sprintf(`我吃了: %s

我的信息: %s

请返回JSON格式(不要其他内容):
{
  "is_food": true或false(用户描述的内容是否是食物),
  "foods": [
    {"name": "食物名称", "unit": "kg", "unit_amount": 估算重量(参考该用户身体信息的正常食量，单位为kg的数字，如0.1表示100g), "calories": 卡路里数字, "nutrients": {"protein": 蛋白质g, "carbs": 碳水g, "fat": 脂肪g, "fiber": 膳食纤维g, "sugar": 糖分g, "vitamin_c": 维生素C_mg}},
    {"name": "食物名称", "unit": "kg", "unit_amount": 估算重量, "calories": 卡路里数字, "nutrients": {"protein": 蛋白质g, "carbs": 碳水g, "fat": 脂肪g, "fiber": 膳食纤维g, "sugar": 糖分g, "vitamin_c": 维生素C_mg}}
  ],
  "suggestion": "根据用户信息给出的个性化饮食建议，特别关注糖分摄入对血糖的影响"
}

重要规则:
- is_food字段必填。如果用户描述的不是食物(如"我吃了个手机"、"今天天气不错"等)，设为false，foods为空数组
- unit固定为"kg"
- 如果用户没有明确说食物量，根据用户的身高体重年龄估算一个成人正常食用量作为unit_amount
- 所有营养数据必须基于unit_amount(估算重量)来计算
- 糖分(sugar)字段必填，单位为克，这对糖尿病患者非常重要
- 维生素C(vitamin_c)如果食物含有则填写，单位为mg，不含则为0`, description, userInfo),
			},
		},
	}

	return s.callAPI(ctx, payload)
}

// AnalyzeDish analyzes a single dish name and returns nutritional data for caching.
func (s *DeepSeekService) AnalyzeDish(ctx context.Context, dishName string) (*DishAnalysisResult, error) {
	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]interface{}{
			{
				"role": "system",
				"content": "你是一个专业的营养师AI助手。根据菜品名称，分析其营养成分。请始终以JSON格式回复。",
			},
			{
				"role": "user",
				"content": fmt.Sprintf(`菜品名称: %s

请返回JSON格式(不要其他内容):
{
  "is_food": true或false(这个名称是否是一种食物),
  "name": "%s",
  "category": "分类(中式/西式/日式/韩式/其他)",
  "unit": "kg",
  "unit_amount": 成人正常一份的重量(kg数字,如0.2表示200g),
  "calories": 一份的卡路里(kcal数字),
  "nutrients": {"protein": 蛋白质g, "carbs": 碳水g, "fat": 脂肪g, "fiber": 膳食纤维g, "sugar": 糖分g, "vitamin_c": 维生素C_mg}
}

注意:
- is_food字段必填。如果名称不是食物，设为false
- unit固定为"kg"
- 营养数据基于unit_amount(成人正常一份)计算
- sugar字段必填，这对糖尿病患者很重要`, dishName, dishName),
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", s.BaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil || len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("parse dish analysis response failed")
	}

	content := extractJSON(apiResp.Choices[0].Message.Content)

	var result DishAnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("parse dish JSON failed: %w", err)
	}
	return &result, nil
}

func (s *DeepSeekService) callAPI(ctx context.Context, payload map[string]interface{}) (*AnalyzeResult, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", s.BaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %s", string(respBody))
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from deepseek")
	}

	content := extractJSON(apiResp.Choices[0].Message.Content)

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

func genderStr(g int) string {
	switch g {
	case 1:
		return "男"
	case 2:
		return "女"
	default:
		return "未知"
	}
}

func extractJSON(s string) string {
	re := regexp.MustCompile("(?s)```(?:json)?\\s*\n?(.*?)\n?```")
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return strings.TrimSpace(s)
}
