package util

import (
	"fmt"
	"strconv"
	"strings"
)

// GetStringByLimit 如果字符串长度超过 limit, 截取字符串
func GetStringByLimit(foo string, limit int) string {
	if len(foo) > limit {
		return foo[:limit]
	}
	return foo
}

// IntArrayJoin 将整型数据的元素连接成一个字符串，使用 delim 分割各个元素
func IntArrayJoin(arr []int, delim string) string {
	b := make([]string, len(arr))
	for i, v := range arr {
		b[i] = strconv.Itoa(v)
	}

	return strings.Join(b, delim)
}

// Int64ArrayJoin 将整型数据的元素连接成一个字符串，使用 delim 分割各个元素
func Int64ArrayJoin(arr []int64, delim string) string {
	b := make([]string, len(arr))
	for i, v := range arr {
		b[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(b, delim)
}

//Interface2String 将int、float64等类型转为string类型，若该类型不在case里面则返回空字符串
func Interface2String(inter interface{}) string {

	var a string
	switch inter.(type) {

	case string:
		a = fmt.Sprint(inter.(string))
		break
	case int:
		a = fmt.Sprint(inter.(int))
		break
	case int8:
		a = fmt.Sprint(inter.(int8))
		break
	case int16:
		a = fmt.Sprint(inter.(int16))
		break
	case int32:
		a = fmt.Sprint(inter.(int32))
		break
	case int64:
		a = fmt.Sprint(inter.(int64))
		break
	case float32:
		a = fmt.Sprint(inter.(float32))
		break
	case float64:
		a = fmt.Sprint(inter.(float64))
		break
	}

	return a

}
