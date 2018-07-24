//initialization
var cms_database = fetchDB(); //fetch db object (not used currentlys)
var server_get = new restRequest();
var server_post = new postRequest();
var appTray = document.getElementById("allicons");
var filterParams = [selects, searchField];
var swapOutContainer = document.getElementById("swapOutContainer");
var globalViewButton = document.getElementById('globalViewButton');
var globalViewDataJSON = null;

//each int in array is mapped to corresponding appConfig
var globalConfigArray = [];

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

//initialization with server post requests
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
    applyFilters();
});



updateFilterValues();
updateGlobalConfigArray();

function updateGlobalConfigArray(){
    console.log("updateGlobalConfigArray – Getting all apps from server")
    var postRequestJSON = JSON.parse('{"functionToCall" : "getAllAppConfigs", "data" : {}}');
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(allAppConfigs) {
        console.log("updateGlobalConfigArray – Server Response:");
        console.log(allAppConfigs);
        globalConfigArray = [];
        var dummyConfig = {"dummy" : "dummy"};
        globalConfigArray.push(dummyConfig);
        for(var i = 0; i < allAppConfigs.appConfigs.length; i++){
            var appConfig = allAppConfigs.appConfigs[i];
            globalConfigArray.push(appConfig);
        }
        console.log("updateGlobalConfigArray – updated array:");
        console.log(globalConfigArray);
    });
}


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
        console.log("SHOW_WEBAPPS – Adding "+webapps[o].ModifiableName+" iconContainer to the HTML | MAXGO");
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

function toggleGlobalView(){
    globalViewButton.classList.toggle('globalViewON');
    filters.classList.toggle("hidden");
    if(globalViewButton.classList.contains('globalViewON')){ //GLOBAL VIEW IS ON
        updateGlobalConfigArray();
        var swapinHTML =  "<hr>";
        console.log("toggleGlobalView – Global view turned on, requesting globalData from server...");
        var postRequestJSON = JSON.parse('{"functionToCall" : "globalView", "data" : {}}');
        swapOutContainer.innerHTML = swapinHTML;
        server_post.post(post_url, postRequestJSON, function(globalData) {
            console.log("toggleGlobalView – Success! Server returned:");
            console.log(globalData);
            generateGlobalViewHTML(globalData);
        });
    } else {
        swapOutContainer.innerHTML = '<main><div id ="appTray"></div></main>';
        swapOutContainer.children[0].children[0].appendChild(appTray);
        applyFilters();
    }
}
function generateGlobalViewHTML(globalData){
    globalViewDataJSON = globalData;
    var globalView = document.createElement('div');
    globalView.className = 'globalView';
    swapOutContainer.appendChild(globalView);
    for(var i = 0; i < globalViewDataJSON.GlobalDataApps.length; i++){
        var app = document.createElement('div');
        app.className = 'globalViewApp';
        app.classList.add('collapsed');
        var appTitle = document.createElement('div');
        appTitle.className = 'description';
        var downArrow = document.createElement('div');
        downArrow.innerHTML = '<div onclick="toggleGlobalApp(this.parentElement)" style="background-image: url(\'/images/arrow_drop_down.svg\'); background-repeat: no-repeat; background-size:100%;"></div>'
        downArrow = downArrow.children[0];
        downArrow.className = "rowImage";
        var appContents = document.createElement('div');
        appContents.className = 'globalViewAppContents';
        appTitle.innerText = globalViewDataJSON.GlobalDataApps[i].OriginalName;
        var appConfigs = document.createElement('div');
        appConfigs.className = "globalViewAppConfigs";
        for(var o = 0; o < globalViewDataJSON.GlobalDataApps[i].ConfigNumbers.length; o++){
            var appConfig = document.createElement('div');
            var configNumber = globalViewDataJSON.GlobalDataApps[i].ConfigNumbers[o];

            appConfig.innerText = (configNumber);
            appConfig.className = "globalViewAppConfig";
            setConfigHover(appConfig, configNumber);
            appConfigs.appendChild(appConfig);
        }

        app.appendChild(appTitle);
        app.appendChild(downArrow);
        app.appendChild(appConfigs);
        app.appendChild(appContents);
        globalView.appendChild(app);
    }
}
function setConfigHover(appConfig, configNumber){
    appConfig.onmouseover =  function() {showAppConfigOnHover(this, configNumber)};
    appConfig.onmouseout = function() {hideAppConfigOnHover(this, configNumber)};
}
function showAppConfigOnHover(appConfig, configNumber) {
    console.log(appConfig);
    var appConfigHoverContents = document.createElement('div');
    appConfigHoverContents.className = 'appConfigHoverContent';
    appConfigHoverContents.innerText = generateAppConfigHoverContents(configNumber);
    console.log("showAppConfigOnHover – hoverContents innerHTML = " + appConfigHoverContents.innerText);
    var configFeaturedLocs = document.createElement('div');
    configFeaturedLocs.className = 'configFeaturedLocations';
    configFeaturedLocs.innerHTML = '<div class=\'loading\'></div>';
    configFeaturedLocs = configFeaturedLocs.children[0];
    appConfigHoverContents.appendChild(configFeaturedLocs);

    console.log("showAppConfigOnHover – Getting featuredLocations for config " + configNumber)
    var postRequestText = '{"functionToCall" : "getFeaturedLocations", "data" : {'
        + ' "Config_ID" : "'+ configNumber + '"'
        +'}}';
    var postRequestJSON = JSON.parse(postRequestText);
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(featuredLocations) {
        console.log("showAppConfigOnHover – server responded with:");
        console.log(featuredLocations);
        var newHTML = 'Featured Locations : ';
        for(i in featuredLocations) {
            newHTML += (featuredLocations[i] + ", ");
        }
        var parent = configFeaturedLocs.parentElement;
        if(document.getElementsByClassName('loading')[0]!=null){
            document.getElementsByClassName('loading')[0].remove();
        }
        parent.innerHTML+= newHTML;
    });


    appConfigHoverContents.classList.toggle('hidden');
    if(appConfig.children[0] != null){
        appConfig.children[0].remove();
    }

    appConfig.prepend(appConfigHoverContents);
    appConfig.children[0].classList.toggle('hidden');
}


function hideAppConfigOnHover(appConfig) {
    appConfig.children[0].classList.toggle('hidden');
}
function generateAppConfigHoverContents(configNumber){
    console.log("generateAppConfigHoverContents – Setting appconfig contents for config #" + configNumber);
    var appConfig = globalConfigArray[configNumber];
    var returnString = "";
    for (var element in appConfig) {
        var val = appConfig[element];
        returnString+=(element + " : " + val + "\n");
    }
    return returnString;
}

function toggleGlobalApp(appElement) {
    console.log("toggleGlobalApp – \t\tElement clicked is");
    console.log(appElement);
    var appOriginalName = appElement.children[0].innerText;
    appElement.classList.toggle('collapsed');
    if(appElement.classList.contains('collapsed')){ //collapse app
        appElement.children[3].innerHTML="";
    } else { //expand app
        var loading = document.createElement('div');
        loading.className = 'loading';
        appElement.children[3].prepend(loading);
        var postRequestJSON = JSON.parse('{"functionToCall" : "globalView", "data" : {'
        + ' "App_OriginalName" : "'+ appOriginalName + '"'
        +'}}');
        server_post.post(post_url, postRequestJSON, function(appData) {
            console.log("toggleGlobalView – Success! Server returned:");
            console.log(appData);
            appElement.children[3].children[0].remove(); //removes loading;
            for(var i = 0; i < appData.GlobalDataCountries.length; i++) {
                var country = document.createElement('div');
                country.className = 'globalViewCountry';
                country.classList.add('collapsed');
                var countryTitle = document.createElement('div');
                countryTitle.className = 'description';
                countryTitle.innerText = appData.GlobalDataCountries[i].name;
                countryTitle.setAttribute("id", appData.GlobalDataCountries[i].Country_ID);
                var downArrow = document.createElement('div');
                downArrow.innerHTML = '<div onclick="toggleGlobalCountry(this.parentElement)" style="background-image: url(\'/images/arrow_drop_down.svg\'); background-repeat: no-repeat; background-size:100%;"></div>'
                downArrow = downArrow.children[0];
                downArrow.className = "rowImage";
                var countryContents = document.createElement('div');
                var appConfigs = document.createElement('div');
                appConfigs.className = "globalViewAppConfigs";
                for(var o = 0; o < appData.GlobalDataCountries[i].ConfigNumbers.length; o++){
                    var configNumber = appData.GlobalDataCountries[i].ConfigNumbers[o];
                    var appConfig = document.createElement('div');
                    appConfig.className = "globalViewAppConfig";
                    appConfig.innerText = (appData.GlobalDataCountries[i].ConfigNumbers[o]);
                    setConfigHover(appConfig, configNumber);

                    appConfigs.appendChild(appConfig);
                }
                country.appendChild(countryTitle);
                country.appendChild(appConfigs);
                if(appData.GlobalDataCountries[i].operatorRows != null) {
                    country.appendChild(downArrow);
                }
                country.appendChild(countryContents);
                appElement.children[3].appendChild(country);
            }
        });
    }
}
function toggleGlobalCountry(countryElement) {
    console.log("toggleGlobalCountry – \t\tElement clicked is");
    console.log(countryElement);
    var countryName = countryElement.children[0].innerText;
    var appName = countryElement.parentElement.parentElement.children[0].innerText;
    var country_ID = countryElement.children[0].id;
    countryElement.classList.toggle('collapsed');
    if(countryElement.classList.contains('collapsed')) {
        countryElement.children[3].innerHTML = "";
    } else {
        var loading = document.createElement('div');
        loading.className = 'loading';
        countryElement.children[3].prepend(loading);
        var postRequestJSON = JSON.parse('{"functionToCall" : "globalView", "data" : {'
        + ' "App_OriginalName" : "'+ appName + '",'
        + ' "Country_ID" : "'+ country_ID + '"'
        +'}}');
        console.log(postRequestJSON);
        server_post.post(post_url, postRequestJSON, function(countryData) {
            console.log("toggleGlobalCountry – \t\tFound country data:")
            countryElement.children[3].children[0].remove();
            console.log(countryData);
            for(var i = 0; i < countryData.operatorRows.length; i++){
                console.log(countryData.operatorRows[i]);
                var operator = document.createElement('div');
                operator.className = 'globalViewOperator';
                var operatorTitle = document.createElement('div');
                operatorTitle.className = 'description';
                operatorTitle.innerText = "(" + countryData.operatorRows[i].MCCMNC_ID + ") ";
                operatorTitle.innerText += countryData.operatorRows[i].Operator_Name;
                var operatorConfigs = document.createElement('div');
                operatorConfigs.className = "globalViewAppConfigs";

                for(var o =0; o <countryData.operatorRows[i].ConfigNumbers.length; o++) {
                    var operatorConfig = document.createElement('div');
                    operatorConfig.className = "globalViewAppConfig";
                    operatorConfig.innerText = countryData.operatorRows[i].ConfigNumbers[o];
                    setConfigHover(operatorConfig, countryData.operatorRows[i].ConfigNumbers[o]);
                    operatorConfigs.appendChild(operatorConfig);
                }


                operator.appendChild(operatorTitle);
                operator.appendChild(operatorConfigs);
                countryElement.children[3].appendChild(operator);
            }
        });
    }
}
function findGlobalViewCountry(appName, countryName){
    console.log("findGlobalViewCountry – \t\tSearching for country: " + appName + " | " + countryName);
    var appData = {};//findGlobalViewApp(appName);
    console.log("findGlobalViewCountry – \t\tFound app with name  " + appName + ": ");
    console.log(appData);
    for(var i = 0; i<appData.Countries.length; i++){
        if(appData.Countries[i].name === countryName){
            return appData.Countries[i];
        }
    }
    return null;
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
        // window.history.pushState("", "", '/ultra/' + app.OriginalName); //changes url without reloading page, might be useful
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
    console.log("SUBMIT_NEW_APP – Taking the popup form + filter status and adding app for current filter configuration... ")
    addUltraApp(form); //writes to DB ->> new Add App View should use user-specified filterParams within the Add App view, not the appTray filters.
    console.log("SUBMIT_NEW_APP – Closing popup window...")
    closeAddAppPopup();
    //add App to app tray//showWebapps_Old() again?
}
function addUltraApp(form)
{
    console.log("addUltraApp – Adding " + form.children[0].children[0].value + " Ultra for the current filter configuration. (Not implemented yet...)");
    var countriesList= "";
    var operatorsList = "";
    // console.log(configMappings.children[0].textContent);
    var configMappings = form.children[1].children[1];
    var existsEverywhere = false;
    if(configMappings.children[0].textContent === "ALL COUNTRIES"){ //insert app globally
        console.log("addUltraApp – ALL COUNTRIES DETECTED");
        existsEverywhere = true;
    }
    else {
        for(var i = 0; i < configMappings.children.length; i++) { //iterates through all bubbles
            var allOperatorsChecked = true;

            if(configMappings.children[i].children.length > 2) { //operators specified
                for(var o = 0; o < configMappings.children[i].children[2].children.length; o++){ //iterates through all operators, looking 4 unchecked
                    if(!configMappings.children[i].children[2].children[o].classList.contains("checked")) {
                        console.log(configMappings.children[i].children[2].children[o].textContent);
                        allOperatorsChecked = false;
                        console.log("addUltraApp – Not all operators in " + configMappings.children[i].children[2].children[o].textContent +" are checked, adding specified operators to operatorList." )
                    }
                }
            }
            if(allOperatorsChecked) {
                console.log("addUltraApp – Adding " + configMappings.children[i].children[0].textContent + "to list of countries");
                countriesList += ("\""+configMappings.children[i].id+"\"" + ", ");
            } else {
                for(var o = 0; o < configMappings.children[i].children[2].children.length; o++){ //iterates through all operators,
                    if(configMappings.children[i].children[2].children[o].classList.contains("checked")) {
                        console.log("addUltraApp – " + configMappings.children[i].children[2].children[o].textContent);
                        operatorsList += ("\""+configMappings.children[i].children[2].children[o].id+"\"" + ", ");
                    }
                }
            }
        }

        console.log("addUltraApp – Countries List: " + countriesList + " | Operators List: " + operatorsList );
    }
    countriesList = countriesList.replace(/,\s*$/, "");
    operatorsList = operatorsList.replace(/,\s*$/, "");
    var json = ('{"functionToCall" : "addNewConfig", "data" : {'
        + ' "App_ModifiableName" : "'+ form.children[0].children[0].value+ '",'
        + ' "App_OriginalName" : "'+ form.children[0].children[0].value.toLowerCase() + '",'
        + ' "App_Rank" : "'+ form.children[0].children[1].value + '",'
        + ' "App_HomeURL" : "'+ form.children[0].children[2].value + '",'
        + ' "App_NativeURL" : "'+ form.children[0].children[5].value + '",'
        + ' "App_IconURL" : "'+ form.children[0].children[6].value + '",'
        + ' "App_ExistsEverywhere" : '+ existsEverywhere + ','
        + ' "App_ConfigurationMappings" : { '
            + ' "Countries" : ['
                + countriesList
            + '], "Operators" : ['
                + operatorsList
            + '], "FeaturedLocations" : ['
                + '"maxGo", "homescreen", "max", "folder"'
            + ']'
        +'}'
    +'}}');
    console.log(json);
    json = JSON.parse(json);
    console.log("addUltraApp – Sending post request with the following JSON:");
    console.log(json);
    server_post.post(post_url, json, function(message) {
        console.log("addUltraApp – POST REQUEST SUCCESS!!! RESPONSE:");
        console.log(message);
        applyFilters();
    });
}

//================================== "ADD APP" POPUP WINDOW ========================================//
function showAddAppPopup(){ //shows Add App popup window
    var addAppPopup = document.getElementsByClassName("addAppPopup")[0];
    addAppPopup.classList.toggle("show");
    document.getElementsByTagName("BODY")[0].classList.toggle("overflow-hidden");
    var popupHTML = '<div class ="closeButton"';
    popupHTML += (" onclick=\"closeAddAppPopup()\"></div>");
    popupHTML += generateAddAppPopupInputFields();
    var addAppPopupContents = addAppPopup.children[0];
    addAppPopupContents.innerHTML = popupHTML;

    var title = document.createElement("div");
    title.innerHTML = "<div class = 'popupTitle'>Create new Ultra App Configuration</div>";
    addAppPopupContents.parentElement.appendChild(title);

    console.log("ADD_APP_WINDOW – Showed popup window");
}
function closeAddAppPopup(){ //closes the AddApp Popup window
    var addAppPopup = document.getElementsByClassName("addAppPopup")[0];
    addAppPopup.classList.toggle("show");
    document.getElementsByTagName("BODY")[0].classList.toggle("overflow-hidden");
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
                webAppHTML += "Icon URL";
            webAppHTML += ("</div>");
            webAppHTML += "<div class='rowValue'>";
                webAppHTML += app.IconUrl;
                webAppHTML += "<div class='rowImage' style='background-image: url(\"../" + app.IconUrl + "\"); background-repeat: no-repeat; background-size:100%;'>";
                webAppHTML += ("</div>");
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
                webAppHTML += "Featured Location";
            webAppHTML += ("</div>");
            webAppHTML += "<div class='rowValue'>";
                webAppHTML += app.featuredLocationName;
            webAppHTML += ("</div>");
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
        addAppViewHTML += '<div class="countrySearchResults"><div class="rowValue">ALL COUNTRIES</div></div>'
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
    returnHTML += '<div class="rowValue" id="'+country.Country_ID+'"><div class="countryBubbleTitle">'+country.name+'</div><div class="rowImage" onClick="toggleCountryBubble(this.parentElement)" style="background-image: url(\'/images/arrow_drop_down.svg\'); background-repeat: no-repeat; background-size:100%;"></div></div>'
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
                    // if(operator.Operator_Name!=""){ //should probably get rid of this check, and get rid of null entries in DB
                        html+= ("<div value='"+operator.Operator_Name+"' id ='"+operator.MCCMNC_ID+"' class='operator checked' onClick='toggleOperator(this);'>"+operator.Operator_Name + " (" +operator.MCCMNC_ID+")</div>");
                    // }
                }
                html+= ("</div>");
                countryBubble.innerHTML += html;
        });
    }
}
function toggleOperator(operator){
    operator.classList.toggle("checked");
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
