package alter

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/go-errors/errors"
	"github.com/xbonlinenet/alter/lib"
	"github.com/xbonlinenet/goup/frame/data"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/util"
)

var (
	client *lib.Client
)

// InitAlter 初始化报警模块
func InitAlter() {
	gateway := data.MustGetRedis("gateway")
	users := viper.GetStringSlice("alter.users")
	robotUrls := viper.GetStringSlice("alter.robot-urls")
	if len(users) == 0 && len(robotUrls) == 0 {
		panic(errors.New("Havn't config alter users or robots"))
	}

	log.Default().Info(fmt.Sprintf("Alter users: %s", users))

	var err error
	client, err = lib.NewClientV2(gateway, users, robotUrls, os.Args[0])
	util.CheckError(err)
}

// Notify 通知异常错误
func Notify(message string, detail string, errorID string) {
	client.Alter(message, detail, errorID)
}

// NotifyError 通知错误
func NotifyError(message string, err error) {
	goErr := errors.Wrap(err, 1)

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Exception: %s \n", err.Error()))
	sb.Write(goErr.Stack())

	errorStr := fmt.Sprintf("%s:%s %d", goErr.StackFrames()[0].File, goErr.StackFrames()[0].Name, goErr.StackFrames()[0].LineNumber)
	errorID := util.CalcMD5(errorStr)
	log.Default().Info(fmt.Sprintf("Occur error: %s", message))
	log.Default().Info(fmt.Sprintf("strack: %s", sb.String()))
	client.Alter(message, sb.String(), errorID)
}
