package pgsql

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	"unicode"

	"gorm.io/gorm/schema"
)

// TestAutoMigrateModelsHaveChineseColumnComments 防止新增或修改持久化字段时遗漏中文数据库注释。
func TestAutoMigrateModelsHaveChineseColumnComments(t *testing.T) {
	for _, model := range autoMigrateModels() {
		parsed, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
		if err != nil {
			t.Fatalf("解析模型 %T：%v", model, err)
		}
		for _, field := range parsed.Fields {
			if field.IgnoreMigration {
				continue
			}
			if strings.TrimSpace(field.Comment) == "" {
				t.Errorf("模型 %s 的字段 %s 缺少数据库注释", reflect.TypeOf(model).Elem().Name(), field.Name)
				continue
			}
			if !strings.ContainsFunc(field.Comment, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
				t.Errorf("模型 %s 的字段 %s 未使用中文数据库注释：%q", reflect.TypeOf(model).Elem().Name(), field.Name, field.Comment)
			}
		}
	}
}
