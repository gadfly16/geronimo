package core

import (
	"encoding/json"
	"io"

	mk "github.com/gadfly16/geronimo/msgKinds"
)

var OKMsg = Msg{Kind: mk.OK}

type E struct{}

type DC chan E

type Pipe chan *Msg

func (p Pipe) MarshalJSON() ([]byte, error) {
	return json.Marshal("Pipe")
}

type Msg struct {
	Kind    mk.MK
	User    *Tag
	Payload any

	resp Pipe
}

func (t *Tag) Ask(mk mk.MK, u *Tag, pl ...any) Msg {
	// For sake of comfort, if there's only one payload, we'll use it directly.
	var epl any = pl
	if len(pl) == 1 {
		epl = pl[0]
	}
	m := &Msg{
		Kind:    mk,
		Payload: epl,
		User:    u,
		resp:    make(Pipe),
	}
	t.In <- m
	return *<-m.resp
}

func (t *Tag) AskMsg(m *Msg) Msg {
	m.resp = make(Pipe)
	t.In <- m
	return *<-m.resp
}

func (t *Tag) Notify(mk mk.MK, u *Tag, pl ...any) {
	var epl any = pl
	if len(pl) == 1 {
		epl = pl[0]
	}
	m := &Msg{
		Kind:    mk,
		Payload: epl,
		User:    u,
	}
	t.In <- m
}

func (t *Tag) NotifyMsg(m *Msg) {
	t.In <- m
}

func (q *Msg) Answer(mk mk.MK, pl ...any) {
	var epl any = pl
	if len(pl) == 1 {
		epl = pl[0]
	}
	a := &Msg{
		Kind:    mk,
		Payload: epl,
		User:    q.User,
	}
	q.resp <- a
}

func (q *Msg) AnswerMsg(a *Msg) {
	q.resp <- a
}

func (q *Msg) AnswerOK() {
	q.resp <- &OKMsg
}

func NewErrorMsg(err error) *Msg {
	return &Msg{
		Kind:    mk.Error,
		Payload: err.Error(),
	}
}

func (m *Msg) ErrorMsg() string {
	return m.Payload.(string)
}

func (m *Msg) KindName() string {
	return mk.Names[m.Kind]
}

func UnmarshalMsg(k mk.MK, b io.ReadCloser) (m *Msg, err error) {
	m = &Msg{}
	switch k {
	case mk.Update:
		m.Payload = H{}
	case mk.CreateChild, mk.RenameChild:
		m.Payload = []any{}
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
	m.Kind = k
	return
}
