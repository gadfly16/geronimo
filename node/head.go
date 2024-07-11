package node

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gadfly16/geronimo/msg"
	"gorm.io/gorm"
)

var commonMsgHandlers map[msg.Kind]func(*Head, *msg.Msg) *msg.Msg
var nodeMsgHandlers = map[Kind]map[msg.Kind]func(Node, *msg.Msg) *msg.Msg{}

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

	path     string
	parent   msg.Pipe
	children map[string]msg.Pipe
}

func (h *Head) name() string {
	return h.Name
}

func (h *Head) setParentID(pid int) {
	h.ParentID = pid
}

func (h *Head) load() (in msg.Pipe, err error) {
	h.In = make(msg.Pipe)
	h.children = make(map[string]msg.Pipe)
	n, err := Kinds[h.Kind].loadBody(h)
	if err != nil {
		return
	}

	chs := []*Head{}
	if err = Db.Where("parent_id = ?", h.ID).Find(&chs).Error; err != nil {
		return
	}
	for _, ch := range chs {
		ch.parent = h.In
		ch.path = h.path + "/" + ch.Name
		var chin msg.Pipe
		chin, err = ch.load()
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

func (h *Head) handleMsg(n Node, m *msg.Msg) (r *msg.Msg) {
	hf, ok := commonMsgHandlers[m.Kind]
	if ok {
		return hf(h, m)
	}
	nf, ok := nodeMsgHandlers[h.Kind][m.Kind]
	if ok {
		return nf(n, m)
	}
	return msg.NewErrorMsg(fmt.Errorf(""))
}

// func (h *Head) commonMsg(m *msg.Msg) (r *msg.Msg) {
// 	f, ok := commonMsgHandlers[m.Kind]
// 	if !ok {
// 		return nil
// 	}
// 	return f(h, m)
// }

func createHandler(h *Head, m *msg.Msg) (r *msg.Msg) {
	var ch msg.Pipe
	var err error
	n := m.Payload.(Node)
	if _, ok := h.children[n.name()]; ok {
		return msg.NewErrorMsg(fmt.Errorf("node '%s' already exists", n.name()))
	}
	n.setParentID(h.ID)
	switch pl := m.Payload.(type) {
	case *UserNode:
		if len(h.children) == 0 {
			pl.Parms.Admin = true
		}
	}
	ch, err = n.create(h.path)
	if err != nil {
		return msg.NewErrorMsg(err)
	}
	h.children[n.name()] = ch
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
