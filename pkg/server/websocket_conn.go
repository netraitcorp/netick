package server

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/netraitcorp/netick/pkg/util"

	"github.com/gorilla/websocket"
	"github.com/netraitcorp/netick/pkg/log"
)

type WebsocketConn struct {
	srv       Server
	conn      *websocket.Conn
	ping      pingt
	connID    string
	buf       chan []byte
	handler   Handler
	cancelCtx context.CancelFunc
	closed    bool
	mu        sync.Mutex
}

func NewWebsocketConn(conn *websocket.Conn, srv Server) *WebsocketConn {
	rawConnKey := fmt.Sprintf("tcp:%s <-> tcp:%s", conn.RemoteAddr().String(), conn.LocalAddr().String())

	c := &WebsocketConn{
		srv:    srv,
		conn:   conn,
		connID: util.Sha1(rawConnKey + strconv.Itoa(util.RandInt())),
		buf:    make(chan []byte, 0x40),
		closed: false,
	}
	c.handler = NewReadHandler(c)
	c.handler.CreateConn()

	c.conn.SetPongHandler(c.pongHandler)
	c.startPingTimer()

	log.Info("NewWebsocketConn: %s, cid: %s", rawConnKey, c.ConnID())

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

func (c *WebsocketConn) Closed() bool {
	return c.closed
}

func (c *WebsocketConn) Close() error {
	if c.closed {
		return nil
	}
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()

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

	log.Info("CloseWebsocketConn: cid: %s", c.ConnID())

	return c.conn.Close()
}

func (c *WebsocketConn) Write(data []byte) error {
	if c.closed {
		return fmt.Errorf("WebsocketConn.Write: connection closed")
	}
	select {
	case c.buf <- data:
		return nil
	default:
		return fmt.Errorf("WebsocketConn.Write: write buf full")
	}
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
			if e, ok := err.(*websocket.CloseError); ok {
				log.Info("LoopRead closeFrame: cid: %s, code: %d, text: %s", c.ConnID(), e.Code, e.Text)
			} else {
				log.Error("LoopRead error: cid: %s, err: %s", c.ConnID(), err.Error())
			}
			_ = c.Close()
			return
		}

		if c.handler != nil {
			if err := c.handler.ReadData(data); err != nil {
				log.Error("Handler.ReadData error: cid: %s, err:%s", c.ConnID(), err.Error())
				_ = c.Close()
				return
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
				log.Error("Write error: cid: %s, err: %s", c.ConnID(), err.Error())

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

	if c.closed {
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
		log.Warn("LoopPingTimer: writePingMessage failed, cid: %s, %s", c.ConnID(), err.Error())
		_ = c.Close()
		return
	}

	dd := c.srv.Options().PingInterval
	c.ping.timer = time.AfterFunc(dd, c.loopPingTimer)
}

func (c *WebsocketConn) pongHandler(string) error {
	c.ping.lastp = time.Now()

	log.Debug("PongHandler: %s", c.ping.lastp.String())
	return nil
}
