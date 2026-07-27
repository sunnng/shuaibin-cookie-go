package biscuit

import (
	"strconv"
	"strings"
)

// effectSlotCount 脆饼固定 4 条副词条。
const effectSlotCount = 4

// Effect 一条副词条的 OCR 解析结果。
type Effect struct {
	Name  string
	Value float64
	Raw   string // 原始片段（含 %），补位条目为空
}

// extractNumber 从字符串末尾反向提取数字（支持小数）。
// 返回数值与去掉数字后 trim 过的名称；尾部无数字或数字无法解析时 ok=false。
// 移植自 Lua extractNumber：按字节从尾向前扫描 [0-9.]，UTF-8 中文名的
//
//	continuation 字节均 >=0x80，不会误判。
func extractNumber(s string) (value float64, name string, ok bool) {
	if s == "" {
		return 0, "", false
	}
	start := len(s)
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if (c >= '0' && c <= '9') || c == '.' {
			start = i
		} else {
			break
		}
	}
	if start >= len(s) {
		return 0, s, false
	}
	name = strings.TrimSpace(s[:start])
	v, err := strconv.ParseFloat(s[start:], 64)
	if err != nil {
		return 0, name, false
	}
	return v, name, true
}

// parseRaw 按 % 拆分 OCR 原文并逐段解析为词条。
// 移植自 Lua parseRaw：无数字或无名称的段被丢弃。
func parseRaw(text string) []Effect {
	if text == "" {
		return nil
	}
	var result []Effect
	for _, part := range strings.Split(text, "%") {
		if part == "" {
			continue
		}
		value, name, ok := extractNumber(part)
		if ok && name != "" {
			result = append(result, Effect{Name: name, Value: value, Raw: part + "%"})
		}
	}
	return result
}
