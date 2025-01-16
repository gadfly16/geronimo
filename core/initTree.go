package core

import (
	"errors"
	"log/slog"

	mk "github.com/gadfly16/geronimo/msgKinds"
)

func InitTree(sdb string, rp RootParms) (err error) {
	if err = initDb(sdb); err != nil {
		slog.Error("Failed to create db. Exiting.", "error", err.Error())
		return
	}
	if err = connectDB(sdb); err != nil {
		slog.Error("Failed to connect to db. Exiting.", "error", err.Error())
		return
	}

	if err = initRootNode(&rp); err != nil {
		slog.Error("Failed to create root node. Exiting.", "error", err.Error())
		return
	}
	r := Tree.Sys.Root.Ask(mk.CreateChild, SystemUser, UsersKind, "Users")
	if r.Kind == mk.Error {
		return errors.New("init: users creation failed")
	}
	r = Tree.Sys.Root.Ask(mk.CreateChild, SystemUser, GroupKind, "System")
	if r.Kind == mk.Error {
		return errors.New("init: system group creation failed")
	}
	return nil
}
