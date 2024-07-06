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
})(nodeKinds || (nodeKinds = {}));
export var payloadKinds;
(function (payloadKinds) {
    payloadKinds[payloadKinds["UserNodePayload"] = 0] = "UserNodePayload";
})(payloadKinds || (payloadKinds = {}));
export let NodeTypeName = ["root", "user", "account", "broker", "group", "pocket"];
