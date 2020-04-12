package xrpc

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/xbonlinenet/goup/frame/log"
	"golang.org/x/net/context"

	jsoniter "github.com/json-iterator/go"
)

// requestLatency 接口延迟
var requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "xrpc_http_request_latency",
	Help: "stat request latency by seconds",
}, []string{"host", "path"})

func init() {
	prometheus.MustRegister(requestLatency)

}

var initHttClientOnce sync.Once
var httpClient *http.Client

func initHttpClient() {
	initHttClientOnce.Do(func() {

		maxConnsPreHost := viper.GetInt("xrpc.http.max-conns-pre-host")
		if maxConnsPreHost == 0 {
			maxConnsPreHost = 100
		}
		maxIdleConns := viper.GetInt("xrpc.http.max-idle-conns")

		httpClient = &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 20,
				MaxConnsPerHost:     maxConnsPreHost,
				MaxIdleConns:        maxIdleConns,
			},
		}
	})

}

var Json = jsoniter.ConfigCompatibleWithStandardLibrary

func HttpPostWithJson(url string, data interface{}, timeout time.Duration) ([]byte, error) {
	initHttpClient()

	start := time.Now()

	reqBytes, err := Json.Marshal(&data)
	if err != nil {
		log.Default().Sugar().Errorf("Requst data Marshal json error: %s", err.Error())
		return []byte{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBytes))
	if err != nil {
		log.Default().Sugar().Errorf("New Request Error: %s", err.Error())

		return []byte{}, err
	}

	host := req.URL.Hostname()
	path := req.URL.Path

	defer requestLatency.WithLabelValues(host, path).Observe(time.Since(start).Seconds())

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Default().Sugar().Warnf("Do Request %s, Error: %s", url, err.Error())
		return []byte{}, err
	}
	defer resp.Body.Close()

	log.Default().Sugar().Infof("Request %s, Status Code: %d", url, resp.StatusCode)

	return ioutil.ReadAll(resp.Body)
}
