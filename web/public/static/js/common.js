export var msgKinds;
(function (msgKinds) {
    msgKinds[msgKinds["OK"] = 0] = "OK";
    msgKinds[msgKinds["Error"] = 1] = "Error";
    msgKinds[msgKinds["Stop"] = 2] = "Stop";
    msgKinds[msgKinds["Stopped"] = 3] = "Stopped";
    msgKinds[msgKinds["Update"] = 4] = "Update";
    msgKinds[msgKinds["Parms"] = 5] = "Parms";
    msgKinds[msgKinds["GetParms"] = 6] = "GetParms";
    msgKinds[msgKinds["Create"] = 7] = "Create";
    msgKinds[msgKinds["AuthUser"] = 8] = "AuthUser";
    msgKinds[msgKinds["GetTree"] = 9] = "GetTree";
    msgKinds[msgKinds["Tree"] = 10] = "Tree";
    msgKinds[msgKinds["GetCopy"] = 11] = "GetCopy";
    msgKinds[msgKinds["GetDisplay"] = 12] = "GetDisplay";
    msgKinds[msgKinds["Display"] = 13] = "Display";
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
})(WSMsg || (WSMsg = {}));
export var nodeKinds;
(function (nodeKinds) {
    nodeKinds[nodeKinds["Root"] = 0] = "Root";
    nodeKinds[nodeKinds["Group"] = 1] = "Group";
    nodeKinds[nodeKinds["User"] = 2] = "User";
    nodeKinds[nodeKinds["Account"] = 3] = "Account";
    nodeKinds[nodeKinds["Broker"] = 4] = "Broker";
})(nodeKinds || (nodeKinds = {}));
// export enum payloadKinds {
//   Empty = 0,
//   UserNode,
//   Parms,
// }
export let NodeKindName = ["root", "group", "user", "broker", "account", "broker", "pocket"];
