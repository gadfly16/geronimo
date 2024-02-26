export interface GuiMessage {
    Type: string,
    Payload: any
  }

export let guiMessageType = {
    getUserTree: "getUserTree",
    userTree: "userTree"
}