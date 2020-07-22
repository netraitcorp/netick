package server

import (
	"context"
	"crypto/sha1"
	"fmt"
	"net"

	"github.com/netraitcorp/netick/pkg/safe"

	"github.com/gorilla/websocket"
)

type ReadHandler func(c Conn, data []byte) error

type Conn interface {
	UniqID() string
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Accept()
	Close() error
	Write(data []byte)
	Bind(ReadHandler)
}

type WebsocketConn struct {
	conn      *websocket.Conn
	uniqID    string
	buf       chan []byte
	cancelCtx context.CancelFunc
	closed    safe.AtomicBool
	handler   ReadHandler
}

func NewWebsocketConn(conn *websocket.Conn) *WebsocketConn {
	h := sha1.New()
	_, _ = h.Write([]byte(fmt.Sprintf("tcp:%s:%s", conn.RemoteAddr(), conn.LocalAddr())))
	connUniqID := fmt.Sprintf("%x", h.Sum(nil))

	c := &WebsocketConn{
		conn:   conn,
		uniqID: connUniqID,
		buf:    make(chan []byte, 0x40),
	}

	return c
}

func (c *WebsocketConn) Bind(h ReadHandler) {
	c.handler = h
}

func (c *WebsocketConn) UniqID() string {
	return c.uniqID
}

func (c *WebsocketConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *WebsocketConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *WebsocketConn) Accept() {
	ctx, cancelCtx := context.WithCancel(context.Background())

	c.cancelCtx = cancelCtx

	readCtx, _ := context.WithCancel(ctx)
	go c.loopRead(readCtx)

	writeCtx, _ := context.WithCancel(ctx)
	go c.loopWrite(writeCtx)
}

func (c *WebsocketConn) Close() error {
	if c.closed.IsSet() {
		return nil
	}
	c.closed.Set()
	if c.cancelCtx != nil {
		c.cancelCtx()
	}
	return c.conn.Close()
}

func (c *WebsocketConn) Write(data []byte) {
	c.buf <- data
}

func (c *WebsocketConn) loopRead(ctx context.Context) {
	for {
		_, data, err := c.conn.ReadMessage()
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err != nil {
			_ = c.Close()
			return
		}

		if c.handler != nil {
			if err := c.handler(c, data); err != nil {
				_ = c.Close()
			}
		}
	}
}

func (c *WebsocketConn) loopWrite(ctx context.Context) {
	for {
		select {
		case data := <-c.buf:
			if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				c.Close()
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
