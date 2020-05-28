// +build linux

package netpoll

import "golang.org/x/sys/unix"

const (
	// InitEvents poller 的事件列表的初始化大小.
	InitEvents = 128
	// ErrEvents 表示非读/写的异常事件，如套接字被关闭，从被关闭的socket上进行读写
	ErrEvents = unix.EPOLLERR | unix.EPOLLHUP | unix.EPOLLRDHUP
	// OutEvents 结合EPOLLOUT事件和一些异常事件
	OutEvents = ErrEvents | unix.EPOLLOUT
	// InEvents 结合EPOLLIN/EPOLLPRI事件和一些异常事件
	InEvents = ErrEvents | unix.EPOLLIN | unix.EPOLLPRI
)

// eventList 事件列表
type eventList struct {
	size   int
	events []unix.EpollEvent
}

// newEventList 创建一个事件列表
func newEventList(size int) *eventList {
	return &eventList{size, make([]unix.EpollEvent, size)}
}

// increase 增加事件初始化大小, 扩容两倍
func (el *eventList) increase() {
	el.size <<= 1
	el.events = make([]unix.EpollEvent, el.size)
}
