package demo

import (
	"time"

	"github.com/xbonlinenet/goup/frame/gateway"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/util"
	"github.com/xbonlinenet/goup/frame/xrpc"
)

type SleepRequest struct {
	Seconds int64 `json:"seconds" desc:"Sleep seconds before response"`
}

type SleepResponse struct {
	Code int `json:"code" desc:"0: success other: fail"`
}

type SleepHandler struct {
	Request  SleepRequest
	Response SleepResponse
}

func (h SleepHandler) Mock() interface{} {
	return SleepResponse{Code: 0}

}

func (h SleepHandler) Handler(c *gateway.ApiContext) (interface{}, error) {

	time.Sleep(time.Duration(h.Request.Seconds) * time.Second)

	data, err := xrpc.HttpPostWithJson("http://localhost:13360/api/demo/echo", map[string]interface{}{"message": "Test"}, time.Second)
	util.CheckError(err)
	log.Default().Sugar().Infof("echo result : %s", string(data))
	return SleepResponse{Code: 0}, nil
}
