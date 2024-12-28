package core

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

var Tree = nodeTree{
	nodes: make(map[int]Pipe),
}

type TreeEntry struct {
	ID       int
	Name     string
	Kind     Kind
	Children []*TreeEntry `json:",omitempty"`
}

type nodeTree struct {
	nodesLock sync.RWMutex
	nodes     map[int]Pipe
	Sys       runtime
}

type runtime struct {
	Root        Pipe
	TreeUpdater Pipe
	Users       Pipe
	System      Pipe
}

func (t *nodeTree) LoadAndRun(sdb string) (err error) {
	if ok := FileExists(sdb); !ok {
		return fmt.Errorf("database '%s' doesn't exist", sdb)
	}
	if err = connectDB(sdb); err != nil {
		return
	}

	rootHead := &Head{}
	if err = Db.First(rootHead, 1).Error; err != nil {
		return
	}
	rootHead.path = "/Root"
	Tree.Sys.Root, err = rootHead.load()
	if err != nil {
		return
	}

	// Still not very nice..
	var ok bool
	Tree.Sys.Users, ok = Tree.GetNode(2)
	if !ok {
		return errors.New("users node can not be found")
	}

	Tree.Sys.System, ok = Tree.GetNode(3)
	if !ok {
		return errors.New("users node can not be found")
	}

	a := Tree.Sys.System.Ask(
		Msg{
			Kind: CreateMsgKind,
			Payload: &CreatePayload{
				Name: "TreeUpdater",
				Kind: TreeUpdaterKind,
			},
		})
	if a.Kind == ErrorMsgKind {
		return errors.New("startup: tree updater creation creation failed")
	}
	tu := a.Payload.(Pipe)
	Tree.Sys.TreeUpdater = tu

	slog.Info("Node tree initialized.", "nnodes", Tree.LenNodes())
	return
}

func (tr *nodeTree) GetNode(id int) (Pipe, bool) {
	tr.nodesLock.RLock()
	n, ok := tr.nodes[id]
	tr.nodesLock.RUnlock()
	return n, ok
}

func (tr *nodeTree) PutNode(id int, n Pipe) {
	tr.nodesLock.Lock()
	tr.nodes[id] = n
	tr.nodesLock.Unlock()
}

func (tr *nodeTree) LenNodes() int {
	tr.nodesLock.RLock()
	l := len(tr.nodes)
	tr.nodesLock.RUnlock()
	return l
}
