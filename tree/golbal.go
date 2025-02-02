package tree

type E struct{}

type DC chan E

type Pipe chan *Msg

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

	NK_Noop
)

// TODO: We need to rename this to NNames.
var NKNames = map[NK]string{
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

var oka = &Msg{Kind: M_OK}

const (
	M_OK MK = iota
	M_Error

	M_Create
	M_Rename
	M_Delete

	M_Get_Parms
	M_Get_Auth
	M_Get_Tree
	M_Get_Copy
	M_Get_Display
	M_Get_Child

	M_Update_Parms
	M_Update_GUI
	M_Update_Tree

	M_Subscribe
	M_Unsubscribe

	M_Stop

	MK_UpdatePath
	MK_InitGUI
)

// MNames is exported for logging purposes.
var MKNames = map[MK]string{
	M_OK:    "OK",
	M_Error: "Error",

	M_Create: "Create",
	M_Rename: "Rename",
	M_Delete: "Delete",

	M_Get_Parms:   "Get_Parms",
	M_Get_Auth:    "Get_Auth",
	M_Get_Tree:    "Get_Tree",
	M_Get_Copy:    "Get_Copy",
	M_Get_Display: "Get_Display",
	M_Get_Child:   "Get_Child",

	M_Update_Parms: "Update_Parms",
	M_Update_GUI:   "Update_GUI",
	M_Update_Tree:  "Update_Tree",

	M_Subscribe:   "Subscribe",
	M_Unsubscribe: "Unsubscribe",

	M_Stop: "Stop",

	MK_UpdatePath: "UpdatePath",
	MK_InitGUI:    "InitGUI",
}
