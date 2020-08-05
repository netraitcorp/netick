package types

type OpCode uint8

const (
	OpUnknown  = 0x00
	OpPing     = 0x01
	OpPong     = 0x02
	OpAuth     = 0x04
	OpAuthRest = 0x05
)
