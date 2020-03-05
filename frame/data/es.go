package data

import (
	localLog "github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/util"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/olivere/elastic.v5"
)

var mutex = &sync.Mutex{}

var esClientMgr *ESClientMgr

// ErrESConfig 配置错误
var ErrESConfig = errors.New("sql config error")

// ErrESNotInited ES 还未初始化
var ErrESNotInited = errors.New("sql not inited")

// ESClientMgr esClient的管理器
type ESClientMgr struct {
	esClientMap map[string]*elastic.Client
	mutex       *sync.Mutex // 用于防止并发调用导致出错
	dbConfig    *viper.Viper
}

// InitESMgr 初始化ESMgr
func InitESMgr() {
	esClientMgr = newESClientMgr(viper.Sub("data.es"))
}

// newESClientMgr 根据配置创建新的数据库连接管理
func newESClientMgr(conf *viper.Viper) *ESClientMgr {
	esClientMgr := &ESClientMgr{
		esClientMap: make(map[string]*elastic.Client),
		mutex:       &sync.Mutex{},
		dbConfig:    conf,
	}
	return esClientMgr
}

// UninitESMgr 反初始化 ES 相关
func UninitESMgr() {
	if esClientMgr != nil {
		esClientMgr.Close()
		esClientMgr = nil
	}
}

// MustGetESClient 获取 ES，如果获取失败，直接报错
func MustGetESClient(name string) *elastic.Client {
	if esClientMgr == nil {
		panic(ErrSQLNotInited)
	}

	return esClientMgr.mustGetESClient(name)
}

// getESClient 根据名称获取数据库连接，抛回error
func (mgr *ESClientMgr) getESClient(name string) (*elastic.Client, error) {
	config := mgr.dbConfig.Sub(name)
	if config == nil {
		return nil, ErrSQLConfig
	}

	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	esClient, ok := mgr.esClientMap[name]
	if ok {
		return esClient, nil
	}

	esClient, err := initESClient(config, name)
	if err != nil {
		return nil, err
	}
	mgr.esClientMap[name] = esClient
	return esClient, nil
}

// mustGetESClient 根据名称获取数据库连接，不抛回error
func (mgr *ESClientMgr) mustGetESClient(name string) *elastic.Client {
	config := mgr.dbConfig.Sub(name)
	if config == nil {
		panic(ErrSQLConfig)
	}

	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	esClient, ok := mgr.esClientMap[name]
	if ok {
		return esClient
	}

	esClient, err := initESClient(config, name)
	util.CheckError(err)

	mgr.esClientMap[name] = esClient
	return esClient
}

// Close 关闭管理器，释放数据库连接
func (mgr *ESClientMgr) Close() {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	for _, esClient := range mgr.esClientMap {
		// close
		esClient.Stop()
	}
	mgr.esClientMap = make(map[string]*elastic.Client)
}

func initESClient(config *viper.Viper, name string) (*elastic.Client, error) {
	host := config.GetString("url")
	//var host = "http://ad.com:9200/"
	esClient, err := elastic.NewClient(
		elastic.SetURL(host),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10*time.Second),
		elastic.SetGzip(true),
		// TODO 日志可以优化
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
		elastic.SetInfoLog(log.New(os.Stdout, "", log.LstdFlags)))

	if err != nil {
		return nil, err
	}
	localLog.Default().Info(fmt.Sprintf("%s es: %s", name, host))
	return esClient, nil
}
