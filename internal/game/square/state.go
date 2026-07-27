package square

import (
	"encoding/json"
	"fmt"
	"time"

	"app/internal/store"
)

// store 键名与 Lua 广场_会话.lua 保持一致。
const (
	doneDateKey = "cuckoo_square_done_date"
	activeKey   = "cuckoo_square_active"
)

// Active 一轮广场挂机会话（Lua Session 的 active 表）。
type Active struct {
	StartedAt      int64  `json:"startedAt"`
	AccumulatedSec int64  `json:"accumulatedSec"`
	LastEnterAt    int64  `json:"lastEnterAt"` // 0 = 未在计时（Lua 的 nil）
	CheckedDate    string `json:"checkedDate"`
}

// State 布谷鸟广场会话：今日完成标记 + 广场/离开弹窗内有效停留累计。
type State struct {
	store *store.Store
}

func NewState(store *store.Store) *State {
	return &State{store: store}
}

func today() string {
	return time.Now().Format("2006-01-02")
}

// IsDoneToday 今日是否已完成（Lua Session.isDoneToday）。
func (s *State) IsDoneToday() bool {
	v, _ := s.store.Get(doneDateKey, "").(string)
	return v != "" && v == today()
}

// MarkDoneToday 标记今日完成并清掉挂机会话（Lua Session.markDoneToday）。
func (s *State) MarkDoneToday() {
	_ = s.store.Set(doneDateKey, today())
	s.Clear()
}

// GetActive 读取挂机会话；无会话或数据损坏时 ok=false。
func (s *State) GetActive() (Active, bool) {
	v := s.store.Get(activeKey, nil)
	if v == nil {
		return Active{}, false
	}
	// store 内存里可能是 Active（本进程写入）或 map[string]any（磁盘读回），
	// 统一走 JSON 往返解码。
	raw, err := json.Marshal(v)
	if err != nil {
		return Active{}, false
	}
	var a Active
	if err := json.Unmarshal(raw, &a); err != nil {
		return Active{}, false
	}
	return a, true
}

// IsActive 是否存在挂机会话（Lua Session.isActive）。
func (s *State) IsActive() bool {
	_, ok := s.GetActive()
	return ok
}

// save 写回挂机会话（Lua Session.save）。
func (s *State) save(a Active) {
	_ = s.store.Set(activeKey, a)
}

// Clear 删除挂机会话（Lua Session.clear）。
func (s *State) Clear() {
	_ = s.store.Delete(activeKey)
}

// ClearAll 清空完成标记与会话（Lua Session.clearAll）。
func (s *State) ClearAll() {
	_ = s.store.Delete(doneDateKey)
	s.Clear()
}

// Ensure 无会话时创建并返回（Lua Session.ensure）。
func (s *State) Ensure() Active {
	if a, ok := s.GetActive(); ok {
		return a
	}
	a := Active{StartedAt: time.Now().Unix()}
	s.save(a)
	return a
}

// MarkCheckedToday 记录今日已完成首次弹窗检查（Lua Session.markCheckedToday）。
func (s *State) MarkCheckedToday() {
	a := s.Ensure()
	a.CheckedDate = today()
	s.save(a)
}

// HasCheckedToday 今日是否已做过首次弹窗检查（Lua Session.hasCheckedToday）。
func (s *State) HasCheckedToday() bool {
	a, ok := s.GetActive()
	return ok && a.CheckedDate == today()
}

// StartStay 开始/恢复广场有效停留计时；已在计时则不动（Lua Session.startStay）。
func (s *State) StartStay() {
	a := s.Ensure()
	if a.LastEnterAt == 0 {
		a.LastEnterAt = time.Now().Unix()
		s.save(a)
	}
}

// PauseStay 暂停计时，并结算当前已停留秒数（Lua Session.pauseStay）。
func (s *State) PauseStay() {
	a, ok := s.GetActive()
	if !ok || a.LastEnterAt == 0 {
		return
	}
	a.AccumulatedSec += max(0, time.Now().Unix()-a.LastEnterAt)
	a.LastEnterAt = 0
	s.save(a)
}

// ResetStayTimer 重置一轮奖励结算所需的有效停留计时（Lua Session.resetStayTimer）。
func (s *State) ResetStayTimer() {
	a := s.Ensure()
	a.AccumulatedSec = 0
	a.LastEnterAt = time.Now().Unix()
	s.save(a)
}

// StayElapsed 有效停留时长：已结算秒数 + 正在计时的一段（Lua Session.stayElapsed）。
func (s *State) StayElapsed() time.Duration {
	a, ok := s.GetActive()
	if !ok {
		return 0
	}
	sec := a.AccumulatedSec
	if a.LastEnterAt != 0 {
		sec += max(0, time.Now().Unix()-a.LastEnterAt)
	}
	return time.Duration(sec) * time.Second
}

// StayRemaining 距 required 还差多少有效停留（Lua Session.stayRemaining）。
func (s *State) StayRemaining(required time.Duration) time.Duration {
	rem := required - s.StayElapsed()
	if rem < 0 {
		return 0
	}
	return rem
}

// CheckReady 供调度器 CheckReady 接线：今日未完成即就绪；remain 恒 0
// （Lua 调度语义：checkReady = not Session.isDoneToday()）。
func (s *State) CheckReady() (bool, time.Duration) {
	if s.IsDoneToday() {
		return false, 0
	}
	return true, 0
}

// Describe 一行会话描述，供 UI 详情展示（Lua Session.describe）。
func (s *State) Describe() string {
	if s.IsDoneToday() {
		return "今日已完成"
	}
	a, ok := s.GetActive()
	if !ok {
		return "今日未完成，无挂机会话"
	}
	checked := "未初检"
	if a.CheckedDate == today() {
		checked = "已初检"
	}
	return fmt.Sprintf("今日未完成，%s，有效停留 %ds", checked, int64(s.StayElapsed().Seconds()))
}
