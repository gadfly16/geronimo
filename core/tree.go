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
	guis map[int]map[Pipe]bool // Maps a user id to a gui
}

func (t *nodeTree) LoadAndRun(sdb string) (err error) {
	if err = ConnectDB(sdb); err != nil {
		return
	}

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
	for m := range tu.in {
		switch m.Kind {
		case SubscribeMsgKind:
			t := m.Payload.(Tag)
			_, ok := tu.guis[t.ID]
			if !ok {
				tu.guis[t.ID] = make(map[Pipe]bool, 0)
			}
			tu.guis[t.ID][t.Node] = true
			slog.Debug("TU registered new GUI for updates.", "guis", tu.guis)
			m.Answer(&OKMsg)
		case UnsubscribeMsgKind:
			t := m.Payload.(Tag)
			delete(tu.guis[t.ID], t.Node)
			slog.Debug("TU unregistered GUI from updates.", "guis", tu.guis)
			m.Answer(&OKMsg)
		case TreeNodeRenameMsgKind:
			h := m.Payload.(Head)
			for g := range tu.guis[h.OwnerID] {
				g.Notify(*m)
				slog.Debug("GUI notified about tree node rename", "GUI", h.OwnerID)
			}
			for g := range tu.guis[0] {
				g.Notify(*m)
				slog.Debug("Admin GUI notified about tree node rename", "GUI", h.OwnerID)
			}
			slog.Debug("Tree node rename reveived", "user", h.OwnerID)
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
