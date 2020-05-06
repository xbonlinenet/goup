package data

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql" // apply mysql driver
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/util"
)

var (
	// ErrSQLConfig 配置错误
	ErrSQLConfig = errors.New("sql Cfg error")

	// ErrSQLNotInited 还未初始化
	ErrSQLNotInited = errors.New("sql not inited")
)

var sqlMgr *SQLDBMgr

// InitSQLMgr 初始化 sqlMgr
func InitSQLMgr() {
	sqlMgr = newSQLDBMgr(viper.Sub("data.mysql"))
}

// UnInitSQLMgr 反初始化 sqlMgr 相关
func UnInitSQLMgr() {
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

type dbConn struct {
	DB     *gorm.DB
	Cfg    *mysql.Config
	AddrIP net.IP
}

func (c *dbConn) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}

	return nil
}

// newSQLDBMgr 根据配置创建新的数据库连接管理
func newSQLDBMgr(conf *viper.Viper) *SQLDBMgr {
	dbMgr := &SQLDBMgr{
		connMap:  make(map[string]*dbConn),
		mutex:    &sync.Mutex{},
		dbConfig: conf,
	}

	// start monitoring db host ip
	go dbMgr.monitorDBIP(time.Second * 5)

	return dbMgr
}

// SQLDBMgr 数据库连接管理
type SQLDBMgr struct {
	connMap  map[string]*dbConn
	mutex    *sync.Mutex
	dbConfig *viper.Viper
	isClosed bool
}

// getDB 根据名称获取数据库连接
func (mgr *SQLDBMgr) getDB(name string) (*gorm.DB, error) {
	config := mgr.dbConfig.Sub(name)
	if config == nil {
		return nil, ErrSQLConfig
	}

	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	conn, ok := mgr.connMap[name]
	if ok {
		return conn.DB, nil
	}

	conn, err := initDBConn(config, name)
	if err != nil {
		return nil, err
	}

	mgr.connMap[name] = conn
	return conn.DB, nil
}

// mustGetDB 根据名称获取数据库连接
func (mgr *SQLDBMgr) mustGetDB(name string) *gorm.DB {
	db, err := mgr.getDB(name)
	util.CheckError(err)
	return db
}

// Close 关闭管理器，释放数据库连接
func (mgr *SQLDBMgr) Close() {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	for _, conn := range mgr.connMap {
		if closeErr := conn.Close(); closeErr != nil {
			log.Default().Error("closeConnErr", zap.Error(closeErr))
		}
	}

	mgr.connMap = make(map[string]*dbConn)
	mgr.isClosed = true
}

func (mgr *SQLDBMgr) updateDBConn(name string) error {
	oldConn, ok := mgr.connMap[name]
	if !ok {
		return errors.New("oldConn not found")
	}

	config := mgr.dbConfig.Sub(name)
	if config == nil {
		return ErrSQLConfig
	}

	newConn, err := initDBConn(config, name)
	if err != nil {
		return err
	}

	mgr.mutex.Lock()

	if !mgr.isClosed {
		mgr.connMap[name] = newConn
		log.Default().Info(
			"updateDBConn",
			zap.String("name", name),
			zap.String("old_ip", oldConn.AddrIP.String()),
			zap.String("new_ip", newConn.AddrIP.String()),
		)
	}

	mgr.mutex.Unlock()

	if err = oldConn.Close(); err != nil {
		log.Default().Error("close old conn err", zap.Error(err))
	}

	return nil
}

// monitorDBIP 监控域名的 IP 是否发生变化
func (mgr *SQLDBMgr) monitorDBIP(period time.Duration) {
	for !mgr.isClosed {

		var waitUpdate []string

		// iter over connections and check ip
		for dbName, dbConn := range mgr.connMap {
			addrIP, err := lookupAddrIP(dbConn.Cfg.Addr)
			if err != nil {
				log.Default().Error("LookupIPErr", zap.String("Addr", dbConn.Cfg.Addr), zap.Error(err))
				continue
			}

			// compare to current conn ip
			if !bytes.Equal(dbConn.AddrIP, addrIP) {
				waitUpdate = append(waitUpdate, dbName)
			}
		}

		// update connections
		for _, dbName := range waitUpdate {
			if err := mgr.updateDBConn(dbName); err != nil {
				log.Default().Error("UpdateConnErr", zap.String("name", dbName), zap.Error(err))
			}
		}

		// sleep period
		time.Sleep(period)
	}
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
	log.Default().Info(fmt.Sprintf("%s DB: maxIdleConn:%d, maxOpenConn: %d", name, maxIdleConn, maxOpenConn))
	return db, nil
}

func initDBConn(config *viper.Viper, name string) (*dbConn, error) {
	dbDSN := config.GetString("url")
	dbCfg, err := mysql.ParseDSN(dbDSN)

	if err != nil {
		return nil, err
	}

	addrIP, err := lookupAddrIP(dbCfg.Addr)
	if err != nil {
		return nil, err
	}

	db, err := initDB(config, name)
	if err != nil {
		return nil, err
	}

	conn := &dbConn{
		DB:     db,
		Cfg:    dbCfg,
		AddrIP: addrIP,
	}

	return conn, nil
}

func lookupAddrIP(addr string) ([]byte, error) {

	// parse host from addr
	addrSplits := strings.Split(addr, ":")
	addrHost := addrSplits[0]

	// lookup IP by host
	addrIPs, err := net.LookupIP(addrHost)
	if err != nil {
		return nil, err
	}

	if len(addrIPs) == 0 {
		return nil, errors.New("no IP")
	}

	return addrIPs[0], nil
}
