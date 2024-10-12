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
export let WSMsg = {
    Credentials: "Credentials",
    Subscribe: "Subscribe",
    Unsubscribe: "Unsubscribe",
    Update: "Update",
};
export var nodeKinds;
(function (nodeKinds) {
    nodeKinds[nodeKinds["Root"] = 0] = "Root";
    nodeKinds[nodeKinds["Group"] = 1] = "Group";
    nodeKinds[nodeKinds["User"] = 2] = "User";
    nodeKinds[nodeKinds["Broker"] = 3] = "Broker";
    nodeKinds[nodeKinds["Account"] = 4] = "Account";
})(nodeKinds || (nodeKinds = {}));
// export enum payloadKinds {
//   Empty = 0,
//   UserNode,
//   Parms,
// }
export let NodeKindName = ["root", "group", "user", "broker", "account", "broker", "pocket"];
