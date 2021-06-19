package cclocal

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cast"
	"github.com/xbonlinenet/goup/frame/cc"
	"github.com/xbonlinenet/goup/frame/log"
	"go.uber.org/zap"
)


func GetCurrentPath() (string, error) {
	return cc.GetCurrentPath()
}

func InitConfigCenter(servers []string) {
	cc.InitConfigCenter(servers)
}

func UnInitConfigCenter() {
	cc.UnInitConfigCenter()
}

func GetModule(path string) LocalConfigReader {
	return GetLocalConfig(path)
}

func Get(module LocalConfigReader, key string, defaultValue interface{}) interface{} {
	value := module.Get(key)
	if value == nil {
		return defaultValue
	}
	return value
}

func GetInt(module LocalConfigReader, key string, defaultValue int) int {
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

func GetBool(module LocalConfigReader, key string, defaultValue bool) bool {
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

func GetString(module LocalConfigReader, key string, defaultValue string) string {
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

func GetFloat(module LocalConfigReader, key string, defaultValue float64) float64 {
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

func GetStringMap(module LocalConfigReader, key string, defaultValue map[string]interface{}) map[string]interface{} {
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

func GetStringMapString(module LocalConfigReader, key string, defaultValue map[string]string) map[string]string {
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

func GetRawString(module LocalConfigReader, key string, defaultValue string) string {
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
