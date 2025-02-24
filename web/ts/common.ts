export interface msg {
  Kind: number
  Payload: any
}

export enum msgKinds {
  OK = 0,
  Error,

  Create, // nk.NK, nm
  Rename,
  Delete,

  Get_Parms,
  Get_Auth,
  Get_Tree,
  Get_Display,
  Get_Child,

  Update_Parms,
  Update_GUI,
  Update_Tree,

  Subscribe,
  Unsubscribe,

  Stop,

  UpdatePath,
  InitGUI,
}

export enum WSMsg {
  Credentials = 0,
  Subscribe,
  Unsubscribe,
  Update,
  Error,
  ClientShutdown,
  Heartbeat,
  TreeNodeRename,
  TreeNodeCreate,
  TreeNodeDelete,
}

export enum nodeKinds {
  Root = 0,
  Group,
  User,
  Account,
  Trader,
  TreeUpdater,
  Users,
  GUI,
}

export let nodeKindName = [
  "root",
  "group",
  "user",
  "account",
  "broker",
  "tree updater",
  "users",
  "GUI",
]
