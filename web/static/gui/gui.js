import { guiMessageType } from "../shared/gui_types.js";
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
        let displayElement = document.querySelector("#displayTemplate").content.cloneNode(true).querySelector(".display");
        let displayName = displayElement.querySelector(".displayName");
        displayName.textContent = data.Name;
        displayName.classList.add(data.Detail.Type);
        const settings = data.Detail.Settings;
        let settingsElement = document.querySelector("#settingsTemplate").content.cloneNode(true);
        const settingFieldTemplate = document.querySelector("#settingFieldTemplate");
        for (const s in settings) {
            let field = settingFieldTemplate.content.cloneNode(true);
            let label = field.querySelector(".settingLabel");
            let input = field.querySelector(".settingInput");
            label.textContent = s;
            switch (typeof settings[s]) {
                case "string":
                    console.log("String value of ", s, settings[s]);
                    input.setAttribute("type", "text");
                    input.setAttribute("value", settings[s]);
                    break;
                case "number":
                    console.log("Number value of ", s, settings[s]);
                    input.setAttribute("type", "text");
                    input.setAttribute("value", settings[s].toString());
                    break;
                default:
                    alert("Unknown setting type:" + s + " " + settings[s]);
                    break;
            }
            settingsElement.appendChild(field);
        }
        displayElement.appendChild(settingsElement);
        const displayBox = document.querySelector("#displayBox");
        displayBox.removeChild(displayBox.querySelector(".display"));
        displayBox.appendChild(displayElement);
    })
        .catch((e) => {
        alert(e);
    });
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
    }).then((data) => {
        let treeRoot = document.querySelector("#tree");
        treeRoot.appendChild(buildTree(data, ""));
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
