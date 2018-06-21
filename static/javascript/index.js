//initialization
var cms_database = fetchDB(); //fetch db object (not used currentlys)
var server = new restRequest();
var selects = document.getElementsByTagName('select');
var maxCheckbox = document.getElementById("maxCheck");
var folderCheckbox = document.getElementById("folderCheck");
var homescreenCheckbox = document.getElementById("homescreenCheck");
var searchField = document.getElementsByClassName('search')[0];
var allAppsContainer = document.getElementById("allicons");
var filterParams = [selects, maxCheckbox, folderCheckbox, homescreenCheckbox, searchField];

// var platformSelect = document.getElementsByName("platform")[0];

// Sending rest request for a json of all ultra apps at /rest/allApps

var url = "/rest/allApps";
server.get(url, function(allApps) {
    showWebapps(allAppsContainer, allApps);
});

function selectChange(selectObject){
    console.log("SELECTCHANGE – Change detected in: " + selectObject.name);
    if (selectObject.oldvalue===undefined){
        selectObject.oldvalue="star"; //handles first case
    }
    console.log("SELECTCHANGE – Old Value is " + selectObject.oldvalue + ", New Value is " + selectObject.value);

    if(selectObject.oldvalue==="star")
    {
        starOFF();
    }
    if(selectObject.value==="star") //we should only check the rest when a select filter switches to a star
    {
        if(allStars(selects) === true) //if all select filter values are stars, starON();
        {
            starON();
        }
    }

    //get all select filter values + checkbox values
    applyFilters();

    //package all values into a json for request
    //send


}

function allStars(selectsToCheckforStars)
{
    for(var z=0; z<selectsToCheckforStars.length; z++){
        if(selectsToCheckforStars[z].value!="star")
        {
            return false;
        }
    }
    console.log("ALLSTARS – ALL FILTERS HAVE STARS");
    return true;
} //hi

function applyFilters()
{
    console.log("APPLYFILTERS – Current filter status:");
    console.log(filterParams);
    console.log("APPLYFILTERS – APPLYING FILTERS...")


    url = "/rest/allApps";
    server.get(url, function(allApps) {
        showWebapps(allAppsContainer, allApps);
    });
}


//=====================FUNCTION DECLARATIONS============================//
//responsible for fetching database object
function fetchDB() {
    var db;
    fetch('cms-database.json').then(function(response) {
        if(response.ok)
        {
            response.json().then(function(json) {
                db = json;
            });
        }
        else
        {
            console.log('FETCHDB – Network request failed with response ' + response.status + ': ' + response.statusText);
        }
    });
    return db;
}

//input an app container + a json of webapps, and this func will display them in the container with proper nesting
function showWebapps(allAppsContainer, webapps) {
    var webAppsHTML = "";  //set webAppsHTML string to null, so we can += to it later
    for(var o= 0; o < webapps.length; o++){
        console.log("SHOWWEBAPPS – Adding "+webapps[o].name+" iconContainer to the HTML");
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
function starOFF()
{
    document.querySelector('#star').classList.remove("clicked");
    document.querySelector('#star').classList.add("unclicked");
}
function starON()
{
    document.querySelector('#star').classList.remove("unclicked");
    document.querySelector('#star').classList.add("clicked");
}
function setAllSelectstoStar()
{
    var selects = document.getElementsByTagName('select');
    for(var z=0; z<selects.length; z++){
        if(selects[z].value!="star")
        {
            selects[z].value="star";
        }
    }
    console.log("ALL FILTERS HAVE STARS!!!!!!!");
}

// CHECKBOXES + STAR TOGGLING LOGIC
document.querySelector('#maxCheck').onclick = function(){
    if(this.classList.contains("checked")) {
        this.classList.remove("checked");
    }
    else {
        this.classList.add("checked");
    }
    applyFilters();
};
document.querySelector('#folderCheck').onclick = function(){
    if(this.classList.contains("checked")) {
        this.classList.remove("checked");
    }
    else {
        this.classList.add("checked");
    }
    applyFilters();
};
document.querySelector('#homescreenCheck').onclick = function(){
    if(this.classList.contains("checked")) {
        this.classList.remove("checked");
    }
    else {
        this.classList.add("checked");
    }
    applyFilters();
};

document.querySelector('#star').onclick = function(){
    if(allStars(selects)===false) //change star on click ONLY if all select filters are NOT stars
    {
        if(this.classList.contains("clicked")) {
            this.classList.remove("clicked");
            this.classList.add("unclicked");
        }
        else { //STAR WAS CLICKED
            this.classList.remove("unclicked");
            this.classList.add("clicked");
            setAllSelectstoStar();
        }
    }
};
