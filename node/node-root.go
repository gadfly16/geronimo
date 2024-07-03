package node

import (
	"log/slog"

	"github.com/gadfly16/geronimo/msg"
	"gorm.io/gorm"
)

type RootParms struct {
	parmModel
	LogLevel string
	HTTPAddr string
	DbKey    string
}

type RootNode struct {
	*Head
	Parms *RootParms
}

var rootMsgHandlers = map[msg.Kind]func(Node, *msg.Msg) *msg.Msg{
	msg.UpdateKind:   rootUpdateHandler,
	msg.GetParmsKind: rootGetParmsHandler,
	msg.StopRootKind: rootStopTreeHandler,
}

func (n *RootNode) run() {
	slog.Info("Running Root node.", "name", n.Head.Name, "logLevel", n.Parms.LogLevel)
	for q := range n.In {
		slog.Info("Message received.", "node", n.Name, "kind", q.KindName())
		r := n.Head.commonMsg(q)
		if r == nil {
			r = rootMsgHandlers[q.Kind](n, q)
		}
		r.Answer(q)
		if r.Kind == msg.StopRootKind {
			break
		}
		slog.Info("Message answered.", "node", n.Name, "query", q.KindName(), "resp", r.KindName())
	}
	slog.Info("Stopped Root node.")
}

func (t *RootNode) load(h *Head) (n Node, err error) {
	h.In = make(msg.Pipe)
	rn := &RootNode{
		Head:  h,
		Parms: &RootParms{},
	}
	if err = Db.Where("node_id = ?", h.ID).Order("created_at desc").Take(rn.Parms).Error; err != nil {
		return
	}

	return rn, nil
}

func (n *RootNode) create() (in msg.Pipe, err error) {
	err = Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&n.Head).Error; err != nil {
			return err
		}
		n.Parms.NodeID = n.Head.ID
		if err := tx.Create(&n.Parms).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}
	go n.run()
	n.Head.register()
	Tree.Root = n.Head.In
	slog.Info("Created Root node.", "name", n.Head.Name)
	return n.Head.In, nil
}

func (n *RootNode) Init() (err error) {
	_, err = n.create()
	return err
}

func rootStopTreeHandler(ni Node, m *msg.Msg) (r *msg.Msg) {
	n := ni.(*RootNode)
	for _, ch := range n.Head.children {
		close(ch)
	}
	for range len(Tree.Nodes) - 1 {
		<-n.Head.In
	}
	return m
}

func rootGetParmsHandler(ni Node, m *msg.Msg) (r *msg.Msg) {
	n := ni.(*RootNode)
	return &msg.Msg{
		Kind:    msg.ParmsKind,
		Payload: *n.Parms,
	}
}

func rootUpdateHandler(ni Node, m *msg.Msg) (r *msg.Msg) {
	// if v, ok := m.Payload["parms"]; ok {
	// 	if n.Parms, ok = v.(*RootParms); !ok {
	// 		return fmt.Errorf("update failed: wrong parms type %T", n.Parms)
	// 	}
	// }
	// Db.Transaction(func(tx *gorm.DB) error {
	// 	return nil
	// })
	return
}
