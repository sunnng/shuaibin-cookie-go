package main

import "strings"

// splitList 把面板的逗号分隔文本解析为条目列表（兼容中英文逗号，去空白与空项）。
func splitList(s string) []string {
	var out []string
	for _, part := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == '，' }) {
		if item := strings.TrimSpace(part); item != "" {
			out = append(out, item)
		}
	}
	return out
}

// joinList 列表拼回逗号分隔文本（面板显示用）。
func joinList(items []string) string {
	return strings.Join(items, ",")
}
