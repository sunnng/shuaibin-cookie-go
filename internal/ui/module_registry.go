package ui

import "sync"

// ModuleCategory 任务分类，用于 Task Hub 左侧过滤。
type ModuleCategory string

const (
	CategoryDaily ModuleCategory = "daily"
	CategoryEvent ModuleCategory = "event"
	CategoryMaint ModuleCategory = "maint"
)

// ModuleDef 可注册的任务模块。RenderDetail 仅在 Android ImGui 面板调用；桌面 stub 可为空。
type ModuleDef struct {
	ID          string
	Title       string
	Category    ModuleCategory
	EnabledKey  string
	Summary     func(*Store) string
	RenderDetail func(*Store)
}

var (
	moduleMu   sync.RWMutex
	moduleList []ModuleDef
	moduleByID = map[string]ModuleDef{}
)

// RegisterModule 注册（或覆盖同 ID）模块。应在 RunShell 前完成。
func RegisterModule(m ModuleDef) {
	if m.ID == "" {
		return
	}
	moduleMu.Lock()
	defer moduleMu.Unlock()
	if _, ok := moduleByID[m.ID]; ok {
		for i := range moduleList {
			if moduleList[i].ID == m.ID {
				moduleList[i] = m
				break
			}
		}
	} else {
		moduleList = append(moduleList, m)
	}
	moduleByID[m.ID] = m
}

// ClearModules 清空注册表（测试用）。
func ClearModules() {
	moduleMu.Lock()
	defer moduleMu.Unlock()
	moduleList = nil
	moduleByID = map[string]ModuleDef{}
}

// Modules 按注册顺序返回副本。
func Modules() []ModuleDef {
	moduleMu.RLock()
	defer moduleMu.RUnlock()
	out := make([]ModuleDef, len(moduleList))
	copy(out, moduleList)
	return out
}

// ModuleByID 查找模块。
func ModuleByID(id string) (ModuleDef, bool) {
	moduleMu.RLock()
	defer moduleMu.RUnlock()
	m, ok := moduleByID[id]
	return m, ok
}

// CountEnabled 统计 EnabledKey 为 true 的模块数。
func CountEnabled(store *Store) (enabled, total int) {
	moduleMu.RLock()
	defer moduleMu.RUnlock()
	total = len(moduleList)
	if store == nil {
		return 0, total
	}
	for _, m := range moduleList {
		if m.EnabledKey != "" && store.GetBool(m.EnabledKey) {
			enabled++
		}
	}
	return enabled, total
}

// FilterModules 按分类与关键字过滤（关键字匹配 Title / ID，大小写不敏感由调用方规范化）。
func FilterModules(cat ModuleCategory, query string) []ModuleDef {
	moduleMu.RLock()
	defer moduleMu.RUnlock()
	out := make([]ModuleDef, 0, len(moduleList))
	for _, m := range moduleList {
		if cat != "" && cat != "all" && m.Category != cat {
			continue
		}
		if query != "" && !moduleMatchQuery(m, query) {
			continue
		}
		out = append(out, m)
	}
	return out
}

func moduleMatchQuery(m ModuleDef, query string) bool {
	if containsFold(m.ID, query) || containsFold(m.Title, query) {
		return true
	}
	return false
}

func containsFold(s, substr string) bool {
	if substr == "" {
		return true
	}
	return len(s) >= len(substr) && (s == substr || indexFold(s, substr) >= 0)
}

func indexFold(s, substr string) int {
	// 简易 ASCII/UTF-8 子串：转小写比较（中文 Title 用原文包含）。
	ls, lsub := toLowerASCII(s), toLowerASCII(substr)
	for i := 0; i+len(lsub) <= len(ls); i++ {
		if ls[i:i+len(lsub)] == lsub {
			return i
		}
	}
	// 中文等：原文包含
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func toLowerASCII(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
