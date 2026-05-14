package service

import (
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

// IdentifyFoodItem is a single food item identified from an image (name + estimated weight + unit).
type IdentifyFoodItem struct {
	Name       string  `json:"name"`
	Unit       string  `json:"unit"`
	UnitAmount float64 `json:"unit_amount"`
}

// IdentifyResult is the response from the vision model's food identification step.
type IdentifyResult struct {
	IsFood bool               `json:"is_food"`
	Foods  []IdentifyFoodItem `json:"foods"`
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
	// Try markdown code block first
	re := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(.*?)\\n?```")
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	// Fallback: extract between first { and last }
	trimmed := strings.TrimSpace(s)
	first := strings.Index(trimmed, "{")
	last := strings.LastIndex(trimmed, "}")
	if first >= 0 && last > first {
		return trimmed[first : last+1]
	}
	return trimmed
}
