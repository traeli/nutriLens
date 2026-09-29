package model

import (
	"database/sql/driver"
	"encoding/json"
)

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
