package server

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/netraitcorp/netick/pkg/log"
	"github.com/netraitcorp/netick/pkg/util"
)

const (
	HeadPackSizeLen = 4
)

type TCPConn struct {
	srv       Server
	conn      net.Conn
	ping      pingt
	connID    string
	wb        chan []byte
	rb        []byte
	rblen     uint32
	handler   Handler
	cancelCtx context.CancelFunc
	closed    bool
	mu        sync.Mutex
}

func NewTCPConn(rw net.Conn, srv Server) *TCPConn {
	rawConnKey := fmt.Sprintf("tcp:%s <-> tcp:%s", rw.RemoteAddr().String(), rw.LocalAddr().String())

	c := &TCPConn{
		srv:    srv,
		conn:   rw,
		connID: util.Sha1(rawConnKey + strconv.Itoa(util.RandInt())),
		wb:     make(chan []byte, 0x40),
		closed: false,
	}
	c.handler = NewReadHandler(c)
	c.handler.CreateConn()

	log.Info("NewTCPConn: %s, cid: %s", rawConnKey, c.ConnID())

	return c
}

func (c *TCPConn) Server() Server {
	return c.srv
}

func (c *TCPConn) ConnID() string {
	return c.connID
}

func (c *TCPConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *TCPConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *TCPConn) Accept() {
	ctx, cancelCtx := context.WithCancel(context.Background())
	c.cancelCtx = cancelCtx

	readCtx, _ := context.WithCancel(ctx)
	go c.loopRead(readCtx)

	writeCtx, _ := context.WithCancel(ctx)
	go c.loopWrite(writeCtx)
}

func (c *TCPConn) Closed() bool {
	return c.closed
}

func (c *TCPConn) Close() error {
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

	log.Info("CloseTCPConn: cid: %s", c.ConnID())

	return c.conn.Close()
}

func (c *TCPConn) Write(data []byte) error {
	if c.closed {
		return fmt.Errorf("TCPConn.Write: connection closed")
	}
	select {
	case c.wb <- data:
		return nil
	default:
		return fmt.Errorf("TCPConn.Write: write buf full")
	}
}

func (c *TCPConn) loopRead(ctx context.Context) {
	rb := make([]byte, 0xFFFF)
	for {
		n, err := c.conn.Read(rb)
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err != nil {
			log.Error("LoopRead error: cid: %s, err: %s", c.ConnID(), err.Error())
			_ = c.Close()
			return
		}
		c.rb = append(c.rb, rb[:n]...)
		for {
			b := c.unPacket()
			if len(b) == 0 {
				break
			}
			c.readHandler(b)
		}
	}
}

func (c *TCPConn) readHandler(b []byte) {
	if c.handler != nil {
		if err := c.handler.ReadData(b); err != nil {
			log.Error("Handler.ReadData error: cid: %s, err: %s", c.ConnID(), err.Error())
			_ = c.Close()
			return
		}
	}
}

func (c *TCPConn) unPacket() []byte {
	if c.rblen == 0 {
		if len(c.rb) < HeadPackSizeLen {
			return nil
		}
		c.rblen = binary.BigEndian.Uint32(c.rb[:HeadPackSizeLen])
		c.rb = c.rb[HeadPackSizeLen:]
	}
	if int(c.rblen) > len(c.rb) {
		return nil
	}

	rb := c.rb[:c.rblen]

	c.rb = c.rb[c.rblen:]
	c.rblen = 0

	return rb
}

func (c *TCPConn) loopWrite(ctx context.Context) {
	for {
		select {
		case data := <-c.wb:
			d := time.Duration(len(data)/0x19000)*time.Second + writeWait
			_ = c.conn.SetWriteDeadline(time.Now().Add(d))
			for {
				n, err := c.conn.Write(data)
				if err != nil {
					log.Error("Write error: cid: %s, err: %s", c.ConnID(), err.Error())
					_ = c.Close()
					return
				}
				if n < len(data) {
					data = data[n:]
					continue
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
