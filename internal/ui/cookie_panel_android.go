//go:build android && cgo

package ui

import (
	"fmt"

	"github.com/Dasongzi1366/AutoGo/imgui"
)

// settingsStatus 系统区操作反馈（仅面板会话内有效，不落盘）。
var settingsStatus string

// renderCookiePanel 左轨（任务/系统）+ 分类列表 + 详情；模块表见 BuiltinModules。
func renderCookiePanel(opts ShellOptions) {
	store := opts.Store
	if store == nil {
		return
	}
	SeedPanelDefaults(store)

	avail := imgui.ContentRegionAvail()
	railW := float32(120)
	listW := float32(260)
	if avail.X < 520 {
		listW = avail.X * 0.42
		if listW < 160 {
			listW = 160
		}
	}

	imgui.PushStyleColorVec4(imgui.ColChildBg, QQBlueRailBg())
	imgui.BeginChildStrV("panel_rail", imgui.Vec2{X: railW, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
	renderPanelRail(store)
	imgui.EndChild()
	imgui.PopStyleColor()

	imgui.SameLine()

	switch store.GetString(KeyPanelNav) {
	case PanelNavSystem:
		imgui.BeginChildStrV("panel_system", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderSystemPage(opts)
		imgui.EndChild()
	default:
		imgui.BeginChildStrV("panel_list", imgui.Vec2{X: listW, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderModuleList(store)
		imgui.EndChild()

		imgui.SameLine()

		imgui.BeginChildStrV("panel_detail", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderModuleDetail(store)
		imgui.EndChild()
	}
}

func renderPanelRail(store *Store) {
	nav := store.GetString(KeyPanelNav)
	railButton(store, PanelNavTasks, "任务", nav == PanelNavTasks)
	railButton(store, PanelNavSystem, "系统", nav == PanelNavSystem)

	imgui.Dummy(imgui.Vec2{X: 0, Y: 12})
	en, total := CountEnabled(store, BuiltinModules())
	imgui.TextDisabled(fmt.Sprintf("%d/%d 启用", en, total))
}

func railButton(store *Store, id, label string, active bool) {
	const padX, padY = float32(10), float32(10)
	_, btnH := fitButtonSize(label, padX, padY)
	if btnH < 40 {
		btnH = 40
	}
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: padX, Y: padY})
	if active {
		// 选中态：天蓝底 + 白字（QQ 风选中项）
		imgui.PushStyleColorVec4(imgui.ColButton, QQBlueAccent())
		imgui.PushStyleColorVec4(imgui.ColButtonHovered, QQBlueAccent())
		imgui.PushStyleColorVec4(imgui.ColButtonActive, QQBlueAccent())
		imgui.PushStyleColorVec4(imgui.ColText, QQBlueWhite())
	}
	if imgui.ButtonV(label+"##rail_"+id, imgui.Vec2{X: -1, Y: btnH}) {
		store.SetString(KeyPanelNav, id)
	}
	if active {
		imgui.PopStyleColorV(4)
	}
	imgui.PopStyleVar()
}

func renderModuleList(store *Store) {
	cat := store.GetString(KeyPanelCat)
	if cat == "" {
		cat = PanelCatAll
	}
	renderCatChips(store, cat)

	imgui.Separator()
	imgui.TextDisabled("启用  模块")
	imgui.Separator()

	mods := FilterByCategory(BuiltinModules(), store.GetString(KeyPanelCat))
	selected := store.GetString(KeyPanelSelected)
	for _, m := range mods {
		on := m.EnabledKey != "" && store.GetBool(m.EnabledKey)
		mark := "□ "
		if on {
			mark = "■ "
		}
		label := mark + m.Title
		selectedHere := selected == m.ID
		if selectedHere {
			imgui.PushStyleColorVec4(imgui.ColHeader, QQBlueAccent())
			imgui.PushStyleColorVec4(imgui.ColHeaderHovered, QQBlueAccent())
			imgui.PushStyleColorVec4(imgui.ColHeaderActive, QQBlueAccent())
			imgui.PushStyleColorVec4(imgui.ColText, QQBlueWhite())
		}
		rowH := measureLabelSize(label).Y + 12
		if rowH < 28 {
			rowH = 28
		}
		if imgui.SelectableBoolV(label+"##mod_"+m.ID, selectedHere, imgui.SelectableFlagsNone, imgui.Vec2{X: 0, Y: rowH}) {
			store.SetString(KeyPanelSelected, m.ID)
		}
		if selectedHere {
			imgui.PopStyleColorV(4)
		}
		if sum := moduleSummary(store, m); sum != "" {
			imgui.TextDisabled("   " + sum)
		}
	}
	if len(mods) == 0 {
		imgui.TextDisabled("（该分类暂无模块）")
	}
}

func renderCatChips(store *Store, cat string) {
	chips := []struct{ id, label string }{
		{PanelCatAll, "全部"},
		{PanelCatDaily, "日常"},
		{PanelCatEvent, "活动"},
		{PanelCatMaint, "维护"},
	}
	const padX, padY = float32(8), float32(6) // padX 收窄，保证 4 个 chip 在最小列宽内不换行
	const gap = float32(6)
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: padX, Y: padY})
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: gap, Y: gap})
	defer imgui.PopStyleVarV(2)

	lineStart := true
	for _, c := range chips {
		w, h := fitButtonSize(c.label, padX, padY)
		if !lineStart {
			remain := imgui.ContentRegionAvail().X
			if remain < w+gap {
				lineStart = true
			} else {
				imgui.SameLine()
			}
		}
		active := cat == c.id
		if active {
			imgui.PushStyleColorVec4(imgui.ColButton, QQBlueAccent())
			imgui.PushStyleColorVec4(imgui.ColText, QQBlueWhite())
		}
		if imgui.ButtonV(c.label+"##cat_"+c.id, imgui.Vec2{X: w, Y: h}) {
			store.SetString(KeyPanelCat, c.id)
		}
		if active {
			imgui.PopStyleColorV(2)
		}
		lineStart = false
	}
}

func renderModuleDetail(store *Store) {
	mods := BuiltinModules()
	id := store.GetString(KeyPanelSelected)
	m, ok := FindModule(mods, id)
	if !ok {
		if len(mods) == 0 {
			imgui.TextDisabled("无模块")
			return
		}
		m = mods[0]
		store.SetString(KeyPanelSelected, m.ID)
	}

	imgui.Text(m.Title)
	imgui.SameLine()
	on := m.EnabledKey != "" && store.GetBool(m.EnabledKey)
	if on {
		imgui.TextColored(QQBlueAccent(), "已启用")
	} else {
		imgui.TextDisabled("未启用")
	}
	imgui.TextDisabled(fmt.Sprintf("分类=%s", categoryLabel(m.Category)))
	imgui.Separator()

	switch m.ID {
	case "arena":
		renderArenaDetail(store)
	default:
		imgui.TextDisabled("无详情渲染")
	}
}

func renderArenaDetail(store *Store) {
	UI_创建复选框(store, KeyArenaEnabled, "启用")
	UI_创建数字输入框(store, KeyArenaMaxBattles, "每日战斗上限", "0=不限", float32(-1), float32(0), "#1f3a52", float64(1), float64(0), float64(999))
	UI_创建数字输入框(store, KeyArenaAutoBuy, "自动购买次数", "0", float32(-1))
	UI_创建数字输入框(store, KeyArenaTrophyDiff, "奖杯差阈值", "0", float32(-1))
}

func moduleSummary(store *Store, m Module) string {
	switch m.ID {
	case "arena":
		max := int(store.GetFloat(KeyArenaMaxBattles))
		buy := int(store.GetFloat(KeyArenaAutoBuy))
		diff := int(store.GetFloat(KeyArenaTrophyDiff))
		maxLabel := "不限"
		if max > 0 {
			maxLabel = fmt.Sprintf("%d", max)
		}
		return fmt.Sprintf("上限 %s · 购买 %d · 奖杯差 %d", maxLabel, buy, diff)
	default:
		return ""
	}
}

func renderSystemPage(opts ShellOptions) {
	store := opts.Store
	imgui.Text("系统")
	imgui.TextDisabled("配置持久化")
	imgui.Separator()

	path := opts.ConfigPath
	if path == "" {
		path = DefaultConfigPath
	}
	imgui.TextDisabled("配置文件  " + path)

	UI_创建按钮("save_config", "保存配置", func() {
		if err := store.SaveConfig(path); err != nil {
			settingsStatus = fmt.Sprintf("保存失败：%v", err)
			return
		}
		settingsStatus = "配置已保存"
	}, float32(-2), float32(-2))
	imgui.SameLine()
	UI_创建按钮("clear_cache", "清除缓存", func() {
		if opts.Controller != nil {
			st := opts.Controller.State()
			if st == StateRunning || st == StatePaused {
				opts.Controller.Stop()
			}
		}
		if err := ClearPanelCache(store, path, opts.DataStorePath, opts.Reseed); err != nil {
			settingsStatus = fmt.Sprintf("清除失败：%v", err)
			return
		}
		SeedPanelDefaults(store)
		settingsStatus = "缓存已清除，默认配置已恢复"
	}, float32(-2), float32(-2))

	if settingsStatus != "" {
		imgui.Spacing()
		imgui.TextWrapped(settingsStatus)
	}
}
