package server

import "github.com/netraitcorp/netick/pkg/log"

type Handler interface {
	DealData(data []byte) error
}

type ReadHandler struct {
	conn       Conn
	authorized bool
}

func (h *ReadHandler) DealData(data []byte) (err error) {
	if len(data) == 0 {
		return
	}
	log.Info()
	if !h.authorized {
		return h.authorize()
	}

	return
}

func (h *ReadHandler) authorize() error {
	h.authorized = true

	return nil
}

func NewReadHandler(c Conn) *ReadHandler {
	return &ReadHandler{
		conn: c,
	}
}
