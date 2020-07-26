package server

import "github.com/netraitcorp/netick/pkg/log"

type Handler interface {
	CreateConn()
	Close()
	ReadData(data []byte) error
}

type ReadHandler struct {
	conn       Conn
	authorized bool
}

func (h *ReadHandler) CreateConn() {

}

func (h *ReadHandler) Close() {

}

func (h *ReadHandler) ReadData(data []byte) (err error) {
	if len(data) == 0 {
		return
	}

	log.Debug("ReadData: connID: %s, data: %s", h.conn.ConnID(), data)

	if !h.authorized {
		return h.authorize()
	}

	return
}

func (h *ReadHandler) publish() {

}

func (h *ReadHandler) subscribe() {

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
