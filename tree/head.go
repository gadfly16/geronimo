package tree

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID     NodeID `gorm:"primarykey"`
	Kind   NK
	In     Pipe `gorm:"-"`
	Admin  bool `gorm:"-"`
	Parent *Tag `gorm:"-"`
	Owner  *Tag `gorm:"-"`
}

type Head struct {
	*Tag
	ParentID  NodeID // This is only used for database storage.
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name string

	path     string
	children map[string]*Tag

	guiSubs map[*Tag]E
}

func (h *Head) load() (nt *Tag, err error) {
	h.In = make(Pipe)
	var ok bool
	h.Parent, ok = getNode(h.ParentID)
	if !ok && h.ParentID != 0 {
		return nil, fmt.Errorf("parent (%d) not found for node (%d)", h.ParentID, h.ID)
	}
	h.children = make(map[string]*Tag)
	if h.Kind == NK_User {
		h.Owner = h.Tag
	}
	n, err := nkTemplates[h.Kind].loadBody(h)
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

func (h *Head) askChildren(k MK, u *Tag, pl ...any) {
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

func (h *Head) updateGUIs() {
	for gt := range h.guiSubs {
		slog.Debug("sending updated msg to GUI", "gui_id", gt.ID)
		gt.Notify(SystemUser, M_Refresh_Node, h.Tag)
	}
}
