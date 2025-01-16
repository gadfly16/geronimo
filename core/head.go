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
	CreateChildMsgKind: createChildHandler,
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
	*Tag
	ParentID  NodeID // This is only used for database storage.
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name  string
	Owner *Tag `gorm:"-"`

	path     string
	children map[string]*Tag

	guiSubs map[*Tag]E
}

func (h *Head) load() (nt *Tag, err error) {
	h.In = make(Pipe)
	var ok bool
	h.Parent, ok = Tree.GetNode(h.ParentID)
	if !ok && h.ParentID != 0 {
		return nil, fmt.Errorf("parent (%d) not found for node (%d)", h.ParentID, h.ID)
	}
	h.children = make(map[string]*Tag)
	if h.Kind == UserKind {
		h.Owner = h.Tag
	}
	n, err := KindTemplates[h.Kind].loadBody(h)
	if err != nil {
		return
	}
	slog.Debug("Loaded node.", "node", h.path)
	Tree.PutNode(h.ID, h.Tag)

	chs := []*Head{}
	if err = Db.Where("parent_id = ?", h.ID).Find(&chs).Error; err != nil {
		return
	}
	for _, ch := range chs {
		ch.path = h.path + "/" + ch.Name
		ch.Owner = h.Owner
		var cht *Tag
		cht, err = ch.load()
		if err != nil {
			return
		}
		h.children[ch.Name] = cht
	}
	go n.run()
	return h.Tag, err
}

func (h *Head) initNew() {
	h.children = make(map[string]*Tag)
	h.In = make(Pipe)
	Tree.PutNode(h.ID, h.Tag)
}

func (h *Head) handleMsg(n Node, q *Msg) (a *Msg) {
	if q.User != h.Owner && !q.User.Admin {
		slog.Debug("MSG unauthorized.", "n", h.path, "no", h.Owner.ID, "qu", q.User.ID, "qua", q.User.Admin, "qk", q.KindName())
		return NewErrorMsg(fmt.Errorf("unathorized request"))
	}
	slog.Debug("MSG received.", "node", n.getPath(), "kind", q.KindName())

	nmh, ok := nodeMsgHandlers[h.Kind][q.Kind]
	if ok {
		a = nmh(n, q)
		if a != nil {
			q.AnswerMsg(a)
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
			q.AnswerMsg(a)
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

func createChildHandler(h *Head, m *Msg) (r *Msg) {
	pl := m.Payload.([]any)
	nnk := pl[0].(Kind)
	nnm := pl[1].(string)
	if nnk == RootKind || nnk == UserKind {
		return NewErrorMsg(fmt.Errorf("%s kind can not be created", kindNames[nnk]))
	}
	nm := nnm
	if nm == "" {
		nm = ("New" + kindNames[nnk])
	}
	if _, ok := h.children[nm]; ok {
		return NewErrorMsg(fmt.Errorf("node '%s' already exists", nm))
	}
	nn := NewNodeKind(nnk)
	if nn == nil {
		return NewErrorMsg(fmt.Errorf("node kind '%s' not implemented yet", kindNames[nnk]))
	}
	nn.setKind(nnk)
	nn.setName(nm)
	nn.setParentID(h.Tag)
	nn.setOwnerID(h.Owner)

	nnt, err := nn.create(pl[2:])
	if err != nil {
		return NewErrorMsg(err)
	}
	//Name might have been changed by create.
	nm = nn.getName()
	h.children[nm] = nnt
	nn.setPath(h.path + "/" + nm)

	if Tree.Sys.TreeUpdater != nil {
		Tree.Sys.TreeUpdater.Notify(TreeNodeCreateMsgKind, SystemUser, nnt, nm, h.Owner)
	}

	slog.Debug("NODE created.", "node", nn.getPath(), "kind", nn.kindName())
	return &Msg{Kind: OKMsgKind, Payload: nnt}
}

func stopHandler(h *Head, m *Msg) (r *Msg) {
	h.askChildrenMsg(m)
	return &Msg{Kind: StoppedMsgKind, Payload: h.ID}
}

func (h *Head) askChildren(k MsgKind, u *Tag, pl ...any) {
	var epl any = pl
	if len(pl) == 1 {
		epl = pl[0]
	}
	q := &Msg{
		Kind:    k,
		User:    u,
		Payload: epl,
		resp:    make(Pipe),
	}
	for _, ch := range h.children {
		ch.In <- q
	}
	for range len(h.children) {
		<-q.resp
	}
}

func (h *Head) askChildrenMsg(q *Msg) {
	// Children must receive a copy of the message, to avoid reponding to upstream callers.
	chq := *q
	chq.resp = make(Pipe)
	for _, ch := range h.children {
		ch.In <- &chq
	}
	for range len(h.children) {
		<-chq.resp
	}
}

func updatePathHandler(h *Head, m *Msg) (r *Msg) {
	h.path = m.Payload.(string) + "/" + h.Name
	h.askChildren(UpdatePathMsgKind, m.User, h.path)
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
		ch.In <- &chm
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
		h.guiSubs = make(map[*Tag]E)
	}
	gui := m.Payload.(*Tag)
	h.guiSubs[gui] = E{}
	slog.Debug("GUI subscribed", "node", h.path, "gui", gui.ID)
	return nil
}

func unsubscribeHandler(h *Head, m *Msg) (r *Msg) {
	gt := m.Payload.(*Tag)
	delete(h.guiSubs, gt)
	slog.Debug("GUI unsubscribed", "node", h.path, "gui", gt.ID)
	return nil
}

func renameChildHandler(h *Head, q *Msg) *Msg {
	pl := q.Payload.([]any)
	nm := pl[0].(string)
	nnm := pl[1].(string)
	ch, ok := h.children[nm]
	if !ok {
		return NewErrorMsg(fmt.Errorf("node has no children named '%s'", nm))
	}
	if _, ok := h.children[nnm]; ok {
		return NewErrorMsg(fmt.Errorf("node already has a children named '%s'", nnm))
	}

	a := ch.Ask(RenameMsgKind, q.User, nnm)
	if a.Kind == ErrorMsgKind {
		return &a
	}

	h.children[nnm] = ch
	delete(h.children, nm)
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
	h.askChildren(UpdatePathMsgKind, m.User, h.path)
	h.updateGUIs()
	Tree.Sys.TreeUpdater.Notify(TreeNodeRenameMsgKind, SystemUser, h.Tag, h.Name, h.Owner)
	return &OKMsg
}

func deleteChildHandler(h *Head, q *Msg) *Msg {
	nm := q.Payload.(string)
	ch, ok := h.children[nm]
	if !ok {
		return NewErrorMsg(fmt.Errorf("DELETE found no children named '%s'", nm))
	}
	a := ch.Ask(StopMsgKind, q.User)
	if a.Kind == ErrorMsgKind {
		slog.Error("HEAD couldn't delete child.", "node", h.path, "name", nm)
		return &a
	}

	Tree.Sys.TreeUpdater.Notify(TreeNodeDeleteMsgKind, SystemUser, ch, nm, h.Owner)

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
	for gt := range h.guiSubs {
		slog.Debug("sending updated msg to GUI", "gui_id", gt.ID)
		gt.Notify(NodeUpdateMsgKind, SystemUser, h.Tag)
	}
}
