console.log("Hello world!")

# Create WebSocket connection.
socket = new WebSocket("ws://127.0.0.1:8088/socket")

# Connection opened
socket.addEventListener("open", (event) -> 
    stateReq =
        ID: 1
        Type: "FullStateRequest"
        Payload: "{}"
    socket.send(JSON.stringify(stateReq))
)

# Listen for messages
socket.addEventListener("message", (event) ->
  console.log("Message from server ", event.data)
  console.log(typeof event.data)
)