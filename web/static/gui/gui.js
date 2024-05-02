import { guiMessageType, NodeType, NodeTypeName } from "../shared/gui_types.js";
class Node {
    constructor() {
        this.ID = 0;
        this.Name = "";
        this.DetailType = 0;
        this.ParentID = 0;
    }
}
class NodeTree {
    constructor() {
        this.Root = new Node;
    }
}
class NodeDisplay {
    constructor(displayData) {
        this.Name = "";
        this.DetailType = 0;
        this.path = "";
        this.Display = null;
        this.Name = displayData.Name;
        this.DetailType = displayData.DetailType;
        switch (this.DetailType) {
            case NodeType.Broker:
                this.Display = new BrokerDisplay(displayData);
                break;
            case NodeType.Account:
                this.Display = new AccountDisplay(displayData);
                break;
            case NodeType.User:
                this.Display = new UserDisplay(displayData);
                break;
        }
        console.log("Display object: ", this);
    }
    render() {
        let html = `
      <div class="display">
        <div class="displayHead">
          <div class="displayName ${NodeTypeName[this.DetailType]}">${this.Name}</div>
          <div class="displayPath">${this.path}</div>
        </div>
        ${this.Display.render()}
      </div>`;
        return html;
    }
}
class UserDisplay {
    constructor(displayData) {
        this.Parameters = new ParameterForm;
        this.InfoList = new InfoList;
        let parmDict = displayData.Detail;
        parmDict["Last Modified"] = parmDict.CreatedAt;
        // this.Parameters.add(parmDict, ["Exchange"])
        this.InfoList.add(parmDict, ["Last Modified"]);
    }
    render() {
        let html = this.Parameters.render();
        html += this.InfoList.render();
        return html;
    }
}
class BrokerDisplay {
    constructor(displayData) {
        this.Parameters = new ParameterForm;
        this.InfoList = new InfoList;
        let parmDict = displayData.Detail;
        parmDict["Last Modified"] = parmDict.CreatedAt;
        this.Parameters.add(parmDict, ["Pair", "Base", "Quote", "LowLimit", "HighLimit", "Delta", "MinWait", "MaxWait", "Offset"]);
        this.InfoList.add(parmDict, ["Fee", "Last Modified"]);
    }
    render() {
        let html = this.Parameters.render();
        html += this.InfoList.render();
        return html;
    }
}
class AccountDisplay {
    constructor(displayData) {
        this.Parameters = new ParameterForm;
        this.InfoList = new InfoList;
        let parmDict = displayData.Detail;
        parmDict["Last Modified"] = parmDict.CreatedAt;
        this.Parameters.add(parmDict, ["Exchange"]);
        this.InfoList.add(parmDict, ["Last Modified"]);
    }
    render() {
        let html = this.Parameters.render();
        html += this.InfoList.render();
        return html;
    }
}
class ParameterForm {
    constructor() {
        this.ParameterList = [];
    }
    add(parmDict, parmList = []) {
        if (!parmList.length) {
            parmList = Object.keys(parmDict);
        }
        parmList.forEach(k => {
            this.ParameterList.push(new Parameter(k, parmDict[k]));
        });
    }
    render() {
        let html = `
      ${this.ParameterList.length ? `
      <form class="parameterForm">
        <div class="parameterFormHeadBox">
            <div class="parameterFormTitle">Parameters:</div>
            <div class="parameterFormSubmit">Submit</div>
        </div>        
        ${this.ParameterList.reduce((a, s) => a + s.render(), "")}
      </form>
      ` : ""}
    `;
        return html;
    }
}
class Parameter {
    constructor(name, value) {
        this.Name = "";
        this.Value = 0;
        this.InputType = "";
        this.Name = name;
        this.Value = value;
        this.InputType = typeof this.Value == "string" ? "text" : "number";
    }
    render() {
        let html = `
      <div class="inputBox">
        <label for="${this.Name} class="settingLabel">${this.Name}</label>
        <input
          name="${this.Name}"
          class="settingInput"
          type="${this.InputType}"
          value="${this.Value}"
        />
      </div>`;
        return html;
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
        let html = `
      ${this.InfoList.length ? `
      <div class="infoListBox">
      <div class="infoListHead">Info:</div>
        ${this.InfoList.reduce((a, s) => a + s.render(), "")}
      </div>
      ` : ""}
    `;
        return html;
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
        let html = `
      <div class="infoBox">
        <span class="infoName">${this.Name}:</span>
        <span class="infoValue">${this.Value}</span>
      </div>`;
        return html;
    }
}
function buildTree(treeNode, path) {
    let item;
    path = path + "/" + treeNode.Name;
    if ("children" in treeNode) {
        item = document.createElement("details");
        item.open = true;
        let summary = document.createElement("summary");
        summary.appendChild(document.createTextNode(treeNode.Name));
        summary.setAttribute("data-path", path);
        item.appendChild(summary);
        let children = document.createElement("ul");
        for (let ch of treeNode.children) {
            children.appendChild(buildTree(ch, path));
        }
        item.appendChild(children);
    }
    else {
        item = document.createElement("li");
        item.setAttribute("data-path", path);
        item.appendChild(document.createTextNode(treeNode.Name));
    }
    return item;
}
function loadDisplay() {
    let path = location.href.split("/gui/").at(-1);
    fetch("/api/display/" + path)
        .then((resp) => {
        return resp.json();
    })
        .then((data) => {
        console.log("Display Data: ", data);
        let display = new NodeDisplay(data);
        display.path = path;
        const displayBox = document.getElementById("displayBox");
        displayBox.innerHTML = display.render();
    })
        .catch((e) => {
        alert(e);
    });
}
function displayTemplate(dd) {
    return "";
}
function treeClick(e) {
    let target = e.target;
    if (e.offsetX > target.offsetHeight) {
        e.preventDefault();
    }
    let path = target.getAttribute("data-path");
    let current = new URL(location.href);
    let dest = current.origin + "/gui" + path;
    if (dest != current.href) {
        window.history.pushState({}, "", dest);
        loadDisplay();
    }
}
function getUserTree(userID) {
    fetch("/api/tree?" + new URLSearchParams({
        userid: userID.toString()
    })).then((resp) => {
        return resp.json();
    }).then((treeData) => {
        console.log(treeData);
        let treeRoot = document.querySelector("#tree");
        treeRoot.appendChild(buildTree(treeData, ""));
    }).catch((e) => {
        alert(e);
    });
}
window.onload = () => {
    var _a;
    let userID = parseInt(document.getElementById("user-id").getAttribute("value"));
    console.log("UserID: ", userID);
    let location = window.location.pathname;
    console.log("Location: ", location);
    let gm = {
        Type: guiMessageType.getUserTree,
        Payload: userID
    };
    getUserTree(userID);
    (_a = document.querySelector("#tree")) === null || _a === void 0 ? void 0 : _a.addEventListener("click", treeClick);
    window.addEventListener("popstate", (event) => {
        loadDisplay();
    });
    loadDisplay();
};
