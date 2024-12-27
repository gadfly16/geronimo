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
	RenameChildMsgKind: renameChildHandler,
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

func (h *Head) handleMsg(n Node, q *Msg) (a *Msg) {
	slog.Debug("NODE message received.", "node", n.getPath(), "kind", q.KindName())
	nhf, ok := nodeMsgHandlers[h.Kind][q.Kind]
	if ok {
		a = nhf(n, q)
		if a != nil {
			q.Answer(a)
			slog.Debug("NODE message answered.", "node", n.getPath(), "reqKind", q.KindName(), "ansKind", a.KindName())
		} else {
			slog.Debug("NODE notifycation handled.", "node", n.getPath(), "reqKind", q.KindName())
		}
		return
	}
	chf, ok := commonMsgHandlers[q.Kind]
	if ok {
		a = chf(h, q)
		if a != nil {
			q.Answer(a)
			slog.Debug("NODE common message answered.", "node", n.getPath(), "reqKind", q.KindName(), "ansKind", a.KindName())
		} else {
			slog.Debug("NODE common notifycation handled.", "node", n.getPath(), "reqKind", q.KindName())
		}
		return
	}
	slog.Error("No appropriate handler found.", "path", h.path, "msg_kind", q.KindName())
	return NewErrorMsg(fmt.Errorf("no appropriate handler found for %s on %s", q.KindName(), h.KindName()))
}

func createHandler(h *Head, m *Msg) (r *Msg) {
	cpl := m.Payload.(*CreatePayload)
	if cpl.Kind == RootKind || cpl.Kind == UserKind {
		return NewErrorMsg(fmt.Errorf("%s kind can not be created", kindNames[cpl.Kind]))
	}
	nn := cpl.Name
	if nn == "" {
		nn = ("New" + kindNames[cpl.Kind])
	}
	if _, ok := h.children[nn]; ok {
		return NewErrorMsg(fmt.Errorf("node '%s' already exists", nn))
	}
	// n := Kinds[cpl.Kind]
	n := NewNodeKind(cpl.Kind)
	if n == nil {
		return NewErrorMsg(fmt.Errorf("node kind '%s' not implemented yet", kindNames[cpl.Kind]))
	}
	n.setKind(cpl.Kind)
	n.setName(nn)
	n.setParentID(h.ID)
	n.setPath(h.path + "/" + nn)
	n.setOwnerID(h.OwnerID)

	nin, err := n.create()
	if err != nil {
		return NewErrorMsg(err)
	}
	h.children[n.getName()] = nin
	return &Msg{Kind: OKMsgKind, Payload: nin}
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
		return NewErrorMsg(errors.New("attempt to unscrubscribe non-subscribed GUI"))
	}
	delete(h.subs, guiid)
	slog.Debug("GUI unsubscribed", "node", h.path, "gui", guiid)
	return &OKMsg
}

func renameChildHandler(h *Head, m *Msg) (r *Msg) {
	rchpl := m.Payload.(*renameChildPayload)
	ch, ok := h.children[rchpl.Name]
	if !ok {
		return NewErrorMsg(fmt.Errorf("node has no children named '%s'", rchpl.Name))
	}
	if _, ok := h.children[rchpl.NewName]; ok {
		return NewErrorMsg(fmt.Errorf("node already has a children named '%s'", rchpl.NewName))
	}

	a := ch.Ask(Msg{
		Kind:    RenameMsgKind,
		Payload: rchpl.NewName,
	})
	if a.Kind == ErrorMsgKind {
		return &a
	}

	h.children[rchpl.NewName] = ch
	delete(h.children, rchpl.Name)
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
	Tree.Sys.TreeUpdater.Notify(Msg{
		Kind:    TreeNodeRenameMsgKind,
		Payload: *h,
	})
	return &OKMsg
}

func (h *Head) display() H {
	d := H{
		"Head": H{
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
