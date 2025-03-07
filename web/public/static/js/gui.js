import { WSMsg, nodeKinds, nodeKindName, msgKinds } from "./common.js";
const nodeKindIDs = {
    Group: 1,
    User: 2,
    Account: 3,
    Trader: 4,
};
function ask(mk, tid, pl, f) {
    return fetch(`/api/msg/${mk}/${tid}`, {
        method: "post",
        mode: "same-origin",
        body: pl === null ? null : JSON.stringify(pl),
    })
        .then((resp) => {
        if (!resp.ok) {
            if (resp.status === 401) {
                window.location.replace("/static/login.html" + new URL(location.href).search);
            }
            throw "http error";
        }
        return resp.json();
    })
        .then(f)
        .catch((e) => {
        console.log("Error:", e);
        alert(`${e} at line: ${e.lineNumber}`);
    });
}
function removeChangeAlert(e) {
    const t = e.target;
    t.classList.remove("changeAlert");
}
// Globals
let gui;
// This is not jQuery, but a helper function to turn a html string into a HTMLElement
const _dollarRegexp = /^\s+|\s+$|(?<=\>)\s+(?=\<)/gm;
function $(html) {
    const template = document.createElement("template");
    template.innerHTML = html.replace(_dollarRegexp, "");
    const result = template.content.firstElementChild;
    return result;
}
let parmTypes = new Map([
    ["string", "text"],
    ["number", "number"],
    ["boolean", "checkbox"],
]);
class GUI {
    constructor(rootNodeID) {
        this.tree = null;
        this.nodes = new Map();
        this.selection = new Map();
        this.guiID = 0;
        this.guiOTP = "";
        this.heart = 0;
        this.last_srv_beat = 0;
        this.heart_interval = 5000;
        this.roundtrip = 2000;
        this.htmlTreeView = document.querySelector("#tree-view");
        this.htmlDisplayView = document.querySelector("#display-view");
        this.rootNodeID = rootNodeID;
        this.htmlTreeView.addEventListener("click", this.treeClick.bind(this));
        this.socket = new WebSocket("/socket");
        this.socket.onmessage = this.socketMessageHandler.bind(this);
        this.heart = setInterval(this.heartbeat.bind(this), this.heart_interval);
    }
    heartbeat() {
        let off = Date.now() - this.last_srv_beat;
        if (off > this.heart_interval + this.roundtrip && this.last_srv_beat != 0) {
            console.log("connection to server lost.", off);
        }
        // console.log("Sending heartbeat to gui")
        this.socket.send(JSON.stringify({
            Kind: WSMsg.Heartbeat,
            GUIID: this.guiID,
            OTP: this.guiOTP,
        }));
    }
    socketMessageHandler(event) {
        var _a, _b;
        const wsm = JSON.parse(event.data);
        switch (wsm.Kind) {
            case WSMsg.Heartbeat:
                this.last_srv_beat = Date.now();
                console.log("Heartbeat received.");
                break;
            case WSMsg.Credentials:
                this.guiID = wsm.GUIID;
                this.guiOTP = wsm.OTP;
                this.socket.send(JSON.stringify(wsm));
                console.log(`Credentials received: guiid=${this.guiID}`);
                this.fetchTree(this.rootNodeID);
                break;
            case WSMsg.Update:
                console.log(`Update needed for node id: ${wsm.NodeID}`);
                let node = this.nodes.get(wsm.NodeID);
                if (node) {
                    node.updateDisplay();
                }
                else {
                    console.log(`Update requested for unknown node id: ${wsm.NodeID}`);
                }
                break;
            case WSMsg.TreeNodeRename:
                console.log(`Tree node rename received. id=${wsm.NodeID}, name='${wsm.NodeName}'`);
                gui.nodes.get(wsm.NodeID).renameTreeElement(wsm.NodeName);
                break;
            case WSMsg.TreeNodeCreate:
                console.log(`Tree node create received. msg=${JSON.stringify(wsm)}'`);
                const te = {
                    ID: wsm.NodeID,
                    Name: wsm.NodeName,
                    ParentID: wsm.NodeParentID,
                    Kind: wsm.NodeKind,
                };
                (_a = gui.nodes.get(te.ParentID)) === null || _a === void 0 ? void 0 : _a.createTreeElement(te);
                break;
            case WSMsg.TreeNodeDelete:
                console.log(`Tree node create received. msg=${JSON.stringify(wsm)}'`);
                (_b = gui.nodes.get(wsm.NodeParentID)) === null || _b === void 0 ? void 0 : _b.deleteTreeElement(wsm.NodeName);
                break;
        }
    }
    subscribe(id) {
        this.socket.send(JSON.stringify({
            Kind: WSMsg.Subscribe,
            GUIID: this.guiID,
            OTP: this.guiOTP,
            NodeID: id,
        }));
    }
    unsubscribe(id) {
        this.socket.send(JSON.stringify({
            Kind: WSMsg.Unsubscribe,
            GUIID: this.guiID,
            OTP: this.guiOTP,
            NodeID: id,
        }));
    }
    addNode(node) {
        this.nodes.set(node.ID, node);
    }
    fetchTree(nodeID) {
        ask(msgKinds.Get_Tree, nodeID, null, (treeData) => {
            this.tree = new Node(treeData);
            this.htmlTreeView.appendChild(this.tree.renderTree());
            this.updateSelection();
        });
    }
    treeClick(e) {
        let target = e.target;
        if (e.offsetX > target.offsetHeight) {
            e.preventDefault();
        }
        let nid = target.getAttribute("data-id");
        let loc = new URL(location.href);
        let selection = loc.searchParams.getAll("select");
        if (e.ctrlKey) {
            if (selection.includes(nid)) {
                loc.searchParams.delete("select", nid);
            }
            else {
                loc.searchParams.append("select", nid);
            }
        }
        else {
            if (selection.includes(nid) && selection.length == 1) {
                return;
            }
            else {
                loc.searchParams.delete("select");
                loc.searchParams.append("select", nid);
            }
        }
        window.history.pushState({}, "", loc);
        this.updateSelection();
    }
    updateSelection() {
        let newSelection = new URL(location.href).searchParams.getAll("select");
        this.selection.forEach((node, id) => {
            if (!newSelection.includes(node.ID.toString())) {
                this.selection.delete(id);
                node.deselect();
            }
        });
        for (let nid of newSelection) {
            const id = parseInt(nid);
            if (!Number.isNaN(id)) {
                if (!this.selection.has(id)) {
                    let node = this.nodes.get(id);
                    if (node != undefined) {
                        this.selection.set(id, node);
                        node.select();
                    }
                }
            }
        }
    }
}
class Node {
    constructor(nodeData = null, parentID = 0) {
        this.ID = 0;
        this.Name = "";
        this.Kind = 0;
        this.ParentID = 0;
        this.htmlTreeElem = null;
        this.htmlDisplayElem = null;
        this.display = null;
        this.children = new Map();
        if (nodeData == null)
            return;
        this.ID = nodeData.ID;
        this.ParentID = parentID;
        this.Name = nodeData.Name;
        this.Kind = nodeData.Kind;
        if ("Children" in nodeData) {
            nodeData.Children.forEach((e) => {
                const nn = new Node(e, this.ID);
                this.children.set(nn.Name, nn);
            });
        }
        gui.addNode(this);
    }
    renderTree() {
        let e;
        let state = 'open="true"';
        if (this.children.size === 0) {
            state = "";
        }
        e = $(`
      <details ${state}>
        <summary data-id="${this.ID}">${this.Name}</summary>
        <ul></ul>
      </details>
    `);
        let ule = e.querySelector("ul");
        for (let [nm, n] of this.children) {
            ule.appendChild(n.renderTree());
        }
        this.htmlTreeElem = e;
        return e;
    }
    renameTreeElement(newName) {
        console.log(`Renaming tree element "${this.Name}" to "${newName}".`);
        let e;
        e = this.htmlTreeElem.querySelector("summary");
        e.textContent = newName;
        e.addEventListener("animationend", removeChangeAlert);
        e.classList.add("changeAlert");
    }
    createTreeElement(nd) {
        var _a;
        console.log(`Creating tree element under "${this.Name}".`);
        const nn = new Node(nd, this.ID);
        this.children.set(nn.Name, nn);
        const de = nn.renderTree();
        const s = de.querySelector("summary");
        s.addEventListener("animationend", removeChangeAlert);
        // console.log(this.htmlTreeElem?.querySelector("ul"))
        (_a = this.htmlTreeElem) === null || _a === void 0 ? void 0 : _a.querySelector("ul").appendChild(de);
        if (this.children.size === 1) {
            this.htmlTreeElem.setAttribute("open", "true");
        }
        s.classList.add("changeAlert");
    }
    deleteTreeElement(nm) {
        const ch = this.children.get(nm);
        if (ch === undefined) {
            console.log("Can't find child with this name:", nm);
            return;
        }
        const che = ch.htmlTreeElem;
        che.addEventListener("animationend", this.removeChild.bind(this, ch));
        che.classList.add("fadeOut");
    }
    removeChild(n, e) {
        const ul = this.htmlTreeElem.querySelector("ul");
        ul.removeChild(n.htmlTreeElem);
        this.children.delete(n.Name);
    }
    updateDisplay() {
        console.log(`Updating node ${this.ID}.`);
        ask(msgKinds.Get_Display, this.ID, null, (displayData) => {
            var _a;
            if (displayData.error)
                throw new Error(displayData.error);
            (_a = this.display) === null || _a === void 0 ? void 0 : _a.update(displayData);
        });
    }
    select() {
        this.htmlTreeElem.classList.add("selected");
        ask(msgKinds.Get_Display, this.ID, null, (displayData) => {
            console.log(`Display data received:`, displayData);
            if (displayData.error)
                throw new Error(displayData.error);
            switch (displayData.Head.Kind) {
                case nodeKinds.Trader:
                    this.display = new BrokerDisplay(displayData);
                    break;
                case nodeKinds.Account:
                    this.display = new AccountDisplay(displayData);
                    break;
                case nodeKinds.User:
                    this.display = new UserDisplay(displayData);
                    break;
                case nodeKinds.Group:
                    this.display = new GroupDisplay(displayData);
                    break;
                case nodeKinds.TreeUpdater:
                    this.display = new TreeUpdaterDisplay(displayData);
                    break;
                case nodeKinds.Root:
                    this.display = new RootDisplay(displayData);
                    break;
                case nodeKinds.Users:
                    this.display = new UsersDisplay(displayData);
                    break;
                case nodeKinds.GUI:
                    this.display = new GUIDisplay(displayData);
                    break;
            }
            gui.htmlDisplayView.appendChild(this.display.render());
            gui.subscribe(this.ID);
        });
    }
    deselect() {
        this.htmlTreeElem.classList.remove("selected");
        gui.htmlDisplayView.removeChild(this.display.htmlDisplay);
        this.display = null;
        gui.unsubscribe(this.ID);
    }
}
class NodeDisplay {
    constructor(displayData) {
        this.parms = null;
        this.infos = null;
        this.htmlDisplay = null;
        this.name = displayData.Head.Name;
        this.kind = displayData.Head.Kind;
        this.ID = displayData.Head.ID;
        this.path = displayData.Head.Path;
        this.n_children = displayData.Head.N_children;
        if ("Parms" in displayData) {
            this.parms = new ParameterForm(this, displayData);
        }
    }
    render() {
        const head = this.renderHead();
        const detail = head.querySelector(".nodeDetailsBox");
        if (this.parms)
            detail.appendChild(this.parms.render());
        detail.appendChild(this.renderChildren());
        if (this.infos)
            detail.appendChild(this.infos.render());
        this.htmlDisplay = head;
        return head;
    }
    renderHead() {
        let dispHead = $(`
      <div class="nodeDisplay">
        <div class="nodeKindBox ${nodeKindName[this.kind].replace(/\s/g, "")}">
          <div class="nodeKindText">
            ${nodeKindName[this.kind]}
          </div>
        </div>
        <div class="nodeDetailsBox">
          <div class="displayHead">
            <form class="renameForm">
              <input type="text" class="displayName" value="${this.name}"/>
            </form>
            <div class="displayPath">${this.path}</div>
            <div class="nodeActions">
              Actions
              <div class="nodeActionsMenu">
                <div class="menuItem">Rename</div>
                <div class="menuItem">Delete</div>
              </div>
            </div>
            <div class="renameAction">
              Rename
            </div>
          </div>
        </div>
      </div>
    `);
        dispHead.querySelector(".displayName").addEventListener("input", this.nameChange.bind(this));
        dispHead.querySelector(".renameForm").addEventListener("submit", this.rename.bind(this));
        dispHead.querySelector(".renameAction").addEventListener("click", this.rename.bind(this));
        dispHead.querySelector(".displayName").addEventListener("animationend", removeChangeAlert);
        dispHead.querySelector(".displayPath").addEventListener("animationend", removeChangeAlert);
        return dispHead;
    }
    nameChange(e) {
        const t = e.target;
        const na = this.htmlDisplay.querySelector(".nodeActions");
        const ra = this.htmlDisplay.querySelector(".renameAction");
        if (t.value !== this.name) {
            na.style.display = "none";
            ra.style.display = "block";
        }
        else {
            na.style.display = "block";
            ra.style.display = "none";
        }
    }
    rename(e) {
        var t;
        e.preventDefault();
        if (e.type === "click") {
            t = e.target.parentElement;
        }
        else {
            t = e.target.parentElement;
        }
        const i = t.querySelector(".displayName");
        const na = t.querySelector(".nodeActions");
        const ra = t.querySelector(".renameAction");
        na.style.display = "block";
        ra.style.display = "none";
        i.blur();
        ask(msgKinds.Rename, gui.nodes.get(this.ID).ParentID, [this.name, i.value], (a) => {
            console.log("Rename request answer: ", a);
        });
    }
    renderChildren() {
        let elem = $(`
      <div class="childrenBox">
        <div class="n_children">Children: ${this.n_children}</div>
        <div class="newChildren">
          New
          <div class="newChildrenMenu">
            <div class="menuItem">Group</div>
            <div class="menuItem">User</div>
            <div class="menuItem">Pocket</div>
          </div>
        </div>
      </div>
    `);
        const mis = elem.querySelector(".newChildrenMenu").addEventListener("click", this.createChild.bind(this));
        return elem;
    }
    createChild(ev) {
        const t = ev.target;
        const n = t.textContent;
        const nk = nodeKindIDs[n];
        console.log("clicked create child:", n, this.ID);
        ask(msgKinds.Create, this.ID, [nk, ""], (r) => {
            console.log(r);
        });
    }
    update(displayData) {
        var _a;
        if (displayData.Head.Name !== this.name) {
            const nn = displayData.Head.Name;
            const ne = (_a = this.htmlDisplay) === null || _a === void 0 ? void 0 : _a.querySelector(".displayName");
            ne.setAttribute("value", `${nn}`);
            this.name = nn;
            ne.classList.add("changeAlert");
        }
        if (displayData.Head.Path !== this.path) {
            const pe = this.htmlDisplay.querySelector(".displayPath");
            this.path = displayData.Head.Path;
            pe.textContent = displayData.Head.Path;
            pe.classList.add("changeAlert");
        }
        if (this.parms) {
            this.parms.update(displayData.Parms);
        }
    }
}
class UserDisplay extends NodeDisplay {
    // infoNames = ["Last Modified"]
    constructor(displayData) {
        super(displayData);
        // this.infos = new InfoList(parmDict, this.infoNames)
    }
}
class BrokerDisplay extends NodeDisplay {
    // infoNames = ["Fee", "Last Modified"]
    constructor(displayData) {
        super(displayData);
        // this.infos = new InfoList(parmDict, this.infoNames)
    }
}
class AccountDisplay extends NodeDisplay {
    // infoNames = ["Last Modified"]
    constructor(displayData) {
        super(displayData);
        // this.infos = new InfoList(parmDict, this.infoNames)
    }
}
class GroupDisplay extends NodeDisplay {
    // infoNames = ["Last Modified"]
    constructor(displayData) {
        super(displayData);
        // this.infos = new InfoList(parmDict, this.infoNames)
    }
}
class TreeUpdaterDisplay extends NodeDisplay {
    // infoNames = ["Last Modified"]
    constructor(displayData) {
        super(displayData);
        // this.infos = new InfoList(parmDict, this.infoNames)
    }
}
class RootDisplay extends NodeDisplay {
    // infoNames = ["Last Modified"]
    constructor(displayData) {
        super(displayData);
        // this.infos = new InfoList(parmDict, this.infoNames)
    }
}
class UsersDisplay extends NodeDisplay {
    // infoNames = ["Last Modified"]
    constructor(displayData) {
        super(displayData);
        // this.infos = new InfoList(parmDict, this.infoNames)
    }
}
class GUIDisplay extends NodeDisplay {
    // infoNames = ["Last Modified"]
    constructor(displayData) {
        super(displayData);
        // this.infos = new InfoList(parmDict, this.infoNames)
    }
}
class ParameterForm {
    constructor(nodeDisplay, displayData) {
        this.parms = new Map();
        this.htmlParmForm = null;
        this.submitButton = null;
        this.nodeDisplay = nodeDisplay;
        if ("Parms" in displayData) {
            for (const parmName in displayData.Parms) {
                this.parms.set(parmName, new Parameter(parmName, displayData.Parms[parmName], this));
            }
        }
    }
    submitParms(event) {
        event.preventDefault();
        const formData = new FormData(event.target);
        const newParms = {};
        for (const [name, parm] of this.parms) {
            const value = formData.get(parm.name);
            switch (parm.inputType) {
                case "number":
                    newParms[parm.name] = Number(value);
                    break;
                case "checkbox":
                    newParms[parm.name] = Boolean(value);
                    break;
                case "text":
                    newParms[parm.name] = String(value);
                    break;
            }
        }
        console.log("newParms:", newParms);
        ask(msgKinds.Update_Parms, this.nodeDisplay.ID, newParms, (response) => {
            var _a;
            const diffs = this.htmlParmForm.querySelectorAll(".changed");
            for (const delem of diffs) {
                delem.classList.remove("changed");
            }
            (_a = this.htmlParmForm) === null || _a === void 0 ? void 0 : _a.classList.remove("changed");
        });
        return false;
    }
    render() {
        this.htmlParmForm = $(`
      <form class="parameterForm">
        <div class="parameterFormHeadBox">
            <div class="parameterFormTitle">Parameters:</div>
            <button class="parameterFormSubmit">Submit Parameters</button>
        </div>
      </form>
    `);
        this.htmlParmForm.addEventListener("submit", this.submitParms.bind(this));
        this.submitButton = this.htmlParmForm.querySelector(".parameterFormSubmit");
        for (const [name, parm] of this.parms) {
            this.htmlParmForm.appendChild(parm.render());
        }
        return this.htmlParmForm;
    }
    update(parms) {
        for (const [name, parm] of this.parms) {
            const newValue = parms[name];
            parm.update(newValue);
        }
    }
    checkDifferences() {
        var _a, _b;
        for (const [name, parm] of this.parms) {
            if (parm.changed) {
                (_a = this.htmlParmForm) === null || _a === void 0 ? void 0 : _a.classList.add("changed");
                return;
            }
        }
        (_b = this.htmlParmForm) === null || _b === void 0 ? void 0 : _b.classList.remove("changed");
    }
}
class Parameter {
    constructor(name, value, parmForm) {
        this.htmlParm = null;
        this.name = name;
        this.value = value;
        this.origValue = value;
        this.inputType = parmTypes.get(typeof this.value);
        this.parmForm = parmForm;
        this.changed = false;
    }
    render() {
        var _a;
        this.htmlParm = $(`
      <div class="parmBox">
        <label for="${this.name}" class="settingLabel">${this.name} </label>
        <input
          name="${this.name}"
          class="settingInput"
          type="${this.inputType}"
          ${this.inputType == "number" ? `step="any"` : ``}
          ${this.inputType == "checkbox" ? `${this.value ? "checked" : ""}` : `value="${this.value}"`}
        />
      </div>
    `);
        (_a = this.htmlParm.querySelector("input")) === null || _a === void 0 ? void 0 : _a.addEventListener("input", this.valueChange.bind(this));
        this.htmlParm.addEventListener("animationend", this.removeChangedAlert.bind(this), false);
        return this.htmlParm;
    }
    removeChangedAlert(event) {
        var _a;
        console.log("animation ended");
        (_a = this.htmlParm) === null || _a === void 0 ? void 0 : _a.classList.remove("changeAlert");
    }
    update(newValue) {
        var _a, _b;
        console.log("update request for parm", newValue);
        if (this.origValue != newValue) {
            console.log("update needed for parameter", this.inputType, newValue, typeof newValue);
            this.value = newValue;
            this.origValue = newValue;
            const htmlInput = (_a = this.htmlParm) === null || _a === void 0 ? void 0 : _a.querySelector("input");
            if (this.inputType == "checkbox") {
                htmlInput.checked = newValue;
            }
            else {
                htmlInput.setAttribute("value", `${newValue}`);
            }
            (_b = this.htmlParm) === null || _b === void 0 ? void 0 : _b.classList.add("changeAlert");
        }
    }
    valueChange(event) {
        var _a, _b;
        const target = event.target;
        if (this.inputType == "checkbox") {
            this.value = target.checked;
        }
        else {
            this.value = target.value;
        }
        this.changed = this.value != this.origValue;
        if (this.changed) {
            (_a = this.htmlParm) === null || _a === void 0 ? void 0 : _a.classList.add("changed");
        }
        else {
            (_b = this.htmlParm) === null || _b === void 0 ? void 0 : _b.classList.remove("changed");
        }
        this.parmForm.checkDifferences();
    }
}
class InfoList {
    constructor(parmDict, parmList = []) {
        this.infos = new Map();
        if (!parmList.length) {
            parmList = Object.keys(parmDict);
        }
        for (const name of parmList) {
            this.infos.set(name, new Info(name, parmDict[name]));
        }
    }
    update(parmDict) {
        for (const [name, info] of this.infos) {
            const newValue = parmDict[name];
            info.update(newValue);
        }
    }
    render() {
        let elem = $(`
      <div class="infoListBox">
        <div class="infoListHead">Info:</div>
      </div>
    `);
        console.log(this.infos);
        for (const [name, info] of this.infos) {
            elem.appendChild(info.render());
        }
        return elem;
    }
}
class Info {
    constructor(name, value) {
        this.name = "";
        this.value = 0;
        this.htmlInfo = null;
        this.name = name;
        this.value = value;
    }
    render() {
        this.htmlInfo = $(`
      <div class="infoBox">
        <span class="infoName">${this.name}: </span>
        <span class="infoValue">${this.value}</span>
      </div>
    `);
        return this.htmlInfo;
    }
    update(newValue) {
        var _a;
        if (this.value != newValue) {
            this.value = newValue;
            const htmlValue = this.htmlInfo.querySelector(".infoValue");
            htmlValue.textContent = `${newValue}`;
            (_a = this.htmlInfo) === null || _a === void 0 ? void 0 : _a.classList.add("changeAlert");
        }
    }
}
window.onload = () => {
    let userID = parseInt(document.getElementById("user-id").getAttribute("value"));
    console.log("UserID: ", userID);
    // Select user node in URL if nothing else is selected
    let loc = new URL(location.href);
    let selection = loc.searchParams.getAll("select");
    if (!selection.length) {
        loc.searchParams.append("select", userID.toString());
        window.history.pushState({}, "", loc);
    }
    gui = new GUI(userID);
    // Update display if location URL changes
    window.addEventListener("popstate", (event) => {
        gui.updateSelection();
    });
};
