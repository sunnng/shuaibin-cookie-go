package production

import (
	"app/internal/platform/screen"
)

// Feature 王国生产任务的 UI 特征表（基准分辨率 1600×900）。
// 只描述常量，行为在 page / route。
type Feature struct {
	Board   BoardFeature
	Dialogs DialogsFeature
}

// BoardFeature 王国生产总览/产线界面。
type BoardFeature struct {
	Identify screen.Feature
	Actions  BoardActions
}

type BoardActions struct {
	// CollectAll 一键收取等主操作区；未取色前 Region 为零值。
	CollectAll screen.Region
}

// DialogsFeature 生产相关弹窗特征；未取色的不注册进 Guard。
type DialogsFeature struct {
	// 按需追加 DialogDef。
}

// DialogDef 弹窗比色 + 确认点击区。
type DialogDef struct {
	Identify screen.Feature
	Confirm  screen.Region
}

// DefaultFeature 返回空特征表，供装配与单测使用；真机坐标后续填入。
func DefaultFeature() *Feature {
	return &Feature{}
}
