package demo

import (
	"github.com/xbonlinenet/goup/frame/gateway"
)

type EchoRequest struct {
	Message string `json:"message" desc:"message need echo" binding:"required" `
}

type EchoResponse struct {
	Code    int    `json:"code" desc:"0: success other: fail"`
	Message string `json:"message" desc:"message need echo" binding:"required" `
}

type EchoHandler struct {
	Request  EchoRequest
	Response EchoResponse
}

func (h EchoHandler) Mock() interface{} {
	return EchoResponse{Code: 0, Message: "Mock Response"}

}
func (h EchoHandler) Handler(c *gateway.ApiContext) (interface{}, error) {
	return &EchoResponse{Code: 0, Message: h.Request.Message}, nil
}
