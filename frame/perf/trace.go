package perf

import (
	"encoding/json"
	"fmt"
	"time"
)

type PerfNode struct {
	Name string
	Cost int64
}

type Perf struct {
	StartTick int64
	EndTick   int64
	LastName  string
	LastTick  int64
	Nodes     []*PerfNode
	MaxNode   *PerfNode
	Stat      map[string]interface{}
}

func NewPrefTrace(name string) *Perf {
	perfPtr := &Perf{
		Nodes: make([]*PerfNode, 0, 0),
		Stat:  make(map[string]interface{}),
	}

	perfPtr.Start(name)
	return perfPtr

}

func (p *Perf) Start(name string) {
	p.LastName = name
	p.StartTick = int64(time.Now().UnixNano() / 1000 / 1000)
	p.LastTick = p.StartTick

}

func (p *Perf) Finish() {
	p.EndTick = int64(time.Now().UnixNano() / 1000 / 1000)
}

func (p *Perf) AddStat(key string, value interface{}) {
	p.Stat[key] = value
}

func (p *Perf) Trace(name string) {
	tickNow := int64(time.Now().UnixNano() / 1000 / 1000)
	cost := tickNow - p.LastTick
	nameKey := fmt.Sprintf("%s - %s", p.LastName, name)
	node := &PerfNode{Name: nameKey, Cost: cost}
	p.Nodes = append(p.Nodes, node)
	p.LastName = name
	p.LastTick = int64(time.Now().UnixNano() / 1000 / 1000)

	if p.MaxNode == nil {
		p.MaxNode = node
	} else if cost > p.MaxNode.Cost {
		p.MaxNode = node
	}
}

func (p *Perf) Stats() string {
	str, _ := json.Marshal(*p)
	return string(str)
}

func (p *Perf) Cost() int64 {
	return p.EndTick - p.StartTick
}
