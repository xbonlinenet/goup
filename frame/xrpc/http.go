package xrpc

import (
	"bytes"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"golang.org/x/net/context"

	jsoniter "github.com/json-iterator/go"

	"github.com/xbonlinenet/goup/frame/gateway"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/perf"
)

// requestLatency 接口延迟
var requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "xrpc_http_request_latency",
	Help: "stat request latency by seconds",
}, []string{"host", "path"})

func init() {
	prometheus.MustRegister(requestLatency)
}

const (
	DefaultTimeout = 10 * time.Second
)

type RequestOptions struct {
	// 超时时间
	Timeout time.Duration

	// Headers
	Headers map[string]string
}

// options abs
type RequestOption interface {
	apply(options *RequestOptions)
}

type ApplyOption func(*RequestOptions)

func (f ApplyOption) apply(options *RequestOptions) {
	f(options)
}

func WithHeaders(headers map[string]string) RequestOption {
	return ApplyOption(func(options *RequestOptions) {
		for key, val := range headers {
			options.Headers[key] = val
		}
	})
}

func WithTimeout(timeout time.Duration) RequestOption {
	return ApplyOption(func(options *RequestOptions) {
		options.Timeout = timeout
	})
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

func HttpPostWithOptions(
	c *gateway.ApiContext, url string, data interface{}, options ...RequestOption) ([]byte, error) {

	initHttpClient()

	start := time.Now()

	reqOpts := RequestOptions{
		Timeout: DefaultTimeout,
		Headers: make(map[string]string, 4),
	}

	// apply options
	for _, option := range options {
		option.apply(&reqOpts)
	}

	reqBytes, err := Json.Marshal(&data)
	if err != nil {
		log.Default().Sugar().Errorf("Request Data Marshal json error: %s", err.Error())
		return []byte{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), reqOpts.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBytes))
	if err != nil {
		log.Default().Sugar().Errorf("New Request Error: %s", err.Error())

		return []byte{}, err
	}

	if c != nil {
		req.Header.Add(perf.ReqIdKey, c.ReqId)
		req.Header.Add(perf.ReqLevel, strconv.Itoa(c.ReqLevel+1))
	}

	// attach option headers
	for k, v := range reqOpts.Headers {
		req.Header.Add(k, v)
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

	return nil, nil
}

func HttpPostWithJson(
	c *gateway.ApiContext, url string, data interface{}, timeout time.Duration) ([]byte, error) {

	return HttpPostWithOptions(c, url, data, WithTimeout(timeout))
}
