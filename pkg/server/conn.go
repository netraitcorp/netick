package server

import (
	"bytes"
	"context"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/netraitcorp/netick/pkg/safe"
)

type conn struct {
	srv    *Server
	id     uint64
	mu     sync.Mutex
	rw     net.Conn
	closed safe.AtomicBool
	rb     *bytes.Buffer
	wc     chan []byte
}

var connUniqueIncr uint64

func newConnection(srv *Server, rw net.Conn) *conn {
	c := &conn{
		srv: srv,
		id:  atomic.AddUint64(&connUniqueIncr, 1),
		rw:  rw,
		rb:  connBufferPool.Get().(*bytes.Buffer),
		wc:  make(chan []byte, 16),
	}
	return c
}

func (c *conn) LocalAddr() net.Addr {
	return c.rw.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	return c.rw.RemoteAddr()
}

func (c *conn) ID() uint64 {
	return c.id
}

func (c *conn) Close() error {
	if c.closed.IsSet() {
		return nil
	}
	c.closed.Set()

	return c.rw.Close()
}

func (c *conn) Write(data []byte) {
	c.wc <- data
}

func (c *conn) accept() {
	ctx, cancelCtx := context.WithCancel(context.Background())

	go c.loopRead(cancelCtx)

	go c.loopWrite(ctx)
}

func (c *conn) loopRead(cancelCtx context.CancelFunc) {
	defer func() {
		cancelCtx()
		c.rb.Reset()
		connBufferPool.Put(c.rb)
	}()

	buf := make([]byte, 0x800)
	for {
		n, err := c.rw.Read(buf)
		if c.closed.IsSet() {
			return
		}
		if err != nil {
			_ = c.Close()

			return
		}

		c.rb.Write(buf[:n])

		c.rb.Reset()
	}
}

func (c *conn) loopWrite(ctx context.Context) {
	for {
		select {
		case data := <-c.wc:
			n, err := c.rw.Write(data)
			log.Printf("[DEBUG] write to client, %v\n, %d, %s", data, n, err.Error())

		case <-ctx.Done():
			return
		}
	}
}

var connBufferPool = NewBufferPoll()

func NewBufferPoll() (pool sync.Pool) {
	pool.New = func() interface{} {
		return &bytes.Buffer{}
	}
	return pool
}
