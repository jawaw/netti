// +build linux

package netti

import "golang.org/x/sys/unix"

// acceptNewConnection 接收并创建新的连接
func (svr *server) acceptNewConnection(fd int) error {
	nfd, sa, err := unix.Accept(fd)
	if err != nil {
		if err == unix.EAGAIN {
			return nil
		}
		return err
	}
	// 在io复用中把监听套接字设为非阻塞 https://blog.csdn.net/dashoumeixi/article/details/85256220
	if err := unix.SetNonblock(nfd, true); err != nil {
		return err
	}
	el := svr.subLoopGroup.next()
	c := newTCPConn(nfd, el, sa)
	_ = el.poller.Trigger(func() (err error) {
		if err = el.poller.AddRead(nfd); err != nil {
			return
		}
		el.connections[nfd] = c
		err = el.loopOpen(c)
		return
	})
	return nil
}
