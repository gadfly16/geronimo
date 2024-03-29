window.onload = function() {
    // Attach handlers
    document.getElementById("login-form")!.onsubmit = login;
}

function login(e: SubmitEvent) {
    const data = new FormData(<HTMLFormElement>e.target)
    let user = {
        Email: data.get("Email"),
        Password: data.get("Password"),
    }
    fetch("/login", {
        method: 'post',
        body: JSON.stringify(user),
        mode: 'same-origin',
    }).then((response) => {
        if (response.ok) {
            return response.json();
        } else {
            throw 'unauthorized';
        }
    }).then((data) => {
        const destParam = new URLSearchParams(new URL(location.href).search)
        window.location.replace(destParam.has("dest") ? "/gui" + destParam.get("dest") : "/gui/home")
    }).catch((e) => { alert(e) });
    return false;
}

