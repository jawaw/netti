package netti

import "time"

// EventServer 是EventHandler的一个内置实现, EventHandler 设置的每个方法都有一个默认的实现,
// 当你不想在 EventHandler 中实现所有的方法时, 你可以重写该默认实现。
type EventServer struct {
}

// OnInitComplete 会在服务器配置初始化完成并准备接受新连接的时候触发, 传入的 svr 实例包含服务器
// 相关的信息和多种参数
func (es *EventServer) OnInitComplete(svr Server) (action Action) {
	return
}

// OnOpened 在一个新连接被打开的时候触发, 参数传入为 connection 接口实例, 包含有关连接的本地地址
// 和远端地址等信息, 使用 out 返回值来进行对连接的数据写入操作
func (es *EventServer) OnOpened(c Conn) (out []byte, action Action) {
	return
}

// OnClosed fires when a connection has been closed.
// The err parameter is the last known connection error.
func (es *EventServer) OnClosed(c Conn, err error) (action Action) {
	return
}

// React fires when a connection sends the server data.
// Invoke c.Read() or c.ReadN(n) within the parameter c to read incoming data from client/connection.
// Use the out return value to write data to the client/connection.
func (es *EventServer) React(frame []byte, c Conn) (out []byte, action Action) {
	return
}

// Tick fires immediately after the server starts and will fire again
// following the duration specified by the delay return value.
func (es *EventServer) Tick() (delay time.Duration, action Action) {
	return
}
