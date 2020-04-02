package demo

import (
	"math/rand"

	"github.com/xbonlinenet/goup/frame/data"
	"github.com/xbonlinenet/goup/frame/gateway"
	"github.com/xbonlinenet/goup/frame/util"
)

type RedisRequest struct {
	IncrBy int64 `json:"incrBy" desc:"Incr request count by" binding:"required" `
}

type RedisResponse struct {
	Code  int   `json:"code" desc:"0: success other: fail"`
	Count int64 `json:"count" desc:"request count" `
}

type RedisHandler struct {
	Request  RedisRequest
	Response RedisResponse
}

func (h RedisHandler) Mock() interface{} {
	return RedisResponse{Code: 0, Count: rand.Int63()}

}

func (h RedisHandler) Handler(c *gateway.ApiContext) (interface{}, error) {
	client := data.MustGetRedis("gateway")

	cnt, err := client.IncrBy("demo.redis.request", h.Request.IncrBy).Result()
	util.CheckError(err)

	return RedisResponse{Code: 0, Count: cnt}, nil
}
