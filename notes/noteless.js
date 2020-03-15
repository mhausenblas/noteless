(function () {
    const notes = document.querySelector('#notes')
    const info = document.querySelector('#info')
    const listbtn = document.querySelector('#list')
    const rulesbtn = document.querySelector('#rules')
    const commandsbtn = document.querySelector('#commands')

    listbtn.addEventListener('click', () => {
        // call out to Noteless HTTP API (Lambda functions)
        fetch('http://127.0.0.1:9898/notes')
            .then((res) => {
                res.json().then((content) => {
                    console.log(content);
                    var res = "";
                    for (var i = 0; i < content.length; i++) {
                        res += "<div class='note'><img src='" + content[i].ImageBase64 + "' width='200px' /></div>\n";
                    }
                    notes.innerHTML = res;
                })
            })
        
    })

    rulesbtn.addEventListener('click', () => {
        // call out to Noteless HTTP API (Lambda functions)
        fetch('http://127.0.0.1:9898/rules')
            .then((response) => {
                if (!response.ok) {
                    throw new Error('API call failed');
                }
                return response.text()
            })
            .then((content) => {
                console.log(content);
                info.innerHTML = "<code><pre>" + content+"</pre></code>";
            })
            .catch((error) => {
                console.error('Problem with rules GET:', error);
            });

    })

    commandsbtn.addEventListener('click', () => {
        // call out to Noteless HTTP API (Lambda functions)
        fetch('http://127.0.0.1:9898/commands')
            .then((res) => {
                res.json().then((content) => {
                    console.log(content);
                    var res = "<strong>detected commands</strong>:\n<ul>\n";
                    for (var i = 0; i < content.length; i++) {
                        res += "<li>" + content[i] + "</li>\n";
                    }
                    res += "</ul>";
                    info.innerHTML = res;
                })
            })

    })
})()