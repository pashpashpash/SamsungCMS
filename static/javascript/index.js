//initialization
var cms_database = fetchDB(); //fetch db object (not used currentlys)
var server = new restRequest();
var allAppsContainer = document.getElementById("allicons");
var filterParams = [selects, maxCheckbox, folderCheckbox, homescreenCheckbox, searchField];
var swapOutContainer = document.getElementById("swapOutContainer");
// var platformSelect = document.getElementsByName("platform")[0];

// Sending rest request for a json of all ultra apps at /rest/allApps

var url = "/rest/allApps";
server.get(url, function(allApps) {
    showWebapps(allAppsContainer, allApps);
});

function applyFilters()
{
    console.log("APPLYFILTERS – Current filter status:");
    console.log(filterParams);
    console.log("APPLYFILTERS – Applying filters... (Currently not implemented)")
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
        webAppsHTML += (" onclick=\"swapOut('"+ webapps[o].id +"')\"");
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
function swapOut(appID)
{

    console.log("SWAPOUT – Swapping out app tray for single ultra app view...");
    console.log("SWAPOUT – Current filter status: ");
    console.log(filterParams);
    console.log("SWAPOUT – Figuring out app info based off app ID and current filter status...")

    // Sending rest request for a specific ultra app
    var url = "/rest/ultra/" + appID;

    server.get(url, function(app) {
         //set webAppHTML string to null, so we can += to it later
        console.log("SWAPOUT – Adding "+app.name+" app to the HTML");
        window.history.pushState("", "", '/ultra/' + app.id);
        swapOutContainer.innerHTML = generateAppHTML(app);
        document.getElementById('header').children[1].innerHTML =  app.name + '<span id="smallerText"> Ultra</span>';
        console.log("SWAPOUT – Successfully swapped out html ");
    });
}
function generateAppHTML(app)
{
    console.log("GENERATEAPPHTML – Generating " + app.name + " html...")
    var webAppHTML = "<hr>";
    webAppHTML += '<div class ="webApp">';

        webAppHTML += "<div class='row'>";
            webAppHTML += "<div class='rowDescription'>";
                webAppHTML += "Name";
            webAppHTML += ("</div>");
            webAppHTML += "<div class='rowValue'>";
                webAppHTML += app.name;
            webAppHTML += ("</div>");
            webAppHTML += "<div class='edit'></div>";
        webAppHTML += "</div>";

        webAppHTML += "<div class='row'>";
            webAppHTML += "<div class='rowDescription'>";
                webAppHTML += "Rank";
            webAppHTML += ("</div>");
            webAppHTML += "<div class='rowValue'>";
                webAppHTML += app.rank;
            webAppHTML += ("</div>");
            webAppHTML += "<div class='edit'></div>";
        webAppHTML += "</div>";

        webAppHTML += "<div class='row'>";
            webAppHTML += "<div class='rowDescription'>";
                webAppHTML += "Webapp Link";
            webAppHTML += ("</div>");
            webAppHTML += "<div class='rowValue'>";
                webAppHTML += extractRootDomain(app.homeUrl) + "/...";
            webAppHTML += ("</div>");
            webAppHTML += "<div class='edit'></div>";
        webAppHTML += "</div>";

        if(app.hasOwnProperty('defaultEnabledFeatures')) {
        webAppHTML += "<div class='row'>";
            webAppHTML += "<div class='rowDescription'>";
                webAppHTML += "Enabled Features";
            webAppHTML += ("</div>");
            for(var i = 0; i < app.defaultEnabledFeatures.length; i++)
            {
                webAppHTML += "<div class='rowValue'>";
                webAppHTML += app.defaultEnabledFeatures[i];
                webAppHTML += ("</div>");
            }
            webAppHTML += "<div class='edit'></div>";
        webAppHTML += "</div>";
        }

        if(app.hasOwnProperty('hiddenFeatures')) {
        webAppHTML += "<div class='row'>";
            webAppHTML += "<div class='rowDescription'>";
                webAppHTML += "Hidden Features";
            webAppHTML += ("</div>");
            for(var i = 0; i < app.hiddenFeatures.length; i++)
            {
                webAppHTML += "<div class='rowValue'>";
                webAppHTML += app.hiddenFeatures[i];
                webAppHTML += ("</div>");
            }
            webAppHTML += "<div class='edit'></div>";
        webAppHTML += "</div>";
        }

        if(app.hasOwnProperty('nativeApps')) {
            webAppHTML += "<div class='row'>";
                webAppHTML += "<div class='rowDescription'>";
                    webAppHTML += "Native Apps";
                webAppHTML += ("</div>");
                for(var i = 0; i < app.nativeApps.length; i++)
                {
                    webAppHTML += "<div class='rowValue'>";
                    webAppHTML += app.nativeApps[i].trunc(24);
                    webAppHTML += ("</div>");
                }
                webAppHTML += "<div class='edit'></div>";
            webAppHTML += "</div>";
        }

        webAppHTML += "<div class='row'>";
            webAppHTML += "<div class='rowDescription'>";
                webAppHTML += "Icon URL";
            webAppHTML += ("</div>");
            webAppHTML += "<div class='rowValue'>";
                webAppHTML += app.iconUrl;
            webAppHTML += ("</div>");
            webAppHTML += "<div class='edit'></div>";
        webAppHTML += "</div>";

    webAppHTML += ("</div>");
    webAppHTML += ("<hr>");
    webAppHTML += ("<div class='center'>");
        webAppHTML += ('<div id="headerText"> Timeline <span id="smallerText">Version History</span></div>');
        webAppHTML += ("<div id='timeline'>");
        //not implemented
        webAppHTML += ("</div>");
    webAppHTML += ("</div>");
    return webAppHTML;
}


//HELPER FUNCTIONS
String.prototype.trunc = String.prototype.trunc ||
    function(n){
        return (this.length > n) ? this.substr(0, n-1) + '&hellip;' : this;
    };

// To address those who want the "root domain," use this function:
function extractRootDomain(url) {
    var domain = extractHostname(url),
        splitArr = domain.split('.'),
        arrLen = splitArr.length;

    //extracting the root domain here
    //if there is a subdomain
    if (arrLen > 2) {
        domain = splitArr[arrLen - 2] + '.' + splitArr[arrLen - 1];
        //check to see if it's using a Country Code Top Level Domain (ccTLD) (i.e. ".me.uk")
        if (splitArr[arrLen - 2].length == 2 && splitArr[arrLen - 1].length == 2) {
            //this is using a ccTLD
            domain = splitArr[arrLen - 3] + '.' + domain;
        }
    }
    return domain;
}

function extractHostname(url) {
    var hostname;
    //find & remove protocol (http, ftp, etc.) and get hostname

    if (url.indexOf("://") > -1) {
        hostname = url.split('/')[2];
    }
    else {
        hostname = url.split('/')[0];
    }

    //find & remove port number
    hostname = hostname.split(':')[0];
    //find & remove "?"
    hostname = hostname.split('?')[0];

    return hostname;
}
