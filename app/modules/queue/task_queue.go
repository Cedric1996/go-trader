/*
 * @Author: cedric.jia
 * @Date: 2021-08-12 18:08:17
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-12 23:59:46
 */

package queue

import (
	"fmt"
	"sync"
	"time"
)

type TaskQueue struct {
	name       string
	workerNum  int
	taskSync   sync.WaitGroup
	handleFunc func(data interface{}) error
	pushFunc   func(*chan interface{})
	dataChan   chan interface{}
}

func NewTaskQueue(name string, workerNum int, handleFunc func(data interface{}) error, pushFunc func(*chan interface{})) *TaskQueue {
	return &TaskQueue{
		name:       name,
		workerNum:  workerNum,
		taskSync:   sync.WaitGroup{},
		handleFunc: handleFunc,
		pushFunc:   pushFunc,
		dataChan:   make(chan interface{}),
	}
}
func (q *TaskQueue) Run() error {
	startT := time.Now()
	for i := 0; i < q.workerNum; i++ {
		q.taskSync.Add(1)
		go func() {
			for data := range q.dataChan {
				if err := q.handleFunc(data); err != nil {
					break
				}
			}
			q.taskSync.Done()
		}()
	}
	q.pushFunc(&q.dataChan)
	close(q.dataChan)
	q.taskSync.Wait()
	fmt.Printf("task %s finished successfully, total time: %s", q.name, time.Since(startT).String())
	return nil
}
