interface socketMessage {
  Kind: number
  OTP: string
  GUIID: number
  NodeID: number
  NodeName: string
  NodeKind: number
  NodeParentID: number
}

class WSManager {
  private static instance: WSManager

  private constructor() {}

  static it(): WSManager {
    if (!WSManager.instance) {
      WSManager.instance = new WSManager()
    }
    return WSManager.instance
  }
}
