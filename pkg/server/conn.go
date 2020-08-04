package server

import (
	"net"
	"time"
)

const (
	writeWait = 10 * time.Second
)

type Conn interface {
	ConnID() int64
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Accept()
	Close() error
	Write(data []byte)
	Server() Server
}

type pingt struct {
	lastp time.Time
	timer *time.Timer
}
