package perf

import (
	"time"
)

const ReqIdKey = "goup-req-id"
const ReqLevel = "goup-req-level"

const TraceKafkaTopic = "goup_call_trace"

type Point struct {
	ReqId    string
	Server   string
	Host     string
	ActionAt time.Time
	Level    int
	Name     string
	Action   int // 1: enter, 2: out
}

func InnerCall(reqId string, level int) {}
