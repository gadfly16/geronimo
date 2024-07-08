package node

import (
	"encoding/json"
	"io"
	"log/slog"

	"github.com/gadfly16/geronimo/msg"
	"gorm.io/gorm"
)

func init() {
	nodeMsgHandlers[UserKind] = map[msg.Kind]func(Node, *msg.Msg) *msg.Msg{
		// msg.UpdateKind:   rootUpdateHandler,
		// msg.GetParmsKind: rootGetParmsHandler,
	}
}

type UserParms struct {
	ParmModel
	Admin       bool
	DisplayName string
	Password    string
}

type UserNode struct {
	*Head
	Parms *UserParms
}

func (n *UserNode) run() {
	slog.Debug("Running User node.", "name", n.Head.Name)
	for q := range n.In {
		slog.Debug("Message received.", "node", n.Name, "kind", q.KindName())
		r := n.Head.handleMsg(n, q)
		r.Answer(q)
		slog.Debug("Message answered.", "node", n.Name, "kind", r.KindName())
		if r.Kind == msg.StoppedKind {
			break
		}
	}
	slog.Info("Stopped Root node.")
}

func (t *UserNode) loadBody(h *Head) (n Node, err error) {
	h.In = make(msg.Pipe)
	un := &UserNode{
		Head:  h,
		Parms: &UserParms{},
	}
	if err = Db.Where("head_id = ?", h.ID).Order("created_at desc").Take(un.Parms).Error; err != nil {
		return
	}
	return un, nil
}

func (n *UserNode) create() (in msg.Pipe, err error) {
	n.Parms.Password, err = hashPassword(n.Parms.Password)
	if err != nil {
		return
	}
	err = Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&n.Head).Error; err != nil {
			return err
		}
		n.Parms.HeadID = n.Head.ID
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
	slog.Info("Created User node.", "name", n.Head.Name)
	return n.Head.In, nil
}

func (n *UserNode) UnmarshalMsg(b io.ReadCloser) (m *msg.Msg, err error) {
	m = &msg.Msg{
		Payload: UserNode{},
	}
	d := json.NewDecoder(b)
	err = d.Decode(m)
	return
}

// func rootGetParmsHandler(ni Node, m *msg.Msg) (r *msg.Msg) {
// 	n := ni.(*RootNode)
// 	return &msg.Msg{
// 		Kind:    msg.ParmsKind,
// 		Payload: *n.Parms,
// 	}
// }

// func rootUpdateHandler(ni Node, m *msg.Msg) (r *msg.Msg) {
// 	// if v, ok := m.Payload["parms"]; ok {
// 	// 	if n.Parms, ok = v.(*RootParms); !ok {
// 	// 		return fmt.Errorf("update failed: wrong parms type %T", n.Parms)
// 	// 	}
// 	// }
// 	// Db.Transaction(func(tx *gorm.DB) error {
// 	// 	return nil
// 	// })
// 	return
// }

// func (n *RootNode) setLogLevel() {
// 	LogLevel.Set(slog.Level(n.Parms.LogLevel))
// }
