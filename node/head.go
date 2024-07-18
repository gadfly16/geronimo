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
		msg.CreateKind:  createHandler,
		msg.StopKind:    stopHandler,
		msg.GetTreeKind: getTreeHandler,
	}
}

type Head struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name     string
	Kind     Kind
	OwnerID  int `gorm:"-"`
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
	if h.Kind == UserKind {
		h.OwnerID = h.ID
	}
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
		ch.OwnerID = h.OwnerID
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
	return msg.NewErrorMsg(fmt.Errorf("no appropriate handler found for %s on %s", m.KindName(), h.KindName()))
}

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
	ch, err = n.create(h)
	if err != nil {
		return msg.NewErrorMsg(err)
	}
	h.children[n.name()] = ch
	return &msg.OK
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

func getTreeHandler(h *Head, m *msg.Msg) (r *msg.Msg) {
	if m.Auth.UserID != h.OwnerID && !m.Auth.Admin {
		slog.Debug("unauthorized tree request", "path", h.path, "user", m.Auth.UserID, "owner", h.OwnerID, "admin", m.Auth.Admin)
		return msg.NewErrorMsg(fmt.Errorf("unathorized tree request"))
	}
	tree := &TreeEntry{
		ID:   h.ID,
		Name: h.Name,
		Kind: h.Kind,
	}

	chm := msg.GetTree
	chm.Resp = make(msg.Pipe)
	chm.Auth.UserID = m.Auth.UserID
	chm.Auth.Admin = m.Auth.Admin
	for _, ch := range h.children {
		ch <- &chm
	}

	var cherr bool
	for range len(h.children) {
		chr := <-chm.Resp
		if chr.Kind == msg.ErrorKind {
			cherr = true
		} else {
			slog.Debug("Children gave back tree")
			tree.Children = append(tree.Children, chr.Payload.(*TreeEntry))
		}
	}

	if cherr {
		return msg.NewErrorMsg(fmt.Errorf("unathorized tree request downstream"))
	}

	r = &msg.Msg{
		Kind:    msg.TreeKind,
		Payload: tree}
	return
}
