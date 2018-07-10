var postRequest = function() {
    this.post = function(aUrl, aJSON, aCallback) {
        var xhr = new XMLHttpRequest();
        xhr.open("POST", aUrl, true);
        xhr.setRequestHeader("Content-Type", "application/json");
        xhr.onreadystatechange = function () {
            if (xhr.readyState === 4 && xhr.status === 200) {
                // console.log('POST.JS – NETWORK REQUEST SUCCESS : ' + aUrl);
                var json = JSON.parse(xhr.responseText);
                aCallback(json)
            } else {
                // console.log('POST.JS – REQUEST FAILURE – ' + xhr.status + ' : ' + aUrl);
            }
        };
        var data = JSON.stringify(aJSON);
        xhr.send(data);
    }
}
