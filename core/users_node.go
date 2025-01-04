package core

import (
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func init() {
	nodeMsgHandlers[UsersKind] = map[MsgKind]func(Node, *Msg) *Msg{
		CreateUserMsgKind: createUserHandler,
		AuthUserMsgKind:   authUserHandler,
		GetDisplayMsgKind: usersGetDisplayHandler,
		// msg.UpdateKind:   rootUpdateHandler,
		// msg.GetParmsKind: rootGetParmsHandler,
	}
}

type UsersParms struct {
	ParmModel
	InvitationOnly bool
}

type UsersNode struct {
	*Head
	Parms *UsersParms
}

func (t *UsersNode) loadBody(h *Head) (n Node, err error) {
	// h.In = make(Pipe)
	gn := &UsersNode{
		Head:  h,
		Parms: &UsersParms{},
	}
	if err = Db.Where("head_id = ?", h.ID).Order("created_at desc").Take(gn.Parms).Error; err != nil {
		return
	}
	return gn, nil
}

func (n *UsersNode) run() {
	slog.Info("Running Users node.", "name", n.Head.Name)
	for q := range n.Head.In {
		a := n.Head.handleMsg(n, q)
		if a != nil && a.Kind == StoppedMsgKind {
			break
		}
	}
	slog.Info("Stopped Users node.", "node", n.path)
}

func (n *UsersNode) create() (in Pipe, err error) {
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
	n.Head.initNew()
	slog.Info("Created Users node.", "node", n.Head.path)
	go n.run()
	return n.Head.In, nil
}

func createUserHandler(ni Node, m *Msg) (r *Msg) {
	n := ni.(*UsersNode)
	nu := m.Payload.(*UserNode)
	if nu.Head.Kind != UserKind {
		return NewErrorMsg(fmt.Errorf("user node kind isn't userKind"))
	}
	if nu.Head.Name == "" {
		return NewErrorMsg(fmt.Errorf("new user node must have a name"))
	}
	if _, ok := n.Head.children[nu.Head.Name]; ok {
		return NewErrorMsg(fmt.Errorf("user node '%s' already exists", nu.Head.Name))
	}
	nu.Head.ParentID = n.Head.ID
	nu.Head.path = n.Head.path + "/" + nu.Head.Name
	if len(n.Head.children) == 0 {
		nu.Parms.Admin = true
	}

	nin, err := nu.create()
	if err != nil {
		return NewErrorMsg(err)
	}
	n.Head.children[nu.Head.Name] = nin

	nnpl := &NewTreeNodePL{
		ID:       nu.Head.ID,
		Name:     nu.Head.Name,
		Kind:     nu.Head.Kind,
		ParentID: nu.Head.ParentID,
		OwnerID:  nu.Head.OwnerID,
	}
	Tree.Sys.TreeUpdater.Notify(Msg{
		Kind:    TreeNodeCreateMsgKind,
		Admin:   true,
		Payload: nnpl,
	})

	return &Msg{Kind: OKMsgKind, Payload: nin}
}

func authUserHandler(ni Node, q *Msg) (a *Msg) {
	n := ni.(*UsersNode)
	uc := q.Payload.(*UserNode)
	slog.Debug("AUTH getting User from children", "name", uc.Head.Name)
	u, ok := n.children[uc.Head.Name]
	if !ok {
		return NewErrorMsg(fmt.Errorf("user not found"))
	}
	up := u.Ask(Msg{
		Kind:  GetCopyMsgKind,
		Admin: true,
	}).Payload.(UserNode)

	err := bcrypt.CompareHashAndPassword(up.Parms.Password, uc.Parms.Password)
	if err != nil {
		return NewErrorMsg(err)
	}
	return &Msg{Kind: ParmsMsgKind, Payload: up}
}

func usersGetDisplayHandler(ni Node, _ *Msg) *Msg {
	n := ni.(*UsersNode)
	d := n.Head.display()
	d["Parms"] = H{
		"Invitation Only": n.Parms.InvitationOnly,
	}
	r := &Msg{
		Kind:    DisplayMsgKind,
		Payload: d,
	}
	return r
}
