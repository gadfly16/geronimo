export interface GuiMessage {
    Type: string,
    Payload: any
  }

export let guiMessageType = {
    getUserTree: "getUserTree",
    userTree: "userTree"
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