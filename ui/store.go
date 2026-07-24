package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Store 统一管理所有 UI 控件状态，以 key 为索引存储各类型值。
// Store 非并发安全；调用方需自行同步（当前在 UI goroutine 与启动钩子内串行访问）。
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

func (s *Store) ToJSON() (string, error) {
	b, err := json.Marshal(s.values)
	return string(b), err
}

func (s *Store) SaveConfig(path string) error {
	b, err := json.Marshal(s.values)
	if err != nil {
		return err
	}
	// 设备路径（如 /sdcard/shuaibin-cookie/ui.json）首启时父目录可能不存在；
	// 不 MkdirAll 则 WriteFile 失败，面板「开始」经 StartStop 会静默中止启动。
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, b, 0o644)
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

// Clear 清空内存中的全部控件状态。
func (s *Store) Clear() {
	s.values = map[string]interface{}{}
}
