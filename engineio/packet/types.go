package packet

import (
	"github.com/wudgaby/go-socket.io/engineio/frame"
)

type Frame struct {
	FType frame.Type
	Data  []byte
}

type Packet struct {
	FType frame.Type
	PType Type
	Data  []byte
}
