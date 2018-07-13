//initialization
var cms_database = fetchDB(); //fetch db object (not used currentlys)
var server_get = new restRequest();
var server_post = new postRequest();
var appTray = document.getElementById("allicons");
var filterParams = [selects, searchField];
var swapOutContainer = document.getElementById("swapOutContainer");

window.addEventListener('keydown',function(e){if(e.keyIdentifier=='U+000A'||e.keyIdentifier=='Enter'||e.keyCode==13){if(e.target.nodeName=='INPUT'&&e.target.type=='text'){
    e.preventDefault();
    if(e.srcElement===document.getElementById('countrySearch'))
    {
        console.log("searchForCountry");
        displayCountrySearchResults(document.getElementById('countrySearch').value)
    } else if(e.srcElement===document.getElementsByClassName('search')[0]) {
        console.log("searchForApp");
        applyFilters();
    }
    return false;
}}},true);


var post_url = "/post/";
var site_loaded = false;
var selected_country = filterParams[0][0].options[filterParams[0][0].selectedIndex].value;
var selected_operator = filterParams[0][1].options[filterParams[0][1].selectedIndex].value;
var selected_version = filterParams[0][2].options[filterParams[0][2].selectedIndex].value;
var searchfield_text = filterParams[1].value;
var postRequestJSON = JSON.parse('{"functionToCall" : "loadAppTray", "data" : {'
+ ' "Selected_country" : "'+ selected_country + '",'
+ ' "Selected_operator" : "'+ selected_operator + '",'
+ ' "Selected_version" : "'+ selected_version + '"'
+'}}');
console.log("MAIN – Sending post request with the following JSON:");
console.log(postRequestJSON);
server_post.post(post_url, postRequestJSON, function(appsToLoad) {
    console.log("MAIN – POST REQUEST SUCCESS!!! RESPONSE:");
    console.log(appsToLoad);
    showWebapps(appTray, appsToLoad);
});

updateFilterValues();




function applyFilters()
{
    var post_url = "/post/";
    var selected_country = filterParams[0][0].options[filterParams[0][0].selectedIndex].value;
    var selected_operator = filterParams[0][1].options[filterParams[0][1].selectedIndex].value;
    var selected_version = filterParams[0][2].options[filterParams[0][2].selectedIndex].value;
    var searchfield_text = filterParams[1].value;
    var postRequestJSON = JSON.parse('{"functionToCall" : "loadAppTray", "data" : {'
    + ' "Selected_country" : "'+ selected_country + '",'
    + ' "Selected_operator" : "'+ selected_operator + '",'
    + ' "Selected_version" : "'+ selected_version + '"'
    +'}}');
    console.log("applyFilters – Sending post request with the following JSON:");
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(appsToLoad) {
        console.log("applyFilters – Post request success. Applying filters and updating select filter rows:");
        showWebapps(appTray, appsToLoad);
        updateFilterValues()
    });

}
function searchApplyFilters(searchValue)
{
    var post_url = "/post/";
    var selected_country = filterParams[0][0].options[filterParams[0][0].selectedIndex].value;
    var selected_operator = filterParams[0][1].options[filterParams[0][1].selectedIndex].value;
    var selected_version = filterParams[0][2].options[filterParams[0][2].selectedIndex].value;
    var searchfield_text = filterParams[1].value;
    var postRequestJSON = JSON.parse('{"functionToCall" : "loadAppTray", "data" : {'
    + ' "Selected_country" : "'+ selected_country + '",'
    + ' "Selected_operator" : "'+ selected_operator + '",'
    + ' "Selected_version" : "'+ selected_version + '",'
    + ' "Searchfield_text" : "'+ searchValue + '"'
    +'}}');
    console.log("applyFilters – Sending post request with the following JSON:");
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(appsToLoad) {
        console.log("applyFilters – Post request success. Applying filters and updating select filter rows:");
        showWebapps(appTray, appsToLoad);
        updateFilterValues()
    });

}


//================================== APP TRAY PAGE ========================================//
//input an app container + a json of webapps, and this func will display them in the container with proper nesting
function showWebapps(appTray, webapps) {
    var webAppsHTML = "";  //set webAppsHTML string to null, so we can += to it later
    for(var o= 0; o < webapps.length; o++){
        console.log("SHOW_WEBAPPS – Adding "+webapps[o].ModifiableName+" iconContainer to the HTML");
        webAppsHTML += "<div class='iconContainer' id='" + webapps[o].OriginalName + "'>";
            webAppsHTML += ("<div id='deleteIcon' ");
            webAppsHTML += (" onclick=\"deleteAppfromTray('"+ webapps[o].OriginalName +"')\"");
            webAppsHTML += ("></div>");
            webAppsHTML += ("<img id='icon' src='" + webapps[o].IconUrl + "'");
            webAppsHTML += (" onclick=\"swapOut('"+ webapps[o].OriginalName +"')\"");
            webAppsHTML += (" />");

            webAppsHTML += ("<div id='iconText'>");
                webAppsHTML += (webapps[o].ModifiableName + " Ultra");
            webAppsHTML += ("</div>");
        webAppsHTML += ("</div>");
    }

    webAppsHTML += "<div class='iconContainer'>"; //ADD ULTRA APP ICON
    webAppsHTML += ("<img id='icon' src='" + "/images/add_icon.png" +"'");
    webAppsHTML += (" onclick=\"showAddAppPopup()\"");
    webAppsHTML += (" />");
    webAppsHTML += '<div class="addAppPopup">'
        webAppsHTML += '<div class= "contents">'
        webAppsHTML += '</div>';
    webAppsHTML += '</div>';
    webAppsHTML += ("<div id='iconText'>");
    webAppsHTML += ("Create new Ultra App");
    webAppsHTML += ("</div>");
    webAppsHTML += ("</div>");
    appTray.innerHTML = (webAppsHTML);
}

function swapOut(appID)
{
    console.log("SWAPOUT – Swapping out app tray for single ultra app view...");
    console.log("SWAPOUT – Current filter status: ");
    console.log(filterParams);
    console.log("SWAPOUT – Figuring out app info based off app ID and current filter status...");

    // Sending rest request for a specific ultra app
    var url = "/post/";
    var selected_country = filterParams[0][0].options[filterParams[0][0].selectedIndex].value;
    var selected_operator = filterParams[0][1].options[filterParams[0][1].selectedIndex].value;
    var selected_version = filterParams[0][2].options[filterParams[0][2].selectedIndex].value;
    var searchfield_text = filterParams[1].value;
    var app_name = appID;

    var postRequestJSON = JSON.parse('{"functionToCall" : "appView", "data" : {'
    + ' "Selected_country" : "'+ selected_country + '",'
    + ' "Selected_operator" : "'+ selected_operator + '",'
    + ' "Selected_version" : "'+ selected_version + '",'
    + ' "App_name" : "'+ app_name + '"'
    +'}}');

    server_post.post(post_url, postRequestJSON, function(app) {
         //set webAppHTML string to null, so we can += to it later
        console.log("SWAPOUT – Adding "+app.ModifiableName+" app to the HTML");
        window.history.pushState("", "", '/ultra/' + app.OriginalName);
        var swapinHTML =  "<hr>";
        swapinHTML += generateAppDetailsHTML(app);
        swapOutContainer.innerHTML = swapinHTML;
        document.getElementById('header').children[1].innerHTML =  app.ModifiableName + '<span id="smallerText"> Ultra</span>';
        console.log("SWAPOUT – Successfully swapped out html ");
    });
}

//======================================= READ/EDIT/WRITE ========================================//
function deleteAppfromTray(appID)
{
    console.log("DELETE_APP_FROM_TRAY – Deleting " + appID + "...");
    console.log("DELETE_APP_FROM_TRAY – CURRENT FILTER STATUS:");
    console.log(filterParams);
    deleteUltraApp(filterParams, appID); //writes to DB
    document.getElementById(appID).remove(); //this should be changed later. Need to update DB and refresh showWebapps_Old() again.
}
function deleteUltraApp(filterParameters, appID)
{
    console.log("DELETE_ULTRA_APP – Deleting " + appID + " for current filter setting. (Not implemented yet...)");
}
function submitNewApp(form){
    console.log("SUBMIT_NEW_APP – Submitting new app form... Form: ")
    console.log(form);
    console.log("SUBMIT_NEW_APP – Filter Status: ")
    console.log(filterParams);
    console.log("SUBMIT_NEW_APP – Taking the popup form + filter status and adding app for current filter configuration... ")
    addUltraApp(filterParams, form); //writes to DB ->> new Add App View should use user-specified filterParams within the Add App view, not the appTray filters.
    console.log("SUBMIT_NEW_APP – Closing popup window...")
    closeAddAppPopup();
    //add App to app tray//showWebapps_Old() again?
}
function addUltraApp(filterParameters, addAppForm)
{
    console.log("ADD_ULTRA_APP – Adding " + addAppForm.children[0].value + " Ultra for the current filter configuration. (Not implemented yet...)");
}

//================================== "ADD APP" POPUP WINDOW ========================================//
function showAddAppPopup(){ //shows Add App popup window
    var addAppPopup = document.getElementsByClassName("addAppPopup")[0];
    addAppPopup.classList.toggle("show");

    var popupHTML = '<div class ="closeButton"';
    popupHTML += (" onclick=\"closeAddAppPopup()\"></div>");
    popupHTML += generateAddAppPopupInputFields();

    var addAppPopupContents = addAppPopup.children[0];
    addAppPopupContents.innerHTML = popupHTML;
    console.log("ADD_APP_WINDOW – Showed popup window");
}
function closeAddAppPopup(){ //closes the AddApp Popup window
    var addAppPopup = document.getElementsByClassName("addAppPopup")[0];
    addAppPopup.classList.toggle("show");
    console.log("CLOSE_ADD_APP_POPUP – Closed popup window.");
}

//================================== HTML GENERATOR FUNCTIONS ========================================//
function generateAppDetailsHTML(app) //Responsible for generating app details HTML (in swapout func)
{
    console.log("GENERATE_APP_HTML – Recieved the following app");
    console.log(app);
    var webAppHTML = "";
    console.log("GENERATE_APP_HTML – Generating " + app.ModifiableName + " html...")
    webAppHTML += '<div class ="webApp">';

        webAppHTML += "<div class='row'>";
            webAppHTML += "<div class='rowDescription'>";
                webAppHTML += "Name";
            webAppHTML += ("</div>");
            webAppHTML += "<div class='rowValue'>";
                webAppHTML += app.ModifiableName;
            webAppHTML += ("</div>");
            webAppHTML += "<div class='edit'></div>";
        webAppHTML += "</div>";

        webAppHTML += "<div class='row'>";
            webAppHTML += "<div class='rowDescription'>";
                webAppHTML += "Rank";
            webAppHTML += ("</div>");
            webAppHTML += "<div class='rowValue'>";
                webAppHTML += app.Rank;
            webAppHTML += ("</div>");
            webAppHTML += "<div class='edit'></div>";
        webAppHTML += "</div>";

        webAppHTML += "<div class='row'>";
            webAppHTML += "<div class='rowDescription'>";
                webAppHTML += "Webapp Link";
            webAppHTML += ("</div>");
            webAppHTML += "<div class='rowValue'>";
                webAppHTML += extractRootDomain(app.homeURL) + "/...";
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
                webAppHTML += app.IconUrl;
                webAppHTML += "<div class='rowImage' style='background-image: url(\"../" + app.IconUrl + "\"); background-repeat: no-repeat; background-size:100%;'>";
                webAppHTML += ("</div>");
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


function generateAddAppPopupInputFields(){ //AddApp Popup window helper function
    var addAppViewHTML = '<form id="addAppForm" onsubmit="submitNewApp(this); return false"><div id="appConfig">';
    addAppViewHTML += '<input type="text" placeholder="Ultra App Name" name="name">';
    addAppViewHTML += '<input type="text" placeholder="Ultra App Rank" name="rank">';
    addAppViewHTML += '<input type="text" placeholder="Webapp Link" name="homeUrl">';
    addAppViewHTML += '<div id ="addAppEnabledFeatures">';
        addAppViewHTML += 'Default Enabled Features ';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="savings" checked />';
            addAppViewHTML += '<label for="savings">Savings</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="privacy" checked />';
            addAppViewHTML += '<label for="privacy">Privacy</label>';
        addAppViewHTML += '</div>'
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="adBlock" checked />';
            addAppViewHTML += '<label for="adBlock">Adblock</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="noImages" checked />';
            addAppViewHTML += '<label for="noImages">No Images</label>';
        addAppViewHTML += '</div>';
    addAppViewHTML += '</div>';
    addAppViewHTML += '<div id ="addAppHiddenFeatures">';
        addAppViewHTML += 'Default Hidden Features ';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenFeatures" value="savings" checked />';
            addAppViewHTML += '<label for="savings">Savings</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenFeatures" value="privacy" checked />';
            addAppViewHTML += '<label for="privacy">Privacy</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenFeatures" value="adBlock" checked />';
            addAppViewHTML += '<label for="adBlock">Adblock</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenFeatures" value="noImages" checked />';
            addAppViewHTML += '<label for="noImages">No Images</label>';
        addAppViewHTML += '</div>';
    addAppViewHTML += '</div>';
    addAppViewHTML += '<input type="text" placeholder="Native App Link(s)" name="nativeApps">';
    addAppViewHTML += '<input type="text" placeholder="Icon URL Link" name="iconUrl">';
    addAppViewHTML += '</div>';
    addAppViewHTML += '<div id="configurationMapping">';
        addAppViewHTML += '<input class="search" id="countrySearch" type="text" placeholder="Search..">'
        addAppViewHTML += '<div class="countrySearchResults"><div class="rowValue">ALL COUNTRIES<div class="rowImage" style="background-image: url(\'/images/arrow_drop_down.svg\'); background-repeat: no-repeat; background-size:100%;"></div></div></div>'
    addAppViewHTML += '</div>';
    addAppViewHTML += '<input type="submit" value="Submit"></form>';
    return addAppViewHTML;
}
function displayCountrySearchResults(countrySearchFieldText){
    console.log("displayCountrySearchResults – User input: " + countrySearchFieldText);
    var countrySearchResults = document.getElementsByClassName("countrySearchResults")[0];
    countryBubbleExists = false;
    for(var i = 0; i < countrySearchResults.children.length; i++) {
        if(countrySearchResults.children[i].textContent===countrySearchFieldText) {
            countryBubbleExists = true;
            console.log("displayCountrySearchResults – Country bubble for " +countrySearchFieldText +" already exits!");
            break;
        }
    }
    if(countrySearchFieldText != ""  && !countryBubbleExists)//only want to do things if textfield isn't empty, and bubble doens't already exist
    {
        console.log("displayCountrySearchResults – getting country by name: "+ countrySearchFieldText);
        getCountryByName(countrySearchFieldText, function(country){
            console.log("displayCountrySearchResults – GOT COUNTRY");
            if(country.name != "")//only want to do things if real country
            {
                if(countrySearchResults.children[0].textContent === "ALL COUNTRIES")
                {
                    console.log("displayCountrySearchResults – 'All countries' bubble detected, deleting it")
                    console.log("displayCountrySearchResults – Country returned from search is not null, adding bubble")
                    countrySearchResults.innerHTML = ""; //get rid of all countries if valid country
                    console.log("Contents:");
                    console.log(country);
                }

                var countryBubbleHTML = getCountryBubbleHTML(country);

                console.log("displayCountrySearchResults – Adding to the countrySearchResults html the following bubble:");
                console.log(countryBubbleHTML);
                countrySearchResults.innerHTML += countryBubbleHTML;
            }
        });
    }
}
function getCountryBubbleHTML(country){
    var returnHTML = "";
    returnHTML += '<div class="rowValue" id="'+country.Country_ID+'">'+country.name+'<div class="rowImage" onClick="toggleCountryBubble(this.parentElement)" style="background-image: url(\'/images/arrow_drop_down.svg\'); background-repeat: no-repeat; background-size:100%;"></div></div>'
    return returnHTML;
}
function toggleCountryBubble(countryBubble) {
    console.log("toggleCountryBubble – BUBBLE CLICKED:");
    console.log(countryBubble);

    countryBubble.classList.toggle("wide");
    if(!countryBubble.classList.contains("wide")) { //no longer expanded
        var operatorElement;
        for (var i = 0; i < countryBubble.childNodes.length; i++) {
            if (countryBubble.childNodes[i].className == 'operators') {
              operatorElement = countryBubble.childNodes[i];
              break;
            }
        }
        operatorElement.remove();

    }
    else { //now Expanded
        var postRequestJSON = JSON.parse('{"functionToCall" : "getOperatorsByCountryID", "data" : {'
        + ' "Country_ID" : "'+ countryBubble.id + '"'
        +'}}');

        server_post.post(post_url, postRequestJSON, function(operators) {
            console.log("toggleCountryBubble – Recieved the following JSON: ");
            console.log(operators);
                var html = "";
                html+= ("<div class = 'operators'>");
                for(var i = 0; i < operators.operatorRows.length; i++){
                    var operator = operators.operatorRows[i];
                    html+= (operator.Operator_Name + "<br>");
                }
                html+= ("</div>");
                countryBubble.innerHTML += html;
        });
    }
}

function getCountryByName(countryName, functionUsingCountry){
    var country = {"hi" : "hello"};

    var postRequestJSON = JSON.parse('{"functionToCall" : "getCountryByName", "data" : {'
    + ' "Country_Name" : "'+ countryName + '"'
    +'}}');

    server_post.post(post_url, postRequestJSON, function(countryRow) {
        console.log("getCountryByName – Recieved the following JSON: ");
        console.log(countryRow);
        functionUsingCountry(countryRow);
    });
}


//================================== HELPER FUNCTIONS ========================================//
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

function fetchDB() { //responsible for fetching database object
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
            console.log('FETCH_DB – Network request failed with response ' + response.status + ': ' + response.statusText);
        }
    });
    return db;
}
