// +build linux

package netti

import "netti/internal/netpoll"

// activateMainReactor .
func (svr *server) activateMainReactor() {
	defer svr.signalShutdown()

	svr.logger.Printf("main reactor exits with error:%v\n", svr.mainLoop.poller.Polling(func(fd int, ev uint32) error {
		return svr.acceptNewConnection(fd)
	}))
}

// activateSubReactor .
func (svr *server) activateSubReactor(el *eventloop) {
	defer func() {
		if el.idx == 0 && svr.opts.Ticker {
			close(svr.ticktock)
		}
		svr.signalShutdown()
	}()

	if el.idx == 0 && svr.opts.Ticker {
		go el.loopTicker()
	}

	//svr.logger.Printf("", el.poller.Polling(el.handleEvent))
	svr.logger.Printf("event-loop:%d exits with error:%v\n", el.idx, el.poller.Polling(func(fd int, ev uint32) error {
		if c, ack := el.connections[fd]; ack {
			switch c.outBuffer.IsEmpty() {
			// Don't change the ordering of processing EPOLLOUT | EPOLLRDHUP / EPOLLIN unless you're 100%
			// sure what you're doing!
			// Re-ordering can easily introduce bugs and bad side-effects, as I found out painfully in the past.
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
		return nil
	}))
}
