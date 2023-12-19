interface account {
    ID: number;
    Name: string;
}


function main() {
    let socket = new WebSocket("ws://127.0.0.1:8088/socket")
    var state: account[]

    let details = document.getElementById("details")!

    const rebuild = new Event("rebuild")
    
    details.addEventListener("rebuild", () => {
        console.log("Rebuilding details.")
        state.forEach((acc) => {
            let accElem = document.createElement("div")
            accElem.textContent = acc.Name
            details.appendChild(accElem)
        })
    })

    socket.addEventListener("open", () => {
        let req = {
            ID: 1,
            Type: "FullStateRequest",
            Payload: "{}",
        }
        return socket.send(JSON.stringify(req))
    })
    
    socket.addEventListener("message", (e) => {
        console.log("Message from server: ", e.data)
        console.log(typeof e.data)
        let msg = JSON.parse(e.data)
        state = msg.JSPayload
        details.dispatchEvent(rebuild)
    })
}

main()