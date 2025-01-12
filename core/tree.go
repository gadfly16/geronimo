package core

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

var Tree = nodeTree{
	nodes: make(map[NodeID]Pipe),
}

type nodeTree struct {
	nodesLock sync.RWMutex
	nodes     map[NodeID]Pipe
	Sys       runtime
}

type runtime struct {
	Root        Pipe
	TreeUpdater Pipe
	Users       Pipe
	System      Pipe
}

type TreeEntry struct {
	ID       NodeID
	Name     string
	Kind     Kind
	Children []*TreeEntry `json:",omitempty"`
}

func (t *nodeTree) LoadAndRun(sdb string) (err error) {
	if ok := FileExists(sdb); !ok {
		return fmt.Errorf("database '%s' doesn't exist", sdb)
	}
	if err = connectDB(sdb); err != nil {
		return
	}

	rh := &Head{}
	if err = Db.First(rh, 1).Error; err != nil {
		return
	}
	rh.path = "/Root"
	Tree.Sys.Root, err = rh.load()
	if err != nil {
		return
	}
	slog.Info("Created Root node.", "path", rh.path)

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
			Kind:  CreateMsgKind,
			Admin: true,
			Payload: &CreatePL{
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

func (t *nodeTree) Stop() (err error) {
	a := Tree.Sys.Root.Ask(Msg{
		Kind:  StopMsgKind,
		Admin: true,
	})
	if a.Kind == ErrorMsgKind {
		return errors.New(a.Payload.(string))
	}

	if err = CloseDB(); err != nil {
		slog.Error("PROC couldn't close database.", "err", err)
	}

	return err
}

func (tr *nodeTree) GetNode(id NodeID) (Pipe, bool) {
	tr.nodesLock.RLock()
	n, ok := tr.nodes[id]
	tr.nodesLock.RUnlock()
	return n, ok
}

func (tr *nodeTree) PutNode(id NodeID, n Pipe) {
	tr.nodesLock.Lock()
	tr.nodes[id] = n
	tr.nodesLock.Unlock()
}

func (tr *nodeTree) RemoveNode(id NodeID) {
	tr.nodesLock.Lock()
	delete(tr.nodes, id)
	tr.nodesLock.Unlock()
}

func (tr *nodeTree) LenNodes() int {
	tr.nodesLock.RLock()
	l := len(tr.nodes)
	tr.nodesLock.RUnlock()
	return l
}
