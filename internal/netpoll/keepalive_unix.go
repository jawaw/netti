// Copyright 2020 PittMo. All rights reserved.
// Copyright 2017 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build linux netbsd freebsd dragonfly

package netpoll

import "golang.org/x/sys/unix"

// SetKeepAlive 设置为文件就绪符的保活状态.
func SetKeepAlive(fd, secs int) error {
	// 设置socket选项, 套接字选项（SOL_SOCKET级别）
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1); err != nil {
		return err
	}
	if err := unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, secs); err != nil {
		return err
	}
	return unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, secs)
}
