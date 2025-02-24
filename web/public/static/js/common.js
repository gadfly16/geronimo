export var msgKinds;
(function (msgKinds) {
    msgKinds[msgKinds["OK"] = 0] = "OK";
    msgKinds[msgKinds["Error"] = 1] = "Error";
    msgKinds[msgKinds["Create"] = 2] = "Create";
    msgKinds[msgKinds["Rename"] = 3] = "Rename";
    msgKinds[msgKinds["Delete"] = 4] = "Delete";
    msgKinds[msgKinds["Get_Parms"] = 5] = "Get_Parms";
    msgKinds[msgKinds["Get_Auth"] = 6] = "Get_Auth";
    msgKinds[msgKinds["Get_Tree"] = 7] = "Get_Tree";
    msgKinds[msgKinds["Get_Display"] = 8] = "Get_Display";
    msgKinds[msgKinds["Get_Child"] = 9] = "Get_Child";
    msgKinds[msgKinds["Update_Parms"] = 10] = "Update_Parms";
    msgKinds[msgKinds["Update_GUI"] = 11] = "Update_GUI";
    msgKinds[msgKinds["Update_Tree"] = 12] = "Update_Tree";
    msgKinds[msgKinds["Subscribe"] = 13] = "Subscribe";
    msgKinds[msgKinds["Unsubscribe"] = 14] = "Unsubscribe";
    msgKinds[msgKinds["Stop"] = 15] = "Stop";
    msgKinds[msgKinds["UpdatePath"] = 16] = "UpdatePath";
    msgKinds[msgKinds["InitGUI"] = 17] = "InitGUI";
})(msgKinds || (msgKinds = {}));
export var WSMsg;
(function (WSMsg) {
    WSMsg[WSMsg["Credentials"] = 0] = "Credentials";
    WSMsg[WSMsg["Subscribe"] = 1] = "Subscribe";
    WSMsg[WSMsg["Unsubscribe"] = 2] = "Unsubscribe";
    WSMsg[WSMsg["Update"] = 3] = "Update";
    WSMsg[WSMsg["Error"] = 4] = "Error";
    WSMsg[WSMsg["ClientShutdown"] = 5] = "ClientShutdown";
    WSMsg[WSMsg["Heartbeat"] = 6] = "Heartbeat";
    WSMsg[WSMsg["TreeNodeRename"] = 7] = "TreeNodeRename";
    WSMsg[WSMsg["TreeNodeCreate"] = 8] = "TreeNodeCreate";
    WSMsg[WSMsg["TreeNodeDelete"] = 9] = "TreeNodeDelete";
})(WSMsg || (WSMsg = {}));
export var nodeKinds;
(function (nodeKinds) {
    nodeKinds[nodeKinds["Root"] = 0] = "Root";
    nodeKinds[nodeKinds["Group"] = 1] = "Group";
    nodeKinds[nodeKinds["User"] = 2] = "User";
    nodeKinds[nodeKinds["Account"] = 3] = "Account";
    nodeKinds[nodeKinds["Trader"] = 4] = "Trader";
    nodeKinds[nodeKinds["TreeUpdater"] = 5] = "TreeUpdater";
    nodeKinds[nodeKinds["Users"] = 6] = "Users";
    nodeKinds[nodeKinds["GUI"] = 7] = "GUI";
})(nodeKinds || (nodeKinds = {}));
export let nodeKindName = [
    "root",
    "group",
    "user",
    "account",
    "broker",
    "tree updater",
    "users",
    "GUI",
];
