import {GuiMessage, guiMessageType, NodeType, NodeTypeName} from "../shared/gui_types.js"

class Node {
  ID: number = 0
  Name: string = ""
  DetailType: number = 0
  ParentID: number = 0
}

class NodeTree {
  Root: Node = new Node
}

class NodeDisplay {
  Name: string = ""
  DetailType: number = 0
  path: string = ""
  Display: null | BrokerDisplay | AccountDisplay = null
  
  constructor(displayData: any) {
    this.Name = displayData.Name
    this.DetailType = displayData.DetailType
    switch (this.DetailType) {
      case NodeType.Broker:
        this.Display = new BrokerDisplay(displayData)
        break
      case NodeType.Account:
        this.Display = new AccountDisplay(displayData)
        break
      case NodeType.User:
        this.Display = new UserDisplay(displayData)
        break
    }
    console.log("Display object: ", this)
  }

  render():string {
    let html = `
      <div class="display">
        <div class="displayHead">
          <div class="displayName ${NodeTypeName[this.DetailType]}">${this.Name}</div>
          <div class="displayPath">${this.path}</div>
        </div>
        ${this.Display!.render()}
      </div>`
    
    return html
  }
}

class UserDisplay {
  Parameters = new ParameterForm
  InfoList = new InfoList
  constructor(displayData: any) {
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    // this.Parameters.add(parmDict, ["Exchange"])
    this.InfoList.add(parmDict, ["Last Modified"])
  }

  render(): string {
    let html = this.Parameters.render()
    html += this.InfoList.render()
    return html
  }
}

class BrokerDisplay {
  Parameters = new ParameterForm
  InfoList = new InfoList
  constructor(displayData: any) {
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    this.Parameters.add(parmDict, ["Pair", "Base", "Quote", "LowLimit", "HighLimit", "Delta", "MinWait", "MaxWait", "Offset"])
    this.InfoList.add(parmDict, ["Fee", "Last Modified"])
  }

  render(): string {
    let html = this.Parameters.render()
    html += this.InfoList.render()
    return html
  }
}

class AccountDisplay {
  Parameters = new ParameterForm
  InfoList = new InfoList
  constructor(displayData: any) {
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    this.Parameters.add(parmDict, ["Exchange"])
    this.InfoList.add(parmDict, ["Last Modified"])
  }

  render(): string {
    let html = this.Parameters.render()
    html += this.InfoList.render()
    return html
  }
}

class ParameterForm {
  ParameterList: Parameter[] = []

  add(parmDict: any, parmList: string[] = []) {
    if (!parmList.length) {
      parmList = Object.keys(parmDict)
    }
    parmList.forEach(k => {
      this.ParameterList.push(new Parameter(k, parmDict[k]))
    })
  }

  render():string {
    let html = `
      ${this.ParameterList.length ? `
      <form class="parameterForm">
        <div class="parameterFormHeadBox">
            <div class="parameterFormTitle">Parameters:</div>
            <div class="parameterFormSubmit">Submit</div>
        </div>        
        ${this.ParameterList.reduce((a,s) => a+s.render(),"")}
      </form>
      ` : ""}
    `
    return html
  }
}

class Parameter {
  Name = ""
  Value: number|string = 0
  InputType = ""

  constructor(name: string, value: number|string) {
    this.Name = name
    this.Value = value
    this.InputType = typeof this.Value == "string" ? "text" : "number"
  }

  render():string {
    let html = `
      <div class="inputBox">
        <label for="${this.Name} class="settingLabel">${this.Name}</label>
        <input
          name="${this.Name}"
          class="settingInput"
          type="${this.InputType}"
          value="${this.Value}"
        />
      </div>`
    return html
  }
}

class InfoList {
  InfoList: Info[] = []

  add(parmDict: any, parmList: string[] = []) {
    if (!parmList.length) {
      parmList = Object.keys(parmDict)
    }
    parmList.forEach(k => {
      this.InfoList.push(new Info(k, parmDict[k]))
    })
  }

  render():string {
    let html = `
      ${this.InfoList.length ? `
      <div class="infoListBox">
      <div class="infoListHead">Info:</div>
        ${this.InfoList.reduce((a,s) => a+s.render(),"")}
      </div>
      ` : ""}
    `
    return html
  }
}

class Info {
  Name = ""
  Value: number|string = 0

  constructor(name: string, value: number|string) {
    this.Name = name
    this.Value = value
  }

  render():string {
    let html = `
      <div class="infoBox">
        <span class="infoName">${this.Name}:</span>
        <span class="infoValue">${this.Value}</span>
      </div>`
    return html
  }
}

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
      let display = new NodeDisplay(data)
      display.path = path!

      const displayBox = document.getElementById("displayBox") as HTMLDivElement
      displayBox.innerHTML = display.render()
    })
    .catch((e) => {
      alert(e) 
    })
}

function displayTemplate(dd: any): string {
  return ""
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
  }).then((treeData) => {
    console.log(treeData)
    let treeRoot = document.querySelector("#tree")!
    treeRoot.appendChild(buildTree(treeData, ""))
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