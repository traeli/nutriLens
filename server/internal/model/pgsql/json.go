package pgsql

import (
	"database/sql/driver"
	"encoding/json"
)

// JSONList 将一对多的、只随所属聚合读写的数据保存在 PostgreSQL JSONB 数组中。
// 这些元素没有独立生命周期，避免为标签、媒体和凭证维护额外关系表。
type JSONList[T any] []T

func (items JSONList[T]) Value() (driver.Value, error) {
	if items == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(items)
}

func (items *JSONList[T]) Scan(value any) error {
	if value == nil {
		*items = JSONList[T]{}
		return nil
	}
	var data []byte
	switch typed := value.(type) {
	case []byte:
		data = typed
	case string:
		data = []byte(typed)
	default:
		return nil
	}
	if len(data) == 0 || string(data) == "null" {
		*items = JSONList[T]{}
		return nil
	}
	return json.Unmarshal(data, items)
}

func (items JSONList[T]) MarshalJSON() ([]byte, error) {
	if items == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]T(items))
}

type JSONDocument json.RawMessage

func (j JSONDocument) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return []byte(j), nil
}

func (j *JSONDocument) Scan(value interface{}) error {
	if value == nil {
		*j = JSONDocument("{}")
		return nil
	}
	var bytes []byte
	switch typed := value.(type) {
	case []byte:
		bytes = typed
	case string:
		bytes = []byte(typed)
	default:
		return nil
	}
	*j = append((*j)[:0], bytes...)
	return nil
}

func (j JSONDocument) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return j, nil
}
