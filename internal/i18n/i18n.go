package i18n

import "fmt"

// lang 是当前界面语言："en" 或 "zh"。vortex-cn 分支默认中文。
var lang = "zh"

// SetLang 设置界面语言；仅接受 "en" 或 "zh"，非法值被忽略。
func SetLang(l string) {
	if l == "en" || l == "zh" {
		lang = l
	}
}

// Lang 返回当前语言。
func Lang() string { return lang }

// T 返回 key 的翻译。未命中时返回 key 本身（即英文原文），
// 因此未翻译的文案会自然回退为英文，不会出现空白或 key 泄漏。
func T(key string) string {
	if lang == "zh" {
		if v, ok := zh[key]; ok {
			return v
		}
	}
	return key
}

// Tf 先翻译 format 串，再按 fmt.Sprintf 展开参数。
func Tf(key string, args ...interface{}) string {
	return fmt.Sprintf(T(key), args...)
}
