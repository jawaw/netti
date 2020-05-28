package netti

import "net"

// Conn 客户端连接的接口
type Conn interface {

	// Context 返回用户定义的上下.
	Context() (ctx interface{})

	// SetContext 设置用户定义的上下文
	SetContext(ctx interface{})

	// LocalAddr 连接的本地套接字地址
	LocalAddr() (addr net.Addr)

	// RemoteAddr 是连接的远程对端地址
	RemoteAddr() (addr net.Addr)

	// 在不移动“read”指针的情况下从入站环形缓冲区和事件循环缓冲区读取所有数据, 它并不实际从环形缓冲区中取出数据，
	// 这些数据将会出现在环形缓冲区中，直到调用ResetBuffer方法。
	Read() (buf []byte)

	// ResetBuffer 重置入站环形缓冲区，这意味着已清除入站环形缓冲区中的所有数据
	ResetBuffer()

	// ReadN 在不移动“read”指针的情况下从入站环形缓冲区和事件循环缓冲区读取给定长度的字节, 这意味着在调用ShiftN方法之前，
	// 这个方法不会从缓冲区中删除数据，对于给定的“n”必须满足与可用数据的长度相等，才会从入站环形缓冲区和事件循环缓冲区读取数据,
	// 否则它将不会从入站环形缓冲区读取任何数据。所以你应该使用这个方法，只有当你知道基于协议的后续TCP流的确切长度,
	// 类似于HTTP请求中的Content-Length属性, 它指示你应该读取多少数据, 从入站环形缓冲区。
	ReadN(n int) (size int, buf []byte)

	// ShiftN 按给定长度移动缓冲区中的“读”指针。
	ShiftN(n int) (size int)

	// BufferLength 返回入站环形缓冲区中可用数据的长度
	BufferLength() (size int)

	// SendTo 为UDP套接字写数据, 它允许你发送数据回UDP套接字在单独的goroutines
	SendTo(buf []byte) error

	// AsyncWrite 异步地将数据写入客户端连接，通常你需要在单个goroutine中调用它而不是事件循环中
	AsyncWrite(buf []byte) error

	// Wake 在当前连接触发一次 React event
	Wake() error

	// Close 关闭当前连接.
	Close() error
}
