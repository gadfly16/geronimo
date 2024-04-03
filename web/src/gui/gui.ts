import {GuiMessage, guiMessageType} from "../shared/gui_types.js"

function buildTree(treeNode: any, path: string): any {
  let item: HTMLDetailsElement | HTMLLIElement
  path = path + "/" + treeNode.Name
  if ("children" in treeNode) {
    item = document.createElement("details") as HTMLDetailsElement
    item.open = true
    let summary = document.createElement("summary")
    summary.appendChild(document.createTextNode(treeNode.Name))
    summary.setAttribute("data-path", path)
    item.appendChild(summary)
    let children = document.createElement("ul")
    for (let ch of treeNode.children) {
      children.appendChild(buildTree(ch, path))
    }
    item.appendChild(children)
  } else {
    item = document.createElement("li")
    item.setAttribute("data-path", path)
    item.appendChild(document.createTextNode(treeNode.Name))  
  }
  return item
}

function loadDetail() {
  let path = location.href.split("/gui/").at(-1)
  console.log(path)
  fetch("/api/detail/" + path)
    .then((resp) => {
      return resp.json()
    })
    .then((data) => {
      console.log(data)
    })
    .catch((e) => {
      alert(e) 
    })
}

function treeClick(e: Event) {
  let target = e.target as HTMLDetailsElement | HTMLLIElement
  if ((e as MouseEvent).offsetX > target.offsetHeight) {
    e.preventDefault()
  }
  let path = target.getAttribute("data-path")
  let current = new URL(location.href)
  let dest = current.origin + "/gui" + path
  if (dest != current.href) {
    window.history.pushState({}, "", dest)
    loadDetail()
  }
}

function getUserTree(userID: number) {
  fetch("/api/tree?" + new URLSearchParams({
    userid: userID.toString()
  })).then((resp) => {
      return resp.json()
  }).then((data) => {
    console.log(data)
    let treeRoot = document.querySelector("#tree")!
    treeRoot.appendChild(buildTree(data, ""))
  }).catch((e) => {
    alert(e) 
  })
}

window.onload = () => { 
  let userID = parseInt(document.getElementById("user-id")!.getAttribute("value")!)
  console.log("UserID: ", userID)

  let location = window.location.pathname
  console.log("Location: ", location)

  let gm: GuiMessage = {
    Type: guiMessageType.getUserTree,
    Payload: userID
  } 

  document.querySelector("#tree")?.addEventListener("click", treeClick)

  getUserTree(userID)
}