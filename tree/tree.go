package tree

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

// System user is a special user that can do anything.
var SystemUser = &Tag{0, NK_User, nil, true, nil, nil}

type NodeID int

var Tree = nodeTree{
	nodes: make(map[NodeID]*Tag),
}

type nodeTree struct {
	nodesLock sync.RWMutex
	nodes     map[NodeID]*Tag
	Sys       runtime
}

type runtime struct {
	Root        *Tag
	TreeUpdater *Tag
	Users       *Tag
	System      *Tag
}

type TreeEntry struct {
	ID       NodeID
	Name     string
	Kind     NK
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
	rh.Owner = SystemUser
	Tree.Sys.Root, err = rh.load()
	if err != nil {
		return
	}
	slog.Info("Created Root node.", "path", rh.path)

	// Still not very nice..
	var ok bool
	Tree.Sys.Users, ok = getNode(2)
	if !ok {
		return errors.New("users node can not be found")
	}

	Tree.Sys.System, ok = getNode(3)
	if !ok {
		return errors.New("users node can not be found")
	}

	a := Tree.Sys.System.Ask(SystemUser, M_Create, NK_TreeUpdater, "TreeUpdater")
	if a.Kind == M_Error {
		return errors.New("startup: tree updater creation creation failed")
	}
	tu := a.Payload.(*Tag)
	Tree.Sys.TreeUpdater = tu

	slog.Info("Node tree initialized.", "nnodes", Tree.LenNodes())
	return
}

func (t *nodeTree) Stop() (err error) {
	a := Tree.Sys.Root.Ask(SystemUser, M_Stop)
	if a.Kind == M_Error {
		return errors.New(a.Payload.(string))
	}

	if err = CloseDB(); err != nil {
		slog.Error("PROC couldn't close database.", "err", err)
	}

	return err
}

func getNode(id NodeID) (*Tag, bool) {
	Tree.nodesLock.RLock()
	nt, ok := Tree.nodes[id]
	Tree.nodesLock.RUnlock()
	return nt, ok
}

func (tr *nodeTree) PutNode(id NodeID, nt *Tag) {
	tr.nodesLock.Lock()
	tr.nodes[id] = nt
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
