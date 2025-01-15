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
	defer close(n.Head.In)
	defer Tree.RemoveNode(n.Head.ID)

	slog.Debug("USERS node starting up.", "node", n.Head.path)
	for q := range n.Head.In {
		a := n.Head.handleMsg(n, q)
		if a != nil && a.Kind == StoppedMsgKind {
			// Drain unsubscribe messages
			for range len(n.Head.guiSubs) {
				q := <-n.Head.In
				q.AnswerOK()
			}
			q.AnswerMsg(a)
			break
		}
	}

	slog.Info("Stopped Users node.", "node", n.path)
}

func (n *UsersNode) create(_ any) (_ *Tag, err error) {
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
	go n.run()
	return n.Head.Tag, nil
}

func createUserHandler(ni Node, m *Msg) (r *Msg) {
	n := ni.(*UsersNode)
	nu := m.Payload.(*UserNode)
	if nu.Kind != UserKind {
		return NewErrorMsg(fmt.Errorf("user node kind isn't userKind"))
	}
	if nu.Name == "" {
		return NewErrorMsg(fmt.Errorf("new user node must have a name"))
	}
	if _, ok := n.children[nu.Name]; ok {
		return NewErrorMsg(fmt.Errorf("user node '%s' already exists", nu.Name))
	}
	nu.Parent = n.Tag
	nu.ParentID = n.ID
	nu.path = n.path + "/" + nu.Name
	if len(n.children) == 0 {
		nu.Parms.Admin = true
	}

	nut, err := nu.create(nil)
	if err != nil {
		return NewErrorMsg(err)
	}
	nu.Admin = nu.Parms.Admin
	nu.Owner = nut
	n.children[nu.Name] = nut

	Tree.Sys.TreeUpdater.Notify(TreeNodeCreateMsgKind, SystemUser, nut, nu.Name, nut)

	return &Msg{Kind: OKMsgKind, Payload: nut}
}

func authUserHandler(ni Node, q *Msg) (a *Msg) {
	n := ni.(*UsersNode)
	uc := q.Payload.(*UserNode)
	slog.Debug("AUTH getting User from children", "name", uc.Head.Name)
	u, ok := n.children[uc.Head.Name]
	if !ok {
		return NewErrorMsg(fmt.Errorf("user not found"))
	}
	up := u.Ask(GetCopyMsgKind, SystemUser).Payload.(UserNode)

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
