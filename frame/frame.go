package frame

import (
	"github.com/xbonlinenet/goup/frame/cc"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/xbonlinenet/goup/frame/alter"
	"github.com/xbonlinenet/goup/frame/data"
	"github.com/xbonlinenet/goup/frame/dyncfg"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/util"
	"github.com/go-errors/errors"
	"github.com/spf13/viper"
)

var (
	config string
	start  int64

	ctx    context.Context
	cancel context.CancelFunc
)

func Bootstrap(run func()) {
	InitFramework()
	defer UnInitFramework()
	run()
}

// InitFramework 初始化框架
func InitFramework() {
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

	data.InitSQLMgr()
	data.InitRedisMgr()
	data.InitESMgr()

	users := viper.GetStringSlice("alter.users")
	if len(users) > 0 {
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
		panic("zkServers is empty")
	}
	cc.InitConfigCenter(zkServers)
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
	data.UninitSQLMgr()
	data.UninitRedisMgr()
	data.UninitESMgr()

	cancel()
}

// GetConfig 获取 config 参数
func GetConfig() string {
	return config
}
