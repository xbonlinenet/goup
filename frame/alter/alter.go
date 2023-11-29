package alter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/viper"

	"github.com/go-errors/errors"
	"github.com/xbonlinenet/alter/lib"
	"github.com/xbonlinenet/goup/frame/data"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/util"
)

var (
	client            *lib.Client
	gNotifyFuncHandle NotifyFunc
)

type NotifyFunc func(message string, detail string, errorID string)

func FeishuNotifyDemo(message string, detail string, errorID string) {
	msg := fmt.Sprintf("errorId: %s\nmessage: %s\ndetail:%s\n", errorID, message, detail)
	robotUrl := "https://open.feishu.cn/open-apis/bot/v2/hook/c4e35fdd-7d55-43b2-ba22-0e883a86dd35"
	FeishuNotify(msg, robotUrl)
}

func FeishuNotify(msg, robotUrl string) {
	// robotUrl: 你复制的webhook地址
	payload_message := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]interface{}{
			"text": msg,
		},
	}
	cli := &http.Client{}
	buf, err := json.Marshal(payload_message)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, robotUrl, bytes.NewBuffer(buf))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resBuf, err := cli.Do(req)
	if err != nil {
		return
	}
	defer resBuf.Body.Close()
}

// InitAlter 初始化报警模块
func InitAlter(notifyFuncHandle NotifyFunc) {
	if notifyFuncHandle != nil {
		gNotifyFuncHandle = notifyFuncHandle
	}

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
	if gNotifyFuncHandle != nil {
		gNotifyFuncHandle(message, detail, errorID)
		return
	}

	client.Alter(message, detail, errorID)

	// sentry 通知
	if viper.GetString("sentry.dsn") != "" {
		var evt = sentry.NewEvent()
		evt.Level = sentry.LevelError
		evt.Message = fmt.Sprintf("%s\nErrorId: %s\nDetail:\n%s", message, errorID, detail)
		sentry.CaptureEvent(evt)
	}
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
	Notify(message, sb.String(), errorID)
}
