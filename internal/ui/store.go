package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// Store 统一管理所有 UI 控件状态，以 key 为索引存储各类型值。
type Store struct {
	values map[string]interface{}
}

// NewStore 初始化并返回一个空的 Store。
func NewStore() *Store {
	return &Store{values: map[string]interface{}{}}
}

func (s *Store) GetBool(key string) bool {
	v, _ := s.values[key].(bool)
	return v
}

func (s *Store) SetBool(key string, v bool) {
	s.values[key] = v
}

func (s *Store) GetString(key string) string {
	v, _ := s.values[key].(string)
	return v
}

func (s *Store) SetString(key string, v string) {
	s.values[key] = v
}

func (s *Store) GetFloat(key string) float64 {
	v, _ := s.values[key].(float64)
	return v
}

func (s *Store) SetFloat(key string, v float64) {
	s.values[key] = v
}

func (s *Store) HasKey(key string) bool {
	_, ok := s.values[key]
	return ok
}

func (s *Store) ToMap() map[string]interface{} {
	return s.values
}

func (s *Store) ToJSON() (string, error) {
	b, err := json.Marshal(s.values)
	return string(b), err
}

func (s *Store) ToStringMap() map[string]string {
	result := make(map[string]string, len(s.values))
	for k, v := range s.values {
		switch val := v.(type) {
		case bool:
			if val {
				result[k] = "true"
			} else {
				result[k] = "false"
			}
		case float64:
			result[k] = strconv.FormatFloat(val, 'f', -1, 64)
		case string:
			result[k] = val
		default:
			result[k] = fmt.Sprintf("%v", val)
		}
	}
	return result
}

func (s *Store) SaveConfig(path string) error {
	b, err := json.Marshal(s.values)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o777)
}

func (s *Store) LoadConfig(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var loaded map[string]interface{}
	if err := json.Unmarshal(b, &loaded); err != nil {
		return err
	}
	for k, v := range loaded {
		s.values[k] = v
	}
	return nil
}
