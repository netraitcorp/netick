package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/netraitcorp/netick/pb"
	"github.com/netraitcorp/netick/pkg/log"
	"github.com/netraitcorp/netick/pkg/types"
	"github.com/netraitcorp/netick/pkg/util"
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
	timer      *time.Timer
}

func (r *ReadHandler) CreateConn() {
	r.timer = time.AfterFunc(r.conn.Server().Options().Auth.Timeout, r.authorizeTimeoutCheck)
}

func (r *ReadHandler) Close() {
	if r.timer != nil {
		r.timer.Stop()
		r.timer = nil
	}

	accounts.RemoveAccount(r.conn.ConnID())
}

func (r *ReadHandler) ReadData(data []byte) (err error) {
	if len(data) == 0 {
		return
	}

	opCode, payload, err := packet.Unmarshal(data)
	if err != nil {
		return err
	}
	switch opCode {
	case types.OpAuth:
		err = r.authorize(payload.(*pb.AuthReq))
	}
	return
}

func (r *ReadHandler) authorize(req *pb.AuthReq) error {
	pass := r.conn.Server().Options().Auth.Password
	if pass != "" {
		if req.GetPassword() == "" {
			return fmt.Errorf("ReadHandler.authorize: password empty, cid: %s", r.conn.ConnID())
		}
		if req.GetPassword() != util.Sha1(pass) {
			return fmt.Errorf("ReadHandler.authorize: password incorrect, cid: %s", r.conn.ConnID())
		}
	}
	r.authorized = true

	accounts.AddAccount(NewAccount(r.conn))

	data, err := packet.Marshal(types.OpAuthRest, &pb.AuthRet{
		ConnId:     r.conn.ConnID(),
		Authorized: true,
	})
	if err != nil {
		return err
	}

	if err := r.conn.Write(data); err != nil {
		return err
	}

	log.Info("ReadHandler.authorize: verified, cid: %s", r.conn.ConnID())
	return nil
}

func (r *ReadHandler) authorizeTimeoutCheck() {
	if r.authorized {
		return
	}
	log.Info("ReadHandler.authorizeTimeoutCheck: cid: %s", r.conn.ConnID())
	_ = r.conn.Close()
}

func NewReadHandler(c Conn) *ReadHandler {
	return &ReadHandler{
		conn: c,
	}
}
