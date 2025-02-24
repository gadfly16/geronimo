package tree

import (
	"errors"
	"log/slog"
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
	r := Tree.Sys.Root.Ask(SystemUser, M_Create, NK_Users, "Users")
	if r.Kind == M_Error {
		return errors.New("init: users creation failed")
	}
	r = Tree.Sys.Root.Ask(SystemUser, M_Create, NK_Group, "System")
	if r.Kind == M_Error {
		return errors.New("init: system group creation failed")
	}

	if err = CloseDB(); err != nil {
		slog.Error("DB failed to close.", "err", err)
		return
	}
	return nil
}
