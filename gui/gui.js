function main() {
    var socket = new WebSocket("ws://127.0.0.1:8088/socket");
    var state;
    var details = document.getElementById("details");
    var rebuild = new Event("rebuild");
    details.addEventListener("rebuild", function () {
        console.log("Rebuilding details.");
        state.forEach(function (acc) {
            var accElem = document.createElement("div");
            accElem.textContent = acc.Name;
            details.appendChild(accElem);
        });
    });
    socket.addEventListener("open", function () {
        var req = {
            ID: 1,
            Type: "FullStateRequest",
            Payload: "{}",
        };
        return socket.send(JSON.stringify(req));
    });
    socket.addEventListener("message", function (e) {
        console.log("Message from server: ", e.data);
        console.log(typeof e.data);
        var msg = JSON.parse(e.data);
        state = msg.JSPayload;
        details.dispatchEvent(rebuild);
    });
}
main();
