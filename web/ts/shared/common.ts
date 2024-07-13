export interface msg {
  Kind: Number,
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
}

export let WSMsg = {
    Credentials: "Credentials",
    Subscribe: "Subscribe",
    Unsubscribe: "Unsubscribe",
    Update: "Update",
  }

export enum nodeKinds {
	Root = 0,
  Group,
	User,
  Broker,
  Account,
}

export enum payloadKinds {
  UserNodePayload = 0,
}

export let NodeTypeName = ["root", "user", "account", "broker", "group", "pocket"]