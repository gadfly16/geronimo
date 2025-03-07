package tree

import (
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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

func (n *UsersNode) create(_ []any) (_ *Tag, err error) {
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

func (n *UsersNode) run() {
	defer close(n.Head.In)
	defer Tree.RemoveNode(n.Head.ID)

	slog.Debug("USERS node starting up.", "node", n.Head.path)
	for q := range n.Head.In {
		a := handleMsg(n, q)
		if a != nil && a.Kind == M_Stop {
			// Drain unsubscribe messages
			for range len(n.Head.guiSubs) {
				q := <-n.Head.In
				q.AnswerOK()
			}
			q.AnswerMsg(a)
			break
		}
	}
	slog.Info("USERS stopped.", "node", n.path)
}

func (n *UsersNode) authUser(pl any) (*Tag, error) {
	pls := pl.([]any)
	nm := pls[0].(string)
	pwd := pls[1].(string)
	u, ok := n.children[nm]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	up := u.Ask(SystemUser, M_Get_Parms).Payload.(UserParms)
	err := bcrypt.CompareHashAndPassword(up.Password, []byte(pwd))
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (n *UsersNode) getDisplay(d H) H {
	d["Parms"] = H{
		"Invitation Only": n.Parms.InvitationOnly,
	}
	return d
}

func (n *UsersNode) allowedChildren(nnk NK) bool {
	return nnk == NK_User
}
