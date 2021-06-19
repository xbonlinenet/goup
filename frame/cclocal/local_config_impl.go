package cclocal


import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"io/ioutil"
	"sync"
)

type LocalConfigReader interface{
	SetConfigType(configType string)
	Get(key string) interface{}
	GetInt(key string) int
	GetString(key string) string
	GetStringSlice(key string) []string
	GetBool(key string) bool
	GetFloat64(key string) float64
	GetStringMapString(key string) map[string]string
	GetStringMap(key string) map[string]interface{}
	GetAll() map[string]interface{}
	Raw() []byte
}

type localConfigMgr struct {
	lock          sync.Mutex
	cfg2reader    map[string]LocalConfigReader
}

var (
	gLocalCfgMgr = new(localConfigMgr)
)

type localConfigContext struct {
	cfgFile       string
	buf           []byte
	cfgType 	  string
	cfg           *viper.Viper
	loaded 		  bool
	lock          sync.RWMutex
}

func GetLocalConfig(cfgFile string) LocalConfigReader {
	gLocalCfgMgr.lock.Lock()
	defer gLocalCfgMgr.lock.Unlock()

	fmt.Printf("GetLoalConfig(%s) called\n", cfgFile)

	if gLocalCfgMgr.cfg2reader == nil {
		gLocalCfgMgr.cfg2reader = make(map[string]LocalConfigReader)
	}

	if reader, ok := gLocalCfgMgr.cfg2reader[cfgFile]; ok {
		return reader
	}

	cfgCtx := &localConfigContext{
		cfgFile: cfgFile,
		cfg: viper.New(),
	}

	gLocalCfgMgr.cfg2reader[cfgFile] = cfgCtx
	return cfgCtx
}

func (c *localConfigContext) loadConfigData() error {
	if c.loaded {
		return nil
	}
	if c.cfg == nil {
		return errors.New("viper instance not created before loadConfigData()")
	}
	if c.cfgType != "" {
		c.cfg.SetConfigType(c.cfgType)
	}
	c.cfg.AddConfigPath(".")
	fmt.Println("local config file is:", c.cfgFile)
	c.cfg.SetConfigFile(c.cfgFile)
	if err := c.cfg.ReadInConfig(); err != nil {
		return err
	}
	c.cfg.WatchConfig()
	c.cfg.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		if err := c.readRawData(); err != nil {
			fmt.Println("read raw data error:", err)
		}
	})

	if err := c.readRawData(); err != nil {
		return err
	}

	c.loaded = true
	return nil
}

func (c *localConfigContext) readRawData() error {
	if data, err := ioutil.ReadFile(c.cfgFile); err != nil {
		return err
	}else{
		c.buf = make([]byte, len(data))
		copy(c.buf, data)
	}
	return nil
}

func (c *localConfigContext) SetConfigType(configType string) {
	if configType != "" {
		c.cfgType = configType
	}
}

func (c *localConfigContext) Get(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.loadConfigData()
	return c.cfg.Get(key)
}

func (c *localConfigContext) GetInt(key string) int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.loadConfigData()
	return c.cfg.GetInt(key)
}

func (c *localConfigContext) GetString(key string) string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.loadConfigData()
	return c.cfg.GetString(key)
}

func (c *localConfigContext) GetStringSlice(key string) []string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.loadConfigData()
	return c.cfg.GetStringSlice(key)
}

func (c *localConfigContext) GetBool(key string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.loadConfigData()
	return c.cfg.GetBool(key)
}

func (c *localConfigContext) GetFloat64(key string) float64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.loadConfigData()
	return c.cfg.GetFloat64(key)
}

func (c *localConfigContext) GetStringMapString(key string) map[string]string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.loadConfigData()
	return c.cfg.GetStringMapString(key)
}

func (c *localConfigContext) GetStringMap(key string) map[string]interface{}{
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.loadConfigData()
	return c.cfg.GetStringMap(key)
}

func (c *localConfigContext) GetAll() map[string]interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.loadConfigData()
	return c.cfg.AllSettings()
}

func (c *localConfigContext) Raw() []byte {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.loadConfigData()
	return c.buf
}
