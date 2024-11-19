package core

import (
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
	nodesLock   sync.RWMutex
	nodes       map[int]Pipe
	Root        Pipe
	TreeUpdater Pipe
}

type treeUpdater struct {
	in   Pipe
	guis map[int]map[Pipe]bool
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

	tu := &treeUpdater{
		in:   make(Pipe),
		guis: make(map[int]map[Pipe]bool),
	}
	Tree.TreeUpdater = tu.in
	go tu.run()

	slog.Info("Node tree initialized.", "nnodes", Tree.LenNodes())
	return
}

func (tu *treeUpdater) run() {
	slog.Debug("Running tree updater.")
	for q := range tu.in {
		switch q.Kind {
		case SubscribeMsgKind:
			tupl := q.Payload.(SubscribePayload)
			_, ok := tu.guis[tupl.ID]
			if !ok {
				tu.guis[tupl.ID] = make(map[Pipe]bool, 0)
			}
			tu.guis[tupl.ID][tupl.Node] = true
			slog.Debug("registered new GUI for tree updates", "guis", tu.guis)
			q.Answer(&OKMsg)
		case UnsubscribeMsgKind:
			tupl := q.Payload.(SubscribePayload)
			delete(tu.guis[tupl.ID], tupl.Node)
			slog.Debug("unregistered new GUI for tree updates", "guis", tu.guis)
			q.Answer(&OKMsg)
		default:
			slog.Debug("Unhandled msg received by treeUpdater.")
		}
	}
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
