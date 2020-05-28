package netti

import (
	"netti/pkg/log"
	"time"
)

// Action 在事件完成后发生的操作
type Action int

const (
	// None 表示在事件之后不应该发生任何操作.
	None Action = iota

	// Close 关闭连接
	Close

	// Shutdown 关闭服务器
	Shutdown
)

var defaultLogger = log.NewLogger()

// EventHandler 表示服务调用的服务器事件的回调,每个事件都有一个用于管理连接和服务器的状态的动作返回值。
type EventHandler interface {

	// OnInitComplete 当服务器准备接受连接时触发, 服务器参数包含服务器参数信息
	OnInitComplete(server Server) (action Action)

	// OnOpened 在连接被打开时触发
	OnOpened(c Conn) (out []byte, action Action)

	// OnClosed 在连接被关闭时触发
	OnClosed(c Conn, err error) (action Action)

	// React fires when a connection sends the server data.
	// Invoke c.Read() or c.ReadN(n) within the parameter c to read incoming data from client/connection.
	// Use the out return value to write data to the client/connection.y
	React(frame []byte, c Conn) (out []byte, action Action)

	// Tick 服务器启动后每隔delay时间后触发
	Tick() (delay time.Duration, action Action)
}
