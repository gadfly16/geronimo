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
	GetChildMsgKind:    getChildHandler,
	DeleteChildMsgKind: deleteChildHandler,
}

type Head struct {
	ID        NodeID `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name     string
	Kind     Kind
	ParentID NodeID
	OwnerID  NodeID `gorm:"-"`
	In       Pipe   `gorm:"-"`

	path     string
	children map[string]Pipe

	guiSubs map[NodeID]Pipe
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
	if q.UserID != h.OwnerID && !q.Admin {
		slog.Debug("MSG unauthorized.",
			"node", h.path, "owner", h.OwnerID, "qUser", q.UserID, "qAdmin", q.Admin, "qKind", q.KindName())
		return NewErrorMsg(fmt.Errorf("unathorized request"))
	}
	slog.Debug("MSG received.", "node", n.getPath(), "kind", q.KindName())

	nmh, ok := nodeMsgHandlers[h.Kind][q.Kind]
	if ok {
		a = nmh(n, q)
		if a != nil {
			q.Answer(a)
			slog.Debug("MSG answered.", "node", n.getPath(), "qKind", q.KindName(), "aKind", a.KindName())
		} else {
			slog.Debug("MSG notification handled.", "node", n.getPath(), "qKind", q.KindName())
		}
		return
	}

	cmh, ok := commonMsgHandlers[q.Kind]
	if ok {
		a = cmh(h, q)
		if a != nil {
			if a.Kind == StoppedMsgKind {
				slog.Debug("MSG answer for stop message delayed.", "node", n.getPath())
				return
			}
			q.Answer(a)
			slog.Debug("MSG common answered.",
				"node", n.getPath(), "qKind", q.KindName(), "aKind", a.KindName())
		} else {
			slog.Debug("MSG common notification handled.",
				"node", n.getPath(), "qKind", q.KindName())
		}
		return
	}

	slog.Error("MSG no appropriate handler found.", "node", h.path, "qKind", q.KindName())
	return NewErrorMsg(fmt.Errorf("no appropriate handler found for %s on %s", q.KindName(), h.KindName()))
}

func createHandler(h *Head, m *Msg) (r *Msg) {
	cpl := m.Payload.(*CreatePL)
	if cpl.Kind == RootKind || cpl.Kind == UserKind {
		return NewErrorMsg(fmt.Errorf("%s kind can not be created", kindNames[cpl.Kind]))
	}
	nm := cpl.Name
	if nm == "" {
		nm = ("New" + kindNames[cpl.Kind])
	}
	if _, ok := h.children[nm]; ok {
		return NewErrorMsg(fmt.Errorf("node '%s' already exists", nm))
	}
	n := NewNodeKind(cpl.Kind)
	if n == nil {
		return NewErrorMsg(fmt.Errorf("node kind '%s' not implemented yet", kindNames[cpl.Kind]))
	}
	n.setKind(cpl.Kind)
	n.setName(nm)
	n.setParentID(h.ID)
	n.setOwnerID(h.OwnerID)

	nin, err := n.create(cpl.Payload)
	if err != nil {
		return NewErrorMsg(err)
	}
	nm = n.getName()
	h.children[nm] = nin
	n.setPath(h.path + "/" + nm)

	nnpl := &NewTreeNodePL{
		ID:       n.getID(),
		Name:     nm,
		Kind:     cpl.Kind,
		ParentID: h.ID,
		OwnerID:  h.OwnerID,
	}
	if Tree.Sys.TreeUpdater != nil {
		Tree.Sys.TreeUpdater.Notify(Msg{
			Kind:    TreeNodeCreateMsgKind,
			Payload: nnpl,
			Admin:   true,
		})
	}

	slog.Debug("NODE created.", "node", n.getPath(), "kind", n.kindName())
	return &Msg{Kind: OKMsgKind, Payload: nin}
}

func stopHandler(h *Head, m *Msg) (r *Msg) {
	h.askChildren(m)
	return &Msg{Kind: StoppedMsgKind, Payload: h.ID}
}

func (h *Head) askChildren(m *Msg) {
	chm := *m
	chm.resp = make(Pipe)
	for _, ch := range h.children {
		ch <- &chm
	}
	for range len(h.children) {
		<-chm.resp
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

func getChildHandler(h *Head, m *Msg) (r *Msg) {
	chnm := m.Payload.(string)
	ch, ok := h.children[chnm]
	if !ok {
		return NewErrorMsg(fmt.Errorf("children '%s' not found", chnm))
	}
	return &Msg{Kind: OKMsgKind, Payload: ch}
}

func getTreeHandler(h *Head, m *Msg) (r *Msg) {
	tree := &TreeEntry{
		ID:   h.ID,
		Name: h.Name,
		Kind: h.Kind,
	}

	chm := *m
	chm.resp = make(Pipe)
	for _, ch := range h.children {
		ch <- &chm
	}

	var cherr bool
	for range len(h.children) {
		chr := <-chm.resp
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
	if h.guiSubs == nil {
		h.guiSubs = make(map[NodeID]Pipe)
	}
	gui := m.Payload.(Tag)
	h.guiSubs[gui.ID] = gui.Node
	slog.Debug("GUI subscribed", "node", h.path, "gui", gui.ID)
	return &OKMsg
}

func unsubscribeHandler(h *Head, m *Msg) (r *Msg) {
	guiid := m.Payload.(NodeID)
	_, ok := h.guiSubs[guiid]
	if !ok {
		slog.Error("Can't unscrubsibe GUI that's not subscribed", "node", h.path, "gui", guiid)
		return NewErrorMsg(errors.New("attempt to unscrubscribe non-subscribed GUI"))
	}
	delete(h.guiSubs, guiid)
	slog.Debug("GUI unsubscribed", "node", h.path, "gui", guiid)
	return &OKMsg
}

func renameChildHandler(h *Head, m *Msg) (r *Msg) {
	rchpl := m.Payload.(*renameChildPL)
	ch, ok := h.children[rchpl.Name]
	if !ok {
		return NewErrorMsg(fmt.Errorf("node has no children named '%s'", rchpl.Name))
	}
	if _, ok := h.children[rchpl.NewName]; ok {
		return NewErrorMsg(fmt.Errorf("node already has a children named '%s'", rchpl.NewName))
	}

	a := ch.Ask(Msg{RenameMsgKind, m.UserID, m.Admin, rchpl.NewName, nil})
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
		UserID:  m.UserID,
		Admin:   m.Admin,
	})
	h.updateGUIs()
	Tree.Sys.TreeUpdater.Notify(Msg{
		Kind:    TreeNodeRenameMsgKind,
		Payload: &Tag{ID: h.ID, Name: h.Name, OwnerID: h.OwnerID},
		Admin:   true,
	})
	return &OKMsg
}

func deleteChildHandler(h *Head, q *Msg) *Msg {
	nm := q.Payload.(string)
	ch, ok := h.children[nm]
	if !ok {
		return NewErrorMsg(fmt.Errorf("DELETE found no children named '%s'", nm))
	}
	a := ch.Ask(Msg{StopMsgKind, q.UserID, q.Admin, nil, nil})
	if a.Kind == ErrorMsgKind {
		slog.Error("HEAD couldn't delete child.", "node", h.path, "name", nm)
		return &a
	}

	Tree.Sys.TreeUpdater.Notify(Msg{
		Kind:    TreeNodeDeleteMsgKind,
		Payload: &Tag{ParentID: h.ID, Name: nm, OwnerID: h.OwnerID},
		Admin:   true,
	})

	delete(h.children, nm)
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
	for id, g := range h.guiSubs {
		slog.Debug("sending updated msg to GUI", "gui_id", id)
		g.Notify(Msg{Kind: NodeUpdateMsgKind, Payload: h.ID, Admin: true})
	}
}
