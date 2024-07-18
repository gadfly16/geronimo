package node

import (
	"log/slog"
	"sync"

	"github.com/gadfly16/geronimo/msg"
)

var Tree = nodeTree{
	Nodes: make(map[int]msg.Pipe),
}

type TreeEntry struct {
	ID       int
	Name     string
	Kind     Kind
	Children []*TreeEntry `json:",omitempty"`
}

type nodeTree struct {
	NodeLock sync.RWMutex
	Nodes    map[int]msg.Pipe
	Root     msg.Pipe
}

func (t *nodeTree) Load(sdb string) (err error) {
	ConnectDB(sdb)
	rootHead := &Head{}
	if err = Db.First(rootHead, 1).Error; err != nil {
		return
	}
	rootHead.path = "/Root"
	Tree.Root, err = rootHead.load()
	if err != nil {
		return
	}
	slog.Info("Node tree initialized.", "nnodes", len(Tree.Nodes))
	return
}
