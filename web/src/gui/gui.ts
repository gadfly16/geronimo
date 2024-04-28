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

function loadDisplay() {
  let path = location.href.split("/gui/").at(-1)
  fetch("/api/display/" + path)
    .then((resp) => {
      return resp.json()
    })
    .then((data) => {
      console.log("Display Data: ", data)
      let displayElement = ((document.querySelector("#displayTemplate") as HTMLTemplateElement).content.cloneNode(true) as DocumentFragment).querySelector(".display") as HTMLDivElement
      let displayName = displayElement.querySelector(".displayName")! as HTMLDivElement
      displayName.textContent = data.Name
      displayName.classList.add(data.Detail.Type)

      const settings = data.Detail.Settings
      let settingsElement = (document.querySelector("#settingsTemplate") as HTMLTemplateElement).content.cloneNode(true) as HTMLDivElement
      const settingFieldTemplate = document.querySelector("#settingFieldTemplate") as HTMLTemplateElement
      for (const s in settings) {
        let field = settingFieldTemplate.content.cloneNode(true) as HTMLDivElement
        let label = field.querySelector(".settingLabel")! as HTMLDivElement
        let input = field.querySelector(".settingInput")! as HTMLDivElement
        label.textContent = s
        switch (typeof settings[s]) {
          case "string":
            console.log("String value of ", s, settings[s])
            input.setAttribute("type", "text")
            input.setAttribute("value", settings[s])
            break
          case "number":
            console.log("Number value of ", s, settings[s])
            input.setAttribute("type", "text")
            input.setAttribute("value", settings[s].toString())
            break
          default:
            alert("Unknown setting type:" + s + " " + settings[s])
            break
        }
        settingsElement.appendChild(field)
      }
      displayElement.appendChild(settingsElement)

      const displayBox = document.querySelector("#displayBox") as HTMLDivElement
      displayBox.removeChild(displayBox.querySelector(".display")!)
      displayBox.appendChild(displayElement)
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
    loadDisplay()
  }
}

function getUserTree(userID: number) {
  fetch("/api/tree?" + new URLSearchParams({
    userid: userID.toString()
  })).then((resp) => {
      return resp.json()
  }).then((data) => {
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

  getUserTree(userID)
  document.querySelector("#tree")?.addEventListener("click", treeClick)

  window.addEventListener("popstate", (event) => {
    loadDisplay()
  })

  loadDisplay()
}