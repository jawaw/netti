// +build linux

package netti

import (
	"net"
	"netti/internal/netpoll"
	"time"

	"golang.org/x/sys/unix"
)

type eventloop struct {
	idx          int             // 事件循环组中的唯一序号
	svr          *server         // 时间循环中的服务器实例
	codec        ICodec          // TCP数据包编解码器
	packet       []byte          // read packet buffer
	poller       *netpoll.Poller // epoll or iocp
	connections  map[int]*conn   // loop connections fd -> conn
	eventHandler EventHandler    // 事件回调处理接口
}

// loopRun .
func (el *eventloop) loopRun() {
	defer func() {
		if el.idx == 0 && el.svr.opts.Ticker {
			close(el.svr.ticktock)
		}
		el.svr.signalShutdown()
	}()

	if el.idx == 0 && el.svr.opts.Ticker {
		go el.loopTicker()
	}

	el.svr.logger.Printf("event-loop:%d exits with error: %v\n", el.idx, el.poller.Polling(el.handleEvent))
}

// handleEvent .
func (el *eventloop) handleEvent(fd int, ev uint32) error {
	if c, ok := el.connections[fd]; ok {
		switch c.outBuffer.IsEmpty() {
		case false:
			if ev&netpoll.OutEvents != 0 {
				return el.loopWrite(c)
			}
			return nil
		case true:
			if ev&netpoll.InEvents != 0 {
				return el.loopRead(c)
			}
			return nil
		}
	}
	return el.loopAccept(fd)
}

// loopAccept .
func (el *eventloop) loopAccept(fd int) error {
	if fd == el.svr.ln.fd {
		if el.svr.ln.pconn != nil {
			return el.loopReadUDP(fd)
		}
		nfd, sa, err := unix.Accept(fd)
		if err != nil {
			if err == unix.EAGAIN {
				return nil
			}
			return err
		}
		if err = unix.SetNonblock(nfd, true); err != nil {
			return err
		}
		c := newTCPConn(nfd, el, sa)
		if err = el.poller.AddRead(c.fd); err == nil {
			el.connections[c.fd] = c
			return el.loopOpen(c)
		}
		return err
	}
	return nil
}

// loopOpen .
func (el *eventloop) loopOpen(c *conn) error {
	c.opened = true
	c.localAddr = el.svr.ln.lnaddr
	c.remoteAddr = netpoll.SockaddrToTCPOrUnixAddr(c.sa)
	out, action := el.eventHandler.OnOpened(c)
	if el.svr.opts.TCPKeepAlive > 0 {
		if _, ok := el.svr.ln.ln.(*net.TCPListener); ok {
			_ = netpoll.SetKeepAlive(c.fd, int(el.svr.opts.TCPKeepAlive/time.Second))
		}
	}
	if out != nil {
		c.open(out)
	}

	if !c.outBuffer.IsEmpty() {
		_ = el.poller.AddWrite(c.fd)
	}

	return el.handleAction(c, action)
}

// loopRead .
func (el *eventloop) loopRead(c *conn) error {
	n, err := unix.Read(c.fd, el.packet)
	if n == 0 || err != nil {
		if err == unix.EAGAIN {
			return nil
		}
		return el.loopCloseConn(c, err)
	}
	c.buffer = el.packet[:n]

	for inFrame, _ := c.read(); inFrame != nil; inFrame, _ = c.read() {
		out, action := el.eventHandler.React(inFrame, c)
		if out != nil {
			outFrame, _ := el.codec.Encode(c, out)
			c.write(outFrame)
		}
		switch action {
		case None:
		case Close:
			_ = el.loopWrite(c)
			return el.loopCloseConn(c, nil)
		case Shutdown:
			_ = el.loopWrite(c)
			return ErrServerShutdown
		}
		if !c.opened {
			return nil
		}
	}
	_, _ = c.inBuffer.Write(c.buffer)

	return nil
}

// loopWrite .
func (el *eventloop) loopWrite(c *conn) error {
	// todo 使用环形缓冲区避免写拷贝，扩容问题可以使用环形压缩链表解决
	head, tail := c.outBuffer.LazyReadAll()
	n, err := unix.Write(c.fd, head)
	if err != nil {
		if err == unix.EAGAIN {
			return nil
		}
		return el.loopCloseConn(c, err)
	}
	c.outBuffer.Shift(n)

	if len(head) == n && tail != nil {
		n, err = unix.Write(c.fd, tail)
		if err != nil {
			if err == unix.EAGAIN {
				return nil
			}
			return el.loopCloseConn(c, err)
		}
		c.outBuffer.Shift(n)
	}

	if c.outBuffer.IsEmpty() {
		_ = el.poller.ModRead(c.fd)
	}
	return nil
}

// loopCloseConn .
func (el *eventloop) loopCloseConn(c *conn, err error) error {
	// todo 可能导致一处内存泄露
	err0, err1 := el.poller.Delete(c.fd), unix.Close(c.fd)
	if err0 == nil && err1 == nil {
		delete(el.connections, c.fd)
		switch el.eventHandler.OnClosed(c, err) {
		case Shutdown:
			return ErrServerShutdown
		}
		c.releaseTCP()
	} else {
		if err0 != nil {
			el.svr.logger.Printf("failed to delete fd:%d from poller, error:%v\n", c.fd, err0)
		}
		if err1 != nil {
			el.svr.logger.Printf("failed to close fd:%d, error:%v\n", c.fd, err1)
		}
	}
	return nil
}

// loopWake .
func (el *eventloop) loopWake(c *conn) error {
	//if co, ok := el.connections[c.fd]; !ok || co != c {
	//	return nil // ignore stale wakes.
	//}
	out, action := el.eventHandler.React(nil, c)
	if out != nil {
		frame, _ := el.codec.Encode(c, out)
		c.write(frame)
	}
	return el.handleAction(c, action)
}

// loopTicker .
func (el *eventloop) loopTicker() {
	var (
		delay time.Duration
		open  bool
		err   error
	)
	for {
		err = el.poller.Trigger(func() (err error) {
			delay, action := el.eventHandler.Tick()
			el.svr.ticktock <- delay
			switch action {
			case None:
			case Shutdown:
				err = ErrServerShutdown
			}
			return
		})
		if err != nil {
			el.svr.logger.Printf("failed to awake poller with error:%v, stopping ticker\n", err)
			break
		}
		if delay, open = <-el.svr.ticktock; open {
			time.Sleep(delay)
		} else {
			break
		}
	}
}

// handleAction .
func (el *eventloop) handleAction(c *conn, action Action) error {
	switch action {
	case None:
		return nil
	case Close:
		_ = el.loopWrite(c)
		return el.loopCloseConn(c, nil)
	case Shutdown:
		_ = el.loopWrite(c)
		return ErrServerShutdown
	default:
		return nil
	}
}

// loopReadUDP .
func (el *eventloop) loopReadUDP(fd int) error {
	n, sa, err := unix.Recvfrom(fd, el.packet, 0)
	if err != nil || n == 0 {
		if err != nil && err != unix.EAGAIN {
			el.svr.logger.Printf("failed to read UPD packet from fd:%d, error:%v\n", fd, err)
		}
		return nil
	}
	c := newUDPConn(fd, el, sa)
	out, action := el.eventHandler.React(el.packet[:n], c)
	if out != nil {
		_ = c.sendTo(out)
	}
	switch action {
	case Shutdown:
		return ErrServerShutdown
	}
	c.releaseUDP()
	return nil
}
