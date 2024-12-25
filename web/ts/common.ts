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
  Create,
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
}

export enum nodeKinds {
  Root = 0,
  Group,
  User,
  Account,
  Trader,
  TreeUpdater,
}

export let nodeKindName = ["root", "group", "user", "account", "broker", "tree updater"]
