package cc

import (
	"github.com/xbonlinenet/goup/frame/log"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cast"
	"github.com/xbonlinenet/go_config_center"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	center *go_config_center.ConfigCenter
)

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	return filepath.Dir(filepath.Dir(path)), nil
}

func InitConfigCenter(servers []string) {

	path, err := GetCurrentPath()
	if err != nil {
		path = "../config_cache"
	} else {
		path = filepath.Join(path, "config_cache")
	}

	center = go_config_center.NewConfigCenter("", servers, path, "json")
}

func UnInitConfigCenter() {
	if center != nil {
		center.Close()
	}
}

func GetModule(path string) *go_config_center.ConfigModule {
	return center.GetModule(path)
}

func Get(module *go_config_center.ConfigModule, key string, defaultValue interface{}) interface{} {
	value := module.Get(key)
	if value == nil {
		return defaultValue
	}
	return value
}

func GetInt(module *go_config_center.ConfigModule, key string, defaultValue int) int {
	value := module.Get(key)
	if value == nil {
		return defaultValue
	}

	i, err := cast.ToIntE(value)
	if err != nil {
		return defaultValue
	}
	return i
}

func GetBool(module *go_config_center.ConfigModule, key string, defaultValue bool) bool {
	value := module.Get(key)
	if value == nil {
		return defaultValue
	}

	i, err := cast.ToBoolE(value)
	if err != nil {
		return defaultValue
	}
	return i
}

func GetString(module *go_config_center.ConfigModule, key string, defaultValue string) string {
	value := module.Get(key)
	if value == nil {
		return defaultValue
	}
	str, err := cast.ToStringE(value)
	if err != nil {
		return defaultValue
	}
	return str
}

func GetFloat(module *go_config_center.ConfigModule, key string, defaultValue float64) float64 {
	value := module.Get(key)
	if value == nil {
		return defaultValue
	}
	f, err := cast.ToFloat64E(value)
	if err != nil {
		return defaultValue
	}
	return f
}

func GetStringMap(module *go_config_center.ConfigModule, key string, defaultValue map[string]interface{}) map[string]interface{} {
	value := module.Get(key)
	if value == nil {
		return defaultValue
	}
	f, err := cast.ToStringMapE(value)
	if err != nil {
		return defaultValue
	}
	return f
}

func GetStringMapString(module *go_config_center.ConfigModule, key string, defaultValue map[string]string) map[string]string {
	value := module.Get(key)
	if value == nil {
		return defaultValue
	}
	f, err := cast.ToStringMapStringE(value)
	if err != nil {
		return defaultValue
	}
	return f
}

func GetRawString(module *go_config_center.ConfigModule, key string, defaultValue string) string {
	value := module.Get(key)
	if value == nil {
		return defaultValue
	}
	data, err := jsoniter.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func GetRaw(path string, defaultValue string) string {
	module := GetModule(path)
	buf := module.Raw()
	log.Default().Debug("GetRaw", zap.String("path", path), zap.String("buf", string(buf)))
	if len(buf) == 0 {
		return defaultValue
	}
	return string(buf)
}

func ConvertStringToMapStruct(json string) map[string]interface{} {
	var value map[string]interface{}
	err := jsoniter.Unmarshal([]byte(json), &value)
	if err != nil {
		panic(err)
	}
	return value
}
