package server

import (
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type TCPServer struct {
	opts *Options
	addr string
}

func (srv *TCPServer) ListenAndServe() error {
	ln, err := net.Listen("tcp", srv.addr)
	if err != nil {
		return err
	}
	return srv.serve(ln)
}

func (srv *TCPServer) serve(ln net.Listener) error {
	var tempDelay time.Duration
	for {
		rw, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				//log.Printf("[ERROR] tcp: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		tempDelay = 0

		c := srv.createConn(rw)
		c.Accept()
	}
}

func (srv *TCPServer) createConn(rw net.Conn) Conn {
	return NewTCPConn(rw, srv)
}

func (srv *TCPServer) Options() *Options {
	return srv.opts
}

func RunTCPServer(opts *Options) error {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	srv := &WebsocketServer{
		opts:     opts,
		addr:     opts.Websocket.Addr,
		rt:       opts.Websocket.ReadTimeout,
		wt:       opts.Websocket.WriteTimeout,
		upgrader: upgrader,
	}
	return srv.ListenAndServe()
}
