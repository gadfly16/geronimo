package tree

import (
	"sync"
)

// System user is a special user that can do anything.
var SystemUser = &Tag{0, NK_User, nil, true, nil, nil}

type NodeID = int

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
