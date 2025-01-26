package tree

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
)

func (p Pipe) MarshalJSON() ([]byte, error) {
	return json.Marshal("Pipe")
}

type Msg struct {
	Kind    MK
	User    *Tag
	Payload []any

	resp Pipe
}

func (t *Tag) Ask(u *Tag, mk MK, pl ...any) Msg {
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

func (t *Tag) Notify(u *Tag, mk MK, pl ...any) {
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

func (q *Msg) Answer(mk MK, pl ...any) {
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
	q.resp <- oka
}

func NewErrorMsg(err error) *Msg {
	return &Msg{
		Kind:    M_Error,
		Payload: err,
	}
}

func (m *Msg) KindName() string {
	return MKNames[m.Kind]
}

func UnmarshalMsg(k MK, b io.ReadCloser) (m *Msg, err error) {
	m = &Msg{}
	switch k {
	case M_Update_Parms:
		m.Payload = H{}
	case M_Rename, M_Create:
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
	if k == M_Create {
		// Now this is ugly...
		pl := m.Payload.([]any)
		plc := []any{MK(pl[0].(float64)), pl[1]}
		m.Payload = plc
		slog.Debug("JSON message unparshaled.", "m", m, "pl0type", fmt.Sprintf("%T", m.Payload.([]any)[0]))
	}
	return
}
