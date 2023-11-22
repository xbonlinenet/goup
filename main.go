package main

func Test() {}

// pkg cmd demo main

// import (
// 	"context"
// 	"fmt"
// 	"net/http"

// 	"go.uber.org/zap"

// 	"github.com/gin-gonic/gin"
// 	"github.com/jinzhu/gorm"

// 	"github.com/xbonlinenet/goup/demo"
// 	"github.com/xbonlinenet/goup/frame"
// 	"github.com/xbonlinenet/goup/frame/alter"
// 	"github.com/xbonlinenet/goup/frame/data"
// 	"github.com/xbonlinenet/goup/frame/flags"
// 	"github.com/xbonlinenet/goup/frame/gateway"
// 	"github.com/xbonlinenet/goup/frame/log"
// )

// func main() {

// 	customSqlConf := map[string]*data.SQLConfig{
// 		"custom": {
// 			URL: "test:test@tcp(192.168.0.22:3306)/goup?parseTime=True&loc=Local&multiStatements=true&charset=utf8mb4",
// 		},
// 	}

// 	ctx := context.Background()
// 	frame.BootstrapServer(
// 		ctx,
// 		frame.BeforeServerRun(registerRouter),
// 		frame.Version(version),
// 		frame.CustomRouter(customRouter),
// 		frame.ReportApi("http://192.168.0.22:14000/api/doc/report"),
// 		frame.CustomSqlConf(customSqlConf),
// 		frame.BeforeServerExit(beforeServerExit),
// 		frame.SetDbErrorCallback(setDbErrorCallback),
// 		frame.SetNotifyFuncHandle(feishuNotify),
// 	)
// }

// func beforeServerExit() {
// 	fmt.Println("it will be done before server shutdown")
// }

// func setDbErrorCallback(name, queryType, sql string, err error, scope *gorm.Scope) {
// 	if gorm.IsRecordNotFoundError(err) {
// 		fmt.Printf("[DbErrorCallback] record not found, data name: %s, query type: %s, sql: %s, err: %s\n", name, queryType, sql, err)
// 	}

// 	fmt.Printf("[DbErrorCallback] data name: %s, query type: %s, sql: %s, err: %s\n", name, queryType, sql, err)
// }

// func feishuNotify(message string, detail string, errorID string) {
// 	msg := fmt.Sprintf("errorId: %s\nmessage: %s\ndetail:%s\n", errorID, message, detail)
// 	robotUrl := "https://open.feishu.cn/open-apis/bot/v2/hook/c4e35fdd-7d55-43b2-ba22-0e883a86dd35"
// 	alter.FeishuNotify(msg, robotUrl)
// }

// func version(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"version":    flags.GitCommit,
// 		"build_time": flags.BuildTime,
// 		"build_env":  flags.BuildEnv,
// 	})
// }

// func customRouter(r *gin.Engine) {
// 	r.GET("hello", func(c *gin.Context) {
// 		db := data.MustGetDB("custom")
// 		fmt.Printf("%v\n", db)
// 		c.String(200, "Hello world!")
// 	})
// }

// //如果是加密请求需要进行解密,再加一个defer函数来解决加密改写问题
// func decryptImpl(c *gin.Context) bool {
// 	//	判断header是否有Digest
// 	log.Default().Info("into decryptImpl")
// 	return true
// }
// func encryptImpl(c *gin.Context, d interface{}) string {
// 	//	判断header是否有Digest
// 	log.Default().Info("into encryptImpl")
// 	return "x"
// }
// func signCheckImpl(c *gin.Context) bool {
// 	log.Default().Info("into signCheckImpl", zap.Any("c", c.Request.Header.Get("sign")))
// 	return true
// }

// func registerRouter() {
// 	ilikeCORSHandler := gateway.NewCORSHandler([]string{"www.ilikee.vn", "ilikee.vn", "0.0.0.0:8000"})
// 	// 请求加解密
// 	cryptoHandler := gateway.NewCryptoHandler(encryptImpl, decryptImpl)
// 	gateway.WithCryptoHandler(cryptoHandler)
// 	//签名校验
// 	signCheck := gateway.NewSignCheckHandler(signCheckImpl)
// 	gateway.WithSignCheckHandler(signCheck)

// 	gateway.RegisterAPI("demo", "echo", "Demo for echo", demo.EchoHandler{})
// 	gateway.RegisterAPI("demo", "cors_echo", "Demo for cors",
// 		demo.EchoHandler{},
// 		gateway.WithCORSHandler(ilikeCORSHandler),
// 	)
// 	gateway.RegisterAPI("demo", "redis", "Demo for reids incr", demo.RedisHandler{}, gateway.ExtInfo(map[string]string{"test": "<a href='https://www.null.com'>test</a>"}))
// 	gateway.RegisterAPI("demo", "mysql", "Demo for mysql ", demo.MysqlHandler{})
// 	gateway.RegisterAPI("demo", "postgres", "Demo for postgres ", demo.PostgresHandler{})
// 	gateway.RegisterAPI("demo", "config", "Demo for config center ", demo.ConfigHandler{})
// 	gateway.RegisterAPI("demo", "pre", "Demo for pre handler, normally used in login filter",
// 		demo.PreHandler{},
// 		gateway.HandlerFunc(demoPreHandler),
// 		gateway.HandlerFunc(demoPreHandler2),
// 		gateway.HandlerFunc(demoPreHandler3),
// 		gateway.WithCryptoHandler(cryptoHandler),
// 		gateway.WithSignCheckHandler(signCheck),
// 	)
// 	gateway.RegisterAPI("demo", "sleep", "Demo for sleep", demo.SleepHandler{})
// 	gateway.RegisterAPI("demo", "doc", "Demo complex sturct", demo.DocHandler{})
// 	gateway.RegisterAPI("demo", "reactHtml", "Demo test for react", demo.ReactHandler{})
// }

// func demoPreHandler(c *gin.Context, ctx *gateway.ApiContext) *gateway.Resp {
// 	log.Default().Info("into demoPreHandler")

// 	ctx.Keys["message"] = "This has handled by demoPreHandler"
// 	return nil
// }

// func demoPreHandler2(c *gin.Context, ctx *gateway.ApiContext) *gateway.Resp {
// 	log.Default().Info("into demoPreHandler2")

// 	if _, ok := ctx.Keys["message"]; ok {
// 		message := ctx.Keys["message"].(string)
// 		newMsg := fmt.Sprintf("%s\nalso has handled by demoPreHandler2", message)
// 		ctx.Keys["message"] = newMsg
// 	}

// 	return nil
// }

// func demoPreHandler3(c *gin.Context, ctx *gateway.ApiContext) *gateway.Resp {
// 	log.Default().Info("into demoPreHandler3")

// 	if _, ok := ctx.Keys["message"]; ok {
// 		message := ctx.Keys["message"].(string)
// 		newMsg := fmt.Sprintf("%s\nhandled by demoPreHandler3 too", message)
// 		ctx.Keys["message"] = newMsg
// 	}

// 	return nil
// }
