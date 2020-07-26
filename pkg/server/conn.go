package server

import (
	"context"
	"crypto/sha1"
	"fmt"
	"net"
	"time"

	"github.com/netraitcorp/netick/pkg/log"

	"github.com/gorilla/websocket"
	"github.com/netraitcorp/netick/pkg/safe"
)

type Conn interface {
	ConnID() string
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Accept()
	Close() error
	Write(data []byte)
	Server() Server
}

type pingt struct {
	lastp        time.Time
	timer        *time.Timer
	pingOutTimes int
}

type WebsocketConn struct {
	srv       Server
	conn      *websocket.Conn
	ping      pingt
	connID    string
	buf       chan []byte
	handler   Handler
	cancelCtx context.CancelFunc
	closed    safe.AtomicBool
}

func NewWebsocketConn(conn *websocket.Conn, srv Server) *WebsocketConn {
	rawConnKey := fmt.Sprintf("tcp:%s <-> tcp:%s", conn.RemoteAddr(), conn.LocalAddr())
	h := sha1.New()
	_, _ = h.Write([]byte(rawConnKey))
	connUniqID := fmt.Sprintf("%x", h.Sum(nil))
	c := &WebsocketConn{
		srv:    srv,
		conn:   conn,
		connID: connUniqID,
		buf:    make(chan []byte, 0x40),
	}
	c.handler = NewReadHandler(c)
	c.conn.SetPongHandler(c.pongHandler)

	if c.handler != nil {
		c.handler.CreateConn()
	}

	log.Info("NewWebsocketConn: %s, connID: %s", rawConnKey, c.ConnID())

	return c
}

func (c *WebsocketConn) Server() Server {
	return c.srv
}

func (c *WebsocketConn) ConnID() string {
	return c.connID
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

	if c.handler != nil {
		c.handler.Close()
	}

	log.Info("CloseWebsocketConn: connID: %s", c.ConnID())

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
			log.Error("LoopRead error: connID: %s, err: %s", c.ConnID(), err.Error())

			_ = c.Close()
			return
		}

		if c.handler != nil {
			if err := c.handler.ReadData(data); err != nil {
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
				log.Error("LoopWrite error: connID: %s, err: %s", c.ConnID(), err.Error())

				_ = c.Close()
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *WebsocketConn) startPingTimer() {
	d := c.srv.Options().PingInterval
	if d > 0 {
		c.ping.timer = time.AfterFunc(d, c.loopPingTimer)
	}
}

func (c *WebsocketConn) loopPingTimer() {
	if c.closed.IsSet() {
		return
	}
	if c.ping.pingOutTimes+1 > c.srv.Options().MaxPingOutTimes {

	}

}

func (c *WebsocketConn) pongHandler(appData string) error {
	return nil
}
