package biscuit

import (
	"fmt"

	"app/internal/store"
)

// graduatedKey 持久化毕业标记：对应 Lua 毕业时 UserConfig.set("biscuit",
// {enabled=false})。Go 侧 ui.json 会在下次 Start 时重新写回 enabled，因此
// 毕业状态落在内部 store，重启后依然生效。
const graduatedKey = "biscuit_graduated"

type State struct {
	store *store.Store

	Rolls int // 本轮已洗次数（内存态，不持久化）
}

func NewState(store *store.Store) *State {
	return &State{store: store}
}

// IsGraduated 是否已毕业（含历史运行持久化的标记）。
func (s *State) IsGraduated() bool {
	v, ok := s.store.Get(graduatedKey, false).(bool)
	return ok && v
}

func (s *State) MarkGraduated() {
	_ = s.store.Set(graduatedKey, true)
}

// ClearGraduated 清除毕业标记（用户想重新洗时由编排者/面板调用）。
func (s *State) ClearGraduated() {
	_ = s.store.Delete(graduatedKey)
}

func (s *State) Reset() {
	s.Rolls = 0
}

// StatusText 生成给灵动岛展示的一行状态："洗脆饼 当前/上限"，
// 毕业/达上限时追加后缀（对齐 Lua StatusHud.setBiscuitReroll 的 extra）。
func (s *State) StatusText(cfg *Config) string {
	maxRolls := 0
	if cfg != nil {
		maxRolls = cfg.MaxRolls
	}
	text := fmt.Sprintf("洗脆饼 %d/%d", s.Rolls, maxRolls)
	if s.IsGraduated() {
		text += " · 已毕业"
	} else if maxRolls > 0 && s.Rolls >= maxRolls {
		text += " · 已达上限"
	}
	return text
}
