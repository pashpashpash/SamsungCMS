//initialization
document.getElementById("maxCheck").checked = "false";
document.getElementById("folderCheck").checked = "false";
document.getElementById("homescreenCheck").checked = "false";
var webAppsHTML = "";  //set webAppsHTML string to null, so we can += to it later
var cms_database = fetchDB(); //fetch db object (not used currentlys)

// Sending rest request for all apps
var xmlhttp = new XMLHttpRequest();
var url = "/rest/allApps";
xmlhttp.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
        var allApps = JSON.parse(this.responseText);
        allAppsContainer = document.getElementById("allicons");
        showWebapps(allAppsContainer, allApps);
        console.log('Network request for allApps.json succeeded. JSON:');
        console.log(allApps);
    }
    else {
        console.log('Network request for allApps.json failed with response ' + this.status);
    }
};
xmlhttp.open("GET", url, true);
xmlhttp.send();


//FUNCTION DECLARATIONS
//responsible for fetching database object
function fetchDB() {
    var cms_database;
    fetch('cms-database.json').then(function(response) {
        if(response.ok)
        {
            response.json().then(function(json) {
                cms_database = json;
            });
        }
        else
        {
            console.log('Network request failed with response ' + response.status + ': ' + response.statusText);
        }
    });
    return cms_database;
}

//input an app container, and a json of webapps, and this func will display them in the container with proper nesting
function showWebapps(allAppsContainer, webapps) {
    for(var o= 0; o < webapps.length; o++){
        console.log("Adding "+webapps[o].name+" iconContainer to the HTML");
        webAppsHTML += "<div id='iconContainer'>";
        webAppsHTML += ("<img id='icon' src='" + webapps[o].iconUrl + "'");
        webAppsHTML += (" onclick=\"javascript:window.location.href='/ultra/"+ webapps[o].id +"'; return false;\"");
        webAppsHTML += (" />");

        webAppsHTML += ("<div id='iconText'>");
        webAppsHTML += (webapps[o].name + " Ultra");
        webAppsHTML += ("</div>");
        webAppsHTML += ("</div>");
    }

    webAppsHTML += "<div id='iconContainer'>";
    webAppsHTML += ("<img id='icon' src='" + "/images/add_icon.png" +"' />");
    webAppsHTML += ("<div id='iconText'>");
    webAppsHTML += ("Create new Ultra App");
    webAppsHTML += ("</div>");
    webAppsHTML += ("</div>");

    allAppsContainer.innerHTML = (webAppsHTML);
}


// CHECKBOXES + STAR TOGGLING LOGIC
document.querySelector('#maxCheck').onclick = function(){
    if(this.classList.contains("checked")) {
        this.classList.remove("checked");
    }
    else {
        this.classList.add("checked");
    }
};
document.querySelector('#folderCheck').onclick = function(){
    if(this.classList.contains("checked")) {
        this.classList.remove("checked");
    }
    else {
        this.classList.add("checked");
    }
};
document.querySelector('#homescreenCheck').onclick = function(){
    if(this.classList.contains("checked")) {
        this.classList.remove("checked");
    }
    else {
        this.classList.add("checked");
    }
};
document.querySelector('#star').onclick = function(){
    if(this.classList.contains("clicked")) {
        this.classList.remove("clicked");
        this.classList.add("unclicked");
    }
    else {
        this.classList.remove("unclicked");
        this.classList.add("clicked");

    }
};
