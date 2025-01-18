export var msgKinds;
(function (msgKinds) {
    msgKinds[msgKinds["OK"] = 0] = "OK";
    msgKinds[msgKinds["Error"] = 1] = "Error";
    msgKinds[msgKinds["Stop"] = 2] = "Stop";
    msgKinds[msgKinds["Stopped"] = 3] = "Stopped";
    msgKinds[msgKinds["Update"] = 4] = "Update";
    msgKinds[msgKinds["Parms"] = 5] = "Parms";
    msgKinds[msgKinds["GetParms"] = 6] = "GetParms";
    msgKinds[msgKinds["CreateChild"] = 7] = "CreateChild";
    msgKinds[msgKinds["AuthUser"] = 8] = "AuthUser";
    msgKinds[msgKinds["GetTree"] = 9] = "GetTree";
    msgKinds[msgKinds["Tree"] = 10] = "Tree";
    msgKinds[msgKinds["GetCopy"] = 11] = "GetCopy";
    msgKinds[msgKinds["GetDisplay"] = 12] = "GetDisplay";
    msgKinds[msgKinds["Display"] = 13] = "Display";
    msgKinds[msgKinds["Subscribe"] = 14] = "Subscribe";
    msgKinds[msgKinds["Unsubscribe"] = 15] = "Unsubscribe";
    msgKinds[msgKinds["NodeUpdate"] = 16] = "NodeUpdate";
    msgKinds[msgKinds["Rename"] = 17] = "Rename";
    msgKinds[msgKinds["TreeNodeRename"] = 18] = "TreeNodeRename";
    msgKinds[msgKinds["UpdatePath"] = 19] = "UpdatePath";
    msgKinds[msgKinds["RenameChild"] = 20] = "RenameChild";
    msgKinds[msgKinds["CreateUser"] = 21] = "CreateUser";
    msgKinds[msgKinds["SubscribeTree"] = 22] = "SubscribeTree";
    msgKinds[msgKinds["UnsubscribeTree"] = 23] = "UnsubscribeTree";
    msgKinds[msgKinds["TreeNodeCreate"] = 24] = "TreeNodeCreate";
    msgKinds[msgKinds["GetChild"] = 25] = "GetChild";
    msgKinds[msgKinds["InitGUI"] = 26] = "InitGUI";
    msgKinds[msgKinds["DeleteChild"] = 27] = "DeleteChild";
    msgKinds[msgKinds["TreeNodeDelete"] = 28] = "TreeNodeDelete";
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
