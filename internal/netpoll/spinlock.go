package netpoll

import (
	"runtime"
	"sync/atomic"
)

// spinlock 模仿linux 自旋锁, 实现无锁化编程.
type spinlock struct{ lock uintptr }

// Lock .
func (l *spinlock) Lock() {
	for !atomic.CompareAndSwapUintptr(&l.lock, 0, 1) {
		runtime.Gosched()
	}
}

// Unlock .
func (l *spinlock) Unlock() {
	atomic.StoreUintptr(&l.lock, 0)
}
