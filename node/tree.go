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
	NodeLock    sync.RWMutex
	Nodes       map[int]msg.Pipe
	Root        msg.Pipe
	TreeUpdater msg.Pipe
}

type treeUpdater struct {
	in   msg.Pipe
	guis map[int]map[msg.Pipe]bool
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
		in:   make(msg.Pipe),
		guis: make(map[int]map[msg.Pipe]bool),
	}
	Tree.TreeUpdater = tu.in
	go tu.run()

	slog.Info("Node tree initialized.", "nnodes", len(Tree.Nodes))
	return
}

func (tu *treeUpdater) run() {
	slog.Debug("Running tree updater.")
	for q := range tu.in {
		switch q.Kind {
		case msg.SubscribeKind:
			tupl := q.Payload.(SubscribePayload)
			_, ok := tu.guis[tupl.ID]
			if !ok {
				tu.guis[tupl.ID] = make(map[msg.Pipe]bool, 0)
			}
			tu.guis[tupl.ID][tupl.Node] = true
			slog.Debug("registered new GUI for tree updates", "guis", tu.guis)
			q.Answer(&msg.OK)
		case msg.UnsubscribeKind:
			tupl := q.Payload.(SubscribePayload)
			delete(tu.guis[tupl.ID], tupl.Node)
			slog.Debug("unregistered new GUI for tree updates", "guis", tu.guis)
			q.Answer(&msg.OK)
		default:
			slog.Debug("Unhandled msg received by treeUpdater.")
		}
	}
}
