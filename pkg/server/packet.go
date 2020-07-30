package server

import (
	"fmt"

	"github.com/netraitcorp/netick/pb"

	"github.com/netraitcorp/netick/pkg/types"
	"google.golang.org/protobuf/proto"
)

type Packet struct{}

func (*Packet) Marshal(code types.OpCode, payload interface{}) ([]byte, error) {
	pack, ok := payload.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("Packet.Marshal: payload type is not proto.Message")
	}
	data, err := proto.Marshal(pack)
	if err != nil {
		return nil, err
	}

	return append([]byte{byte(code)}, data...), nil
}

func (*Packet) Unmarshal(payload []byte) (types.OpCode, interface{}, error) {
	if len(payload) < 2 {
		return types.OpUnknown, nil, fmt.Errorf("Packet.Unmarshal failed: unknown OpCode")
	}
	opCode := types.OpCode(payload[0])

	var (
		unpack interface{}
		err    error
	)

	switch opCode {
	case types.OpAuth:
		unpack = &pb.AuthReq{}
		err = proto.Unmarshal(payload[1:], unpack.(*pb.AuthReq))
	}
	return opCode, unpack, err
}

var packet = &Packet{}
