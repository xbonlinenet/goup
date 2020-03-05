package main

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

func TestHandleGet(t *testing.T) {
	e := httpexpect.New(t, "https://www.baidu.com")

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}

	// overwrite client
	println("test")
	e.GET("/").WithClient(client).
		Expect().
		Status(http.StatusOK)

}
