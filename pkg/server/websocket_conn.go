package server

import (
	"context"
	"crypto/sha1"
	"fmt"
	"net"
	"time"

	"github.com/gorilla/websocket"
	"github.com/netraitcorp/netick/pkg/log"
	"github.com/netraitcorp/netick/pkg/safe"
)

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
	rawConnKey := fmt.Sprintf("tcp:%s <-> tcp:%s", conn.RemoteAddr().String(), conn.LocalAddr().String())
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
	c.startPingTimer()

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
	if c.handler != nil {
		if err := c.handler.CreateConn(); err != nil {
			log.Error("Handler.CreateConn: %s", err.Error())
			_ = c.Close()
			return
		}
	}

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

	if c.ping.timer != nil {
		c.ping.timer.Stop()
		c.ping.timer = nil
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
			d := time.Duration(len(data)/0x19000)*time.Second + writeWait
			_ = c.conn.SetWriteDeadline(time.Now().Add(d))
			if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				log.Error("Write error: connID: %s, err: %s", c.ConnID(), err.Error())

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
	c.ping.timer = time.AfterFunc(d, c.loopPingTimer)
	c.ping.lastp = time.Now()
}

func (c *WebsocketConn) loopPingTimer() {
	c.ping.timer = nil

	if c.closed.IsSet() {
		return
	}

	d := c.srv.Options().PingInterval * time.Duration(c.srv.Options().MaxPingOutTimes)
	n := time.Now().Add(-d)
	if c.ping.lastp.Before(n) {
		_ = c.Close()
		return
	}

	err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
	if err != nil {
		log.Warn("LoopPingTimer: writePingMessage failed, connID: %s, %s", c.ConnID(), err.Error())
		_ = c.Close()
		return
	}

	dd := c.srv.Options().PingInterval
	c.ping.timer = time.AfterFunc(dd, c.loopPingTimer)
}

func (c *WebsocketConn) pongHandler(appData string) error {
	c.ping.lastp = time.Now()

	log.Debug("PongHandler: %s", c.ping.lastp.String())
	return nil
}
