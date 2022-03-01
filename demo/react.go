package demo

import (
	"github.com/xbonlinenet/goup/frame/gateway"
)

// ReactRequest ...
type ReactRequest struct {
	RespType int `json:"resp_type"`
}

type ReactResponse struct {
	Code    int    `json:"code" desc:"0: success other: fail"`
	Message string `json:"message" desc:"message get from api context" `
}

type ReactHandler struct {
	Request  ReactRequest
	Response ReactResponse
}

func (h ReactHandler) Mock() interface{} {
	return ReactResponse{Code: 0, Message: "Hello world"}

}

// Handler ....
func (h ReactHandler) Handler(c *gateway.ApiContext) (interface{}, error) {

	switch h.Request.RespType {
	case int(gateway.TextHtmlType):
		return h.GetTextHTML(), nil
	case int(gateway.OctetStreamType):
		return h.GetOctetStream(), nil
	case int(gateway.JsonStreamType):
		return h.GetJsonStream(), nil
	default:
		return h.GetDefault(), nil
	}
}

func (h ReactHandler) GetDefault() interface{} {
	str := struct {
		Key string `json:"key"`
		Tag string `json:"tag"`
	}{Key: "localhost", Tag: "advert"}
	return str
}

// GetTextHTML ...
func (h ReactHandler) GetTextHTML() []byte {
	data := `<!DOCTYPE html>
           <html lang="en">
           <head>
               <meta charset="UTF-8">
               <meta http-equiv="X-UA-Compatible" content="IE=edge">
               <meta name="viewport" content="width=device-width, initial-scale=1.0">
               <title>Document</title>
           </head>
           <body>
               Hello world
           </body>
           </html>`
	return []byte(data)
}

// GetOctetStream ...
func (h ReactHandler) GetOctetStream() []byte {

	return []byte("hello world")

}

// GetJsonStream ...
func (h ReactHandler) GetJsonStream() []byte {

	data := `
       {
         "html":"Success<br>",
         "address":"127.0.0.1"
           }
        `
	return []byte(data)

}
