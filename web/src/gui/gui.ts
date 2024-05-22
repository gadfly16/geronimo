import {GuiMessage, guiMessageType, NodeType, NodeTypeName} from "../shared/gui_types.js"

// UI Globals
let gui: GUI

// This is not jQuery, but a helper function to turn a html string into a HTMLElement
let _dollarRegexp = /^\s+|\s+$|(?<=\>)\s+(?=\<)/gm
function $(html: string): HTMLElement {
  const template = document.createElement('template');
  template.innerHTML = html.replace(_dollarRegexp,'');
  const result = template.content.firstElementChild;
  return result as HTMLElement;
}

class GUI {
  tree: Node | null = null
  htmlTreeView: HTMLElement
  htmlDisplayView: HTMLElement
  selection = new Map<number, Node>()
  nodes = new Map<string,Node>()

  constructor(rootNodeID: number) {
    this.htmlTreeView = document.querySelector("#tree-view")!
    this.htmlDisplayView = document.querySelector("#display-view")!
    this.loadTree(rootNodeID)
    this.htmlTreeView.addEventListener("click", this.treeClick.bind(this))
  }

  addNode(node: Node) {
    this.nodes.set(node.ID.toString(), node)
  }

  loadTree(nodeID: number) {
    fetch("/api/tree?" + new URLSearchParams({
      userid: nodeID.toString()
    })).then((resp) => {
        return resp.json()
    }).then((treeData) => {
      this.tree = new Node(treeData)
      this.htmlTreeView.appendChild(this.tree.renderTree())
      this.updateSelection()
    }).catch((e) => {
      alert(e)
    })
  }

  treeClick(e: MouseEvent) {
    let target = e.target as HTMLElement
    if ((e as MouseEvent).offsetX > target.offsetHeight) {
      e.preventDefault()
    }
    let nid = target.getAttribute("data-id")!
    let loc = new URL(location.href)
    let selection = loc.searchParams.getAll("select")
    if (e.ctrlKey) {
      if (selection.includes(nid)) {
        loc.searchParams.delete("select", nid)
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
    this.updateSelection()
  }

  updateSelection() {
    let newSelection = new URL(location.href).searchParams.getAll("select")
    this.selection.forEach((node, id) => {
      if (!newSelection.includes(node.ID.toString())) {
        this.selection.delete(id)
        node.deselect()
      }
    })
    for (let nid of newSelection) {
      const id = parseInt(nid)
      if (!Number.isNaN(id)) {
        if (!(this.selection.has(id))) {
          let node = this.nodes.get(nid)
          if (node != undefined) {
            this.selection.set(id, node)
            node.select()
          }
        }
      }
    }
  }
}

class Node {
  ID: number = 0
  Name: string = ""
  DetailType: number = 0
  ParentID: number = 0
  htmlTreeElem: HTMLElement | null = null
  htmlDisplayElem: HTMLElement | null = null
  display: NodeDisplay | null = null

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
    gui.addNode(this)
  }

  renderTree(): HTMLElement {
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
        ule.appendChild(n.renderTree())
      }
    } else {
      e = $(`
        <div>
          <li data-id="${this.ID}">${this.Name}</li>
        </div>
      `)
    }
    this.htmlTreeElem = e
    return e
  }

  select() {
    this.htmlTreeElem!.classList.add("selected")
    fetch(`/api/display?select=${this.ID.toString()}`)
    .then((resp) => {
      return resp.json()
    })
    .then((displayDataList) => {
      if (displayDataList.error) throw new Error(displayDataList.error)
      for (let dd of displayDataList) {
        switch (dd.DetailType) {
          case NodeType.Broker:
            this.display = new BrokerDisplay(dd)
            break
          case NodeType.Account:
            this.display = new AccountDisplay(dd)
            break
          case NodeType.User:
            this.display = new UserDisplay(dd)
            break
        }
      }
      gui.htmlDisplayView.appendChild(this.display!.render())
    })
    .catch((e: Error) => {
      if (e.message == "unauthorized") {
        window.location.replace("/login" + new URL(location.href).search)
      }
      alert(e + " at line: " + (e as any).lineNumber) 
    })
  }

  deselect() {
    this.htmlTreeElem!.classList.remove("selected")
    gui.htmlDisplayView.removeChild(this.display!.htmlDisplay!)
    this.display = null
  }
}

// class Tree {
//   root: Node
//   htmlRoot: HTMLElement
//   nodes: {}

//   constructor() {
//     this.root = new Node()
//     this.nodes = {0: this.root}
//     this.htmlRoot = document.querySelector("#tree")!
//     this.htmlRoot.addEventListener("click", this.click)
//   }

// }

class DisplayBox {
  DisplayList: NodeDisplay[] = []
  displayBox = document.getElementById("displayBox") as HTMLDivElement

  update() {
    fetch("/api/display" + new URL(location.href).search)
    .then((resp) => {
      return resp.json()
    })
    .then((displayDataList) => {
      if (displayDataList.error) throw new Error(displayDataList.error)
      this.DisplayList = []
      for (let dd of displayDataList) {
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
      // tree.updateSelection()
    })
    .catch((e: Error) => {
      if (e.message == "unauthorized") {
        window.location.replace("/login" + new URL(location.href).search)
      }
      alert(e + " at line: " + (e as any).lineNumber) 
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
  name: string
  detailType: number
  id: number
  path: string
  parameters: ParameterForm | null = null
  infos: InfoList | null = null
  htmlDisplay: HTMLElement | null = null
  
  constructor(displayData: any) {
    this.name = displayData.Name
    this.detailType = displayData.DetailType
    this.id = displayData.ID
    this.path = displayData.Path
  }

  render():HTMLElement {
    let disp = this.renderHead()
    if (this.parameters) disp.appendChild(this.parameters.render())
    if (this.infos) disp.appendChild(this.infos.render())
    this.htmlDisplay = disp
    return disp
  }

  renderHead():HTMLElement {
    let dispHead = $(`
      <div class="display">
        <div class="displayHead">
          <div class="displayName ${NodeTypeName[this.detailType]}">${this.name}</div>
          <div class="displayPath">${this.path}</div>
        </div>
      </div>
    `)
    return dispHead
  }
}

class UserDisplay extends NodeDisplay{
  constructor(displayData: any) {
    super(displayData)
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    this.infos = new InfoList
    this.infos.add(parmDict, ["Last Modified"])
  }
}

class BrokerDisplay extends NodeDisplay{
  constructor(displayData: any) {
    super(displayData)
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    this.parameters = new ParameterForm(this)
    this.parameters.add(parmDict, ["Pair", "Base", "Quote", "LowLimit", "HighLimit", "Delta", "MinWait", "MaxWait", "Offset"])
    this.infos = new InfoList
    this.infos.add(parmDict, ["Fee", "Last Modified"])
  }
}

class AccountDisplay extends NodeDisplay{
  constructor(displayData: any) {
    super(displayData)
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    this.parameters = new ParameterForm(this)
    this.parameters.add(parmDict, ["Exchange"])
    this.infos = new InfoList
    this.infos.add(parmDict, ["Last Modified"])
  }
}

class ParameterForm {
  ParameterList: Parameter[] = []
  formElem: HTMLFormElement | null = null
  submitButton: HTMLElement | null = null
  nodeDisplay: NodeDisplay

  constructor(nodeDisplay: NodeDisplay) {
    this.nodeDisplay = nodeDisplay
  }

  add(parmDict: any, parmList: string[] = []) {
    if (!parmList.length) {
      parmList = Object.keys(parmDict)
    }
    parmList.forEach(k => {
      this.ParameterList.push(new Parameter(k, parmDict[k], this))
    })
  }

  submit(event: SubmitEvent) {
    event.preventDefault()
    const data = new FormData(event.target as HTMLFormElement)

    const detail: {[k: string]: any} = {}
    for (const parm of this.ParameterList) {
      const value = data.get(parm.name)
      detail[parm.name] = parm.inputType == "number" ? Number(value) : value
    }

    const apiUpdatePath = `/api/update/${NodeTypeName[this.nodeDisplay.detailType]}`
    console.log(apiUpdatePath)
    const msg = {
      Type: "Update",
      Path: this.nodeDisplay.path,
      Payload: detail
    }
    console.log(msg)
    fetch(apiUpdatePath, {
      method: 'post',
      body: JSON.stringify(msg),
      mode: 'same-origin',
    }).then((response) => {
      if (response.ok) {
        console.log(response)
      } else {
        throw 'failed'
      }
    }).catch((e) => { alert(e) })
    return false
  }

  render():HTMLElement {
    this.formElem = $(`
      <form class="parameterForm">
        <div class="parameterFormHeadBox">
            <div class="parameterFormTitle">Parameters:</div>
            <button class="parameterFormSubmit">Submit Parameters</button>
        </div>        
      </form>
    `) as HTMLFormElement
    this.formElem.addEventListener("submit", this.submit.bind(this))
    this.submitButton = this.formElem.querySelector(".parameterFormSubmit")!
    for (let parm of this.ParameterList) {
      this.formElem.appendChild(parm.render())
    }
    return this.formElem
  }

  checkDifferences() {
    for (const parm of this.ParameterList) {
      if (parm.isDifferent) {
        this.formElem?.classList.add("different")
        return
      }
    }
    this.formElem?.classList.remove("different")
  }
}

class Parameter {
  name: string
  value: number|string
  origValue: number|string
  inputType: string
  parmForm: ParameterForm
  isDifferent: boolean
  elem: HTMLElement | null = null

  constructor(name: string, value: number|string, parmForm: ParameterForm) {
    this.name = name
    this.value = value
    this.origValue = value
    this.inputType = typeof this.value == "string" ? "text" : "number"
    this.parmForm = parmForm
    this.isDifferent = false
  }

  render():HTMLElement {
    this.elem = $(`
      <div class="inputBox">
        <label for="${this.name} class="settingLabel">${this.name}</label>
        <input
          name="${this.name}"
          class="settingInput"
          type="${this.inputType}"
          ${this.inputType == "number" ? `step="any"` : ``}
          value="${this.value}"
        />
      </div>
    `)
    this.elem.querySelector("input")?.addEventListener("change", this.valueChange.bind(this))
    return this.elem
  }

  valueChange(event: Event) {
    const target = event.target as HTMLInputElement
    this.value = target.value
    this.isDifferent = (this.value != this.origValue)
    if (this.isDifferent) {
      this.elem?.classList.add("different")
    } else {
      this.elem?.classList.remove("different")
    }
    this.parmForm.checkDifferences()
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
        <span class="infoValue"><b>${this.Value}</b></span>
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

  // Select user node in URL if nothing else is selected
  let loc = new URL(location.href)
  let selection = loc.searchParams.getAll("select")
  if (!selection.length) {
    loc.searchParams.append("select", userID.toString())
    window.history.pushState({}, "", loc)
  }

  gui = new GUI(userID)

  // Update display if location URL changes
  window.addEventListener("popstate", (event) => {
    gui.updateSelection()
  })
}