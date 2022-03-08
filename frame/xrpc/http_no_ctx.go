package xrpc

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/cast"
	"github.com/xbonlinenet/goup/frame/alter"
	"github.com/xbonlinenet/goup/frame/log"
	"github.com/xbonlinenet/goup/frame/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func HttpPostWithJsonResp(apiUrl string, postParam interface{},  resp interface{}, options ...RequestOption) error {
	finalOptions := make([]RequestOption, 0, len(options)+1)
	for _, option := range options {
		finalOptions = append(finalOptions, option)
	}
	finalOptions = append(finalOptions, WithHttpMethod(http.MethodPost))

	return HttpCallWithJsonResp(apiUrl, postParam, resp, finalOptions...)
/*
	start := time.Now()

	reqOpts := RequestOptions{
		Timeout: HttpTimeoutLevel2,
		Headers: make(map[string]string, 4),
	}
	// apply options
	for _, option := range options {
		option.apply(&reqOpts)
	}
	ctx, cancel := context.WithTimeout(context.Background(), reqOpts.Timeout)

	defer func() {
		timeCost := time.Since(start)
		if timeCost.Seconds() > 1.0 {
			log.Default().Warn("HttpPostWithJsonResp rpc is slow", zap.String("apiUrl", apiUrl), zap.Duration("cost", timeCost))
			if reqOpts.SlowAlertAllowed && timeCost.Seconds() > 2.0 {
				alter.Notify("HttpPostWithJsonResp rpc is slow", fmt.Sprintf("apiUrl：%s, cost: %.2f secs", apiUrl, timeCost.Seconds()), "community-common")
			}
		}else{
			log.Default().Debug("HttpPostWithJsonResp rpc cost", zap.String("apiUrl", apiUrl), zap.Duration("cost", timeCost))
		}
		cancel()
	}()

	initHttpClient()

	var reqBody []byte
	var err error

	if reqOpts.MarshalType == MarshalByJsonEncode {
		reqBody, err = Json.Marshal(&postParam)
	}else if reqOpts.MarshalType == MarshalByFormEncode{
		reqBody, err = formEncodeParams(postParam)
	}else if reqOpts.MarshalType == MarshalFromRawBytes {
		if rawBody, ok := postParam.([]byte); !ok {
			log.Default().Error("postParam(http body) is not type of []byte")
			return errors.New("postParam(http body) is not type of []byte")
		}else{
			reqBody = rawBody
		}
	}else{
		log.Default().Error("Unknown marshalType when try to encode http body", zap.Int("marshalType", reqOpts.MarshalType))
		return errors.New("unknown marshalType when try to encode http body")
	}
	if err != nil {
		log.Default().Sugar().Errorf("HttpPostWithJsonResp build http body error: %s", err.Error())
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Default().Error("http.NewRequestWithContext() err", zap.Error(err),
			zap.String("url", apiUrl), zap.String("req", string(reqBody)))
		return err
	}
	// attach option headers
	for k, v := range reqOpts.Headers {
		httpReq.Header.Add(k, v)
	}

	// 发起请求
	respObj, err := httpClient.Do(httpReq)
	if err != nil {
		log.Default().Error("httpClient.Do() err", zap.Error(err),
			zap.String("url", apiUrl), zap.String("req", string(reqBody)))
		return err
	}
	defer respObj.Body.Close()

	body, _ := ioutil.ReadAll(respObj.Body)
	_, err = io.Copy(ioutil.Discard, respObj.Body) // 长连接需要手动丢弃读取完毕的数据
	util.CheckError(err)

	if respObj.StatusCode != http.StatusOK {
		log.Default().Error("NewRequestWithContext response code != 200",  zap.String("url", apiUrl),
			zap.Int("resp_code", respObj.StatusCode),	zap.String("req", string(reqBody)),
			zap.String("resp", string(body)))
		return fmt.Errorf("NewRequestWithContext on '%s' response code is %d", apiUrl, respObj.StatusCode)
	}

	err = Json.Unmarshal(body, resp)
	if err != nil {
		log.Default().Error("Json.Unmarshal() error in HttpPostWithJsonResp func", zap.Error(err), zap.String("url", apiUrl), zap.String("resp", string(body)))
		return err
	}

	return nil
 */
}

func HttpGetWithJsonResp(apiUrl string, getParams map[string]interface{}, resp interface{}, options ...RequestOption) error {
	finalOptions := make([]RequestOption, 0, len(options)+1)
	for _, option := range options {
		finalOptions = append(finalOptions, option)
	}
	finalOptions = append(finalOptions, WithHttpMethod(http.MethodGet))

	return HttpCallWithJsonResp(apiUrl, getParams, resp, finalOptions...)
/*	
	start := time.Now()

	reqOpts := RequestOptions{
		Timeout: HttpTimeoutLevel2,
		Headers: make(map[string]string, 4),
	}
	// apply options
	for _, option := range options {
		option.apply(&reqOpts)
	}
	ctx, cancel := context.WithTimeout(context.Background(), reqOpts.Timeout)

	defer func() {
		timeCost := time.Since(start)
		if timeCost.Seconds() > 1.0 {
			log.Default().Warn("HttpGetWithJsonResp rpc is slow", zap.String("apiUrl", apiUrl), zap.Duration("cost", timeCost))
			if reqOpts.SlowAlertAllowed && timeCost.Seconds() > 2.0 {
				alter.Notify("HttpGetWithJsonResp rpc is slow", fmt.Sprintf("apiUrl：%s, cost: %.2f secs", apiUrl, timeCost.Seconds()), "biz_server")
			}
		}else{
			log.Default().Debug("HttpGetWithJsonResp rpc cost", zap.String("apiUrl", apiUrl), zap.Duration("cost", timeCost))
		}
		cancel()
	}()

	initHttpClient()

	finalUrl := apiUrl
	if len(getParams) > 0 {
		urlValues := url.Values{}
		for name, value := range getParams {
			urlValues.Add(name, cast.ToString(value))
		}
		finalUrl = apiUrl + "?" + urlValues.Encode()
	}

	httpReq, err := http.NewRequestWithContext(ctx,"GET", finalUrl, nil)
	if err != nil {
		log.Default().Error("http.NewRequestWithContext() err", zap.Error(err), zap.String("url", finalUrl))
		return err
	}
	// attach option headers
	for k, v := range reqOpts.Headers {
		httpReq.Header.Add(k, v)
	}

	// 发起请求
	respObj, err := httpClient.Do(httpReq)
	if err != nil {
		log.Default().Error("httpClient.Do() err", zap.Error(err), zap.String("url", finalUrl))
		return err
	}
	defer respObj.Body.Close()

	body, _ := ioutil.ReadAll(respObj.Body)
	_, err = io.Copy(ioutil.Discard, respObj.Body) // 长连接需要手动丢弃读取完毕的数据
	util.CheckError(err)

	err = Json.Unmarshal(body, resp)
	if err != nil {
		log.Default().Error("Json.Unmarshal() error in HttpGetWithJsonResp func", zap.Error(err), zap.String("url", finalUrl), zap.String("resp", string(body)))
		return err
	}

	return nil
 */
}

func formEncodeParams(postParam interface{}) ([]byte, error) {
	if postParam == nil {
		return nil, nil
	}

	mapParams, err := cast.ToStringMapE(postParam)
	if err != nil {
		return nil, err
	}
	values := url.Values{}
	for k, v := range mapParams {
		values.Add(k, cast.ToString(v))
	}

	return []byte(values.Encode()), nil
}


func HttpCallWithJsonResp(apiUrl string, params interface{},  resp interface{}, options ...RequestOption) error {
	start := time.Now()

	reqOpts := RequestOptions{
		Timeout: HttpTimeoutLevel2,
		Headers: make(map[string]string, 4),
	}
	// apply options
	for _, option := range options {
		option.apply(&reqOpts)
	}
	ctx, cancel := context.WithTimeout(context.Background(), reqOpts.Timeout)

	defer func() {
		timeCost := time.Since(start)
		slowCost := float64(reqOpts.Timeout) * 1.0/3
		verySlowCost := slowCost * 2
		if timeCost.Seconds() > slowCost {
			log.Default().Warn("HttpCallWithJsonResp rpc is slow", zap.String("apiUrl", apiUrl), zap.Duration("cost", timeCost))
			if reqOpts.SlowAlertAllowed && timeCost.Seconds() > verySlowCost {
				alter.Notify("HttpCallWithJsonResp rpc is slow", fmt.Sprintf("apiUrl：%s, cost: %.2f secs", apiUrl, timeCost.Seconds()), "community-common")
			}
		}else{
			log.Default().Debug("HttpCallWithJsonResp rpc cost", zap.String("apiUrl", apiUrl), zap.Duration("cost", timeCost))
		}
		cancel()
	}()

	initHttpClient()

	var reqBody []byte
	var err error

	if reqOpts.HttpMethod == http.MethodPost || reqOpts.HttpMethod == http.MethodPut || reqOpts.HttpMethod == http.MethodPatch {
		reqBody, err = encodeHttpBody(params, reqOpts.MarshalType)
	}else if params != nil {
		urlSuffix := buildUrlQueryValueSuffix(params)
		if urlSuffix != "" {
			apiUrl = apiUrl + "?" + urlSuffix
		}
	}

	if err != nil {
		return err
	}

	var bodyBuf io.Reader
	if len(reqBody) > 0 {
		bodyBuf = bytes.NewBuffer(reqBody)
	}
	httpReq, err := http.NewRequestWithContext(ctx,reqOpts.HttpMethod, apiUrl, bodyBuf)
	if err != nil {
		log.Default().Error("http.NewRequestWithContext() err", zap.Error(err),
			zap.String("url", apiUrl), zap.String("req", string(reqBody)))
		return err
	}
	// attach option headers
	for k, v := range reqOpts.Headers {
		httpReq.Header.Add(k, v)
	}

	// 发起请求
	respObj, err := httpClient.Do(httpReq)
	if err != nil {
		log.Default().Error("httpClient.Do() err", zap.Error(err), zap.String("url", apiUrl), zap.String("req", string(reqBody)))
		return err
	}
	defer respObj.Body.Close()

	body, _ := ioutil.ReadAll(respObj.Body)
	_, err = io.Copy(ioutil.Discard, respObj.Body) // 长连接需要手动丢弃读取完毕的数据
	util.CheckError(err)

	if respObj.StatusCode != http.StatusOK {
		log.Default().Error("NewRequestWithContext response code != 200",  zap.String("url", apiUrl),
			zap.Int("resp_code", respObj.StatusCode),	zap.String("req", string(reqBody)),
			zap.String("resp", string(body)))
		return fmt.Errorf("NewRequestWithContext on '%s' response code is %d", apiUrl, respObj.StatusCode)
	}

	err = Json.Unmarshal(body, resp)
	if err != nil {
		log.Default().Error("Json.Unmarshal() error in HttpCallWithJsonResp func", zap.Error(err), zap.String("url", apiUrl), zap.String("resp", string(body)))
		return err
	}

	return nil
}

func encodeHttpBody(params interface{}, marshalType int) ([]byte, error) {
	var reqBody []byte
	var err error

	if marshalType == MarshalByJsonEncode {
		reqBody, err = Json.Marshal(&params)
	}else if marshalType == MarshalByFormEncode{
		reqBody, err = formEncodeParams(params)
	}else if marshalType == MarshalFromRawBytes {
		if rawBody, ok := params.([]byte); !ok {
			log.Default().Error("params(http body) is not type of []byte")
			return nil,  errors.New("params(http body) is not type of []byte")
		}else{
			reqBody = rawBody
		}
	}else{
		log.Default().Error("Unknown marshalType when try to encode http body", zap.Int("marshalType", marshalType))
		return nil, errors.New("unknown marshalType when try to encode http body")
	}
	if err != nil {
		log.Default().Sugar().Errorf("encode http body error: %s", err.Error())
		return nil, err
	}

	return reqBody, nil
}

func buildUrlQueryValueSuffix(params interface{}) string {
	if params == nil {
		return ""
	}
	getParams, ok := params.(map[string]interface{})
	if !ok {
		return ""
	}
	if len(getParams) > 0 {
		urlValues := url.Values{}
		for name, value := range getParams {
			urlValues.Add(name, cast.ToString(value))
		}
		return urlValues.Encode()
	}
	return ""
}