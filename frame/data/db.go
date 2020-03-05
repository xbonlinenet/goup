package data

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/util"
	_ "github.com/go-sql-driver/mysql" // 必须包含 mysql 的驱动
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

var sqlMgr *SQLDBMgr

// ErrSQLConfig 配置错误
var ErrSQLConfig = errors.New("sql config error")

// ErrSQLNotInited Redis 还未初始化
var ErrSQLNotInited = errors.New("sql not inited")

// InitSQLMgr 初始化 Redis
func InitSQLMgr() {
	sqlMgr = newSQLDBMgr(viper.Sub("data.mysql"))
}

// UninitSQLMgr 反初始化 Redis 相关
func UninitSQLMgr() {
	if sqlMgr != nil {
		sqlMgr.Close()
		sqlMgr = nil
	}
}

// GetDB 获取 DB
func GetDB(name string) (*gorm.DB, error) {
	if sqlMgr == nil {
		panic(ErrSQLNotInited)
	}

	return sqlMgr.getDB(name)
}

// MustGetDB 获取 DB，如果获取失败，直接报错
func MustGetDB(name string) *gorm.DB {
	if sqlMgr == nil {
		panic(ErrSQLNotInited)
	}

	return sqlMgr.mustGetDB(name)
}

// newSQLDBMgr 根据配置创建新的数据库连接管理
func newSQLDBMgr(conf *viper.Viper) *SQLDBMgr {
	dbMgr := &SQLDBMgr{
		dbMap:    make(map[string]*gorm.DB),
		mutex:    &sync.Mutex{},
		dbConfig: conf,
	}
	return dbMgr
}

// SQLDBMgr 数据库连接管理
type SQLDBMgr struct {
	dbMap    map[string]*gorm.DB
	mutex    *sync.Mutex
	dbConfig *viper.Viper
}

// getDB 根据名称获取数据库连接
func (mgr *SQLDBMgr) getDB(name string) (*gorm.DB, error) {
	config := mgr.dbConfig.Sub(name)
	if config == nil {
		return nil, ErrSQLConfig
	}

	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	db, ok := mgr.dbMap[name]
	if ok {
		return db, nil
	}

	db, err := initDB(config, name)
	if err != nil {
		return nil, err
	}
	mgr.dbMap[name] = db
	return db, nil
}

// mustGetDB 根据名称获取数据库连接
func (mgr *SQLDBMgr) mustGetDB(name string) *gorm.DB {
	config := mgr.dbConfig.Sub(name)
	if config == nil {
		panic(ErrSQLConfig)
	}

	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	db, ok := mgr.dbMap[name]
	if ok {
		return db
	}

	db, err := initDB(config, name)
	util.CheckError(err)

	mgr.dbMap[name] = db
	return db
}

// Close 关闭管理器，释放数据库连接
func (mgr *SQLDBMgr) Close() {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	for _, db := range mgr.dbMap {
		db.Close()
	}
	mgr.dbMap = make(map[string]*gorm.DB)
}

func initDB(config *viper.Viper, name string) (*gorm.DB, error) {
	url := config.GetString("url")
	db, err := gorm.Open("mysql", url)
	if err != nil {
		return nil, err
	}
	debug := viper.GetBool("application.debug")
	db.LogMode(debug)

	db.DB().SetConnMaxLifetime(2 * time.Hour)
	maxIdleConn := config.GetInt("max-idle-conn")
	if maxIdleConn != 0 {
		db.DB().SetMaxIdleConns(maxIdleConn)
	}
	maxOpenConn := config.GetInt("max-open-conn")
	if maxOpenConn != 0 {
		db.DB().SetMaxOpenConns(maxOpenConn)
	}

	if err := db.DB().Ping(); err != nil {
		return db, err
	}
	log.Default().Info(fmt.Sprintf("%s db: maxIdleConn:%d, maxOpenConn: %d", name, maxIdleConn, maxOpenConn))
	return db, nil
}
