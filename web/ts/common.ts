export interface msg {
  Kind: number
  Payload: any
}

export enum msgKinds {
  OK = 0,
  Error,
  Stop,
  Stopped,
  Update,
  Parms,
  GetParms,
  CreateChild, // nk.NK, nm
  AuthUser,
  GetTree,
  Tree,
  GetCopy,
  GetDisplay,
  Display,
  Subscribe,
  Unsubscribe,
  NodeUpdate,
  Rename,
  TreeNodeRename,
  UpdatePath,
  RenameChild,
  CreateUser,
  SubscribeTree,
  UnsubscribeTree,
  TreeNodeCreate,
  GetChild,
  InitGUI,
  DeleteChild,
  TreeNodeDelete,
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
