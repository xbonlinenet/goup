package util

import "github.com/go-redis/redis"

// CheckError 检查错误，如果有错误会 Painc
func CheckError(err error) {
	if err != nil && err != redis.Nil {
		panic(err)
	}
}

// Max 返回最大值
func Max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// IntMax 返回最大值
func IntMax(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// Min 返回最小值
func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

// IntMin 返回最小值
func IntMin(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func StringArrayContains(arr []string, foo string) bool {
	for _, a := range arr {
		if a == foo {
			return true
		}
	}
	return false
}

func Int64Contains(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func InterfaceToInt64Slice(s []interface{}) []int64 {
	res := make([]int64, len(s))
	for i := range s {
		res[i] = s[i].(int64)
	}

	return res
}

func Int64ToInterfaceSlice(s []int64) []interface{} {
	res := make([]interface{}, len(s))
	for i := range s {
		res[i] = s[i]
	}
	return res
}
