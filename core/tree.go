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
	for m := range tu.in {
		switch m.Kind {
		case SubscribeMsgKind:
			spl := m.Payload.(SubscribePayload)
			_, ok := tu.guis[spl.ID]
			if !ok {
				tu.guis[spl.ID] = make(map[Pipe]bool, 0)
			}
			tu.guis[spl.ID][spl.Node] = true
			slog.Debug("registered new GUI for tree updates", "guis", tu.guis)
			m.Answer(&OKMsg)
		case UnsubscribeMsgKind:
			tupl := m.Payload.(SubscribePayload)
			delete(tu.guis[tupl.ID], tupl.Node)
			slog.Debug("unregistered new GUI for tree updates", "guis", tu.guis)
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
