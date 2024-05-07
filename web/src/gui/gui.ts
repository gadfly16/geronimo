import {GuiMessage, guiMessageType, NodeType, NodeTypeName} from "../shared/gui_types.js"

// UI Globals
let tree: Tree
let display: Display

// This is not jQuery, but a helper function to turn a html string into a HTMLElement
let _dollarRegexp = /^\s+|\s+$|(?<=\>)\s+(?=\<)/gm
function $(html: string): HTMLElement {
  const template = document.createElement('template');
  template.innerHTML = html.replace(_dollarRegexp,'');
  const result = template.content.firstElementChild;
  return result as HTMLElement;
}

class Node {
  ID: number = 0
  Name: string = ""
  DetailType: number = 0
  ParentID: number = 0

  children: Node[] = []

  constructor(nodeData: any = null, parentID: number = 0) {
    if (nodeData == null) return
    this.ID = nodeData.ID
    this.Name = nodeData.Name
    this.DetailType = nodeData.DetailType
    if ("children" in nodeData) {
      nodeData.children.forEach((e: any) => {
        this.children.push(new Node(e, this.ID))
      });
    }
    // console.log("Node object: ", this)
  }

  render(): HTMLElement {
    let e:HTMLElement
    if (this.children.length) {
      e = $(`
        <details open="true">
          <summary data-id="${this.ID}">${this.Name}</summary>
          <ul></ul>
        </details>
      `)
      let ule = e.querySelector("ul")!
      for (let n of this.children) {
        ule.appendChild(n.render())
      }
    } else {
      e = $(` <li data-id="${this.ID}">${this.Name}</li> `)
    }
    return e
  }
}

class Tree {
  root: Node
  htmlRoot: HTMLElement

  constructor() {
    this.root = new Node()
    this.htmlRoot = document.querySelector("#tree")!
    this.htmlRoot.addEventListener("click", this.click)
  }

  update(treeData: any) {
    this.root = new Node(treeData)
    this.htmlRoot.appendChild(this.root.render())
  }

  fetch(nodeID: number) {
    fetch("/api/tree?" + new URLSearchParams({
      userid: nodeID.toString()
    })).then((resp) => {
        return resp.json()
    }).then((treeData) => {
      this.update(treeData)
    }).catch((e) => {
      alert(e) 
    })
  }

  click(e: MouseEvent) {
    let target = e.target as HTMLElement
    if ((e as MouseEvent).offsetX > target.offsetHeight) {
      e.preventDefault()
    }
    let nid = target.getAttribute("data-id")!
    let loc = new URL(location.href)
    let selection = loc.searchParams.getAll("select")
    if (e.shiftKey) {
      if (selection.includes(nid)) {
        loc.searchParams.delete("select", nid)
        // target.classList.remove("selected")
      } else {
        loc.searchParams.append("select", nid)
      }
    } else {
      if (selection.includes(nid) && selection.length == 1) {
        return
      } else {
        loc.searchParams.delete("select")
        loc.searchParams.append("select", nid)
      }
    }
    window.history.pushState({}, "", loc)
    display.update()
  }

  updateSelection() {
    let selElems = this.htmlRoot.querySelectorAll(".selected") as NodeListOf<HTMLElement>
    let loc = new URL(location.href)
    let selection = loc.searchParams.getAll("select")
    for (let elem of selElems) {
      let nid = elem.getAttribute("selected")!
      if (!selection.includes(nid)) {
        elem.classList.remove("selected")
      }
    }
    for (let nid of selection) {
      let elem = this.htmlRoot.querySelector(`[data-id="${nid}"`)
      if (elem) {
        if (!elem.classList.contains("selected")) {
          elem.classList.add("selected")
        }
      }
    }
  }
}

class Display {
  DisplayList: NodeDisplay[] = []
  displayBox = document.getElementById("displayBox") as HTMLDivElement

  update() {
    fetch("/api/display" + new URL(location.href).search)
    .then((resp) => {
      return resp.json()
    })
    .then((displayDataList) => {
      if (displayDataList.Error) throw new Error(displayDataList.Error)
      this.DisplayList = []
      for (let dd of displayDataList) {
        console.log(dd)
        switch (dd.DetailType) {
          case NodeType.Broker:
            this.DisplayList.push(new BrokerDisplay(dd))
            break
          case NodeType.Account:
            this.DisplayList.push(new AccountDisplay(dd))
            break
          case NodeType.User:
            this.DisplayList.push(new UserDisplay(dd))
            break
        }
      }
      this.draw()
      tree.updateSelection()
    })
    .catch((e) => {
      alert(e + " at line: " + e.lineNumber) 
    })
  }

  draw() {
    this.displayBox.textContent = ""
    for (let disp of this.DisplayList) {
      this.displayBox.appendChild(disp.render())
    } 
  }
}

class NodeDisplay {
  Name: string = ""
  DetailType: number = 0
  path: string = ""
  Parameters: ParameterForm | null = null
  Infos: InfoList | null = null
  
  constructor(displayData: any) {
    this.Name = displayData.Name
    this.DetailType = displayData.DetailType
  }

  render():HTMLElement {
    let disp = this.renderHead()
    if (this.Parameters) disp.appendChild(this.Parameters.render())
    if (this.Infos) disp.appendChild(this.Infos.render())
    return disp
  }

  renderHead():HTMLElement {
    let disp = $(`
      <div class="display">
        <div class="displayHead">
          <div class="displayName ${NodeTypeName[this.DetailType]}">${this.Name}</div>
          <div class="displayPath">${this.path}</div>
        </div>
      </div>
    `)
    return disp
  }
}

class UserDisplay extends NodeDisplay{
  constructor(displayData: any) {
    super(displayData)
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    this.Infos = new InfoList
    this.Infos.add(parmDict, ["Last Modified"])
  }
}

class BrokerDisplay extends NodeDisplay{
  constructor(displayData: any) {
    super(displayData)
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    this.Parameters = new ParameterForm
    this.Parameters.add(parmDict, ["Pair", "Base", "Quote", "LowLimit", "HighLimit", "Delta", "MinWait", "MaxWait", "Offset"])
    this.Infos = new InfoList
    this.Infos.add(parmDict, ["Fee", "Last Modified"])
  }
}

class AccountDisplay extends NodeDisplay{
  constructor(displayData: any) {
    super(displayData)
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    this.Parameters = new ParameterForm
    this.Parameters.add(parmDict, ["Exchange"])
    this.Infos = new InfoList
    this.Infos.add(parmDict, ["Last Modified"])
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

  render():HTMLElement {
    let elem = $(`
      <form class="parameterForm">
        <div class="parameterFormHeadBox">
            <div class="parameterFormTitle">Parameters:</div>
            <div class="parameterFormSubmit">Submit</div>
        </div>        
      </form>
    `)
    for (let parm of this.ParameterList) {
      elem.appendChild(parm.render())
    }
    return elem
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

  render():HTMLElement {
    let elem = $(`
      <div class="inputBox">
        <label for="${this.Name} class="settingLabel">${this.Name}</label>
        <input
          name="${this.Name}"
          class="settingInput"
          type="${this.InputType}"
          value="${this.Value}"
        />
      </div>
    `)
    return elem
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

  render():HTMLElement {
    let elem = $(`
      <div class="infoListBox">
        <div class="infoListHead">Info:</div>
      </div>
    `)
    for (let info of this.InfoList) {
      elem.appendChild(info.render())
    }
    return elem
  }
}

class Info {
  Name = ""
  Value: number|string = 0

  constructor(name: string, value: number|string) {
    this.Name = name
    this.Value = value
  }

  render():HTMLElement {
    let elem = $(`
      <div class="infoBox">
        <span class="infoName">${this.Name}:</span>
        <span class="infoValue">${this.Value}</span>
      </div>
    `)
    return elem
  }
}

function displayTemplate(dd: any): string {
  return ""
}

window.onload = () => { 
  let userID = parseInt(document.getElementById("user-id")!.getAttribute("value")!)
  console.log("UserID: ", userID)

  let gm: GuiMessage = {
    Type: guiMessageType.getUserTree,
    Payload: userID
  } 

  tree = new Tree()
  tree.fetch(userID)

  display = new Display()

  window.addEventListener("popstate", (event) => {
    display.update()
  })

  let loc = new URL(location.href)
  let selection = loc.searchParams.getAll("select")
  if (!selection.length) {
    loc.searchParams.append("select", userID.toString())
    window.history.pushState({}, "", loc)
  }
  display.update()
}