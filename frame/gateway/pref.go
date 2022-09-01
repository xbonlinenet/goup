package gateway

import (
	"reflect"
	"runtime"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/xbonlinenet/goup/frame/data"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/perf"
	"github.com/xbonlinenet/goup/frame/util"
)

func getReqInfoFromHeader(c *gin.Context) (string, int, error) {
	reqId := c.GetHeader(perf.ReqIdKey)

	if len(reqId) == 0 {
		return "", 0, errors.Errorf("ReqId not exists")
	}

	str := c.GetHeader(perf.ReqLevel)
	if len(str) == 0 {
		return "", 0, errors.Errorf("ReqLevel not exists")
	}

	reqLevel, err := strconv.Atoi(str)
	if err != nil {
		log.Default().Sugar().Warnf("%s isn't valid, %s", perf.ReqLevel, str)
	}

	return reqId, reqLevel, nil
}

func getReqInfo(c *gin.Context) (string, int, error) {
	reqId, ok := c.Keys[perf.ReqIdKey]

	if !ok {
		return "", 0, errors.Errorf("ReqId not exists")
	}

	reqLevel, ok := c.Keys[perf.ReqLevel]
	if !ok {
		return "", 0, errors.Errorf("ReqLevel not exists")
	}

	return reqId.(string), reqLevel.(int), nil

}

func WrapFunc(ctx *ApiContext, foo func()) {
	funcName := runtime.FuncForPC(reflect.ValueOf(foo).Pointer()).Name()

	startPoint := perf.Point{
		ReqId:    ctx.ReqId,
		Server:   util.GetServerName(),
		Host:     util.GetHost(),
		ActionAt: time.Now(),
		Level:    ctx.ReqLevel,
		Name:     funcName,
		Action:   1,
	}

	recoredPoint(&startPoint)

	defer func() {
		endPoint := startPoint
		endPoint.ActionAt = time.Now()
		endPoint.Action = 2
		recoredPoint(&endPoint)
	}()

	foo()
}

func wrapRequest(foo func(c *gin.Context, prefix string), c *gin.Context, apiPathPrefix string) {

	reqId, reqLevel, err := getReqInfoFromHeader(c)
	if err != nil {
		reqId = uuid.NewV4().String()
		reqLevel = 1
	}

	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[perf.ReqIdKey] = reqId
	c.Keys[perf.ReqLevel] = reqLevel

	funcName := c.Request.URL.Path

	startPoint := perf.Point{
		ReqId:    reqId,
		Server:   util.GetServerName(),
		Host:     util.GetHost(),
		ActionAt: time.Now(),
		Level:    reqLevel,
		Name:     funcName,
		Action:   1,
	}

	recoredPoint(&startPoint)

	defer func() {
		endPoint := startPoint
		endPoint.ActionAt = time.Now()
		endPoint.Action = 2
		recoredPoint(&endPoint)
	}()

	foo(c, apiPathPrefix)
}

func recoredPoint(point *perf.Point) {

	// fmt.Printf("%v\n", point)

	if true {
		return
	}

	producer := data.MustGetAsyncProducer()

	data, err := json.Marshal(point)
	util.CheckError(err)
	producer.Input() <- &sarama.ProducerMessage{
		Topic: perf.TraceKafkaTopic,
		Key:   sarama.StringEncoder(point.ReqId),
		Value: sarama.ByteEncoder(data),
	}
}
