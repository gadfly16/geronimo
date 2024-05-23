import { WSMsg, NodeType, NodeTypeName } from "../shared/gui_types.js";
// UI Globals
let gui;
// This is not jQuery, but a helper function to turn a html string into a HTMLElement
let _dollarRegexp = /^\s+|\s+$|(?<=\>)\s+(?=\<)/gm;
function $(html) {
    const template = document.createElement('template');
    template.innerHTML = html.replace(_dollarRegexp, '');
    const result = template.content.firstElementChild;
    return result;
}
class GUI {
    constructor(rootNodeID) {
        this.tree = null;
        this.nodes = new Map();
        this.selection = new Map();
        this.guiID = 0;
        this.guiOTP = "";
        this.htmlTreeView = document.querySelector("#tree-view");
        this.htmlDisplayView = document.querySelector("#display-view");
        this.loadTree(rootNodeID);
        this.htmlTreeView.addEventListener("click", this.treeClick.bind(this));
        this.socket = new WebSocket("/socket");
        this.socket.onmessage = this.socketMessageHandler.bind(this);
    }
    socketMessageHandler(event) {
        const msg = JSON.parse(event.data);
        switch (msg.Type) {
            case WSMsg.Credentials:
                this.guiID = msg.GUIID;
                this.guiOTP = msg.OTP;
                this.socket.send(JSON.stringify(msg));
                break;
            case WSMsg.Update:
                console.log(`Update needed for node ${msg.NodeID}`);
        }
    }
    subscribe(id) {
        this.socket.send(JSON.stringify({
            Type: WSMsg.Subscribe,
            GUIID: this.guiID,
            OTP: this.guiOTP,
            NodeID: id,
        }));
    }
    unsubscribe(id) {
        this.socket.send(JSON.stringify({
            Type: WSMsg.Unsubscribe,
            GUIID: this.guiID,
            OTP: this.guiOTP,
            NodeID: id,
        }));
    }
    addNode(node) {
        this.nodes.set(node.ID.toString(), node);
    }
    loadTree(nodeID) {
        fetch("/api/tree?" + new URLSearchParams({
            userid: nodeID.toString()
        })).then((resp) => {
            return resp.json();
        }).then((treeData) => {
            this.tree = new Node(treeData);
            this.htmlTreeView.appendChild(this.tree.renderTree());
            this.updateSelection();
        }).catch((e) => {
            alert(e);
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
                if (!(this.selection.has(id))) {
                    let node = this.nodes.get(nid);
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
        this.DetailType = 0;
        this.ParentID = 0;
        this.htmlTreeElem = null;
        this.htmlDisplayElem = null;
        this.display = null;
        this.children = [];
        if (nodeData == null)
            return;
        this.ID = nodeData.ID;
        this.Name = nodeData.Name;
        this.DetailType = nodeData.DetailType;
        if ("children" in nodeData) {
            nodeData.children.forEach((e) => {
                this.children.push(new Node(e, this.ID));
            });
        }
        gui.addNode(this);
    }
    renderTree() {
        let e;
        if (this.children.length) {
            e = $(`
        <details open="true">
          <summary data-id="${this.ID}">${this.Name}</summary>
          <ul></ul>
        </details>
      `);
            let ule = e.querySelector("ul");
            for (let n of this.children) {
                ule.appendChild(n.renderTree());
            }
        }
        else {
            e = $(`
        <div>
          <li data-id="${this.ID}">${this.Name}</li>
        </div>
      `);
        }
        this.htmlTreeElem = e;
        return e;
    }
    select() {
        this.htmlTreeElem.classList.add("selected");
        fetch(`/api/display?select=${this.ID.toString()}`)
            .then((resp) => {
            return resp.json();
        })
            .then((displayDataList) => {
            if (displayDataList.error)
                throw new Error(displayDataList.error);
            for (let dd of displayDataList) {
                switch (dd.DetailType) {
                    case NodeType.Broker:
                        this.display = new BrokerDisplay(dd);
                        break;
                    case NodeType.Account:
                        this.display = new AccountDisplay(dd);
                        break;
                    case NodeType.User:
                        this.display = new UserDisplay(dd);
                        break;
                }
            }
            gui.htmlDisplayView.appendChild(this.display.render());
            gui.subscribe(this.ID);
        })
            .catch((e) => {
            if (e.message == "unauthorized") {
                window.location.replace("/login" + new URL(location.href).search);
            }
            alert(e + " at line: " + e.lineNumber);
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
        this.parameters = null;
        this.infos = null;
        this.htmlDisplay = null;
        this.name = displayData.Name;
        this.detailType = displayData.DetailType;
        this.id = displayData.ID;
        this.path = displayData.Path;
    }
    render() {
        let disp = this.renderHead();
        if (this.parameters)
            disp.appendChild(this.parameters.render());
        if (this.infos)
            disp.appendChild(this.infos.render());
        this.htmlDisplay = disp;
        return disp;
    }
    renderHead() {
        let dispHead = $(`
      <div class="display">
        <div class="displayHead">
          <div class="displayName ${NodeTypeName[this.detailType]}">${this.name}</div>
          <div class="displayPath">${this.path}</div>
        </div>
      </div>
    `);
        return dispHead;
    }
}
class UserDisplay extends NodeDisplay {
    constructor(displayData) {
        super(displayData);
        let parmDict = displayData.Detail;
        parmDict["Last Modified"] = parmDict.CreatedAt;
        this.infos = new InfoList;
        this.infos.add(parmDict, ["Last Modified"]);
    }
}
class BrokerDisplay extends NodeDisplay {
    constructor(displayData) {
        super(displayData);
        let parmDict = displayData.Detail;
        parmDict["Last Modified"] = parmDict.CreatedAt;
        this.parameters = new ParameterForm(this);
        this.parameters.add(parmDict, ["Pair", "Base", "Quote", "LowLimit", "HighLimit", "Delta", "MinWait", "MaxWait", "Offset"]);
        this.infos = new InfoList;
        this.infos.add(parmDict, ["Fee", "Last Modified"]);
    }
}
class AccountDisplay extends NodeDisplay {
    constructor(displayData) {
        super(displayData);
        let parmDict = displayData.Detail;
        parmDict["Last Modified"] = parmDict.CreatedAt;
        this.parameters = new ParameterForm(this);
        this.parameters.add(parmDict, ["Exchange"]);
        this.infos = new InfoList;
        this.infos.add(parmDict, ["Last Modified"]);
    }
}
class ParameterForm {
    constructor(nodeDisplay) {
        this.ParameterList = [];
        this.formElem = null;
        this.submitButton = null;
        this.nodeDisplay = nodeDisplay;
    }
    add(parmDict, parmList = []) {
        if (!parmList.length) {
            parmList = Object.keys(parmDict);
        }
        parmList.forEach(k => {
            this.ParameterList.push(new Parameter(k, parmDict[k], this));
        });
    }
    submit(event) {
        event.preventDefault();
        const data = new FormData(event.target);
        const detail = {};
        for (const parm of this.ParameterList) {
            const value = data.get(parm.name);
            detail[parm.name] = parm.inputType == "number" ? Number(value) : value;
        }
        const apiUpdatePath = `/api/update/${NodeTypeName[this.nodeDisplay.detailType]}`;
        console.log(apiUpdatePath);
        const msg = {
            Type: "Update",
            Path: this.nodeDisplay.path,
            Payload: detail
        };
        console.log(msg);
        fetch(apiUpdatePath, {
            method: 'post',
            body: JSON.stringify(msg),
            mode: 'same-origin',
        }).then((response) => {
            if (response.ok) {
                console.log(response);
            }
            else {
                throw 'failed';
            }
        }).catch((e) => { alert(e); });
        return false;
    }
    render() {
        this.formElem = $(`
      <form class="parameterForm">
        <div class="parameterFormHeadBox">
            <div class="parameterFormTitle">Parameters:</div>
            <button class="parameterFormSubmit">Submit Parameters</button>
        </div>        
      </form>
    `);
        this.formElem.addEventListener("submit", this.submit.bind(this));
        this.submitButton = this.formElem.querySelector(".parameterFormSubmit");
        for (let parm of this.ParameterList) {
            this.formElem.appendChild(parm.render());
        }
        return this.formElem;
    }
    checkDifferences() {
        var _a, _b;
        for (const parm of this.ParameterList) {
            if (parm.isDifferent) {
                (_a = this.formElem) === null || _a === void 0 ? void 0 : _a.classList.add("different");
                return;
            }
        }
        (_b = this.formElem) === null || _b === void 0 ? void 0 : _b.classList.remove("different");
    }
}
class Parameter {
    constructor(name, value, parmForm) {
        this.elem = null;
        this.name = name;
        this.value = value;
        this.origValue = value;
        this.inputType = typeof this.value == "string" ? "text" : "number";
        this.parmForm = parmForm;
        this.isDifferent = false;
    }
    render() {
        var _a;
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
    `);
        (_a = this.elem.querySelector("input")) === null || _a === void 0 ? void 0 : _a.addEventListener("change", this.valueChange.bind(this));
        return this.elem;
    }
    valueChange(event) {
        var _a, _b;
        const target = event.target;
        this.value = target.value;
        this.isDifferent = (this.value != this.origValue);
        if (this.isDifferent) {
            (_a = this.elem) === null || _a === void 0 ? void 0 : _a.classList.add("different");
        }
        else {
            (_b = this.elem) === null || _b === void 0 ? void 0 : _b.classList.remove("different");
        }
        this.parmForm.checkDifferences();
    }
}
class InfoList {
    constructor() {
        this.InfoList = [];
    }
    add(parmDict, parmList = []) {
        if (!parmList.length) {
            parmList = Object.keys(parmDict);
        }
        parmList.forEach(k => {
            this.InfoList.push(new Info(k, parmDict[k]));
        });
    }
    render() {
        let elem = $(`
      <div class="infoListBox">
        <div class="infoListHead">Info:</div>
      </div>
    `);
        for (let info of this.InfoList) {
            elem.appendChild(info.render());
        }
        return elem;
    }
}
class Info {
    constructor(name, value) {
        this.Name = "";
        this.Value = 0;
        this.Name = name;
        this.Value = value;
    }
    render() {
        let elem = $(`
      <div class="infoBox">
        <span class="infoName">${this.Name}:</span>
        <span class="infoValue"><b>${this.Value}</b></span>
      </div>
    `);
        return elem;
    }
}
function displayTemplate(dd) {
    return "";
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
