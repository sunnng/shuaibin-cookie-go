package starlight

import (
	"time"

	"app/internal/store"
)

// doneDateKey 与 Lua 会话（繁星岛_会话.lua）的 store 键保持一致。
const doneDateKey = "starlight_done_date"

// State 梦幻繁星岛会话：今日完成标记持久化。
type State struct {
	store *store.Store
}

func NewState(store *store.Store) *State {
	return &State{store: store}
}

func today() string {
	return time.Now().Format("2006-01-02")
}

// IsDoneToday 今日是否已完成。
func (s *State) IsDoneToday() bool {
	v, _ := s.store.Get(doneDateKey, "").(string)
	return v == today()
}

func (s *State) MarkDoneToday() {
	_ = s.store.Set(doneDateKey, today())
}

func (s *State) Clear() {
	_ = s.store.Delete(doneDateKey)
}

// Describe 一句中文描述，供日志/状态展示。
func (s *State) Describe() string {
	if s.IsDoneToday() {
		return "今日已完成"
	}
	return "今日未完成"
}

// TimeUntilNextDay 距下一个自然日（本地零点）的剩余时间；已完成当天任务时
// 作为调度器 idle remain 使用。
func (s *State) TimeUntilNextDay() time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return time.Until(next)
}
