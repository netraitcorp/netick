package server

import (
	"net"
	"time"
)

const (
	writeWait = 10 * time.Second
)

type Conn interface {
	ConnID() string

	LocalAddr() net.Addr

	RemoteAddr() net.Addr

	Accept()

	Close() error

	Closed() bool

	Write(data []byte) error

	Server() Server
}

type pingt struct {
	lastp time.Time
	timer *time.Timer
}
