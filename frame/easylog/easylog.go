package easylog

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/bitly/go-simplejson"
	"go.uber.org/zap"

	"github.com/xbonlinenet/goup/frame/log"
)

var (
	reporterCount                 int32 = 0
	ErrReporterExhaust                  = errors.New("easy log reporter exhaust")
	ErrEasyLogServerNotOK               = errors.New("easy log server status not ok")
	ErrEasyLogServerInternalError       = errors.New("easy log server internal error")
)

type LogItem struct {
	Uid  string `json:"uid"`
	Data string `json:"data"`
}

type EasyLog struct {
	Product string     `json:"product"`
	Module  string     `json:"module"`
	Event   string     `json:"event"`
	Logs    *[]LogItem `json:"logs"`
}

type Reporter struct {
	reportURL string
	product   string
	module    string
	event     string

	async            bool
	maxReporterCount int32
}

func NewEventReporter(reportURL string, product string, module string, event string) *Reporter {
	return &Reporter{
		reportURL:        reportURL,
		product:          product,
		module:           module,
		event:            event,
		async:            false,
		maxReporterCount: 0,
	}
}

func NewAsyncEventReporter(reportURL string, product string, module string, event string, maxReporterCount int32) *Reporter {
	return &Reporter{
		reportURL:        reportURL,
		product:          product,
		module:           module,
		event:            event,
		async:            true,
		maxReporterCount: maxReporterCount,
	}
}

func (r *Reporter) ReportItem(userId string, data string) error {
	item := LogItem{
		Uid:  userId,
		Data: data,
	}
	logs := &[]LogItem{item}
	return r.ReporterItems(logs)
}

func (r *Reporter) ReporterItems(logs *[]LogItem) error {

	easyLog := &EasyLog{
		Product: r.product,
		Module:  r.module,
		Event:   r.event,
		Logs:    logs,
	}
	if !r.async {
		return reportToEasyLogServerSync(r.reportURL, easyLog)
	}

	if r.maxReporterCount == 0 {
		reportToEasyLogServerAsync(r.reportURL, easyLog)
		return nil
	}

	atomic.AddInt32(&reporterCount, 1)
	defer atomic.AddInt32(&reporterCount, -1)

	if reporterCount >= r.maxReporterCount {
		return ErrReporterExhaust
	}

	reportToEasyLogServerAsync(r.reportURL, easyLog)
	return nil
}

var httpClientContent = &http.Client{
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 200,
		TLSHandshakeTimeout: 0 * time.Second,
	},
	Timeout: time.Millisecond * 3000,
}

func reportToEasyLogServerSync(reportURL string, easylog *EasyLog) error {

	payload, err := json.Marshal(easylog)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", reportURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil
	}

	resp, err := httpClientContent.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Default().Error("easy_log resp err status:", zap.Any("status", resp.Status))
		return ErrEasyLogServerNotOK
	}

	body, _ := ioutil.ReadAll(resp.Body)
	_, err = io.Copy(ioutil.Discard, resp.Body)

	j, err := simplejson.NewJson(body)
	if err != nil {
		return err
	}
	ret, err := j.Get("ret").Int()
	if err != nil {
		return err
	}

	if ret != 0 {
		return ErrEasyLogServerInternalError
	}
	return nil
}

func reportToEasyLogServerAsync(reportURL string, easylog *EasyLog) {
	go func() {
		if err := reportToEasyLogServerSync(reportURL, easylog); err != nil {
			log.Default().Error("easylogErr", zap.Error(err))
		}
	}()
}
