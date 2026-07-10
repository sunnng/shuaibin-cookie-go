//go:build android && cgo

package ui

import (
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// DefaultCookiePanel 项目默认的灰色玻璃风格配置面板。
func DefaultCookiePanel(store *Store) {
	UI_创建左侧父标签页("main_parent_tabs",
		[]string{"首页", "模块", "设置"},
		[]func(){
			func() { renderHomeTab() },
			func() { renderModulesTab(store) },
			func() { renderSettingsTab(store) },
		},
	)
}

func renderHomeTab() {
	UI_创建标签栏("home_child_tabs",
		[]string{"概览", "说明"},
		[]func(){
			func() {
				imgui.Text("帅宾 Cookie - 王国自动化")
				imgui.Spacing()
				imgui.Text("点击右上角「运行脚本」开始执行。")
				imgui.Text("倒计时结束也会自动启动。")
			},
			func() {
				imgui.Text("运行时会在后台循环调度任务。")
				imgui.Text("竞技场模块支持免费刷新等待与战斗上限。")
			},
		},
	)
}

func renderModulesTab(store *Store) {
	UI_创建标签栏("modules_child_tabs",
		[]string{"竞技场"},
		[]func(){
			func() { renderArenaSettings(store) },
		},
	)
}

func renderArenaSettings(store *Store) {
	UI_创建折叠("arena_collapse", "竞技场选项", true, func() {
		UI_创建复选框(store, KeyArenaEnabled, "启用竞技场")
		UI_创建数字输入框(store, KeyArenaMaxBattles, "每日战斗上限", "0=不限", float32(-1), float32(0), "#ffffff", float64(1), float64(0), float64(999))
		UI_创建数字输入框(store, KeyArenaAutoBuy, "自动购买次数", "0", float32(-1))
		UI_创建数字输入框(store, KeyArenaTrophyDiff, "奖杯差阈值", "0", float32(-1))
	})
}

func renderSettingsTab(store *Store) {
	UI_创建标签栏("settings_child_tabs",
		[]string{"调试"},
		[]func(){
			func() {
				imgui.Text("UI 配置保存在 /sdcard/shuaibin-cookie/ui.json")
				imgui.Spacing()
				UI_创建按钮("dump_config", "打印 Store", func() {
					if json, err := store.ToJSON(); err == nil {
						imgui.Text(json)
					}
				}, float32(-1), float32(-2))
			},
		},
	)
}
