// Package ui 是脚本 UI 框架（ADR-0002/0003）：灵动岛、配置面板、配置绑定、
// 函数组件模型与托管状态。应用以描述符（Task/Field/NavEntry）与框架接缝，
// 框架不感知具体游戏与配置类型，依赖方向只能 应用→框架。
//
// 本包所有无构建标签文件为纯逻辑，可本地测试；绘制层在 *_android.go
// （//go:build android && cgo），只做薄绘制。
package ui
