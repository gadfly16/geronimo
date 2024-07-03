import { NodeType } from "../shared/gui_types.js";
window.onload = function () {
    // Attach handlers
    document.getElementById("signup-form").onsubmit = signup;
};
function signup(e) {
    const data = new FormData(e.target);
    let userNode = {
        DetailType: NodeType.User,
        Name: data.get("Name"),
        Detail: {
            Email: data.get("Email"),
            Password: data.get("Password"),
        }
    };
    // Send the request
    fetch("/signup", {
        method: 'post',
        body: JSON.stringify(userNode),
        mode: 'same-origin',
    }).then((response) => {
        if (response.ok) {
            return response.json();
        }
        else {
            throw 'failed';
        }
    }).then((data) => {
        window.location.replace("login");
    }).catch((e) => { alert(e); });
    return false;
}
