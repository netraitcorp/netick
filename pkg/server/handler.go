package server

import (
	"github.com/netraitcorp/netick/pb"
	"github.com/netraitcorp/netick/pkg/types"
)

type Handler interface {
	CreateConn() error
	Close()
	ReadData(data []byte) error
}

type ReadHandler struct {
	conn       Conn
	authorized bool
}

func (h *ReadHandler) CreateConn() error {
	data, err := packet.Marshal(types.OpConnected, &pb.ConnectedRest{
		ConnId: h.conn.ConnID(),
	})
	if err != nil {
		return err
	}
	h.conn.Write(data)
	return nil
}

func (h *ReadHandler) Close() {
}

func (h *ReadHandler) ReadData(data []byte) (err error) {
	if len(data) == 0 {
		return
	}

	opCode, payload, err := packet.Unmarshal(data)
	if err != nil {
		return err
	}
	switch opCode {
	case types.OpAuth:
		h.authorize(payload.(pb.AuthReq))
	}
	return
}

func (h *ReadHandler) publish() {

}

func (h *ReadHandler) subscribe() {

}

func (h *ReadHandler) authorize(req pb.AuthReq) error {
	h.authorized = true

	return nil
}

func NewReadHandler(c Conn) *ReadHandler {
	return &ReadHandler{
		conn: c,
	}
}
