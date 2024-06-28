package node

import (
	"log/slog"
	"time"

	"github.com/gadfly16/geronimo/msg"
	"gorm.io/gorm"
)

type RootParms struct {
	ID        int `gorm:"primarykey"`
	CreatedAt time.Time
	NodeID    int
	LogLevel  string
	HTTPAddr  string
	DbKey     string
}

type RootNode struct {
	*Head
	Parms *RootParms
}

var msgHandlers = map[msg.Kind]func(*RootNode, *msg.Msg) error{
	msg.UpdateKind:   updateHandler,
	msg.GetParmsKind: getParmsHandler,
}

func (n *RootNode) run() {
	slog.Info("Running root node.", "logLevel", n.Parms.LogLevel)
	for m := range n.In {
		if err := msgHandlers[m.Kind](n, m); err != nil {
			m.Resp <- &msg.Msg{Kind: msg.ErrorKind}
		}

		slog.Info("Message received.", "node", n.Name, "MsgKind", m.Kind)
		m.Resp <- &msg.Msg{Kind: msg.OKKind}
	}
	slog.Info("Stopping root node.")
}

func updateHandler(n *RootNode, m *msg.Msg) (err error) {
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

func getParmsHandler(n *RootNode, m *msg.Msg) (err error) {
	(&msg.Msg{
		Kind:    msg.ParmsKind,
		Payload: *n.Parms,
	}).Answer(m)
	return
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

func (n *RootNode) Create() (err error) {
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
	n.Head.In = make(msg.Pipe)
	Tree.NodeLock.Lock()
	Tree.Nodes[n.Head.ID] = n.Head.In
	Tree.NodeLock.Unlock()
	// go n.run()
	return
}

// func LoadRootParms() (rp *RootParms) {
// 	if err := Db.Where("node_id = ?", 1).Order("created_at desc").Take(rp).Error; err != nil {
// 		return nil
// 	}
// 	return
// }
