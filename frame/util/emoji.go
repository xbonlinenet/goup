package util

import (
	"strings"
)

// IsSystemSupportEmoji  根据系统版本号判断系统是否支持 emoji
// osVersion: android/ios_version
func IsSystemSupportEmoji(osVersion string) bool {

	items := strings.Split(osVersion, "_")
	if len(items) < 2 {
		return false
	}

	if items[0] == "ios" {
		return true
	} else if items[0] == "android" {
		if items[1] >= "6.0.1" {
			return true
		} else {
			return false
		}
	}

	return false
}
