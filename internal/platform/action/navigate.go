package action

import (
	"github.com/Dasongzi1366/AutoGo/motion"
	"github.com/Dasongzi1366/AutoGo/utils"
)

func (e *AndroidExecutor) Back() {
	motion.Back(e.displayId)
}

func (e *AndroidExecutor) Home() {
	motion.Home(e.displayId)
}

func (e *AndroidExecutor) Sleep(ms int) {
	utils.Sleep(ms)
}
