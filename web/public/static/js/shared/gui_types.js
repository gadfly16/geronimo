export var msgKinds;
(function (msgKinds) {
    msgKinds[msgKinds["Create"] = 0] = "Create";
})(msgKinds || (msgKinds = {}));
export let WSMsg = {
    Credentials: "Credentials",
    Subscribe: "Subscribe",
    Unsubscribe: "Unsubscribe",
    Update: "Update",
};
export var NodeKinds;
(function (NodeKinds) {
    NodeKinds[NodeKinds["Root"] = 0] = "Root";
    NodeKinds[NodeKinds["User"] = 1] = "User";
    NodeKinds[NodeKinds["Account"] = 2] = "Account";
    NodeKinds[NodeKinds["Broker"] = 3] = "Broker";
    NodeKinds[NodeKinds["Group"] = 4] = "Group";
    NodeKinds[NodeKinds["Pocket"] = 5] = "Pocket";
})(NodeKinds || (NodeKinds = {}));
export let NodeTypeName = ["root", "user", "account", "broker", "group", "pocket"];
