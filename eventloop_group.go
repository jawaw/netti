package netti

// LoadBalance 设置负载均衡策略.
type LoadBalance int

const (
	// RoundRobin 连接请求将会以轮训的方式随机分配到每一个 loop 上.
	RoundRobin LoadBalance = iota
	// Random 连接请求将会随机分配.
	Random
	// LeastConnections 下一个连接请求分配给活动数量最少的loop
	LeastConnections
)

// IEventLoopGroup represents a set of event-loops.
type IEventLoopGroup interface {
	register(*eventloop)
	next() *eventloop
	iterate(func(int, *eventloop) bool)
	len() int
}

// eventLoopGroup 事件循环组，仿netty
type eventLoopGroup struct {
	nextLoopIndex int
	eventLoops    []*eventloop
	size          int
}

// register .
func (g *eventLoopGroup) register(el *eventloop) {
	g.eventLoops = append(g.eventLoops, el)
	g.size++
}

// TODO: 支持更多负载均衡方式, 目前默认轮询
func (g *eventLoopGroup) next() (el *eventloop) {
	el = g.eventLoops[g.nextLoopIndex]
	if g.nextLoopIndex++; g.nextLoopIndex >= g.size {
		g.nextLoopIndex = 0
	}
	return
}

// iterate .
func (g *eventLoopGroup) iterate(f func(int, *eventloop) bool) {
	for i, el := range g.eventLoops {
		if !f(i, el) {
			break
		}
	}
}

// len .
func (g *eventLoopGroup) len() int {
	return g.size
}
