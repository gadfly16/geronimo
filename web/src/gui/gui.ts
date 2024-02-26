import {GuiMessage, guiMessageType} from "../shared/gui_types.js"

window.onload = () => {
  console.log("Starting webworker.")
  let gui_worker = new SharedWorker("/static/workers/gui_worker.js", {type: "module"})
  gui_worker.port.start()

  gui_worker.port.onmessage = (e) => {
    console.log("Message received from worker:", e.data);
  };

  let userID = parseInt(document.getElementById("user-id")!.getAttribute("value")!)
  console.log("UserID: ", userID)

  let location = window.location.pathname
  console.log("Location: ", location)

  let gm: GuiMessage = {
    Type: guiMessageType.getUserTree,
    Payload: userID
  } 

  gui_worker.port.postMessage(gm)
}