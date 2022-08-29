package data

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql" // apply mysql driver
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/util"
	"go.uber.org/zap"
)

const (
	DBTypeMySQL    = "mysql"
	DBTypePostgres = "postgres"
)

type SQLConfig struct {
	Type        string
	URL         string
	MaxIdelConn int
	MaxOpenConn int
}

const (
	checkDBHostIPInterval    = time.Second * 5
	monitorDbMetricsInterval = time.Second * 15
)

var (
	// ErrSQLConfig 配置错误
	ErrSQLConfig = errors.New("sql Cfg error")

	// ErrSQLNotInited 还未初始化
	ErrSQLNotInited = errors.New("sql not inited")

	// ErrSQLUnknowType 位置类型数据库
	ErrSQLUnknowType = errors.New("sql server type is unknown")
)

var sqlMgr *SQLDBMgr

// InitSQLMgr 初始化 sqlMgr
func InitSQLMgr(custom map[string]*SQLConfig) {
	dbSection := viper.Sub("data.db")
	for item := range dbSection.AllSettings() {
		conf := dbSection.Sub(item)

		cType := conf.GetString("type")
		if cType != "" && !util.StringArrayContains([]string{DBTypeMySQL, DBTypePostgres}, cType) {
			panic(ErrSQLUnknowType)
		}

		dbType := DBTypeMySQL // 如果没有配置, 默认是 mysql, 兼容旧版本
		if cType == DBTypePostgres {
			dbType = DBTypePostgres
		}
		config := &SQLConfig{
			Type:        dbType,
			URL:         conf.GetString("url"),
			MaxIdelConn: conf.GetInt("max-idle-conn"),
			MaxOpenConn: conf.GetInt("max-open-conn"),
		}

		custom[item] = config
	}
	sqlMgr = newSQLDBMgr(custom)
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
	Addr   string
	AddrIP []string
}

func (c *dbConn) Close() error {
	if c.DB != nil {

		// Wait for 30s and close old conn
		for i := 0; i < 30; i++ {
			time.Sleep(time.Second)
			if c.DB.DB().Stats().InUse <= 0 {
				log.Default().Info("close db", zap.String("addr", c.Addr))
				return c.DB.Close()
			}
		}
		log.Default().Warn("close db when connect in use",
			zap.Int("in use", c.DB.DB().Stats().InUse),
			zap.String("addr", c.Addr))
		return c.DB.Close()
	}

	return nil
}

// newSQLDBMgr 根据配置创建新的数据库连接管理
func newSQLDBMgr(conf map[string]*SQLConfig) *SQLDBMgr {
	dbMgr := &SQLDBMgr{
		connMap:            make(map[string]*dbConn),
		dbMetricsCollector: NewDbMetricsCollector(),
		mutex:              &sync.Mutex{},
		dbConfig:           conf,
	}

	// start monitoring db host ip
	go dbMgr.monitorDBIP(checkDBHostIPInterval)

	// start monitoring db metrics
	go dbMgr.monitorDbMetrics(monitorDbMetricsInterval)

	return dbMgr
}

// SQLDBMgr 数据库连接管理
type SQLDBMgr struct {
	connMap            map[string]*dbConn
	dbMetricsCollector *DbMetricsCollector
	mutex              *sync.Mutex
	dbConfig           map[string]*SQLConfig
	// dbConfig *viper.Viper
	isClosed bool
}

// getDB 根据名称获取数据库连接
func (mgr *SQLDBMgr) getDB(name string) (*gorm.DB, error) {
	config, ok := mgr.dbConfig[name]
	if !ok {
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

	config, ok := mgr.dbConfig[name]
	if !ok {
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
			zap.Strings("old_ip", oldConn.AddrIP),
			zap.Strings("new_ip", newConn.AddrIP),
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
			addrIP, err := lookupAddrIP(dbConn.Addr)
			if err != nil {
				log.Default().Error("LookupIPErr", zap.String("Addr", dbConn.Addr), zap.Error(err))
				continue
			}

			// compare to current conn ip

			// if !bytes.Equal(dbConn.AddrIP, addrIP) {

			if !util.IsStringArrayEqual(dbConn.AddrIP, addrIP) {
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

func (mgr *SQLDBMgr) monitorDbMetrics(interval time.Duration) {
	for !mgr.isClosed {
		for dbName, dbConn := range mgr.connMap {
			if db := dbConn.DB.DB(); db != nil {
				mgr.dbMetricsCollector.CollectDbStats(dbName, db.Stats())
			} else {
				mgr.dbMetricsCollector.CollectDbStats(dbName, sql.DBStats{})
			}
		}

		// sleep period
		time.Sleep(interval)
	}
}

func initDB(config *SQLConfig, name string) (*gorm.DB, error) {
	switch config.Type {
	case DBTypeMySQL:
		return initMySQL(config, name)
	case DBTypePostgres:
		return initPostgres(config, name)
	default:
		panic(false)
	}
}

func initMySQL(config *SQLConfig, name string) (*gorm.DB, error) {

	db, err := gorm.Open("mysql", config.URL)
	if err != nil {
		return nil, err
	}
	debug := viper.GetBool("application.debug")
	db.SetLogger(NewWriter())
	db.LogMode(debug)

	db.DB().SetConnMaxLifetime(2 * time.Hour)
	if config.MaxIdelConn != 0 {
		db.DB().SetMaxIdleConns(config.MaxIdelConn)
	}
	if config.MaxOpenConn != 0 {
		db.DB().SetMaxOpenConns(config.MaxOpenConn)
	}

	if err := db.DB().Ping(); err != nil {
		return db, err
	}
	log.Default().Info(fmt.Sprintf("%s DB: maxIdleConn:%d, maxOpenConn: %d",
		name, config.MaxIdelConn, config.MaxOpenConn))
	return db, nil
}

func initPostgres(config *SQLConfig, name string) (*gorm.DB, error) {
	db, err := gorm.Open("postgres", config.URL)
	if err != nil {
		return nil, err
	}
	debug := viper.GetBool("application.debug")
	db.SetLogger(NewWriter())
	db.LogMode(debug)

	db.DB().SetConnMaxLifetime(2 * time.Hour)
	if config.MaxIdelConn != 0 {
		db.DB().SetMaxIdleConns(config.MaxIdelConn)
	}

	if config.MaxOpenConn != 0 {
		db.DB().SetMaxOpenConns(config.MaxOpenConn)
	}

	if err := db.DB().Ping(); err != nil {
		return db, err
	}
	log.Default().Info(fmt.Sprintf("%s DB: maxIdleConn:%d, maxOpenConn: %d",
		name, config.MaxIdelConn, config.MaxOpenConn))
	return db, nil
}

func initDBConn(config *SQLConfig, name string) (*dbConn, error) {
	dbDSN := config.URL

	addr := ""
	if config.Type == DBTypeMySQL {
		dbCfg, err := mysql.ParseDSN(dbDSN)

		if err != nil {
			return nil, err
		}
		addr = dbCfg.Addr
	} else if config.Type == DBTypePostgres {
		dbCfg, err := url.Parse(dbDSN)
		if err != nil {
			return nil, err
		}
		addr = dbCfg.Host
	}

	addrIP, err := lookupAddrIP(addr)
	if err != nil {
		return nil, err
	}

	db, err := initDB(config, name)
	if err != nil {
		return nil, err
	}

	conn := &dbConn{
		DB:     db,
		Addr:   addr,
		AddrIP: addrIP,
	}

	return conn, nil
}

func lookupAddrIP(addr string) ([]string, error) {

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

	addrs := make([]string, 0, len(addrIPs))
	for _, ip := range addrIPs {
		addrs = append(addrs, string(ip))
	}

	sort.Strings(addrs)

	return addrs, nil
}
