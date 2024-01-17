var clientID;
var accounts;
var socket;
// Custom events
var rebuild = new Event("rebuild");
// Message handler functions
var messageHandlers = {
    "FullState": fullStateHandler,
    "NewAccount": newAccountHandler,
};
// Declare DOM mutation points
var details;
// Declare templates
var accountTemplate;
window.onload = function () {
    // Load DOM mutation points
    details = document.getElementById("details");
    // Load templates
    accountTemplate = document.getElementById("accountTemplate");
    // Attach handlers
    document.getElementById("login-form").onsubmit = login;
    document.getElementById("sign-up-form").onsubmit = signup;
    document.getElementById("cancel-login").onclick = cancelLogin;
    document.getElementById("sign-up-button").onclick = switchToSignup;
    // Start GUI
    socket = new WebSocket("ws://127.0.0.1:8088/socket");
    details.addEventListener("rebuild", rebuidDetails);
    socket.addEventListener("message", receiveClientID);
};
function login() {
    var formData = {
        "Email": document.getElementById("login-email").value,
        "Password": document.getElementById("login-password").value
    };
    // Send the request
    fetch("/login", {
        method: 'post',
        body: JSON.stringify(formData),
        mode: 'same-origin',
    }).then(function (response) {
        if (response.ok) {
            return response.json();
        }
        else {
            throw 'unauthorized';
        }
    }).then(function (data) {
        // Now we have a OTP, send a Request to Connect to WebSocket
        // connectWebsocket(data.otp);
    }).catch(function (e) { alert(e); });
    return false;
}
function cancelLogin() {
    document.getElementById("login-screen").style.display = "none";
}
function switchToSignup() {
    cancelLogin();
    document.getElementById("sign-up-screen").style.display = "block";
}
function signup() {
    var formData = {
        "Name": document.getElementById("sign-up-name").value,
        "Email": document.getElementById("sign-up-email").value,
        "Password": document.getElementById("sign-up-password").value
    };
    // Send the request
    fetch("/signup", {
        method: 'post',
        body: JSON.stringify(formData),
        mode: 'same-origin',
    }).then(function (response) {
        if (response.ok) {
            return response.json();
        }
        else {
            throw 'failed';
        }
    }).catch(function (e) { alert(e); });
    return false;
}
function rebuidDetails() {
    console.log("Rebuilding details.");
    details.innerHTML = '';
    accounts.forEach(function (acc) {
        var accElem = accountTemplate.content.cloneNode(true);
        var titleElem = accElem.querySelector(".title");
        titleElem.textContent = "Account #".concat(acc.ID, " ").concat(acc.Name);
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
