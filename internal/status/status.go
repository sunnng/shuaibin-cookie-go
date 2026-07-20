// Package status 在任务 goroutine 与 UI 之间传递一行任务状态文本。
// 任务侧（如竞技场）在执行中写入，UI 每帧读取并展示在灵动岛上。
package status

import "sync"

// Reporter 一行状态文本的线程安全通道。空串表示无任务状态，UI 回退到默认文案。
// 单个文本槽：多任务同时上报时后写覆盖先写（当前只有竞技场一个生产者，够用）。
type Reporter struct {
	mu   sync.RWMutex
	text string
}

func New() *Reporter {
	return &Reporter{}
}

// Set 更新状态文本，传空串清除。
func (r *Reporter) Set(text string) {
	r.mu.Lock()
	r.text = text
	r.mu.Unlock()
}

// Text 读取当前状态文本。
func (r *Reporter) Text() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.text
}
