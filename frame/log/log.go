package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	Logger   lumberjack.Logger
}

func Init() error {
	loggers := viper.GetStringMap("log")
	for k, _ := range loggers {
		var conf Conf
		err := viper.Sub("log").Sub(k).Unmarshal(&conf)
		if err != nil {
			return err
		}
		log := initLogger(&conf)
		logMap[k] = log
	}
	return nil
}

func GetLogger(name string) *zap.Logger {
	return logMap[name]
}

func Default() *zap.Logger {
	if log, ok := logMap["default"]; ok {
		return log
	}
	// For test case
	return zap.NewExample()
}

func initLogger(conf *Conf) *zap.Logger {
	log := &conf.Logger

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(filepath.Dir(ex))
	log.Filename = filepath.Join(exPath, log.Filename)
	fmt.Printf("path: %s\n", log.Filename)

	go func() {
		for {
			<-time.After(conf.Interval)
			log.Rotate()
		}
	}()

	zapLevle := zapcore.InfoLevel
	l := strings.ToLower(conf.Level)
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

	encoder := zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	w := zapcore.AddSync(log)
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoder),
		w,
		zapLevle,
	)

	logger := zap.New(core, zap.AddCaller())
	return logger
}
