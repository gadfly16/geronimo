package node

import (
	"encoding/json"
	"io"
	"time"

	"github.com/gadfly16/geronimo/msg"
)

type Node interface {
	create(*Head) (msg.Pipe, error)
	loadBody(*Head) (Node, error)
	run()
	getName() string
	setName(string)
	setParentID(int)
}

type Group struct {
	Head
}

type Account struct {
	Head
	Exchange string
}

type ParmModel struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	HeadID    int
}

type display map[string]interface{}

// This is not right. It's here bacuse node can not be imported to msg.
// It must be investigated how can we handle payloads better, because now we don't
// exactly determine the type of payloads based on message kind.
func UnmarshalMsg(mk msg.Kind, b io.ReadCloser) (m *msg.Msg, err error) {
	m = &msg.Msg{}
	switch mk {
	case msg.UpdateKind:
		m.Payload = map[string]interface{}{}
	case msg.CreateKind:
		m.Payload = &Head{}
	default:
		m.Payload = nil
	}
	if m.Payload != nil {
		d := json.NewDecoder(b)
		err = d.Decode(&m.Payload)
	}
	if err != nil {
		return nil, err
	}
	m.Kind = mk
	if mk == msg.CreateKind {
		switch m.Payload.(*Head).Kind {
		case GroupKind:
			m.Payload = &GroupNode{
				Head: m.Payload.(*Head),
			}
		case UserKind:
			m.Payload = &UserNode{
				Head: m.Payload.(*Head),
			}
		}
	}
	return
}
