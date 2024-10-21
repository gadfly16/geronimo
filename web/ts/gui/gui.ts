import {WSMsg, nodeKinds, NodeKindName, msgKinds} from "../shared/common.js"

interface socketMessage {
  Kind: number,
  OTP: string,
  GUIID: number,
  NodeID: number,
}

function ask(mk: number, tid: number, pl: any, f: ((r: any)=>any)) {
  return fetch(`/api/msg/${mk}/${tid}`, {
    method: "post",
    mode: "same-origin",
    body: pl === null ? null : JSON.stringify(pl),
  })
  .then((resp) => {
    // console.log('Msg resonse: ', resp)
    if (!resp.ok) {
      if (resp.status === 401) {
        window.location.replace("/static/login.html" + new URL(location.href).search)
      }
      throw 'http error'
    }
    return resp.json()
  })
  .then(f)
  .catch((e: Error) => {
    console.log("Error:", e)
    alert(`${e} at line: ${(e as any).lineNumber}`) 
  })
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

let parmTypes = new Map<string, string>([
  ["string", "text"],
  ["number", "number"],
  ["boolean", "checkbox"],
])

class GUI {
  tree: Node | null = null
  nodes = new Map<string,Node>()

  htmlTreeView: HTMLElement
  htmlDisplayView: HTMLElement
  selection = new Map<number, Node>()

  socket: WebSocket
  guiID: number = 0
  guiOTP: string = ""
  heart: number = 0
  last_srv_beat: number = 0
  readonly heart_interval = 5000
  readonly roundtrip = 2000

  constructor(rootNodeID: number) {
    this.htmlTreeView = document.querySelector("#tree-view")!
    this.htmlDisplayView = document.querySelector("#display-view")!
    this.fetchTree(rootNodeID)
    this.htmlTreeView.addEventListener("click", this.treeClick.bind(this))
    this.socket = new WebSocket("/socket")
    this.socket.onmessage = this.socketMessageHandler.bind(this)
    this.heart = setInterval(this.heartbeat.bind(this), this.heart_interval)
  }

  heartbeat() {
    let off = Date.now() - this.last_srv_beat
    if (off > this.heart_interval + this.roundtrip && this.last_srv_beat != 0) {
      console.log("connection to server lost.", off)
    }
    // console.log("Sending heartbeat to gui")
    this.socket.send(
      JSON.stringify({
        Kind: WSMsg.Heartbeat,
        GUIID: this.guiID,
        OTP: this.guiOTP,
      })
    )
  }

  socketMessageHandler(event: MessageEvent) {
    const wsm = JSON.parse(event.data) as socketMessage
    console.log("Socket message received:", wsm)
    switch (wsm.Kind) {
      case WSMsg.Heartbeat:
        this.last_srv_beat = Date.now()
        break
      case WSMsg.Credentials:
        this.guiID = wsm.GUIID
        this.guiOTP = wsm.OTP
        this.socket.send(JSON.stringify(wsm))
        console.log(`Credentials received: guiid=${this.guiID}`)
        break
      case WSMsg.Update:
        console.log(`Update needed for node id: ${wsm.NodeID}`)
        let node = this.nodes.get(wsm.NodeID.toString())
        if (node) {
          node.updateDisplay()
        } else {
          console.log(`Update requested for unknown node id: ${wsm.NodeID}`)
        }
    }
  }

  subscribe(id: number) {
    this.socket.send(
      JSON.stringify({
        Kind: WSMsg.Subscribe,
        GUIID: this.guiID,
        OTP: this.guiOTP,
        NodeID: id,
      })
    )
  }

  unsubscribe(id: number) {
    this.socket.send(
      JSON.stringify({
        Kind: WSMsg.Unsubscribe,
        GUIID: this.guiID,
        OTP: this.guiOTP,
        NodeID: id,
      })
    )
  }

  addNode(node: Node) {
    this.nodes.set(node.ID.toString(), node)
  }

  fetchTree(nodeID: number) {
    ask(msgKinds.GetTree, nodeID, null,
    (treeData) => {
      this.tree = new Node(treeData)
      this.htmlTreeView.appendChild(this.tree.renderTree())
      this.updateSelection()
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
  Kind: number = 0
  ParentID: number = 0
  htmlTreeElem: HTMLElement | null = null
  htmlDisplayElem: HTMLElement | null = null
  display: NodeDisplay | null = null

  children: Node[] = []

  constructor(nodeData: any = null, parentID: number = 0) {
    if (nodeData == null) return
    this.ID = nodeData.ID
    this.Name = nodeData.Name
    this.Kind = nodeData.Kind
    if ("Children" in nodeData) {
      nodeData.Children.forEach((e: any) => {
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

    ask(msgKinds.GetDisplay, this.ID, null,
      (displayData) => {
        if (displayData.error) throw new Error(displayData.error)
        this.display?.update(displayData)
      })
  }
  
  select() {
    this.htmlTreeElem!.classList.add("selected")

    ask(msgKinds.GetDisplay, this.ID, null,
      (displayData) => {
        console.log(`Display data received:`, displayData)
        if (displayData.error) throw new Error(displayData.error)
          switch (displayData.Head.Kind) {
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
}

  deselect() {
    this.htmlTreeElem!.classList.remove("selected")
    gui.htmlDisplayView.removeChild(this.display!.htmlDisplay!)
    this.display = null
    gui.unsubscribe(this.ID)
  }
}

class NodeDisplay {
  ID: number
  name: string
  kind: number
  path: string
  parms: ParameterForm | null = null
  infos: InfoList | null = null
  htmlDisplay: HTMLElement | null = null
  
  constructor(displayData: any) {
    this.name = displayData.Head.Name
    this.kind = displayData.Head.Kind
    this.ID = displayData.Head.ID
    this.path = displayData.Head.Path
    this.parms = new ParameterForm(this, displayData)
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
          <div class="displayName ${NodeKindName[this.kind]}">${this.name}</div>
          <div class="displayPath">${this.path}</div>
        </div>
      </div>
    `)
    return dispHead
  }

  update(displayData: any) {
    if (this.parms) {
      this.parms.update(displayData.Parms)
    }
    // if (this.infos) {
      //   this.infos.update(parms)
      // }
  }
}

class UserDisplay extends NodeDisplay{
  // infoNames = ["Last Modified"]

  constructor(displayData: any) {
    super(displayData)
    // this.infos = new InfoList(parmDict, this.infoNames)
  }
}

class BrokerDisplay extends NodeDisplay{
  // infoNames = ["Fee", "Last Modified"]

  constructor(displayData: any) {
    super(displayData)
    // this.infos = new InfoList(parmDict, this.infoNames)
  }
}

class AccountDisplay extends NodeDisplay{
  // infoNames = ["Last Modified"]

  constructor(displayData: any) {
    super(displayData)
    // this.infos = new InfoList(parmDict, this.infoNames)
  }
}

class ParameterForm {
  parms = new Map<string, Parameter>()
  htmlParmForm: HTMLFormElement | null = null
  submitButton: HTMLElement | null = null
  nodeDisplay: NodeDisplay

  constructor(nodeDisplay: NodeDisplay, displayData: any) {
    this.nodeDisplay = nodeDisplay
    if ("Parms" in displayData) {
      for (const parmName in displayData.Parms) {
        this.parms.set(parmName, new Parameter(parmName, displayData.Parms[parmName], this))
      }
    }
  }

  submit(event: SubmitEvent) {
    event.preventDefault()
    const formData = new FormData(event.target as HTMLFormElement)

    const newParms: {[k: string]: number|string|boolean} = {}
    for (const [name, parm] of this.parms) {
      const value = formData.get(parm.name)
      switch (parm.inputType) {
        case 'number':
          newParms[parm.name] = Number(value)
          break
        case 'checkbox':
          newParms[parm.name] = Boolean(value)
          break
        case 'text':
          newParms[parm.name] = String(value)
          break
      }
    }
    console.log('newParms:', newParms)

    ask(msgKinds.Update, this.nodeDisplay.ID, newParms,
      (response) => {
        const diffs = this.htmlParmForm!.querySelectorAll(".changed")
        for (const delem of diffs) {
          delem.classList.remove("changed")
        }
        this.htmlParmForm?.classList.remove("changed")
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

  update(parms: any) {
    for (const [name, parm] of this.parms) {
      const newValue = parms[name]
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
  value: number|string|boolean
  origValue: number|string|boolean
  inputType: string
  parmForm: ParameterForm
  changed: boolean
  htmlParm: HTMLElement | null = null

  constructor(name: string, value: number|string|boolean, parmForm: ParameterForm) {
    this.name = name
    this.value = value
    this.origValue = value
    this.inputType = parmTypes.get(typeof this.value)!
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
          ${this.inputType == "checkbox" ? `${this.value ? "checked": ""}` : `value="${this.value}"`}
        />
      </div>
    `)
    this.htmlParm.querySelector("input")?.addEventListener("change", this.valueChange.bind(this))
    this.htmlParm.addEventListener("animationend", this.removeChangedAlert.bind(this), false)
    return this.htmlParm
  }

  removeChangedAlert(event: Event) {
    console.log("animation ended")
    this.htmlParm?.classList.remove("changeAlert")
  }

  update(newValue: number|string|boolean) {
    console.log("update request for parm", newValue)
    if (this.origValue != newValue) {
      console.log("update needed for parameter", this.inputType, newValue, typeof newValue)
      this.value = newValue
      this.origValue = newValue
      const htmlInput = this.htmlParm?.querySelector("input")! 
      if (this.inputType == "checkbox") {
        htmlInput.checked = newValue as boolean
      } else {
        htmlInput.setAttribute("value", `${newValue}`)
      }
      this.htmlParm?.classList.add("changeAlert")
    }
  }

  valueChange(event: Event) {
    const target = event.target as HTMLInputElement
    if (this.inputType == "checkbox") {
      this.value = target.checked
    } else {
      this.value = target.value
    }
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