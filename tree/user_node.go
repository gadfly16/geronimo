package tree

import (
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserParms struct {
	ParmModel
	Admin    bool
	Email    string
	Password []byte
}

type UserNode struct {
	*Head
	Parms *UserParms
}

func (t *UserNode) loadBody(h *Head) (n Node, err error) {
	un := &UserNode{
		Head:  h,
		Parms: &UserParms{},
	}
	if err = Db.Where("head_id = ?", h.ID).Order("created_at desc").Take(un.Parms).Error; err != nil {
		return
	}
	h.Admin = un.Parms.Admin
	return un, nil
}

func (n *UserNode) create(pl []any) (_ *Tag, err error) {
	n.Parms = pl[0].(*UserParms)
	n.Parms.Password, err = bcrypt.GenerateFromPassword(n.Parms.Password, 14)
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

	n.Admin = n.Parms.Admin
	n.Owner = n.Tag
	go n.run()
	n.Head.initNew()
	return n.Head.Tag, nil
}

func (n *UserNode) run() {
	defer close(n.Head.In)
	defer Tree.RemoveNode(n.Head.ID)

	slog.Debug("Running User node.", "node", n.Head.path)
	for q := range n.In {
		a := handleMsg(n, q)
		if a != nil && a.Kind == M_Stop {
			// Drain unsubscribe messages
			for range len(n.Head.guiSubs) {
				q := <-n.Head.In
				slog.Debug("SINK of User reveived msg.", "node", n.Head.path, "kind", q.KindName())
				q.AnswerOK()
				slog.Debug("SINK of User answered msg.", "node", n.Head.path, "kind", q.KindName())
			}
			q.AnswerMsg(a)
			break
		}
	}

	slog.Info("Stopped User node.", "node", n.Head.path)
}

func (n *UserNode) updateParms(pl H) (err error) {
	ps := &UserParms{
		ParmModel: ParmModel{
			HeadID: n.ID,
		},
		Admin:    pl["Admin"].(bool),
		Email:    pl["Email"].(string),
		Password: n.Parms.Password,
	}
	if pw, ok := pl["Password"]; ok {
		ps.Password, err = bcrypt.GenerateFromPassword([]byte(pw.(string)), 14)
		if err != nil {
			return
		}
	}
	err = Db.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(ps).Error; err != nil {
			return err
		}
		return
	})
	if err != nil {
		return
	}
	n.Parms = ps
	n.updateGUIs()
	return
}

func (n *UserNode) getDisplay(d H) H {
	d["Parms"] = H{
		"Email":    n.Parms.Email,
		"Admin":    n.Parms.Admin,
		"Password": "",
	}
	return d
}

func (n *UserNode) allowedChildren(nk NK) bool {
	return nk == NK_Group
}

func (n *UserNode) getParms() any {
	return *n.Parms
}
