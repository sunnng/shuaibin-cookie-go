package hud

import "app/internal/logger"

type HUD struct{}

func New() *HUD { return &HUD{} }

func (h *HUD) SetTask(name, status string) {
	logger.Infof("[HUD] %s: %s", name, status)
}

func (h *HUD) SetIdle() {
	logger.Infof("[HUD] idle")
}

func (h *HUD) SetWait(reason string) {
	logger.Infof("[HUD] wait: %s", reason)
}
