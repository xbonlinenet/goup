package util

import (
	"regexp"

	"golang.org/x/text/unicode/norm"
)

// TextNormalizer 去除越南语隐藏字符,规范声调字符
func TextNormalizer(text string) string {
	text = norm.NFC.String(text)
	re1, _ := regexp.Compile("[\u200b-\u200f\ufeff]")
	text = re1.ReplaceAllString(text, "")
	return text
}

// NormalizerMultiWhiteSpace 将多个空白符替换成一个
func NormalizerMultiWhiteSpace(text string) string {
	re1, _ := regexp.Compile("\\s+")
	text = re1.ReplaceAllString(text, " ")
	return text
}

// RemoveHtmlTag 去除html标签
func RemoveHtmlTag(text string) string {
	re2, _ := regexp.Compile("</?\\w+[^>]*>")
	text = re2.ReplaceAllString(text, "")
	return text
}
