package frame

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/xbonlinenet/goup/frame/alter"
	"github.com/xbonlinenet/goup/frame/data"
	"github.com/xbonlinenet/goup/frame/gateway"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/recovery"
	"github.com/xbonlinenet/goup/frame/util"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func BootstrapServer(ctx context.Context, options ...Option) {

	config := &bootstarpServerConfig{
		customSqlConf:     make(map[string]*data.SQLConfig),
		custonRedisConf:   make(map[string]*data.RedisConfig),
		enableHttpHealthz: true,
		pprofToken:        "hello world",
	}

	for _, opt := range options {
		opt.apply(config)
	}

	if config.beforeInit != nil {
		config.beforeInit()
	}

	initFrameWorkImpl(config)
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
	r.Use(gateway.APIMiddleware(config.customApiPathPrefix))

	if config.customRouter != nil {
		config.customRouter(r)
	}

	if config.versionHandler != nil {
		r.GET("/version", config.versionHandler)
	}

	if config.beforeServerRun != nil {
		config.beforeServerRun()
	}
	// pprof 真正收集数据是在访问服务的时候
	// 默认打开方便后续定位性能问题
	// $ curl -o profile.out https://host/debug/pprof -X 'X-Authorization: $TOKEN'
	// $ go tool pprof profile.out
	debugGroup := r.Group("/debug", func(c *gin.Context) {
		if config.pprofToken == "" {
			c.Next()
		} else {
			if c.Request.Header.Get("Authorization") != config.pprofToken {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Next()
		}
	})
	pprof.RouteRegister(debugGroup, "pprof")

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

	if config.enableHttpHealthz {
		r.GET("/system/healthz", gateway.HttpHealthz)
		r.GET("/healthz", gateway.HttpHealthz)
	}

	addr := viper.GetString("server.addr")

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go server.ListenAndServe()

	// 监听退出信号
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	sig := <-ch
	fmt.Println("got a signal", sig)
	if config.beforeServerExit != nil {
		fmt.Println("executing hook function，server will be soon shutdown after hook finish")
		config.beforeServerExit() // 执行hook函数
	}
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := server.Shutdown(ctx)
	if err != nil {
		fmt.Println("shutdown err: ", err)
	}
	if config.afterServerExit != nil {
		fmt.Println("executing hook function，server has shutdown.")
		config.afterServerExit()
	}
	fmt.Println("----exited-----", time.Since(now))
}

type bootstarpServerConfig struct {
	beforeInit          func()
	beforeServerRun     func()
	customRouter        func(r *gin.Engine)
	versionHandler      func(c *gin.Context)
	reportApiDocAddr    string
	initDbDisabled      bool
	initRedisDisabled   bool
	initEsDisabled      bool
	initKafkaDisabled   bool
	enableHttpHealthz   bool
	middlewareList      []gin.HandlerFunc
	customSqlConf       map[string]*data.SQLConfig   // 自定义的 Mysql 配置
	custonRedisConf     map[string]*data.RedisConfig // 自定义 Redis 配置
	customApiPathPrefix string
	beforeServerExit    func() // 注册hook函数，在服务优雅关闭之前执行
	afterServerExit     func()
	dbErrorCallback     data.DbErrorCallback // DB 错误回调
	pprofToken          string
	NotifyFuncHandler   alter.NotifyFunc
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

	payload, err := json.Marshal(&params)
	util.CheckError(err)

	resp, err := httpClient.Post(reportAddr, "application/json; encoding=utf-8", bytes.NewReader(payload))
	if err == nil {
		resp.Body.Close()
	} else {
		log.Default().Sugar().Error("report doc info error", err.Error())
	}

}
