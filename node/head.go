package node

import (
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

	// cache struct {
	// 	valid bool
	// }
}

// var Loaders = map[Kind]func(*Head) Node{
// 	RootKind: root.Loader,
// }

func (h *Head) Load() (in msg.Pipe, err error) {
	// if err = Db.First(h, h.ID).Error; err != nil {
	// 	return
	// }
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
		// if err = detailLoaders[ch.DetailType](ch); err != nil {
		// 	return
		// }
		// if err = ch.loadChildren(); err != nil {
		// 	return
		// }
	}

	Tree.NodeLock.Lock()
	Tree.Nodes[h.ID] = h.In
	Tree.NodeLock.Unlock()

	go n.run()

	return h.In, err
}
