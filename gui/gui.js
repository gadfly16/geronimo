var accounts;
var clientID;
var details = document.getElementById("details");
var rebuild = new Event("rebuild");
var messageHandlers = {
    "FullState": fullStateHandler,
    "NewAccount": newAccountHandler,
};
// Start GUI
var socket = new WebSocket("ws://127.0.0.1:8088/socket");
details.addEventListener("rebuild", rebuidDetails);
socket.addEventListener("message", receiveClientID);
function rebuidDetails() {
    console.log("Rebuilding details.");
    details.innerHTML = '';
    accounts.forEach(function (acc) {
        var accElem = document.createElement("div");
        accElem.textContent = "Account #".concat(acc.ID, " ").concat(acc.Name);
        details.appendChild(accElem);
    });
}
// Message handlers
function receiveClientID(e) {
    var msg = JSON.parse(e.data);
    if (msg.Type != "ClientID") {
        console.log("Expected ClientID message, got: ", msg.Type);
        return;
    }
    clientID = msg.JSPayload;
    msg.ClientID = clientID;
    console.log("ClientID set to: ", clientID);
    socket.send(JSON.stringify(msg));
    socket.removeEventListener("message", receiveClientID);
    socket.addEventListener("message", receiveMessage);
    var req = {
        ID: 1,
        ClientID: clientID,
        Type: "FullStateRequest",
        Payload: "{}",
    };
    socket.send(JSON.stringify(req));
}
function receiveMessage(e) {
    var msg = JSON.parse(e.data);
    console.log("Received message from server: ", msg.Type);
    messageHandlers[msg.Type](msg);
}
function fullStateHandler(msg) {
    accounts = msg.JSPayload;
    details.dispatchEvent(rebuild);
}
function newAccountHandler(msg) {
    accounts.push(msg.JSPayload);
    details.dispatchEvent(rebuild);
}
