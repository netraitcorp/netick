package server

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/netraitcorp/netick/pkg/log"
)

type Server interface {
	Options() *Options
}

type WebsocketServer struct {
	opts     *Options
	addr     string
	rt       time.Duration
	wt       time.Duration
	httpSrv  *http.Server
	upgrader *websocket.Upgrader
}

func (srv *WebsocketServer) ListenAndServe() error {
	srv.httpSrv = &http.Server{
		Addr:         srv.addr,
		Handler:      srv,
		ReadTimeout:  srv.rt,
		WriteTimeout: srv.wt,
	}
	if err := srv.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (srv *WebsocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsConn, err := srv.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Client connection failed, err: %v", err.Error())
		return
	}
	conn := NewWebsocketConn(wsConn, srv)

	go conn.Accept()
}

func (srv *WebsocketServer) Options() *Options {
	return srv.opts
}

func RunWebsocketServer(opts *Options) error {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	srv := &WebsocketServer{
		opts:     opts,
		addr:     opts.WebsocketOpts.Addr,
		rt:       opts.WebsocketOpts.ReadTimeout,
		wt:       opts.WebsocketOpts.WriteTimeout,
		upgrader: upgrader,
	}
	return srv.ListenAndServe()
}

/*
func (srv *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	return srv.serve(ln)
}

func (srv *Server) serve(ln net.Listener) error {
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

func (srv *Server) createConn(rw net.Conn) *Conn {
	return NewConnection(rw)
}
*/
