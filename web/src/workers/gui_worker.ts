import {GuiMessage, guiMessageType} from "../shared/gui_types.js"

// Worker globals
let state: any = null
let counter = 0

// Populate GUI Message handlers
let guiMessageHandlers: {[key: string]: ((msg: GuiMessage) => Promise<GuiMessage>)} = {}
guiMessageHandlers[guiMessageType.getUserTree] = getUserStateHandler

addEventListener("connect", connectHandler)
console.log("Added event listener.")

// Connect to GUI
function connectHandler(e: Event) {
  console.log("Inside web worker.")
  const port = (<MessageEvent>e).ports[0]
  counter += 1

  port.onmessage = function (e) {
    let gm = e.data as GuiMessage
    guiMessageHandlers[gm.Type](gm)
      .then((resp) => {
        console.log("Worker response:", resp)
        port.postMessage(resp)
      })
      .catch((e) => {console.log(e)})
    }
}

// GUI Message handlers
async function getUserStateHandler(msg: GuiMessage): Promise<GuiMessage> {
  if (state === null) {
    return fetch("/api/state?" + new URLSearchParams({userid: msg.Payload}), {method: "get"})
      .then((resp) => resp.json())
      .then((data) => {
        console.log("Data:", data)
        state = data
        return {Type: guiMessageType.userTree, Payload: data}
      })
  }
  console.log("Returning from state var.")
  return Promise.resolve({Type: guiMessageType.userTree, Payload: state})
}