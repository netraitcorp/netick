package server

import "time"

type Options struct {
	Websocket       *WebsocketOptions
	PingInterval    time.Duration
	MaxPingOutTimes int
	Auth            *AuthOptions
}

type AuthOptions struct {
	Timeout  time.Duration
	Password string
}

type WebsocketOptions struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewOptions() *Options {
	ws := &WebsocketOptions{
		Addr:         "0.0.0.0:2634",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	auth := &AuthOptions{
		Timeout:  10 * time.Second,
		Password: "123456",
	}
	return &Options{
		Websocket:       ws,
		PingInterval:    30 * time.Second,
		MaxPingOutTimes: 3,
		Auth:            auth,
	}
}
