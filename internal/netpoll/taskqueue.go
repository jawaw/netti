package netpoll

import "sync"

// Task 异步任务.
type Task func() error

// AsyncTaskQueue queues pending tasks.
type AsyncTaskQueue struct {
	lock  sync.Locker
	tasks []func() error
}

// NewAsyncTaskQueue 创建一个任务队列.
func NewAsyncTaskQueue() AsyncTaskQueue {
	return AsyncTaskQueue{lock: new(spinlock)}
}

// Push .
func (q *AsyncTaskQueue) Push(task Task) (tasksNum int) {
	q.lock.Lock()
	q.tasks = append(q.tasks, task)
	tasksNum = len(q.tasks)
	q.lock.Unlock()
	return
}

// ForEach 迭代并执行队列中的任务.
func (q *AsyncTaskQueue) ForEach() (err error) {
	q.lock.Lock()
	tasks := q.tasks
	q.tasks = nil
	q.lock.Unlock()
	for i := range tasks {
		if err = tasks[i](); err != nil {
			return err
		}
	}
	return
}
