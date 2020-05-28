// +build linux

package netpoll

import (
	"golang.org/x/sys/unix"
	"log"
)

// Poller poller 负责监控文件描述符.
type Poller struct {
	fd     int    // epoll fd
	wfd    int    // wake fd
	wfdBuf []byte // wfd buffer to read packet
	notes  AsyncTaskQueue
}

// NewPoller instantiates a poller.
func NewPoller() (*Poller, error) {
	poller := new(Poller)
	epollFD, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}
	poller.fd = epollFD
	r0, _, errno := unix.Syscall(unix.SYS_EVENTFD2, unix.O_CLOEXEC, unix.O_NONBLOCK, 0)
	if errno != 0 {
		_ = unix.Close(epollFD)
		return nil, errno
	}
	poller.wfd = int(r0)
	poller.wfdBuf = make([]byte, 8)
	if err = poller.AddRead(poller.wfd); err != nil {
		return nil, err
	}
	poller.notes = NewAsyncTaskQueue()
	return poller, nil
}

// Close closes the poller.
func (p *Poller) Close() error {
	if err := unix.Close(p.wfd); err != nil {
		return err
	}
	return unix.Close(p.fd)
}

// Trigger 唤醒阻塞在等待网络事件中的poller, 并执行 notes 队列中的任务
func (p *Poller) Trigger(task Task) error {
	if p.notes.Push(task) == 1 {
		// 写入8字节后唤醒 epoll
		_, err := unix.Write(p.wfd, []byte{0, 0, 0, 0, 0, 0, 0, 1})
		return err
	}
	return nil
}

// Polling blocks the current goroutine, waiting for network-events.
func (p *Poller) Polling(callback func(fd int, ev uint32) error) (err error) {
	el := newEventList(InitEvents)
	var wakenUp bool
	for {
		n, err0 := unix.EpollWait(p.fd, el.events, -1)
		if err0 != nil && err0 != unix.EINTR {
			log.Println(err0)
			continue
		}
		for i := 0; i < n; i++ {
			if fd := int(el.events[i].Fd); fd != p.wfd {
				if err = callback(fd, el.events[i].Events); err != nil {
					return
				}
			} else {
				wakenUp = true
				_, _ = unix.Read(p.wfd, p.wfdBuf)
			}
		}
		if wakenUp {
			wakenUp = false
			if err = p.notes.ForEach(); err != nil {
				return
			}
		}
		if n == el.size {
			el.increase()
		}
	}
}

//const (
//	readEvents      = unix.EPOLLPRI | unix.EPOLLIN
//	writeEvents     = unix.EPOLLOUT
//	readWriteEvents = readEvents | writeEvents
//)

// AddReadWrite ...
func (p *Poller) AddReadWrite(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN | unix.EPOLLOUT,
		},
	)
}

// AddRead ...
func (p *Poller) AddRead(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN,
		},
	)
}

// AddWrite ...
func (p *Poller) AddWrite(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLOUT,
		},
	)
}

// ModRead ...
func (p *Poller) ModRead(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_MOD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN,
		},
	)
}

// ModReadWrite ...
func (p *Poller) ModReadWrite(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_MOD, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN | unix.EPOLLOUT,
		},
	)
}

// Delete ...
func (p *Poller) Delete(fd int) error {
	return unix.EpollCtl(p.fd, unix.EPOLL_CTL_DEL, fd,
		&unix.EpollEvent{Fd: int32(fd),
			Events: unix.EPOLLIN | unix.EPOLLOUT,
		},
	)
}
