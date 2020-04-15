package demo

import (
	"github.com/xbonlinenet/goup/frame/gateway"
)

type Pointer struct {
	Hello string `json:"hello" desc:"for test pointer"`
}

type MapValue struct {
	Hello string `json:"hello" desc:"for test map value"`
}

type MapValuePointer struct {
	Hello string `json:"hello" desc:"for test map value"`
}

type Array struct {
	Hello string `json:"hello" desc:"for test map value"`
}

type ArrayPointer struct {
	Hello string `json:"hello" desc:"for test map value"`
}

type DocRequest struct {
	Message      string                      `json:"message" desc:"message need Doc" binding:"required" `
	Pointer      *Pointer                    `json:"pointer" desc:"pointer"`
	Map          map[string]MapValue         `json:"map" desc:"map"`
	MapPointer   map[string]*MapValuePointer `json:"mapPointer" desc:"mapPointer"`
	Array        []Array                     `json:"array" desc:"array"`
	ArrayPointer []*ArrayPointer             `json:"arrayPointer" desc:"array"`
}

type DocResponse struct {
	Code    int    `json:"code" desc:"0: success other: fail"`
	Message string `json:"message" desc:"message need Doc" binding:"required" `
}

type DocHandler struct {
	Request  DocRequest
	Response DocResponse
}

func (h DocHandler) Mock() interface{} {
	return DocResponse{Code: 0, Message: "Mock Response"}

}
func (h DocHandler) Handler(c *gateway.ApiContext) (interface{}, error) {
	return &DocResponse{Code: 0, Message: h.Request.Message}, nil
}
