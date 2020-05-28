// +build linux darwin netbsd freebsd openbsd dragonfly

package netti

import (
	"github.com/panjf2000/gnet/ringbuffer"
	"net"
	"netti/internal/netpoll"

	"github.com/panjf2000/gnet/pool/bytebuffer"
	prb "github.com/panjf2000/gnet/pool/ringbuffer"
	"golang.org/x/sys/unix"
)

type conn struct {
	fd         int                    // 文件描述符
	sa         unix.Sockaddr          // 远程套接字地址
	ctx        interface{}            // 用户定义的上下文
	loop       *eventloop             // 连接所处的事件循环
	buffer     []byte                 // 接收数据的临时缓冲区内存重用
	codec      ICodec                 // TCP编解码器
	opened     bool                   // 连接被打开事件会触发
	localAddr  net.Addr               // 本地地址
	remoteAddr net.Addr               // 远程地址
	byteBuffer *bytebuffer.ByteBuffer // bytes buffer for buffering current packet and data in ring-buffer
	inBuffer   *ringbuffer.RingBuffer // 来自 client 数据的缓冲区
	outBuffer  *ringbuffer.RingBuffer // 准备写入client的数据的缓冲区
}

// newTCPConn .
func newTCPConn(fd int, el *eventloop, sa unix.Sockaddr) *conn {
	return &conn{
		fd:        fd,
		sa:        sa,
		loop:      el,
		codec:     el.codec,
		inBuffer:  prb.Get(),
		outBuffer: prb.Get(),
	}
}

// releaseTCP .
func (c *conn) releaseTCP() {
	c.opened = false
	c.sa = nil
	c.ctx = nil
	c.buffer = nil
	c.localAddr = nil
	c.remoteAddr = nil
	prb.Put(c.inBuffer)
	prb.Put(c.outBuffer)
	c.inBuffer = nil
	c.outBuffer = nil
	bytebuffer.Put(c.byteBuffer)
	c.byteBuffer = nil
}

// newUDPConn .
func newUDPConn(fd int, el *eventloop, sa unix.Sockaddr) *conn {
	return &conn{
		fd:         fd,
		sa:         sa,
		localAddr:  el.svr.ln.lnaddr,
		remoteAddr: netpoll.SockaddrToUDPAddr(sa),
	}
}

// releaseUDP .
func (c *conn) releaseUDP() {
	c.ctx = nil
	c.localAddr = nil
	c.remoteAddr = nil
}

// open .
func (c *conn) open(buf []byte) {
	n, err := unix.Write(c.fd, buf)
	if err != nil {
		_, _ = c.outBuffer.Write(buf)
		return
	}

	if n < len(buf) {
		_, _ = c.outBuffer.Write(buf[n:])
	}
}

// read .
func (c *conn) read() ([]byte, error) {
	return c.codec.Decode(c)
}

// write .
func (c *conn) write(buf []byte) {
	if !c.outBuffer.IsEmpty() {
		_, _ = c.outBuffer.Write(buf)
		return
	}
	n, err := unix.Write(c.fd, buf)
	if err != nil {
		if err == unix.EAGAIN {
			_, _ = c.outBuffer.Write(buf)
			_ = c.loop.poller.ModReadWrite(c.fd)
			return
		}
		_ = c.loop.loopCloseConn(c, err)
		return
	}
	if n < len(buf) {
		_, _ = c.outBuffer.Write(buf[n:])
		_ = c.loop.poller.ModReadWrite(c.fd)
	}
}

// sendTo .
func (c *conn) sendTo(buf []byte) error {
	return unix.Sendto(c.fd, buf, 0, c.sa)
}

// Read .
func (c *conn) Read() []byte {
	if c.inBuffer.IsEmpty() {
		return c.buffer
	}
	c.byteBuffer = c.inBuffer.WithByteBuffer(c.buffer)
	return c.byteBuffer.Bytes()
}

// ResetBuffer .
func (c *conn) ResetBuffer() {
	c.buffer = nil
	c.inBuffer.Reset()
	bytebuffer.Put(c.byteBuffer)
	c.byteBuffer = nil
}

// ReadN .
func (c *conn) ReadN(n int) (size int, buf []byte) {
	inBufferLen := c.inBuffer.Length()
	tempBufferLen := len(c.buffer)
	if inBufferLen+tempBufferLen < n || n <= 0 {
		return
	}
	size = n
	if c.inBuffer.IsEmpty() {
		buf = c.buffer[:n]
		return
	}
	head, tail := c.inBuffer.LazyRead(n)
	c.byteBuffer = bytebuffer.Get()
	_, _ = c.byteBuffer.Write(head)
	_, _ = c.byteBuffer.Write(tail)
	if inBufferLen >= n {
		buf = c.byteBuffer.Bytes()
		return
	}

	restSize := n - inBufferLen
	_, _ = c.byteBuffer.Write(c.buffer[:restSize])
	buf = c.byteBuffer.Bytes()
	return
}

// ShiftN .
func (c *conn) ShiftN(n int) (size int) {
	inBufferLen := c.inBuffer.Length()
	tempBufferLen := len(c.buffer)
	if inBufferLen+tempBufferLen < n || n <= 0 {
		c.ResetBuffer()
		size = inBufferLen + tempBufferLen
		return
	}
	size = n
	if c.inBuffer.IsEmpty() {
		c.buffer = c.buffer[n:]
		return
	}
	c.byteBuffer.B = c.byteBuffer.B[n:]
	if c.byteBuffer.Len() == 0 {
		bytebuffer.Put(c.byteBuffer)
		c.byteBuffer = nil
	}
	if inBufferLen >= n {
		c.inBuffer.Shift(n)
		return
	}
	c.inBuffer.Reset()

	restSize := n - inBufferLen
	c.buffer = c.buffer[restSize:]
	return
}

func (c *conn) BufferLength() int {
	return c.inBuffer.Length() + len(c.buffer)
}

func (c *conn) AsyncWrite(buf []byte) (err error) {
	var encodedBuf []byte
	if encodedBuf, err = c.codec.Encode(c, buf); err == nil {
		return c.loop.poller.Trigger(func() error {
			if c.opened {
				c.write(encodedBuf)
			}
			return nil
		})
	}
	return
}

func (c *conn) SendTo(buf []byte) error {
	return c.sendTo(buf)
}

func (c *conn) Wake() error {
	return c.loop.poller.Trigger(func() error {
		return c.loop.loopWake(c)
	})
}

func (c *conn) Close() error {
	return c.loop.poller.Trigger(func() error {
		return c.loop.loopCloseConn(c, nil)
	})
}

func (c *conn) Context() interface{}       { return c.ctx }
func (c *conn) SetContext(ctx interface{}) { c.ctx = ctx }
func (c *conn) LocalAddr() net.Addr        { return c.localAddr }
func (c *conn) RemoteAddr() net.Addr       { return c.remoteAddr }
