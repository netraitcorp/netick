package types

type OpCode uint8

const (
	OpUnknown   = 0x00
	OpPing      = 0x01
	OpPong      = 0x02
	OpAuth      = 0x04
	OpAuthRet   = 0x05
	OpSubscribe = 0x06
)
