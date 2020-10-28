package frame

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/xbonlinenet/goup/frame/gateway"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/recovery"
	"github.com/xbonlinenet/goup/frame/util"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func BootstrapServer(ctx context.Context, options ...Option) {

	config := &bootstarpServerConfig{}

	for _, opt := range options {
		opt.apply(config)
	}

	if config.beforeInit != nil {
		config.beforeInit()
	}

	if config.defaultLocalCachePathEnable {
		defaultConfigCenterLocalCacheDir = "/usr/local/vntop/config_center/local_cache/"
	}
	InitFramework()

	defer UnInitFramework()

	r := gin.New()

	p := ginprometheus.NewPrometheus("gin")
	p.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		url := c.Request.URL.Path
		return url
	}

	p.Use(r)

	// Option: middleware
	for _, middle := range config.middlewareList {
		r.Use(middle)
	}

	r.Use(recovery.Recovery())
	r.Use(gateway.APIMiddleware())

	if config.customRouter != nil {
		config.customRouter(r)
	}

	if config.versionHandler != nil {
		r.GET("/version", config.versionHandler)
	}

	if config.beforeServerRun != nil {
		config.beforeServerRun()
	}

	debug := viper.GetBool("application.debug")
	if debug {
		gin.SetMode("debug")
		r.GET("/doc/list", gateway.ApiList)
		r.GET("/doc/detail", gateway.ApiDetail)

		if len(config.reportApiDocAddr) > 0 {
			addr := viper.GetString("server.addr")
			p := strings.Split(addr, ":")[1]
			port, err := strconv.ParseInt(p, 10, 64)
			util.CheckError(err)

			reportApi(viper.GetString("application.name"), port, config.reportApiDocAddr)
		}

	} else {
		gin.SetMode("release")
	}

	addr := viper.GetString("server.addr")
	err := r.Run(addr)
	util.CheckError(err)
	log.Sugar().Errorf("Occur err: %s", err.Error())
}

type bootstarpServerConfig struct {
	beforeInit                  func()
	beforeServerRun             func()
	customRouter                func(r *gin.Engine)
	versionHandler              func(c *gin.Context)
	reportApiDocAddr            string
	middlewareList              []gin.HandlerFunc
	defaultLocalCachePathEnable bool
}

var httpClient = &http.Client{
	Timeout: time.Second * 5,
}

type ApiInfo struct {
	Key  string `json:"key"`
	Desc string `json:"desc"`
}

func reportApi(server string, port int64, reportAddr string) {

	apis := gateway.GetApiList()

	apiList := make([]ApiInfo, 0, len(apis))
	for _, api := range apis {
		apiList = append(apiList, ApiInfo{
			Key:  api.Group + "." + api.Key,
			Desc: api.Name,
		})
	}
	params := map[string]interface{}{
		"server":  server,
		"port":    port,
		"apiList": apiList,
	}

	payload, err := gateway.Json.Marshal(&params)
	util.CheckError(err)

	resp, err := httpClient.Post(reportAddr, "application/json; encoding=utf-8", bytes.NewReader(payload))
	if err == nil {
		resp.Body.Close()
	} else {
		log.Default().Sugar().Error("report doc info error", err.Error())
	}

}
