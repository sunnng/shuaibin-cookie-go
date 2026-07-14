//go:build android && cgo

package ui

import (
	"fmt"

	"github.com/Dasongzi1366/AutoGo/imgui"
)

// settingsStatus 系统页操作反馈。
var settingsStatus string

// DefaultCookiePanel 方案 E：工业风 Task Hub（左导航 + 模块表 + 详情 / 卡密 / 系统）。
func DefaultCookiePanel(store *Store) {
	EnsureBuiltinModules()
	SeedHubDefaults(store)

	avail := imgui.ContentRegionAvail()
	railW := float32(120)
	listW := float32(260)
	if avail.X < 520 {
		listW = avail.X * 0.42
		if listW < 160 {
			listW = 160
		}
	}

	imgui.BeginChildStrV("hub_rail", imgui.Vec2{X: railW, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
	renderHubRail(store)
	imgui.EndChild()

	imgui.SameLine()

	nav := store.GetString(KeyHubNav)
	switch nav {
	case HubNavLicense:
		imgui.BeginChildStrV("hub_license", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderLicensePage(store)
		imgui.EndChild()
	case HubNavSystem:
		imgui.BeginChildStrV("hub_system", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderSystemPage(store)
		imgui.EndChild()
	default:
		imgui.BeginChildStrV("hub_list", imgui.Vec2{X: listW, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderModuleList(store)
		imgui.EndChild()

		imgui.SameLine()

		imgui.BeginChildStrV("hub_detail", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderModuleDetail(store)
		imgui.EndChild()
	}
}

func renderHubRail(store *Store) {
	nav := store.GetString(KeyHubNav)
	hubRailButton(store, HubNavModules, "任务", nav == HubNavModules)
	hubRailButton(store, HubNavLicense, "卡密", nav == HubNavLicense)
	hubRailButton(store, HubNavSystem, "系统", nav == HubNavSystem)

	imgui.Dummy(imgui.Vec2{X: 0, Y: 12})
	en, total := CountEnabled(store)
	imgui.TextDisabled(fmt.Sprintf("%d/%d 启用", en, total))
}

func hubRailButton(store *Store, id, label string, active bool) {
	const padX, padY = float32(10), float32(10)
	_, btnH := fitButtonSize(label, padX, padY)
	if btnH < 40 {
		btnH = 40
	}
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: padX, Y: padY})
	if active {
		imgui.PushStyleColorVec4(imgui.ColButton, IndustrialAccent())
		imgui.PushStyleColorVec4(imgui.ColButtonHovered, IndustrialAccent())
		imgui.PushStyleColorVec4(imgui.ColButtonActive, IndustrialAccent())
	}
	if imgui.ButtonV(label+"##rail_"+id, imgui.Vec2{X: -1, Y: btnH}) {
		store.SetString(KeyHubNav, id)
	}
	if active {
		imgui.PopStyleColorV(3)
	}
	imgui.PopStyleVar()
}

func renderModuleList(store *Store) {
	cat := store.GetString(KeyHubCat)
	if cat == "" {
		cat = HubCatAll
	}

	renderCatChips(store, cat)

	imgui.PushItemWidth(-1)
	filter := store.GetString(KeyHubFilter)
	if imgui.InputTextWithHint("##hub_filter", "搜索模块…", &filter, 0, nil) {
		store.SetString(KeyHubFilter, filter)
	}
	imgui.PopItemWidth()
	imgui.Separator()

	imgui.TextDisabled("启用  模块")
	imgui.Separator()

	mods := FilterModules(ModuleCategory(cat), store.GetString(KeyHubFilter))
	selected := store.GetString(KeyHubSelected)
	for _, m := range mods {
		on := m.EnabledKey != "" && store.GetBool(m.EnabledKey)
		mark := "□ "
		if on {
			mark = "■ "
		}
		label := mark + m.Title
		selectedHere := selected == m.ID
		if selectedHere {
			imgui.PushStyleColorVec4(imgui.ColHeader, IndustrialAccent())
			imgui.PushStyleColorVec4(imgui.ColHeaderHovered, IndustrialAccent())
			imgui.PushStyleColorVec4(imgui.ColHeaderActive, IndustrialAccent())
		}
		// 行高按中文抬高，避免 Selectable 裁切
		rowH := measureLabelSize(label).Y + 12
		if rowH < 28 {
			rowH = 28
		}
		if imgui.SelectableBoolV(label+"##mod_"+m.ID, selectedHere, imgui.SelectableFlagsNone, imgui.Vec2{X: 0, Y: rowH}) {
			store.SetString(KeyHubSelected, m.ID)
		}
		if selectedHere {
			imgui.PopStyleColorV(3)
		}
		if m.Summary != nil {
			imgui.TextDisabled("   " + m.Summary(store))
		}
	}
	if len(mods) == 0 {
		imgui.TextDisabled("（无匹配模块）")
	}
}

func renderCatChips(store *Store, cat string) {
	chips := []struct{ id, label string }{
		{HubCatAll, "全部"},
		{HubCatDaily, "日常"},
		{HubCatEvent, "活动"},
		{HubCatMaint, "维护"},
	}
	const padX, padY = float32(10), float32(6)
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
				// 本行放不下则换行，避免被中间栏裁切
				lineStart = true
			} else {
				imgui.SameLine()
			}
		}
		active := cat == c.id
		if active {
			imgui.PushStyleColorVec4(imgui.ColButton, IndustrialAccent())
		}
		if imgui.ButtonV(c.label+"##cat_"+c.id, imgui.Vec2{X: w, Y: h}) {
			store.SetString(KeyHubCat, c.id)
		}
		if active {
			imgui.PopStyleColorV(1)
		}
		lineStart = false
	}
}

func renderModuleDetail(store *Store) {
	id := store.GetString(KeyHubSelected)
	m, ok := ModuleByID(id)
	if !ok {
		mods := Modules()
		if len(mods) == 0 {
			imgui.TextDisabled("无已注册模块")
			return
		}
		m = mods[0]
		store.SetString(KeyHubSelected, m.ID)
	}

	imgui.Text(m.Title)
	imgui.SameLine()
	on := m.EnabledKey != "" && store.GetBool(m.EnabledKey)
	if on {
		imgui.TextColored(IndustrialAccent(), "已启用")
	} else {
		imgui.TextDisabled("未启用")
	}
	imgui.TextDisabled(fmt.Sprintf("ID=%s  分类=%s", m.ID, categoryLabel(m.Category)))
	imgui.Separator()

	if m.RenderDetail != nil {
		m.RenderDetail(store)
	} else {
		imgui.TextDisabled("无详情渲染")
	}
}

func categoryLabel(c ModuleCategory) string {
	switch c {
	case CategoryDaily:
		return "日常"
	case CategoryEvent:
		return "活动"
	case CategoryMaint:
		return "维护"
	default:
		return string(c)
	}
}

func renderLicensePage(store *Store) {
	imgui.Text("卡密")
	imgui.TextDisabled("验证占位 · 后续接入服务端校验")
	imgui.Separator()

	status := store.GetString(KeyLicenseStatus)
	statusLabel := "未验证"
	switch status {
	case "ok":
		statusLabel = "已授权"
	case "expired":
		statusLabel = "已过期"
	}
	imgui.Text("状态  " + statusLabel)
	imgui.TextDisabled("到期  --")
	imgui.Spacing()

	UI_创建输入框(store, KeyLicense, "卡密", "请输入卡密", float32(-1))
	UI_创建按钮("license_verify", "验证", func() {
		settingsStatus = "卡密验证尚未接入"
		store.SetString(KeyLicenseStatus, "unverified")
	}, float32(-2), float32(-2))
	imgui.SameLine()
	UI_创建按钮("license_clear", "清除", func() {
		store.SetString(KeyLicense, "")
		store.SetString(KeyLicenseStatus, "unverified")
		settingsStatus = "卡密已清除"
	}, float32(-2), float32(-2))

	if settingsStatus != "" {
		imgui.Spacing()
		imgui.TextWrapped(settingsStatus)
	}
}

func renderSystemPage(store *Store) {
	imgui.Text("系统")
	imgui.TextDisabled("配置持久化 / 调试")
	imgui.Separator()

	path := cookiePanelOpts.ConfigPath
	if path == "" {
		path = "/sdcard/shuaibin-cookie/ui.json"
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
		if cookiePanelOpts.Controller != nil {
			st := cookiePanelOpts.Controller.State()
			if st == StateRunning || st == StatePaused {
				cookiePanelOpts.Controller.Stop()
			}
		}
		cfgPath := path
		dataPath := cookiePanelOpts.DataStorePath
		if err := ClearPanelCache(store, cfgPath, dataPath, cookiePanelOpts.Reseed); err != nil {
			settingsStatus = fmt.Sprintf("清除失败：%v", err)
			return
		}
		settingsStatus = "缓存已清除，默认配置已恢复"
	}, float32(-2), float32(-2))
	imgui.SameLine()
	UI_创建按钮("dump_config", "打印 Store", func() {
		if json, err := store.ToJSON(); err == nil {
			settingsStatus = json
		}
	}, float32(-2), float32(-2))

	if settingsStatus != "" {
		imgui.Spacing()
		imgui.TextWrapped(settingsStatus)
	}
}
