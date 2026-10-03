package service

import (
	"context"
	"errors"
	"testing"
)

type nutritionAnalyzerStub struct {
	content string
	err     error
}

func (s nutritionAnalyzerStub) AnalyzeNutrition(context.Context, string) (string, error) {
	return s.content, s.err
}

func TestNutritionServiceAnalyze(t *testing.T) {
	service := NewNutritionService(nil, nutritionAnalyzerStub{content: `{
		"foods":[{"name":"苹果","amount":"1个（约180g）","calories":95,"protein_grams":0.5,"fat_grams":0.3,"carbohydrate_grams":25}],
		"calories":95,"protein_grams":0.5,"fat_grams":0.3,"carbohydrate_grams":25,
		"advice":"可搭配一份蛋白质食物，让加餐更均衡。"
	}`})
	result, err := service.analyze(context.Background(), "我吃了一个苹果")
	if err != nil {
		t.Fatalf("analyze() error = %v", err)
	}
	if len(result.Foods) != 1 || result.Foods[0].Name != "苹果" || result.Calories != 95 || result.Advice == "" {
		t.Fatalf("analyze() = %+v", result)
	}
}

func TestNutritionServiceRejectsInvalidModelResult(t *testing.T) {
	service := NewNutritionService(nil, nutritionAnalyzerStub{content: `{"foods":[],"calories":-1,"protein_grams":0,"fat_grams":0,"carbohydrate_grams":0,"advice":"建议均衡饮食。"}`})
	_, err := service.analyze(context.Background(), "不完整的结果")
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("analyze() error = %v, want ErrUnavailable", err)
	}
}

func TestNutritionServiceRequiresAnalyzer(t *testing.T) {
	service := NewNutritionService(nil, nil)
	_, err := service.analyze(context.Background(), "一个苹果")
	if !errors.Is(err, ErrNutritionAnalysisNotConfigured) {
		t.Fatalf("analyze() error = %v, want ErrNutritionAnalysisNotConfigured", err)
	}
}
