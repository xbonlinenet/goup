package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	logMap = make(map[string]*zap.Logger)
)

type Conf struct {
	Interval time.Duration
	Level    string
	Console  bool
	Encoder  string //输出日志的编码方式 console (默认), json
	Logger   lumberjack.Logger
}

func Init() error {
	loggers := viper.GetStringMap("log")

	if isRunningInDockerContainer() {
		logger, err := zap.NewProduction()
		if err != nil {
			return err
		}
		for loggerName := range loggers {
			logMap[loggerName] = logger
		}

	} else {
		forceLogToStdout := viper.GetBool("application.forceLog2Stdout")
		for k, _ := range loggers {
			var conf Conf
			err := viper.Sub("log").Sub(k).Unmarshal(&conf)
			if err != nil {
				return err
			}
			log := initLogger(&conf, forceLogToStdout)
			logMap[k] = log
		}
	}

	return nil
}

func GetLogger(name string) *zap.Logger {
	return logMap[name]
}

var testLogger *zap.Logger
var testLoggerOnce sync.Once

func Default() *zap.Logger {
	if log, ok := logMap["default"]; ok {
		return log
	}

	testLoggerOnce.Do(func() {
		if isRunningInDockerContainer() {
			testLogger, _ = zap.NewProduction()
		} else {
			config := zap.NewProductionConfig()
			config.Encoding = "console"
			config.OutputPaths = []string{"stdout"}
			var err error
			testLogger, err = config.Build()
			if err != nil {
				panic(err)
			}
		}

	})
	// For test case
	return testLogger
}

func initLogger(conf *Conf, forceLogStdout bool) *zap.Logger {
	log := &conf.Logger

	ex, err := os.Executable()
	if err != nil {
		fmt.Printf("initLogger panic: %s", err.Error())
		panic(err)
	}
	exPath := filepath.Dir(filepath.Dir(ex))
	log.Filename = filepath.Join(exPath, log.Filename)
	fmt.Printf("log path: %s\n", log.Filename)

	go func() {
		for {
			<-time.After(conf.Interval)
			if !forceLogStdout {
				log.Rotate()
			}
		}
	}()

	zapLevel := transformLevel(conf.Level)

	encoder := zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:       "T",
		LevelKey:      "L",
		NameKey:       "N",
		CallerKey:     "C",
		MessageKey:    "M",
		StacktraceKey: "S",
		// FunctionKey:    "F",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var fileWriter zapcore.WriteSyncer
	if !forceLogStdout {
		if conf.Console {
			fileWriter = zapcore.NewMultiWriteSyncer(zapcore.AddSync(log), os.Stdout)
		} else {
			fileWriter = zapcore.AddSync(log)
		}
	} else {
		fileWriter = zapcore.AddSync(os.Stdout)
	}

	var coder zapcore.Encoder
	if conf.Encoder == "json" {
		coder = zapcore.NewJSONEncoder(encoder)
	} else {
		coder = zapcore.NewConsoleEncoder(encoder)
	}

	core := zapcore.NewCore(coder, fileWriter, zapLevel)

	logger := zap.New(core, zap.AddCaller())
	return logger
}

func transformLevel(level string) zapcore.Level {
	zapLevle := zapcore.InfoLevel
	l := strings.ToLower(level)
	switch l {
	case "debug":
		zapLevle = zapcore.DebugLevel
	case "info":
		zapLevle = zapcore.InfoLevel
	case "warn":
		zapLevle = zapcore.WarnLevel
	case "error":
		zapLevle = zapcore.ErrorLevel
	case "painc":
		zapLevle = zapcore.PanicLevel
	case "fatal":
		zapLevle = zapcore.FatalLevel
	}
	return zapLevle
}

// GetLogFields 传入key value对 key1，value1,key2，value2，key3，value3
// key为string，value为任意的
// length不是偶数的话，最后一个会被赋值为空
func GetLogFields(keyValuePairs ...interface{}) []zap.Field {
	length := len(keyValuePairs)
	n := length / 2
	fields := make([]zap.Field, 0, n)
	for i := 1; i < length; i += 2 {
		key := keyValuePairs[i-1]
		value := keyValuePairs[i]

		field := zap.Any(fmt.Sprint(key), value)
		fields = append(fields, field)
	}

	if length%2 != 0 {
		key := keyValuePairs[length-1]
		fields = append(fields, zap.Any(fmt.Sprint(key), ""))
	}

	return fields
}

func isRunningInDockerContainer() bool {
	// docker creates a .dockerenv file at the root
	// of the directory tree inside the container.
	// if this file exists then the viewer is running
	// from inside a container so return true

	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	return false
}
