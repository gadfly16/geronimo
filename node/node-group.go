package node

import (
	"fmt"
	"log/slog"

	"github.com/gadfly16/geronimo/msg"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func init() {
	nodeMsgHandlers[GroupKind] = map[msg.Kind]func(Node, *msg.Msg) *msg.Msg{
		msg.AuthUserKind: groupAuthUserHandler,
		// msg.UpdateKind:   rootUpdateHandler,
		// msg.GetParmsKind: rootGetParmsHandler,
	}
}

type GroupNode struct {
	*Head
}

func (t *GroupNode) loadBody(h *Head) (n Node, err error) {
	// h.In = make(msg.Pipe)
	gn := &GroupNode{
		Head: h,
	}

	return gn, nil
}

func (n *GroupNode) run() {
	slog.Info("Running Group node.", "name", n.Head.Name)
	for m := range n.Head.In {
		slog.Info("Message received.", "node", n.path, "kind", m.KindName())
		r := n.Head.handleMsg(n, m)
		r.Answer(m)
		slog.Info("Message answered.", "node", n.path, "kind", r.KindName())
		if r.Kind == msg.StoppedKind {
			break
		}
	}
	slog.Info("Stopped Group node.", "node", n.path)
}

func (n *GroupNode) create(pp string) (in msg.Pipe, err error) {
	n.Head.path = pp + "/" + n.Head.Name
	err = Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&n.Head).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return
	}
	n.Head.register()
	slog.Info("Created Group node.", "node", n.Head.path)
	go n.run()
	return n.Head.In, nil
}

func groupAuthUserHandler(ni Node, m *msg.Msg) (r *msg.Msg) {
	n := ni.(*GroupNode)
	uc := m.Payload.(*UserNode)
	slog.Debug("getting user from children", "user_credentials", uc)
	u, ok := n.children[uc.Name]
	if !ok {
		return msg.NewErrorMsg(fmt.Errorf("user not found"))
	}
	up := u.Ask(msg.GetParms).Payload.(UserParms)
	err := bcrypt.CompareHashAndPassword(up.Password, uc.Parms.Password)
	if err != nil {
		return msg.NewErrorMsg(err)
	}
	return &msg.Msg{Kind: msg.ParmsKind, Payload: up}
}
