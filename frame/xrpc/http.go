package xrpc

import (
	"bytes"
	"io/ioutil"
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
	HttpTimeoutLevel2 = 5 * time.Second

	MarshalByJsonEncode = 0
	MarshalByFormEncode = 1
	MarshalFromRawBytes = 2
)

// RequestOptions 请求设置
type RequestOptions struct {
	// 超时时间
	Timeout time.Duration

	// Headers
	Headers map[string]string

	//ReqBodyFormEncoded	bool
	Verbose bool
	SlowAlertAllowed bool
	MarshalType int
	HttpMethod string
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

func WithSlowAlert(alertAllowed bool) RequestOption {
	return ApplyOption(func(options *RequestOptions){
		options.SlowAlertAllowed = alertAllowed
	})
}

func WithTimeout(timeout time.Duration) RequestOption {
	return ApplyOption(func(options *RequestOptions) {
		options.Timeout = timeout
	})
}

func WithVerbose(verbose bool) RequestOption {
	return ApplyOption(func(options *RequestOptions) {
		options.Verbose = verbose
	})
}

func WithFormEncoded(formEncoded bool) RequestOption {
	return ApplyOption(func(options *RequestOptions){
		// options.ReqBodyFormEncoded = formEncoded
		options.MarshalType = MarshalByFormEncode
	})
}

func WithRawBody() RequestOption {
	return ApplyOption(func(options *RequestOptions){
		options.MarshalType = MarshalFromRawBytes
	})
}

func WithHttpMethod(method string) RequestOption {
	return ApplyOption(func(options *RequestOptions){
		if isValidHttpMethod(method) {
			options.HttpMethod = method
		}else{
			options.HttpMethod = http.MethodPost // default value
		}
	})
}

func isValidHttpMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead ||
		method == http.MethodPost || method == http.MethodPut ||
		method == http.MethodPatch || method == http.MethodDelete ||
		method == http.MethodOptions || method == http.MethodTrace
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

// HttpPostWithOptions
//     headers := map[string]string{"X-Key": "val"}
//     respBytes, err := HttpPostWithOptions(ctx, url, data, WithTimeout(10*time.Second), WithHeaders(headers))

func HttpPostRawWithOptions(
	c *gateway.ApiContext, url string, reqBytes []byte, options ...RequestOption) ([]byte, error) {
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

	ctx, cancel := context.WithTimeout(context.Background(), reqOpts.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBytes))
	if err != nil {
		log.Default().Sugar().Errorf("New Request Error: %s", err.Error())

		return []byte{}, err
	}

	req.Header.Add("Content-type", "application/json")
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

	if reqOpts.Verbose {
		log.Default().Sugar().Infof("Request %s, Status Code: %d", url, resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

func HttpPostWithOptions(
	c *gateway.ApiContext, url string, data interface{}, options ...RequestOption) ([]byte, error) {
	reqBytes, err := Json.Marshal(&data)
	if err != nil {
		log.Default().Sugar().Errorf("Request Data Marshal json error: %s", err.Error())
		return []byte{}, err
	}

	return HttpPostRawWithOptions(c, url, reqBytes, options...)
}

func HttpPostWithJson(
	c *gateway.ApiContext, url string, data interface{}, timeout time.Duration) ([]byte, error) {

	return HttpPostWithOptions(c, url, data, WithTimeout(timeout))
}

func HttpGetWithOptions(c *gateway.ApiContext, url string,
	options ...RequestOption) ([]byte, error) {
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

	ctx, cancel := context.WithTimeout(context.Background(), reqOpts.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	if reqOpts.Verbose {
		log.Default().Sugar().Infof("Request %s, Status Code: %d", url, resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
