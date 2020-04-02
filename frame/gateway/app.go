package gateway

import (
	"encoding/json"

	"github.com/xbonlinenet/goup/frame/cc"
	"github.com/xbonlinenet/goup/frame/util"
)

// AppInfo 应用配置信息
type AppInfo struct {
	Key     string   `json:"key"`
	AppID   string   `json:"appid"`
	Secret  string   `json:"secret"`
	Partner string   `json:"partner"`
	Apis    []string `json:"apis"`
}

func GetSystemAppInfo() *AppInfo {
	return &AppInfo{
		Key:    "sys",
		Secret: "a65869bb96d77e3d9612c2682f0870690b6259ef",
	}
}

func GetAppInfo(key string) *AppInfo {
	if key == "sys" {
		return GetSystemAppInfo()
	}

	raw := cc.GetRaw("/adfly/appkey.json", `[{"key":"demo","appid":"test","secret":"ac4a10da9e7f62adb59dbe7f62adb59dbe770e8d","partner":"yomanga", "apis":["ig.entry", "demo"]}]`)

	var apps []AppInfo
	err := json.Unmarshal([]byte(raw), &apps)
	util.CheckError(err)

	appMap := make(map[string]*AppInfo)
	for i, app := range apps {
		appMap[app.Key] = &apps[i]
	}

	if app, ok := appMap[key]; ok {
		return app
	}
	return nil
}
