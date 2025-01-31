package tree

import (
	"fmt"
	"io"
	"log/slog"
)

func AskJSON(tid, uid NodeID, b io.ReadCloser) (err error) {
	t, ok := getNode(tid)
	if !ok {
		return fmt.Errorf("target node can not be found")
	}
	u, ok := getNode(uid)
	if !ok {
		return fmt.Errorf("user node can not be found")
	}
	if t.Owner != u && !u.Admin {
		return fmt.Errorf("unauthorized message")
	}
	return nil
}

func GetServerParms() (rp RootParms, err error) {
	a := Tree.Sys.Root.Ask(SystemUser, M_Get_Parms)
	if a.Kind == M_Error {
		return rp, a.Payload.(error)
	}
	return a.Payload.(RootParms), nil
}

func CreateUser(nm, email, pwd string) error {
	a := Tree.Sys.Users.Ask(SystemUser, M_Create, NK_User, nm,
		&UserParms{Email: email, Password: []byte(pwd)})
	if a.Kind == M_Error {
		return a.Payload.(error)
	}
	nu := a.Payload.(*Tag)
	a = nu.Ask(nu, M_Create, NK_Group, "GUI")
	if a.Kind == M_Error {
		return a.Payload.(error)
	}
	slog.Info("SIGNUP created a new user.", "name", nm)
	return nil
}

func AuthUser(nm, pwd string) (uid NodeID, admin bool, err error) {
	a := Tree.Sys.Users.Ask(SystemUser, M_Get_Auth, nm, pwd)
	if a.Kind == M_Error {
		return 0, false, a.Payload.(error)
	}
	ut := a.Payload.(*Tag)
	return ut.ID, ut.Admin, nil
}
