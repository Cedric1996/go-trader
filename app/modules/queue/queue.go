/*
 * @Author: cedric.jia
 * @Date: 2021-08-04 15:11:31
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-12 18:07:47
 */

package queue

import (
	"context"
	"fmt"
	"sync"
)

// Data defines an type of queuable data
type Data interface{}

// ExecuteFunc is a function that takes a variable amount of data and processes it
type ExecuteFunc func(interface{}) (interface{}, error)

// HandleFunc is a function that store a variable number of data
type HandleFunc func([]interface{}) error

// Queue defines an interface of a queue-like item
// Queues will handle their own contents in the Run method
type Queue interface {
	Run(atTerminate func(context.Context, func()))
	Push([]Data) error
	Close()
}

type ChannelQueue struct {
	name        string
	workerNum   int
	batchSize   int
	dataChan    chan Data
	finishChan  chan bool
	resChan     chan interface{}
	workerGroup sync.WaitGroup
	executeFunc ExecuteFunc
	handleFunc  HandleFunc
	finishNum   int
}

// NewQueue takes a queue Type, ExecuteFunc, some options and possibly an exemplar and returns a Queue or an error
func NewQueue(name string, workerNum, batchSize int, executeFunc ExecuteFunc, handleFunc HandleFunc) (*ChannelQueue, error) {
	queue := &ChannelQueue{
		name:        name,
		workerNum:   workerNum,
		executeFunc: executeFunc,
		handleFunc:  handleFunc,
		batchSize:   batchSize,
		dataChan:    make(chan Data, workerNum),
		resChan:     make(chan interface{}, workerNum),
		finishChan:  make(chan bool),
	}
	go queue.Run()
	return queue, nil
}

// Push will push data into the queue if the data is not already in the queue
func (q *ChannelQueue) Push(data Data) {
	q.dataChan <- data
}

// Run starts to run the queue
func (q *ChannelQueue) Run() {
	fmt.Printf("ChannelQueue: %s Starting:\n", q.name)
	go q.handle()
	for i := 0; i < q.workerNum; i++ {
		q.workerGroup.Add(1)
		go q.execute()
	}
	q.workerGroup.Wait()
	close(q.resChan)
}

// Run starts to run the queue
func (q *ChannelQueue) Close() {
	close(q.dataChan)
	for {
		finished := <-q.finishChan
		if finished {
			break
		}
	}
	fmt.Printf("ChannelQueue: %s execute %d tasks\n", q.name, q.finishNum)
}

// Execute starts worker to execute task
func (q *ChannelQueue) execute() {
	for data := range q.dataChan {
		res, err := q.executeFunc(data)
		if err != nil {
			fmt.Errorf("ChannelQueue: %s execute with error: %v", q.name, err)
		}
		if res != nil {
			q.resChan <- res
		}
	}
	q.workerGroup.Done()
}

func (q *ChannelQueue) handle() {
	count := 0
	res := make([]interface{}, q.batchSize)
	for datum := range q.resChan {
		res[count] = datum
		count++
		if count == q.batchSize {
			if err := q.handleFunc(res); err != nil {
				return
			}
			q.finishNum += count
			res = make([]interface{}, q.batchSize)
			count = 0
		}
	}
	if count > 0 {
		res = res[:count]
		if len(res) > 0 {
			if err := q.handleFunc(res); err != nil {
				fmt.Printf("error: handle data in channel queue %s\n", q.name)
			}
		}
	}
	fmt.Printf("handle %d data in channel queue %s, then task will be finished..\n", count, q.name)
	q.finishNum += count
	q.finishChan <- true
}
