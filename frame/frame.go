package frame

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/go-errors/errors"
	"github.com/spf13/viper"

	"github.com/xbonlinenet/goup/frame/cc"
	"github.com/xbonlinenet/goup/frame/flags"

	"github.com/xbonlinenet/goup/frame/alter"
	"github.com/xbonlinenet/goup/frame/data"
	"github.com/xbonlinenet/goup/frame/dyncfg"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/util"
)

var (
	config string
	start  int64

	ctx    context.Context
	cancel context.CancelFunc
)

func Bootstrap(run func(), options ...Option) {
	config := &bootstarpServerConfig{
		customSqlConf:   make(map[string]*data.SQLConfig),
		custonRedisConf: make(map[string]*data.RedisConfig),
	}
	for _, opt := range options {
		opt.apply(config)
	}

	if config.beforeInit != nil {
		config.beforeInit()
	}

	initFrameWorkImpl(config)
	defer UnInitFramework()
	run()
}

func initFrameWorkImpl(serverConfig *bootstarpServerConfig) {
	flags.DisplayCompileTimeFlags()

	start = time.Now().Unix()

	flag.StringVar(&config, "config", "./conf/application.yml", "")
	flag.Parse()

	ctx, cancel = context.WithCancel(context.Background())

	dir := path.Dir(config)

	if !path.IsAbs(dir) {
		wd, err := os.Getwd()
		util.CheckError(err)
		dir = path.Join(wd, dir)
	}
	filePath := path.Join(dir, path.Base(config))
	viper.AddConfigPath(".")
	viper.SetConfigFile(filePath)
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	util.CheckError(err)

	includes := viper.GetStringMapString("include")
	if len(includes) > 0 {
		loadIncludeConfigFiles(includes, dir)
	}
	log.Init()

	v, _ := json.Marshal(viper.AllSettings())

	log.Default().Info(fmt.Sprintf("config: %s", v))

	util.InitGlobeInfo()

	callInitFuncByConfigCondition(func() {
		data.InitSQLMgr(serverConfig.customSqlConf)
	}, "data.InitSQLMgr", serverConfig.initDbDisabled, "data.mysql")

	callInitFuncByConfigCondition(func() {
		data.InitRedisMgr(serverConfig.custonRedisConf)
	}, "data.InitRedisMgr", serverConfig.initRedisDisabled, "data.redis")

	callInitFuncByConfigCondition(data.InitESMgr, "data.InitESMgr", serverConfig.initEsDisabled, "data.es")
	callInitFuncByConfigCondition(func() {
		data.InitKafka(ctx)
	}, "data.InitKafka", serverConfig.initKafkaDisabled, "data.kafka")

	users := viper.GetStringSlice("alter.users")
	robotUrls := viper.GetStringSlice("alter.robot-urls")
	if len(users) > 0 || len(robotUrls) > 0 {
		alter.InitAlter()
	}

	log.Default().Info(fmt.Sprintf("Version: %s", util.Version))
	log.Default().Info(fmt.Sprintf("Compile: %s", util.Compile))

	// 初始化动态配置项
	appName := viper.GetString("application.name")
	appDyncfg := viper.GetString("application.dyncfg")
	if len(appDyncfg) != 0 {
		dyncfg.Init(appName, appDyncfg, ctx)
	}

	zkServers := viper.GetStringSlice("data.zk.config.servers")
	if len(zkServers) == 0 {
		if !util.IsRunningInDockerContainer() {
			panic("zkServers is empty")
		} else {
			log.Default().Warn("!!! 'zkServers' not configured, and service is running in container.")
		}
	} else {
		cc.InitConfigCenter(zkServers)
	}
}

// InitFramework 初始化框架
func InitFramework() {
	config := &bootstarpServerConfig{
		customSqlConf:   make(map[string]*data.SQLConfig),
		custonRedisConf: make(map[string]*data.RedisConfig),
	}

	initFrameWorkImpl(config)
}

func callInitFuncByConfigCondition(initFunc func(), funcName string, disabled bool, keyInCfgFile string) {
	if !disabled {
		if viper.Sub(keyInCfgFile) != nil {
			initFunc()
		} else {
			log.Sugar().Warnf("%s not configured, %s() skipped.", keyInCfgFile, funcName)
		}
	} else {
		log.Sugar().Infof("%s not allowed", funcName)
	}
}

func loadIncludeConfigFiles(items map[string]string, basePath string) {
	for _, v := range items {
		p := path.Join(basePath, v)
		viper.SetConfigFile(p)
		viper.MergeInConfig()
	}
}

// UnInitFramework 反初始化框架
func UnInitFramework() {

	if err := recover(); err != nil {
		goErr := errors.Wrap(err, 0)

		alter.NotifyError(goErr.Error(), goErr.Err)
		os.Exit(1)
	}
	cost := time.Now().Unix() - start
	log.Default().Info(fmt.Sprintf("Total Cost: %d", cost))
	cc.UnInitConfigCenter()
	data.UnInitSQLMgr()
	data.UninitRedisMgr()
	data.UninitESMgr()

	cancel()
}

// GetConfig 获取 config 参数
func GetConfig() string {
	return config
}
