package msg

import (
	"io"
)

const (
	UserNodePayload = iota
)

type PayloadKind int

type Payloader interface {
	UnmarshalMsg(io.ReadCloser) (*Msg, error)
}
