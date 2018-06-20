var restRequest = function() {
    this.get = function(aUrl, aCallback) {
        var anHttpRequest = new XMLHttpRequest();
        anHttpRequest.onreadystatechange = function() {
            if (anHttpRequest.readyState == 4 && anHttpRequest.status == 200){
                console.log('REST.JS – NETWORK REQUEST SUCCESS : ' + aUrl);
                aCallback(JSON.parse(anHttpRequest.responseText));
            }
            else {
                console.log('REST.JS – REQUEST FAILURE – ' + anHttpRequest.status + ' : ' + aUrl);
            }
        }
        anHttpRequest.open( "GET", aUrl, true );
        anHttpRequest.send( null );
    }
}
