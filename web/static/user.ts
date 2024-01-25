interface Account {
    ID: number,
    Name: string,
}

interface Message {
    ID: number,
    ReqID: number,
    ClientID: number,
    Type: string,
    JSPayload: any,
}

var clientID: number
var accounts: Account[]
var socket: WebSocket

// Custom events
const rebuild = new Event("rebuild")

// Message handler functions
const messageHandlers = {
    "FullState": fullStateHandler,
    "NewAccount": newAccountHandler, 
}

// Declare DOM mutation points
var details: HTMLDivElement

// Declare templates
var accountTemplate: HTMLTemplateElement 

window.onload = () => {
    // Load DOM mutation points
    details = <HTMLDivElement> document.getElementById("details")!

    // Load templates
    accountTemplate = <HTMLTemplateElement> document.getElementById("accountTemplate")!

    // Attach handlers

    // Start GUI
    // socket = new WebSocket("ws://127.0.0.1:8088/socket")
    
    // details.addEventListener("rebuild", rebuidDetails)
    // socket.addEventListener("message", receiveClientID)

    fetch("/api/accounts")
}

function rebuidDetails() {
    console.log("Rebuilding details.")
    details.innerHTML=''
    accounts.forEach((acc) => {
        var accElem  = <DocumentFragment> accountTemplate.content.cloneNode(true)
        var titleElem = accElem.querySelector(".title")!
        titleElem.textContent = `Account #${acc.ID} ${acc.Name}`
        details.appendChild(accElem)
    })
}

// Message handlers
function receiveClientID(e: MessageEvent) {
    let msg = JSON.parse(e.data)
    if (msg.Type != "ClientID") {
        console.log("Expected ClientID message, got: ", msg.Type)
        return
    }
    clientID = msg.JSPayload
    msg.ClientID = clientID
    console.log("ClientID set to: ", clientID)
    socket.send(JSON.stringify(msg))

    socket.removeEventListener("message", receiveClientID)
    socket.addEventListener("message", receiveMessage)

    let req = {
        ID: 1,
        ClientID: clientID,
        Type: "FullStateRequest",
        Payload: "{}",
    }
    socket.send(JSON.stringify(req))
}

function receiveMessage(e: MessageEvent) {
    let msg = JSON.parse(e.data)
    console.log("Received message from server: ", msg.Type)
    messageHandlers[msg.Type](msg)
}

function fullStateHandler(msg: Message) {
    accounts = msg.JSPayload
    details.dispatchEvent(rebuild)
}

function newAccountHandler(msg: Message) {
    accounts.push(msg.JSPayload)
    details.dispatchEvent(rebuild)
}