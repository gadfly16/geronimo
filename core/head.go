package core

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"
)

var nodeMsgHandlers = map[Kind]map[MsgKind]func(Node, *Msg) *Msg{}

var commonMsgHandlers = map[MsgKind]func(*Head, *Msg) *Msg{
	CreateMsgKind:      createHandler,
	StopMsgKind:        stopHandler,
	GetTreeMsgKind:     getTreeHandler,
	SubscribeMsgKind:   subscribeHandler,
	UnsubscribeMsgKind: unsubscribeHandler,
	RenameMsgKind:      renameHandler,
	UpdatePathMsgKind:  updatePathHandler,
}

type Head struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name     string
	Kind     Kind
	ParentID int
	OwnerID  int  `gorm:"-"`
	In       Pipe `gorm:"-"`

	path     string
	children map[string]Pipe

	subs map[int]Pipe
}

type Tag struct {
	ID   int
	Node Pipe
}

func (h *Head) getName() string {
	return h.Name
}

func (h *Head) setName(n string) {
	h.Name = n
}

func (h *Head) setParentID(pid int) {
	h.ParentID = pid
}

func (h *Head) load() (in Pipe, err error) {
	h.In = make(Pipe)
	h.children = make(map[string]Pipe)
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
		ch.path = h.path + "/" + ch.Name
		ch.OwnerID = h.OwnerID
		var chin Pipe
		chin, err = ch.load()
		if err != nil {
			return
		}
		h.children[ch.Name] = chin
	}
	Tree.PutNode(h.ID, h.In)
	go n.run()
	return h.In, err
}

func (h *Head) initNew() {
	h.children = make(map[string]Pipe)
	h.In = make(Pipe)
	Tree.PutNode(h.ID, h.In)
}

func (h *Head) handleMsg(n Node, m *Msg) (r *Msg) {
	chf, ok := commonMsgHandlers[m.Kind]
	if ok {
		return chf(h, m)
	}
	nhf, ok := nodeMsgHandlers[h.Kind][m.Kind]
	if ok {
		return nhf(n, m)
	}
	slog.Error("No appropriate handler found.", "path", h.path, "msg_kind", m.KindName())
	return NewErrorMsg(fmt.Errorf("no appropriate handler found for %s on %s", m.KindName(), h.KindName()))
}

func createHandler(h *Head, m *Msg) (r *Msg) {
	n := m.Payload.(Node)
	if n.getName() == "" {
		n.setName("NewNode")
	}
	if _, ok := h.children[n.getName()]; ok {
		return NewErrorMsg(fmt.Errorf("node '%s' already exists", n.getName()))
	}
	n.setParentID(h.ID)
	switch pl := m.Payload.(type) {
	case *UserNode:
		if len(h.children) == 0 {
			pl.Parms.Admin = true
		}
	}
	nin, err := n.create(h)
	if err != nil {
		return NewErrorMsg(err)
	}
	h.children[n.getName()] = nin
	return &OKMsg
}

func stopHandler(h *Head, m *Msg) (r *Msg) {
	slog.Info("Stopping children.", "name", h.Name)
	h.askChildren(m)
	return &StoppedMsg
}

func (h *Head) askChildren(m *Msg) {
	chm := *m
	chm.Resp = make(Pipe)
	for _, ch := range h.children {
		ch <- &chm
	}
	for range len(h.children) {
		<-chm.Resp
	}
}

func updatePathHandler(h *Head, m *Msg) (r *Msg) {
	h.path = m.Payload.(string) + "/" + h.Name
	h.askChildren(&Msg{
		Kind:    UpdatePathMsgKind,
		Payload: h.path,
	})
	h.updateGUIs()
	return &OKMsg
}

func getTreeHandler(h *Head, m *Msg) (r *Msg) {
	if m.UserID != h.OwnerID && !m.Admin {
		slog.Debug("unauthorized tree request", "path", h.path, "user", m.UserID, "owner", h.OwnerID, "admin", m.Admin)
		return NewErrorMsg(fmt.Errorf("unathorized tree request"))
	}
	tree := &TreeEntry{
		ID:   h.ID,
		Name: h.Name,
		Kind: h.Kind,
	}

	chm := *m
	chm.Resp = make(Pipe)
	for _, ch := range h.children {
		ch <- &chm
	}

	var cherr bool
	for range len(h.children) {
		chr := <-chm.Resp
		if chr.Kind == ErrorMsgKind {
			cherr = true
		} else {
			tree.Children = append(tree.Children, chr.Payload.(*TreeEntry))
		}
	}

	if cherr {
		return NewErrorMsg(fmt.Errorf("unathorized tree request downstream"))
	}

	r = &Msg{
		Kind:    TreeMsgKind,
		Payload: tree}
	return
}

func subscribeHandler(h *Head, m *Msg) (r *Msg) {
	if h.subs == nil {
		h.subs = make(map[int]Pipe)
	}
	gui := m.Payload.(Tag)
	h.subs[gui.ID] = gui.Node
	slog.Debug("GUI subscribed", "node", h.path, "gui", gui.ID)
	return &OKMsg
}

func unsubscribeHandler(h *Head, m *Msg) (r *Msg) {
	guiid := m.Payload.(int)
	_, ok := h.subs[guiid]
	if !ok {
		slog.Error("Can't unscrubsibe GUI that's not subscribed", "node", h.path, "gui", guiid)
		return NewErrorMsg(errors.New("ettempt to unscrubscribe non-subscribed GUI"))
	}
	delete(h.subs, guiid)
	slog.Debug("GUI unsubscribed", "node", h.path, "gui", guiid)
	return &OKMsg
}

func renameHandler(h *Head, m *Msg) (r *Msg) {
	nn, ok := m.Payload.(string)
	if !ok {
		return NewErrorMsg(errors.New("unusable payload for rename"))
	}
	on := h.Name
	dbr := Db.Model(h).Where("id = ?", h.ID).Update("name", nn)
	if dbr.Error != nil {
		return NewErrorMsg(fmt.Errorf("database error during rename: %w", dbr.Error))
	}
	h.path = strings.TrimSuffix(h.path, on) + nn
	h.askChildren(&Msg{
		Kind:    UpdatePathMsgKind,
		Payload: h.path,
	})
	h.updateGUIs()
	Tree.TreeUpdater.Notify(Msg{
		Kind:    TreeNodeRenameMsgKind,
		Payload: *h,
	})
	return &OKMsg
}

func (h *Head) display() display {
	d := display{
		"Head": display{
			"ID":         h.ID,
			"Name":       h.Name,
			"Kind":       h.Kind,
			"Path":       h.path,
			"Modified":   h.CreatedAt,
			"N_children": len(h.children),
		},
	}
	return d
}

func (h *Head) updateGUIs() {
	for id, g := range h.subs {
		slog.Debug("sending updated msg to GUI", "gui_id", id)
		g.Notify(Msg{Kind: NodeUpdateMsgKind, Payload: h.ID})
	}
}
