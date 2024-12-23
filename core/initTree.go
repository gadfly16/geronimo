package core

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
	r := Tree.Sys.Root.Ask(
		Msg{
			Kind: CreateMsgKind,
			Payload: &GroupNode{
				Head: &Head{
					Name: "Users",
					Kind: GroupKind,
				},
			},
		})
	if r.Kind == ErrorMsgKind {
		return errors.New("init: user group creation failed")
	}
	r = Tree.Sys.Root.Ask(
		Msg{
			Kind: CreateMsgKind,
			Payload: &GroupNode{
				Head: &Head{
					Name: "System",
					Kind: GroupKind,
				},
			},
		})
	if r.Kind == ErrorMsgKind {
		return errors.New("init: system group creation failed")
	}
	return nil
}
