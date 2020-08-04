package server

import (
	"errors"
	"sync"
	"time"

	"github.com/netraitcorp/netick/pb"
	"github.com/netraitcorp/netick/pkg/log"
	"github.com/netraitcorp/netick/pkg/types"
	"github.com/netraitcorp/netick/pkg/util"
)

var (
	ErrAuthPasswordEmpty     = errors.New("authorize: password empty")
	ErrAuthPasswordIncorrect = errors.New("authorize: password incorrect")
)

type Handler interface {
	CreateConn()
	Close()
	ReadData(data []byte) error
}

type ReadHandler struct {
	conn       Conn
	uid        string
	authorized bool
	mu         sync.Mutex
}

func (h *ReadHandler) CreateConn() {
	time.AfterFunc(h.conn.Server().Options().Auth.Timeout, h.authorizeTimeoutCheck)
	/*
		data, err := packet.Marshal(types.OpConnected, &pb.ConnectedRest{
			ConnId: h.conn.ConnID(),
		})
		if err != nil {
			return err
		}
		h.conn.Write(data)
		return nil

	*/
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
	pass := h.conn.Server().Options().Auth.Password
	if pass != "" {
		if req.Password == "" {
			return ErrAuthPasswordEmpty
		}
		if req.Password != util.Sha1(pass) {
			return ErrAuthPasswordIncorrect
		}
	}

	log.Info("ReadHandler.authorize: verified, req: %s", req)

	h.authorized = true

	return nil
}

func (h *ReadHandler) authorizeTimeoutCheck() {
	if h.authorized {
		return
	}
	log.Info("ReadHandler.authorizeTimeoutCheck: connID: %d", h.conn.ConnID())
	_ = h.conn.Close()
}

func NewReadHandler(c Conn) *ReadHandler {
	return &ReadHandler{
		conn: c,
	}
}
