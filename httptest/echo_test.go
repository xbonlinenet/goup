package main

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

const topnewsLocalHost = "http://127.0.0.1:13360"
const topnewsProductHost = "https://api.adfly.vn"

func TestDemoEcho(t *testing.T) {
	e := httpexpect.New(t, topnewsLocalHost)

	path := "/api/demo/echo"

	ret := e.POST(path).WithJSON(map[string]string{"message": "hello world"}).
		Expect().Status(http.StatusOK).JSON().Object()
	ret.Value("code").Equal(0)
	ret.Value("message").Equal("hello world")

}
