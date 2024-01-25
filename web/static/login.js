window.onload = function () {
    // Attach handlers
    document.getElementById("login-form").onsubmit = login;
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
        window.location.replace("/gui/user/" + data.userid);
    }).catch(function (e) { alert(e); });
    return false;
}
