import {WSMsg, nodeKinds, NodeTypeName} from "../shared/common.js"

interface socketMessage {
  Type: string,
  OTP: string,
  GUIID: number,
  NodeID: number,
}

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
  nodes = new Map<string,Node>()

  htmlTreeView: HTMLElement
  htmlDisplayView: HTMLElement
  selection = new Map<number, Node>()

  socket: WebSocket
  guiID: number = 0
  guiOTP: string = ""

  constructor(rootNodeID: number) {
    this.htmlTreeView = document.querySelector("#tree-view")!
    this.htmlDisplayView = document.querySelector("#display-view")!
    this.loadTree(rootNodeID)
    this.htmlTreeView.addEventListener("click", this.treeClick.bind(this))
    this.socket = new WebSocket("/socket")
    this.socket.onmessage = this.socketMessageHandler.bind(this)
  }

  socketMessageHandler(event: MessageEvent) {
    const msg = JSON.parse(event.data) as socketMessage
    switch (msg.Type) {
      case WSMsg.Credentials:
        this.guiID = msg.GUIID
        this.guiOTP = msg.OTP
        this.socket.send(JSON.stringify(msg))
        break
      case WSMsg.Update:
        console.log(`Update needed for node id: ${msg.NodeID}`)
        let node = this.nodes.get(msg.NodeID.toString())
        if (node) {
          node.updateDisplay()
        } else {
          console.log(`Update requested for unknown node id: ${msg.NodeID}`)
        }
    }
  }

  subscribe(id: number) {
    this.socket.send(
      JSON.stringify({
        Type: WSMsg.Subscribe,
        GUIID: this.guiID,
        OTP: this.guiOTP,
        NodeID: id,
      })
    )
  }

  unsubscribe(id: number) {
    this.socket.send(
      JSON.stringify({
        Type: WSMsg.Unsubscribe,
        GUIID: this.guiID,
        OTP: this.guiOTP,
        NodeID: id,
      })
    )
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

  updateDisplay() {
    console.log(`Updating node ${this.ID}.`)
    fetch(`/api/display/${this.ID.toString()}`)
    .then((resp) => {
      return resp.json()
    })
    .then((displayData) => {
      if (displayData.error) throw new Error(displayData.error)
      this.display?.update(displayData)
    })
    .catch((e: Error) => {
      if (e.message == "unauthorized") {
        window.location.replace("/login" + new URL(location.href).search)
      }
      alert(e + " at line: " + (e as any).lineNumber) 
    })
  }
  
  select() {
    this.htmlTreeElem!.classList.add("selected")
    fetch(`/api/display/${this.ID.toString()}`)
    .then((resp) => {
      return resp.json()
    })
    .then((displayData) => {
      if (displayData.error) throw new Error(displayData.error)
        switch (displayData.DetailType) {
          case nodeKinds.Broker:
            this.display = new BrokerDisplay(displayData)
            break
          case nodeKinds.Account:
            this.display = new AccountDisplay(displayData)
            break
          case nodeKinds.User:
            this.display = new UserDisplay(displayData)
            break
        }
      gui.htmlDisplayView.appendChild(this.display!.render())
      gui.subscribe(this.ID)
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
    gui.unsubscribe(this.ID)
  }
}

class NodeDisplay {
  name: string
  detailType: number
  id: number
  path: string
  parms: ParameterForm | null = null
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
    if (this.parms) disp.appendChild(this.parms.render())
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

  update(displayData: any) {
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    if (this.infos) {
      this.infos.update(parmDict)
    }
    if (this.parms) {
      this.parms.update(parmDict)
    }
  }
}

class UserDisplay extends NodeDisplay{
  infoNames = ["Last Modified"]

  constructor(displayData: any) {
    super(displayData)
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    this.infos = new InfoList(parmDict, this.infoNames)
  }
}

class BrokerDisplay extends NodeDisplay{
  parmNames = ["Pair", "Base", "Quote", "LowLimit", "HighLimit", "Delta", "MinWait", "MaxWait", "Offset"]
  infoNames = ["Fee", "Last Modified"]

  constructor(displayData: any) {
    super(displayData)
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    this.parms = new ParameterForm(this, parmDict, this.parmNames)
    this.infos = new InfoList(parmDict, this.infoNames)
  }
}

class AccountDisplay extends NodeDisplay{
  parmNames = ["Exchange"]
  infoNames = ["Last Modified"]

  constructor(displayData: any) {
    super(displayData)
    let parmDict = displayData.Detail
    parmDict["Last Modified"] = parmDict.CreatedAt
    this.parms = new ParameterForm(this, parmDict, this.parmNames)
    this.infos = new InfoList(parmDict, this.infoNames)
  }
}

class ParameterForm {
  parms = new Map<string, Parameter>()
  htmlParmForm: HTMLFormElement | null = null
  submitButton: HTMLElement | null = null
  nodeDisplay: NodeDisplay

  constructor(nodeDisplay: NodeDisplay, parmDict: any, parmNames: string[] = []) {
    this.nodeDisplay = nodeDisplay
    if (!parmNames.length) {
      parmNames = Object.keys(parmDict)
    }
    for (const parmName of parmNames) {
      this.parms.set(parmName, new Parameter(parmName, parmDict[parmName], this))
    }
  }

  submit(event: SubmitEvent) {
    event.preventDefault()
    const formData = new FormData(event.target as HTMLFormElement)

    const detail: {[k: string]: any} = {}
    for (const [name, parm] of this.parms) {
      const value = formData.get(parm.name)
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
    })
    .then((response) => {
      if (response.ok) {
        const diffs = this.htmlParmForm!.querySelectorAll(".changed")
        for (const delem of diffs) {
          delem.classList.remove("changed")
        }
        this.htmlParmForm?.classList.remove("changed")
      } else {
        throw 'failed'
      }
    })
    .catch((e) => {
      alert(e)
    })
    return false
  }

  render():HTMLElement {
    this.htmlParmForm = $(`
      <form class="parameterForm">
        <div class="parameterFormHeadBox">
            <div class="parameterFormTitle">Parameters:</div>
            <button class="parameterFormSubmit">Submit Parameters</button>
        </div>        
      </form>
    `) as HTMLFormElement
    this.htmlParmForm.addEventListener("submit", this.submit.bind(this))
    this.submitButton = this.htmlParmForm.querySelector(".parameterFormSubmit")!
    for (const [name, parm] of this.parms) {
      this.htmlParmForm.appendChild(parm.render())
    }
    return this.htmlParmForm
  }

  update(parmDict: any) {
    for (const [name, parm] of this.parms) {
      const newValue = parmDict[name]
      parm.update(newValue)
    }
  }

  checkDifferences() {
    for (const [name, parm] of this.parms) {
      if (parm.changed) {
        this.htmlParmForm?.classList.add("changed")
        return
      }
    }
    this.htmlParmForm?.classList.remove("changed")
  }
}

class Parameter {
  name: string
  value: number|string
  origValue: number|string
  inputType: string
  parmForm: ParameterForm
  changed: boolean
  htmlParm: HTMLElement | null = null

  constructor(name: string, value: number|string, parmForm: ParameterForm) {
    this.name = name
    this.value = value
    this.origValue = value
    this.inputType = typeof this.value == "string" ? "text" : "number"
    this.parmForm = parmForm
    this.changed = false
  }

  render():HTMLElement {
    this.htmlParm = $(`
      <div class="parmBox">
        <label for="${this.name}" class="settingLabel">${this.name} </label>
        <input
          name="${this.name}"
          class="settingInput"
          type="${this.inputType}"
          ${this.inputType == "number" ? `step="any"` : ``}
          value="${this.value}"
        />
      </div>
    `)
    this.htmlParm.querySelector("input")?.addEventListener("change", this.valueChange.bind(this))
    return this.htmlParm
  }

  update(newValue: number|string) {
    if (this.value != newValue) {
      this.value = newValue
      const htmlInput = this.htmlParm?.querySelector("input")! 
      htmlInput.setAttribute("value", `${newValue}`)
      this.htmlParm?.classList.add("changeAlert")
    }
  }

  valueChange(event: Event) {
    const target = event.target as HTMLInputElement
    this.value = target.value
    this.changed = (this.value != this.origValue)
    if (this.changed) {
      this.htmlParm?.classList.add("changed")
    } else {
      this.htmlParm?.classList.remove("changed")
    }
    this.parmForm.checkDifferences()
  }
}

class InfoList {
  infos = new Map<string, Info>()

  constructor(parmDict: any, parmList: string[] = []) {
    if (!parmList.length) {
      parmList = Object.keys(parmDict)
    }
    for (const name of parmList) {
      this.infos.set(name, new Info(name, parmDict[name]))
    }
  }

  update (parmDict: any) {
    for (const [name, info] of this.infos) {
      const newValue = parmDict[name]
      info.update(newValue)
    }
  }

  render():HTMLElement {
    let elem = $(`
      <div class="infoListBox">
        <div class="infoListHead">Info:</div>
      </div>
    `)
    console.log(this.infos)
    for (const [name, info] of this.infos) {
      elem.appendChild(info.render())
    }
    return elem
  }
}

class Info {
  name = ""
  value: number|string = 0
  htmlInfo: HTMLElement|null = null

  constructor(name: string, value: number|string) {
    this.name = name
    this.value = value
  }

  render():HTMLElement {
    this.htmlInfo = $(`
      <div class="infoBox">
        <span class="infoName">${this.name}: </span>
        <span class="infoValue">${this.value}</span>
      </div>
    `)
    return this.htmlInfo
  }

  update(newValue: number|string) {
    if (this.value != newValue) {
      this.value = newValue
      const htmlValue = this.htmlInfo!.querySelector(".infoValue")! 
      htmlValue.textContent = `${newValue}`
      this.htmlInfo?.classList.add("changeAlert")
    }
  } 
}

function displayTemplate(dd: any): string {
  return ""
}

window.onload = () => { 
  let userID = parseInt(document.getElementById("user-id")!.getAttribute("value")!)
  console.log("UserID: ", userID)

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