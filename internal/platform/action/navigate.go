package action

import (
	"github.com/Dasongzi1366/AutoGo/motion"
	"github.com/Dasongzi1366/AutoGo/utils"
)

func (e *AndroidExecutor) Back() error {
	motion.Back(e.displayId)
	return nil
}

func (e *AndroidExecutor) Home() error {
	motion.Home(e.displayId)
	return nil
}

func (e *AndroidExecutor) Sleep(ms int) {
	utils.Sleep(ms)
}
