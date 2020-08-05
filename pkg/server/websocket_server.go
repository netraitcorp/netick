package server

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/netraitcorp/netick/pkg/log"
)

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
		addr:     opts.Websocket.Addr,
		rt:       opts.Websocket.ReadTimeout,
		wt:       opts.Websocket.WriteTimeout,
		upgrader: upgrader,
	}
	return srv.ListenAndServe()
}
