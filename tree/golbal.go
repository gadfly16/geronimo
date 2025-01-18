package tree

type NK int

const (
	NK_Root NK = iota
	NK_Group
	NK_User
	NK_Account
	NK_Trader
	NK_TreeUpdater
	NK_Users
	NK_GUI
)

var Names = map[NK]string{
	NK_Root:        "Root",
	NK_Group:       "Group",
	NK_User:        "User",
	NK_Account:     "Account",
	NK_Trader:      "Broker",
	NK_TreeUpdater: "TreeUpdater",
	NK_Users:       "Users",
	NK_GUI:         "GUI",
}

type MK = int

const (
	MK_OK MK = iota
	MK_Error
	MK_Stop
	MK_Stopped
	MK_Update
	MK_Parms
	MK_GetParms
	MK_CreateChild // PL: [k nk.NK, nm string]
	MK_AuthUser
	MK_GetTree
	MK_Tree
	MK_GetCopy
	MK_GetDisplay
	MK_Display
	MK_Subscribe
	MK_Unsubscribe
	MK_NodeUpdate
	MK_Rename
	MK_TreeNodeRename // PL: [t *Tag, nm string, ot *Tag]
	MK_UpdatePath
	MK_RenameChild // PL: [nm string, nnm string]
	MK_CreateUser
	MK_SubscribeTree
	MK_UnsubscribeTree
	MK_TreeNodeCreate // PL: [t *Tag, nm string, ot *Tag]
	MK_GetChild
	MK_InitGUI
	MK_DeleteChild
	MK_TreeNodeDelete
)

var MKNames = map[MK]string{
	MK_OK:              "OK",
	MK_Error:           "Error",
	MK_Stop:            "Stop",
	MK_Stopped:         "Stopped",
	MK_Update:          "Update",
	MK_Parms:           "Parms",
	MK_GetParms:        "GetParms",
	MK_CreateChild:     "CreateChild",
	MK_AuthUser:        "AuthUser",
	MK_GetTree:         "GetTree",
	MK_Tree:            "Tree",
	MK_GetCopy:         "GetCopy",
	MK_GetDisplay:      "GetDisplay",
	MK_Display:         "Display",
	MK_Subscribe:       "Subscribe",
	MK_Unsubscribe:     "Unsubscribe",
	MK_NodeUpdate:      "NodeUpdate",
	MK_Rename:          "Rename",
	MK_TreeNodeRename:  "TreeNodeRename",
	MK_UpdatePath:      "UpdatePath",
	MK_RenameChild:     "RenameChild",
	MK_CreateUser:      "CreateUser",
	MK_SubscribeTree:   "SubscribeTree",
	MK_UnsubscribeTree: "UnsubscribeTree",
	MK_TreeNodeCreate:  "TreeNodeCreate",
	MK_GetChild:        "GetChild",
	MK_InitGUI:         "InitGUI",
	MK_DeleteChild:     "DeleteChild",
	MK_TreeNodeDelete:  "TreeNodeDelete",
}
