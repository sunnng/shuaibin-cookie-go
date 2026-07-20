package production

import (
	"fmt"

	"app/internal/store"
)

// State 王国生产的任务状态（易失计数 + 可选持久化）。
// 命名对齐 CONTEXT.md「任务状态」，避免 Session。
type State struct {
	store *store.Store

	// Collected 本脚本内已收取次数（易失；停止再开清零）。
	Collected int
}

// NewState 创建任务状态；store 可为 nil（仅内存字段可用）。
func NewState(s *store.Store) *State {
	return &State{store: s}
}

// StatusText 灵动岛一行展示。
func (s *State) StatusText() string {
	if s == nil {
		return "生产"
	}
	return fmt.Sprintf("生产 %d", s.Collected)
}

// Reset 清零本脚本易失字段。
func (s *State) Reset() {
	if s == nil {
		return
	}
	s.Collected = 0
}
