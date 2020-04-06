package demo

import (
	"github.com/bxcodec/faker/v3"
	"github.com/xbonlinenet/goup/frame/cc"
	"github.com/xbonlinenet/goup/frame/gateway"
	"github.com/xbonlinenet/goup/frame/util"
)

type ConfigRequest struct {
	UserId  int64  `json:"userId" desc:"message need Config" binding:"required" `
	Version string `json:"version" desc:"client version"  binding:"required"`
}

type ConfigResponse struct {
	Code   int          `json:"code" desc:"0: success other: fail"`
	Config ClientConfig `json:"config" desc:"message need Config" binding:"required" `
}

type UpdateInfo struct {
	NeedUpdate bool   `json:"needUpdate" message:"是否需要更新版本"`
	Url        string `json:"url" message:"更新时打开的url地址"`
}

type ClientConfig struct {
	CDNHosts   map[string]string `json:"cndHosts" message:"关键模块使用的域名"`
	UpdateInfo UpdateInfo        `json:"updateInfo" message:"客户端强制更新的配置信息"`
}

type ConfigHandler struct {
	Request  ConfigRequest
	Response ConfigResponse
}

func (h ConfigHandler) Mock() interface{} {

	var config ClientConfig
	_ = faker.FakeData(&config)

	return ConfigResponse{Code: 0, Config: config}

}
func (h ConfigHandler) Handler(c *gateway.ApiContext) (interface{}, error) {
	module := cc.GetModule("/nrsqu.json")

	cdnHosts := cc.GetStringMapString(module, "cndHosts", map[string]string{"video": "s3.video.vnnn.vn", "image": "s3.image.vnnn.vn"})

	minVersion := cc.GetString(module, "minVersion", "2.2.9")
	updateUrl := cc.GetString(module, "updateUrl", "https://play.google.com/store/apps/details?id=com.yanonline.yannews")

	var updateInfo UpdateInfo
	if util.VersionCompare(h.Request.Version, minVersion) < 0 {
		updateInfo.NeedUpdate = true
		updateInfo.Url = updateUrl
	}

	return ConfigResponse{
		Code: 0,
		Config: ClientConfig{
			CDNHosts:   cdnHosts,
			UpdateInfo: updateInfo},
	}, nil

}
