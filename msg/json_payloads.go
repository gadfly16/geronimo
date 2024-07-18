package msg

import (
	"encoding/json"
	"io"
)

type JSONPayloadKind int

const (
	EmptyJSONPayload JSONPayloadKind = iota
	UserNodeJSONPayload
)

type JSONPayloader interface {
	UnmarshalMsg(io.ReadCloser) (*Msg, error)
}

type EmptyPayload struct{}

func (pl *EmptyPayload) UnmarshalMsg(b io.ReadCloser) (m *Msg, err error) {
	m = &Msg{
		Payload: nil,
	}
	d := json.NewDecoder(b)
	err = d.Decode(m)
	return
}
