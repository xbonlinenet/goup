package util

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"strings"
)

// CalcDocID 计算 URL 的文档 ID
func CalcDocID(url string) string {
	lowerURL := strings.ToLower(url)
	data := []byte(lowerURL)
	sha1 := sha1.Sum(data)
	return fmt.Sprintf("%x", sha1)
}

// CalcMD5 计算字符串的 MD5 值
func CalcMD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	cipherStr := h.Sum(nil)
	return fmt.Sprintf("%x", cipherStr)
}
