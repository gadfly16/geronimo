export interface msg {
  Kind: number,
  Payload: any,
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
}

export enum WSMsg {
    Credentials = 0,
    Subscribe,
    Unsubscribe,
    Update,
    Error,
    ClientShutdown,
    Heartbeat,
  }

export enum nodeKinds {
	Root = 0,
  Group,
	User,
  Account,
  Broker,
}

// export enum payloadKinds {
//   Empty = 0,
//   UserNode,
//   Parms,
// }

export let NodeKindName = ["root", "group", "user", "broker", "account", "broker", "pocket"]