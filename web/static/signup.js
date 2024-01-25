window.onload = function () {
    // Attach handlers
    document.getElementById("signup-form").onsubmit = signup;
};
function signup() {
    var formData = {
        "Name": document.getElementById("signup-name").value,
        "Email": document.getElementById("signup-email").value,
        "Password": document.getElementById("signup-password").value
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
