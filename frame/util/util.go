package util

import (
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

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

func VersionCompare(v1, v2 string) int {
	v1List := strings.Split(v1, ".")
	v2List := strings.Split(v2, ".")
	if v1 == v2 {
		return 0
	}

	i := 0
	for ; i < len(v1List) && i < len(v2List); i++ {
		n1, _ := strconv.Atoi(v1List[i])
		n2, _ := strconv.Atoi(v2List[i])
		if n1 < n2 {
			return -1
		}

		if n1 > n2 {
			return +1
		}
	}

	if i < len(v1List) {
		return +1
	}

	return -1
}
