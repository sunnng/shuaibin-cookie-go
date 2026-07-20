package scheduler

import (
	"fmt"
	"sort"
	"time"

	"app/internal/logger"
)

type Task struct {
	Name      string
	Condition func() bool
	Action    func() error
}

type Scheduler struct {
	tasks         []Task
	idleProviders map[string]func() (time.Duration, string)
}

func New() *Scheduler {
	return &Scheduler{
		idleProviders: make(map[string]func() (time.Duration, string)),
	}
}

func (s *Scheduler) Add(name string, condition func() bool, action func() error) {
	s.tasks = append(s.tasks, Task{Name: name, Condition: condition, Action: action})
	logger.Debugf("[Scheduler] registered task: %s", name)
}

func (s *Scheduler) AddIdleProvider(name string, provider func() (time.Duration, string)) {
	s.idleProviders[name] = provider
	logger.Debugf("[Scheduler] registered idle provider: %s", name)
}

func (s *Scheduler) Clear() {
	s.tasks = nil
	s.idleProviders = make(map[string]func() (time.Duration, string))
}

func (s *Scheduler) Run(stopOnError bool) (bool, error) {
	hasWork := false
	for _, task := range s.tasks {
		if !task.Condition() {
			continue
		}
		hasWork = true
		logger.Infof("[Scheduler] running task: %s", task.Name)
		if err := task.Action(); err != nil {
			logger.Errorf("[Scheduler] task %s failed: %v", task.Name, err)
			if stopOnError {
				return hasWork, fmt.Errorf("task %s: %w", task.Name, err)
			}
		}
	}
	return hasWork, nil
}

// MinIdleWait 返回所有 idle provider 中最小的正等待时长及其标签。
// 取最小值：最早可能就绪的任务决定唤醒时间；没有正等待时返回 (0, "")。
func (s *Scheduler) MinIdleWait() (time.Duration, string) {
	names := make([]string, 0, len(s.idleProviders))
	for name := range s.idleProviders {
		names = append(names, name)
	}
	sort.Strings(names)

	var minWait time.Duration
	var minLabel string
	for _, name := range names {
		wait, label := s.idleProviders[name]()
		if wait <= 0 {
			continue
		}
		if minWait == 0 || wait < minWait {
			minWait, minLabel = wait, label
		}
	}
	return minWait, minLabel
}
