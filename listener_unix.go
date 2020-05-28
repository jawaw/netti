// +build linux

package netti

import (
	"net"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

// listener 不同操作系统有不同的监听就绪符,监听器的具体实现应该放到静态编译过程中进行隔离
type listener struct {
	f             *os.File
	fd            int
	ln            net.Listener
	once          sync.Once
	pconn         net.PacketConn
	lnaddr        net.Addr
	addr, network string
}

// close .
func (ln *listener) close() {
	ln.once.Do(
		func() {
			if ln.f != nil {
				sniffError(ln.f.Close())
			}
			if ln.ln != nil {
				sniffError(ln.ln.Close())
			}
			if ln.pconn != nil {
				sniffError(ln.pconn.Close())
			}
			if ln.network == "unix" {
				sniffError(os.RemoveAll(ln.addr))
			}
		})
}

// setNonBlock 将net监听器从父事件循环中分离出来, 获取文件描述符, 设置成无阻塞
func (ln *listener) setNonBlock() error {
	var err error
	switch netln := ln.ln.(type) {
	case nil:
		switch pconn := ln.pconn.(type) {
		case *net.UDPConn:
			ln.f, err = pconn.File()
		}
	case *net.TCPListener:
		ln.f, err = netln.File()
	case *net.UnixListener:
		ln.f, err = netln.File()
	}
	if err != nil {
		ln.close()
		return err
	}
	ln.fd = int(ln.f.Fd())
	return unix.SetNonblock(ln.fd, true)
}
