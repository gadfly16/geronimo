export interface GuiMessage {
    Type: string,
    Payload: any
  }

export let WSMsg = {
    Credentials: "Credentials",
    Subscribe: "Subscribe",
    Unsubscribe: "Unsubscribe",
    Update: "Update",
  }

export enum NodeType {
	Root = 0,
	User,
	Account,
	Broker,
  Group,
  Pocket
}

export let NodeTypeName = ["root", "user", "account", "broker", "group", "pocket"]