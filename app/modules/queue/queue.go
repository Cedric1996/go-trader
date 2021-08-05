/*
 * @Author: cedric.jia
 * @Date: 2021-08-04 15:11:31
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-04 20:28:27
 */

package queue

import (
	"context"
	"fmt"
	"sync"

	"gitea.com/lunny/log"
)

// Data defines an type of queuable data
type Data interface{}

// HandlerFunc is a function that takes a variable amount of data and processes it
type HandlerFunc func(interface{}) error

// Queue defines an interface of a queue-like item
// Queues will handle their own contents in the Run method
type TaskQueue interface {
	Run(atTerminate func(context.Context, func()))
	Push(Data) error
	Close()
}

type ChannelQueue struct {
	name        string
	workerNum   int
	dataChan    chan Data
	workerGroup sync.WaitGroup
	handleFunc  HandlerFunc
	finishNum   int
}

// NewQueue takes a queue Type, HandlerFunc, some options and possibly an exemplar and returns a Queue or an error
func NewQueue(name string, workerNum int, handleFunc HandlerFunc) (*ChannelQueue, error) {
	queue := &ChannelQueue{
		name:       name,
		workerNum:  workerNum,
		handleFunc: handleFunc,
		dataChan:   make(chan Data, workerNum),
	}
	return queue, nil
}

// Push will push data into the queue if the data is not already in the queue
func (q *ChannelQueue) Push(data Data) {
	q.dataChan <- data
}

// Run starts to run the queue
func (q *ChannelQueue) Run() {
	log.Debug("ChannelUniqueQueue: %s Starting", q.name)
	go func() {
		for i := 0; i < q.workerNum; i++ {
			q.workerGroup.Add(1)
			go q.execute()
		}
		q.workerGroup.Wait()
	}()
}

// Run starts to run the queue
func (q *ChannelQueue) Close() {
	fmt.Printf("ChannelQueue: %s execute %d tasks\n", q.name, q.finishNum)
	close(q.dataChan)
}

// Execute starts worker to execute task
func (q *ChannelQueue) execute() {
	for data := range q.dataChan {
		if err := q.handleFunc(data); err != nil {
			log.Error("ChannelQueue: %s execute with error: %v", q.name, err)
		}
		q.finishNum++
	}
	q.workerGroup.Done()
}
