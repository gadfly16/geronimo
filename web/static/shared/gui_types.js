export let guiMessageType = {
    getUserTree: "getUserTree",
    userTree: "userTree"
};
export var NodeType;
(function (NodeType) {
    NodeType[NodeType["Root"] = 0] = "Root";
    NodeType[NodeType["User"] = 1] = "User";
    NodeType[NodeType["Account"] = 2] = "Account";
    NodeType[NodeType["Broker"] = 3] = "Broker";
})(NodeType || (NodeType = {}));
