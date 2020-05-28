package netti

import (
	"log"
	"net"
	"netti/internal/netpoll"
	"os"
	"runtime"
	"strings"
)

// Serve starts handling events for the specified addresses.
//
// Addresses should use a scheme prefix and be formatted
// like `tcp://192.168.0.10:9851` or `unix://socket`.
// Valid network schemes:
//  tcp   - bind to both IPv4 and IPv6
//  tcp4  - IPv4
//  tcp6  - IPv6
//  udp   - bind to both IPv4 and IPv6
//  udp4  - IPv4
//  udp6  - IPv6
//  unix  - Unix Domain Socket
//
// The "tcp" network scheme is assumed when one is not specified.
func Serve(eventHandler EventHandler, addr string, opts ...Option) error {
	var ln listener
	defer func() {
		ln.close()
		if ln.network == "unix" {
			sniffError(os.RemoveAll(ln.addr))
		}
	}()

	options := loadOptions(opts...)

	ln.network, ln.addr = parseAddr(addr)
	if ln.network == "unix" {
		sniffError(os.RemoveAll(ln.addr))
		if runtime.GOOS == "windows" {
			return ErrProtocolNotSupported
		}
	}
	var err error
	if ln.network == "udp" {
		if options.ReusePort && runtime.GOOS != "windows" {
			ln.pconn, err = netpoll.ReusePortListenPacket(ln.network, ln.addr)
		} else {
			ln.pconn, err = net.ListenPacket(ln.network, ln.addr)
		}
	} else {
		if options.ReusePort && runtime.GOOS != "windows" {
			ln.ln, err = netpoll.ReusePortListen(ln.network, ln.addr)
		} else {
			ln.ln, err = net.Listen(ln.network, ln.addr)
		}
	}
	if err != nil {
		return err
	}
	if ln.pconn != nil {
		ln.lnaddr = ln.pconn.LocalAddr()
	} else {
		ln.lnaddr = ln.ln.Addr()
	}
	if err := ln.setNonBlock(); err != nil {
		return err
	}
	return serve(eventHandler, &ln, options)
}

func parseAddr(addr string) (network, address string) {
	network = "tcp"
	address = addr
	if strings.Contains(address, "://") {
		parts := strings.Split(address, "://")
		network = parts[0]
		address = parts[1]
	}
	return
}

func sniffError(err error) {
	if err != nil {
		log.Println(err)
	}
}
