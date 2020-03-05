package util

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/xbonlinenet/goup/frame/log"
	"github.com/go-errors/errors"
	"github.com/spf13/viper"
)

var (
	// ErrWorkStopped 已停止工作
	ErrWorkStopped = errors.New("work has stopped")

	// ErrMapEmptyResult mapper 空返回
	ErrMapEmptyResult = errors.New("mapper result is empty")

	// ErrUnExpectType 非预期的数据类型
	ErrUnExpectType = errors.New("type not expected")
)

type MapReduceWorkerOption struct {
	Name         string
	MapperCount  int
	ReducerCount int

	JobWaitLenght    int
	ReduceBatchSize  int
	MaxWaitForReduce time.Duration
}

type MapReduceWorker struct {
	Option     *MapReduceWorkerOption
	Mapper     func(interface{}) (interface{}, error)
	ChanMapper func(interface{}, chan interface{})
	Reducer    func([]interface{})

	waitChan            chan interface{}
	reduceChan          chan interface{}
	workingMapperCount  int32
	workingReducerCount int32
	stopped             bool

	mapperCount int64
	reduceCount int64
}

// BuildOption 从 viper 中构建选项, 如：
//   recommend:
// 		mapper-count: 10
// 		reducer-count: 1
// 		job-wait-length: 10000
// 		reduce-batch-size: 1000
// 		max-wait-reduce-duration: 30s
func BuildOption(name string, conf *viper.Viper) MapReduceWorkerOption {
	return MapReduceWorkerOption{
		Name:             name,
		MapperCount:      conf.GetInt("mapper-count"),
		ReducerCount:     conf.GetInt("reducer-count"),
		JobWaitLenght:    conf.GetInt("job-wait-length"),
		ReduceBatchSize:  conf.GetInt("reduce-batch-size"),
		MaxWaitForReduce: conf.GetDuration("max-wait-reduce-duration"),
	}
}

// NewMapReduceWorker 创建批处理工作
func NewMapReduceWorker(
	opt *MapReduceWorkerOption,
	mapper func(interface{}) (interface{}, error),
	reducer func([]interface{})) *MapReduceWorker {

	waitChan := make(chan interface{}, opt.JobWaitLenght)
	reduceChan := make(chan interface{}, opt.ReduceBatchSize*opt.ReducerCount/2)
	worker := &MapReduceWorker{
		Option:     opt,
		Mapper:     mapper,
		Reducer:    reducer,
		waitChan:   waitChan,
		reduceChan: reduceChan,
	}
	return worker
}

// NewChanMapReduceWorker 创建批处理工作
func NewChanMapReduceWorker(
	opt *MapReduceWorkerOption,
	mapper func(interface{}, chan interface{}),
	reducer func([]interface{})) *MapReduceWorker {

	waitChan := make(chan interface{}, opt.JobWaitLenght)
	reduceChan := make(chan interface{}, opt.ReduceBatchSize*opt.ReducerCount/2)
	worker := &MapReduceWorker{
		Option:     opt,
		ChanMapper: mapper,
		Reducer:    reducer,
		waitChan:   waitChan,
		reduceChan: reduceChan,
	}
	return worker
}

// Start 开始工作
func (worker *MapReduceWorker) Start() {

	for i := 0; i < worker.Option.MapperCount; i++ {
		go worker.processMapJobs()
	}

	for i := 0; i < worker.Option.ReducerCount; i++ {
		go worker.processMapResults()
	}
}

// AddJob 添加需要处理的工作
func (worker *MapReduceWorker) AddJob(job interface{}) error {
	if worker.stopped {
		return ErrWorkStopped
	}

	worker.waitChan <- job
	return nil
}

// Stop 停止工作
func (worker *MapReduceWorker) Stop() {
	if !worker.stopped {
		close(worker.waitChan)
	}
}

// WaitFinish 等待 Work 工作完成
func (worker *MapReduceWorker) WaitFinish() {
	closed := false
loop:
	for {
		select {
		case <-time.After(time.Second):
			if worker.workingMapperCount == 0 && worker.workingReducerCount > 0 && !closed {
				close(worker.reduceChan)
				closed = true
			}
			if worker.workingReducerCount == 0 {
				break loop
			}
		}
	}
}

func (worker *MapReduceWorker) processMapJobs() {
	atomic.AddInt32(&worker.workingMapperCount, 1)
	defer atomic.AddInt32(&worker.workingMapperCount, -1)

loop:
	for {
		select {
		case job, ok := <-worker.waitChan:
			if ok {
				if worker.ChanMapper != nil {
					worker.ChanMapper(job, worker.reduceChan)
				} else {
					ret, err := worker.Mapper(job)
					atomic.AddInt64(&worker.mapperCount, 1)
					if err == ErrMapEmptyResult {
						continue
					}
					if err != nil {
						log.Default().Info(fmt.Sprintf("%s occur error when map: %s", worker.Option.Name, err.Error()))
					} else {
						worker.reduceChan <- ret
					}
				}
			} else {
				break loop
			}
		}
	}
}

func (worker *MapReduceWorker) processMapResults() {
	atomic.AddInt32(&worker.workingReducerCount, 1)
	defer atomic.AddInt32(&worker.workingReducerCount, -1)
	items := make([]interface{}, 0, worker.Option.ReduceBatchSize)

loop:
	for {
		select {
		case item, ok := <-worker.reduceChan:
			if !ok {
				break loop
			} else {
				items = append(items, item)
				if len(items) >= worker.Option.ReduceBatchSize {
					worker.Reducer(items)
					atomic.AddInt64(&worker.reduceCount, int64(len(items)))
					items = make([]interface{}, 0, worker.Option.ReduceBatchSize)
				}
			}
		case <-time.After(worker.Option.MaxWaitForReduce):
			worker.Reducer(items)
			atomic.AddInt64(&worker.reduceCount, int64(len(items)))
			items = make([]interface{}, 0, worker.Option.ReduceBatchSize)

			log.Default().Info(fmt.Sprintf("Worker Stat: mapped count %d, reduced count: %d",
				worker.mapperCount, worker.reduceCount))
		}
	}
	if len(items) > 0 {
		worker.Reducer(items)
		atomic.AddInt64(&worker.reduceCount, int64(len(items)))
	}
}
