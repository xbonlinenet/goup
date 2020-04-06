package demo

import (
	"github.com/xbonlinenet/goup/frame/gateway"
)

type PreRequest struct {
}

type PreResponse struct {
	Code    int    `json:"code" desc:"0: success other: fail"`
	Message string `json:"message" desc:"message get from api context" `
}

type PreHandler struct {
	Request  PreRequest
	Response PreResponse
}

func (h PreHandler) Mock() interface{} {
	return PreResponse{Code: 0, Message: "Hello world"}

}

func (h PreHandler) Handler(c *gateway.ApiContext) (interface{}, error) {

	message := "The request hasn't handler by preHandler"
	if _, ok := c.Keys["message"]; ok {
		message = c.Keys["message"].(string)
	}
	return PreResponse{Code: 0, Message: message}, nil
}
