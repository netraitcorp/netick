package server

import "time"

type Options struct {
	WebsocketOpts   *WebsocketOptions
	PingInterval    time.Duration
	MaxPingOutTimes int
}

type WebsocketOptions struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewOptions() *Options {
	wsOpts := &WebsocketOptions{
		Addr:         "0.0.0.0:2634",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	return &Options{
		WebsocketOpts:   wsOpts,
		PingInterval:    30 * time.Second,
		MaxPingOutTimes: 3,
	}
}
