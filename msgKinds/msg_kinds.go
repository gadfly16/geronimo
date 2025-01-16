package msgkinds

type MK = int

const (
	OK MK = iota
	Error
	Stop
	Stopped
	Update
	Parms
	GetParms
	CreateChild
	AuthUser
	GetTree
	Tree
	GetCopy
	GetDisplay
	Display
	Subscribe
	Unsubscribe
	NodeUpdate
	Rename
	TreeNodeRename // PL: [t *Tag, nm string, ot *Tag]
	UpdatePath
	RenameChild
	CreateUser
	SubscribeTree
	UnsubscribeTree
	TreeNodeCreate // PL: [t *Tag, nm string, ot *Tag]
	GetChild
	InitGUI
	DeleteChild
	TreeNodeDelete
)

var Names = map[MK]string{
	OK:              "OK",
	Error:           "Error",
	Stop:            "Stop",
	Stopped:         "Stopped",
	Update:          "Update",
	Parms:           "Parms",
	GetParms:        "GetParms",
	CreateChild:     "CreateChild",
	AuthUser:        "AuthUser",
	GetTree:         "GetTree",
	Tree:            "Tree",
	GetCopy:         "GetCopy",
	GetDisplay:      "GetDisplay",
	Display:         "Display",
	Subscribe:       "Subscribe",
	Unsubscribe:     "Unsubscribe",
	NodeUpdate:      "NodeUpdate",
	Rename:          "Rename",
	TreeNodeRename:  "TreeNodeRename",
	UpdatePath:      "UpdatePath",
	RenameChild:     "RenameChild",
	CreateUser:      "CreateUser",
	SubscribeTree:   "SubscribeTree",
	UnsubscribeTree: "UnsubscribeTree",
	TreeNodeCreate:  "TreeNodeCreate",
	GetChild:        "GetChild",
	InitGUI:         "InitGUI",
	DeleteChild:     "DeleteChild",
	TreeNodeDelete:  "TreeNodeDelete",
}
