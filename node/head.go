package node

import (
	"log/slog"
	"time"

	"github.com/gadfly16/geronimo/msg"
	"gorm.io/gorm"
)

const (
	RootKind = iota
	GroupKind
	NodeAccount
)

var Kinds = map[Kind]Node{
	RootKind: &RootNode{},
}

var commonMsgHandlers map[msg.Kind]func(*Head, *msg.Msg) *msg.Msg

func init() {
	commonMsgHandlers = map[msg.Kind]func(*Head, *msg.Msg) *msg.Msg{
		msg.CreateKind: createHandler,
		msg.StopKind:   stopHandler,
	}
}

type Kind = int

type Head struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name     string
	Kind     Kind
	ParentID int
	In       msg.Pipe `gorm:"-"`

	parent   msg.Pipe
	children map[string]msg.Pipe
}

func (h *Head) Load() (in msg.Pipe, err error) {
	h.In = make(msg.Pipe)
	n, err := Kinds[h.Kind].load(h)
	if err != nil {
		return
	}

	chs := []*Head{}
	if err = Db.Where("parent_id = ?", h.ID).Find(&chs).Error; err != nil {
		return
	}
	for _, ch := range chs {
		ch.parent = h.In
		var chin msg.Pipe
		chin, err = ch.Load()
		if err != nil {
			return
		}
		h.children[ch.Name] = chin
		ch.parent = h.In
	}

	Tree.NodeLock.Lock()
	Tree.Nodes[h.ID] = h.In
	Tree.NodeLock.Unlock()

	go n.run()

	return h.In, err
}

func (h *Head) register() {
	h.children = make(map[string]msg.Pipe)
	h.In = make(msg.Pipe)
	Tree.NodeLock.Lock()
	Tree.Nodes[h.ID] = h.In
	Tree.NodeLock.Unlock()
}

func (h *Head) commonMsg(m *msg.Msg) (r *msg.Msg) {
	f, ok := commonMsgHandlers[m.Kind]
	if !ok {
		return nil
	}
	return f(h, m)
}

func createHandler(h *Head, m *msg.Msg) (r *msg.Msg) {
	var ch msg.Pipe
	var chName string
	var err error
	switch pl := m.Payload.(type) {
	case *GroupNode:
		chName = pl.Head.Name
		ch, err = pl.create()
		if err != nil {
			return msg.NewError(err)
		}
	}
	h.children[chName] = ch
	return msg.OK
}

func stopHandler(h *Head, m *msg.Msg) (r *msg.Msg) {
	slog.Info("Stopping children.", "name", h.Name)
	h.askChildren(&msg.Msg{Kind: msg.StopKind})
	return &msg.Msg{Kind: msg.StoppedKind}
}

func (h *Head) askChildren(m *msg.Msg) {
	m.Resp = make(msg.Pipe)
	for _, ch := range h.children {
		ch <- m
	}
	for range len(h.children) {
		<-m.Resp
	}
}
