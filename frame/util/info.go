package util

import (
	"os"

	"github.com/spf13/viper"
)

var host string
var serverName string

func InitGlobeInfo() {
	host, _ = os.Hostname()
	serverName = viper.GetString("application.name")
}

func GetHost() string {
	return host
}

func GetServerName() string {
	return serverName
}
