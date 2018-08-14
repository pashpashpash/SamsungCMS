//initialization
var cms_database = fetchDB(); //fetch db object (not used currentlys)
var server_get = new restRequest();
var server_post = new postRequest();
var appTray = document.getElementById("allicons");
var filterParams = [selects, searchField];
var swapOutContainer = document.getElementById("swapOutContainer");
var globalViewButton = document.getElementById('globalViewButton');
var settingsViewButton = document.getElementById('settingsViewButton');
//each int in array is mapped to corresponding appConfig
var globalConfigArray = [];
var post_url = "/post/";


var body = document.getElementsByTagName("BODY")[0];
body.classList.add("hidden");
checkLoggedIn();
function checkLoggedIn() {
    console.log("checking if user logged in..")
    var postRequestJSON = JSON.parse('{"functionToCall" : "checkIfLoggedIn", "data" : {}}');
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(loggedIn) {
        console.log("checkLoggedIn – " + loggedIn);
        if(loggedIn === true) {
            body.classList.remove("hidden");
        }
        else {
            body.innerHTML = '<div id ="loginPage"><form method="post" action="/post/login"><input type="text" id="name" name="name" placeholder="username"><input type="password" id="password" name="password"  placeholder="password"><div><button id = "login" type="submit">Login</button></div></form></div>';
            body.classList.remove("hidden");
        }
    });
}


window.addEventListener('keydown',function(e){if(e.keyIdentifier=='U+000A'||e.keyIdentifier=='Enter'||e.keyCode==13){if(e.target.nodeName=='INPUT'&&e.target.type=='text'){
    e.preventDefault();
    if(e.srcElement===document.getElementById('countrySearch'))
    {
        if(!e.srcElement.parentElement.children[1]) {
            console.log("searchForCountry");
            displayCountrySearchResults(document.getElementById('countrySearch').value)
        }
    } else if(e.srcElement===document.getElementsByClassName('search')[0]) {
        console.log("searchForApp");
        applyFilters();
    } else if (e.srcElement===document.getElementById('operatorSearch')) {
        if(!e.srcElement.parentElement.children[1]) {
            console.log("searchForOperator");
            displayOperatorSearchResults(document.getElementById('operatorSearch').value);
        }
    }
    return false;
}}},true);

//initialization with server post requests


var site_loaded = false;
var selected_country = filterParams[0][0].options[filterParams[0][0].selectedIndex].value;
var selected_operator = filterParams[0][1].options[filterParams[0][1].selectedIndex].value;
var searchfield_text = filterParams[1].value;
var postRequestJSON = JSON.parse('{"functionToCall" : "loadAppTray", "data" : {'
+ ' "Selected_country" : "'+ selected_country + '",'
+ ' "Selected_operator" : "'+ selected_operator + '"'
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
    var searchfield_text = filterParams[1].value;
    var postRequestJSON = JSON.parse('{"functionToCall" : "loadAppTray", "data" : {'
    + ' "Selected_country" : "'+ selected_country + '",'
    + ' "Selected_operator" : "'+ selected_operator + '"'
    +'}}');
    console.log("applyFilters – Sending post request with the following JSON:");
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(appsToLoad) {
        console.log("applyFilters – Post request success. Applying filters and updating select filter rows:");
        showWebapps(appTray, appsToLoad);
    });

}

function searchApplyFilters(searchValue)
{
    var post_url = "/post/";
    var selected_country = filterParams[0][0].options[filterParams[0][0].selectedIndex].value;
    var selected_operator = filterParams[0][1].options[filterParams[0][1].selectedIndex].value;
    var searchfield_text = filterParams[1].value;
    var postRequestJSON = JSON.parse('{"functionToCall" : "loadAppTray", "data" : {'
    + ' "Selected_country" : "'+ selected_country + '",'
    + ' "Selected_operator" : "'+ selected_operator + '",'
    + ' "Searchfield_text" : "'+ searchValue + '"'
    +'}}');
    console.log("applyFilters – Sending post request with the following JSON:");
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(appsToLoad) {
        console.log("applyFilters – Post request success. Applying filters and updating select filter rows:");
        showWebapps(appTray, appsToLoad);
    });

}


//================================== APP TRAY PAGE ========================================//
//input an app container + a json of webapps, and this func will display them in the container with proper nesting
function showWebapps(appTray, webapps) {
    if (webapps != null) {
        appTray.innerHTML = "";
        var webAppsHTML = "";  //set webAppsHTML string to null, so we can += to it later
        addAppTraySections(appTray); //creates sections inside of app tray for maxGlobal, max, and maxGo


        for(var o= 0; o < webapps.length; o++){
            addAppToTray(webapps[o], appTray);
        }

        webAppsHTML = "<div class='iconContainer'>"; //ADD ULTRA APP ICON
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

        var addIcon1 = document.createElement("div");
        addIcon1.innerHTML = webAppsHTML;
        addIcon1 = addIcon1.children[0];
        var addIcon2 = document.createElement("div");
        addIcon2.innerHTML = webAppsHTML;
        addIcon2 = addIcon2.children[0];
        var addIcon3 = document.createElement("div");
        addIcon3.innerHTML = webAppsHTML;
        addIcon3 = addIcon3.children[0];

        maxGlobalContent.appendChild(addIcon1);
        maxContent.appendChild(addIcon2);
        maxGoContent.appendChild(addIcon3);


    } else {
        appTray.innerHTML = "";
    }
}
function addAppTraySections(appTray) {
    var webAppsHTML = "";  //set webAppsHTML string to null, so we can += to it later
    var maxGlobal = document.createElement("div");
    maxGlobal.className = "appTraySection";
    maxGlobal.id = "maxGlobal";
    var maxGlobalDescription = document.createElement("div");
    maxGlobalDescription.className = "appTraySectionDescription";
    maxGlobalDescription.innerText = "Max Global";
    var maxGlobalContent = document.createElement("div");
    maxGlobalContent.className = "appTraySectionContent";
    maxGlobalContent.id = "maxGlobalContent";
    var max = document.createElement("div");
    max.className = "appTraySection";
    max.id = "max";
    var maxDescription = document.createElement("div");
    maxDescription.className = "appTraySectionDescription";
    maxDescription.innerText = "Max Preloaded";
    var maxContent = document.createElement("div");
    maxContent.className = "appTraySectionContent";
    maxContent.id = "maxContent";
    var maxGo = document.createElement("div");
    maxGo.className = "appTraySection";
    maxGo.id = "maxGo";
    var maxGoDescription = document.createElement("div");
    maxGoDescription.className = "appTraySectionDescription";
    maxGoDescription.innerText = "Max Go";
    var maxGoContent = document.createElement("div");
    maxGoContent.className = "appTraySectionContent";
    maxGoContent.id = "maxGoContent";

    maxGlobal.appendChild(maxGlobalDescription);
    max.appendChild(maxDescription);
    maxGo.appendChild(maxGoDescription);


    maxGlobal.appendChild(maxGlobalContent);
    max.appendChild(maxContent);
    maxGo.appendChild(maxGoContent);

    appTray.appendChild(maxGlobal);
    appTray.appendChild(max);
    appTray.appendChild(maxGo);
}
function addAppToTray(app, tray){

    console.log("SHOW_WEBAPPS – Adding "+app.ModifiableName+" iconContainer to the HTML");

    webAppsHTML = "<div class='iconContainer' id='" + app.OriginalName + "'>";
        webAppsHTML += ("<img id='icon' src='" + app.IconUrl + "'");
        webAppsHTML += (" onclick=\" window.location = '/configs/"+app.Config_ID+"';\"");
        webAppsHTML += (" />");

        webAppsHTML += ("<div id='iconText'>");
            webAppsHTML += (app.ModifiableName + " Ultra");
        webAppsHTML += ("</div>");
    webAppsHTML += ("</div>");
    var iconContainer = document.createElement("div")
    iconContainer.innerHTML = webAppsHTML;
    iconContainer = iconContainer.children[0];

    if(app.ProductName === "maxGlobal") {
        if(!appAlreadyExistsInSection(app, tray.children[0].children[1])) {
            tray.children[0].children[1].appendChild(iconContainer);
        }
    }
    if(app.ProductName === "max") {
        if(!appAlreadyExistsInSection(app, tray.children[1].children[1])) {
            tray.children[1].children[1].appendChild(iconContainer);
        }
    }
    if(app.ProductName === "maxGo") {
        if(!appAlreadyExistsInSection(app, tray.children[2].children[1])) {
        tray.children[2].children[1].appendChild(iconContainer);
        }
    }

}
function appAlreadyExistsInSection(app, appTraySection) {
    for(var i = 0; i<appTraySection.children.length; i++) {
        if(appTraySection.children[i].id === app.OriginalName) {
            return true;
        }
    }
    return false;
}
function toggleSettingsView(){
    settingsViewButton.classList.toggle('settingsViewON');
    filters.classList.toggle("hidden");
    if(settingsViewButton.classList.contains('settingsViewON')){ //SETTINGS VIEW IS ON
        var swapinHTML =  "<hr>";
        console.log("toggleSettingsView – Settings view turned on, requesting settingsData from server...");
        var postRequestJSON = JSON.parse('{"functionToCall" : "settingsView", "data" : {}}');
        swapOutContainer.innerHTML = swapinHTML;
        server_post.post(post_url, postRequestJSON, function(settingsData) {
            console.log("toggleSettingsView – Success! Server returned:");
            console.log(settingsData);
            generateSettingsViewHTML(settingsData);
        });
    } else {
        swapOutContainer.innerHTML = '<main><div id ="appTray"></div></main>';
        swapOutContainer.children[0].children[0].appendChild(appTray);
        applyFilters();
    }
}
function generateSettingsViewHTML(settingsData){
    var settingsView = document.createElement('div');
    settingsView.className = 'settingsView';


    var operatorGroups = settingsData.OperatorGroups;
    var countries = settingsData.Countries;
    var operatorSection = document.createElement('div');
    var countrySection = document.createElement('div');
    operatorSection.className = "manageOperators"
    countrySection.className = "addCountry"
    for(var i = 0; i < operatorGroups.length; i++) {
        var groupName = operatorGroups[i].operatorRows[0].Operator_Group_Name;
        var operatorGroup = document.createElement('div');
        operatorGroup.className = "settingsViewOperatorGroup"
        operatorGroup.classList.add("collapsed");
        var downArrow = document.createElement('div');
        downArrow.innerHTML = '<div onclick="toggleOperatorGroup(this.parentElement)" style="background-image: url(\'/images/arrow_drop_down.svg\'); background-repeat: no-repeat; background-size:100%;"></div>'
        downArrow = downArrow.children[0];
        downArrow.className = "rowImage";
        var operatorGroupTitle = document.createElement('div');
        operatorGroupTitle.className = "description";
        operatorGroupTitle.innerText = groupName;
        var operatorGroupContents = document.createElement('div');
        operatorGroupContents.className = 'settingsViewAppContents';
        operatorGroupContents.classList.add("hidden");
        for(var o = 0; o < operatorGroups[i].operatorRows.length; o++) {
            var operator = document.createElement('div');
            operator.className = "settingsViewOperator";
            var operatorDescription = document.createElement('div');
            operatorDescription.className = "description";
            operatorDescription.innerText = operatorGroups[i].operatorRows[o].Operator_Name + " | " + operatorGroups[i].operatorRows[o].MCCMNC_ID
            operator.id = operatorGroups[i].operatorRows[o].MCCMNC_ID;
            var deleteButton = document.createElement('div');
            deleteButton.innerHTML = '<div onclick="deleteOperatorFromGroup(this.parentElement)"></div>';
            deleteButton = deleteButton.children[0];
            deleteButton.innerText = "-";
            deleteButton.className = "deleteOperator";
            operator.appendChild(operatorDescription);
            operator.appendChild(deleteButton);
            operatorGroupContents.appendChild(operator);
        }
        var addOperatorToGroupBlock = document.createElement('div');
        addOperatorToGroupBlock.className = "addOperatorToGroup";
        addOperatorToGroupBlockContents = document.createElement('div');
        addOperatorToGroupBlockContents.className = "addOperatorToGroupContents";
        addOperatorToGroupBlockContents.classList.add("hidden");
        addOperatorToGroupBlock.classList.add("collapsed");
        var addOperatorToGroupTitle = document.createElement("div");
        addOperatorToGroupTitle.innerHTML = '<div onclick="addOperatorToGroup(this.parentElement)"></div>';
        addOperatorToGroupTitle = addOperatorToGroupTitle.children[0];
        addOperatorToGroupTitle.innerText = "+ Add New Operator";
        addOperatorToGroupTitle.className = "addOperatorToGroupTitle";
        addOperatorToGroupBlock.appendChild(addOperatorToGroupTitle);
        //insert contents here
        var input = document.createElement("input");
        input.type = "text";
        input.className = "addOperatorName"; // set the CSS class
        input.placeholder = "Operator Name";
        addOperatorToGroupBlockContents.appendChild(input);
        input = document.createElement("input");
        input.type = "text";
        input.className = "addMCCMNC"; // set the CSS class
        input.placeholder = "MCCMNC";
        addOperatorToGroupBlockContents.appendChild(input);
        input = document.createElement("input");
        input.type = "text";
        input.className = "addCountryID"; // set the CSS class
        input.placeholder = "Country ID (two character ID, should exist in [countries] table)";
        addOperatorToGroupBlockContents.appendChild(input);
        input = document.createElement("div");
        input.innerHTML = '<input type="submit" onclick="submitNewOperator(this.parentElement);" />';
        input = input.children[0];
        input.className = "addOperatorSubmit"; // set the CSS class
        addOperatorToGroupBlockContents.appendChild(input);
        //

        addOperatorToGroupBlock.appendChild(addOperatorToGroupBlockContents);
        operatorGroupContents.appendChild(addOperatorToGroupBlock);

        operatorGroup.appendChild(operatorGroupTitle);
        operatorGroup.appendChild(downArrow);
        operatorGroup.appendChild(operatorGroupContents);
        operatorSection.appendChild(operatorGroup);
    }
    var addOperatorBlock = document.createElement('div');
    addOperatorBlock.className = "addOperatorGroup";

    var addOperatorBlockTitle = document.createElement('div');
    addOperatorBlockTitle.innerHTML = '<div onclick="addOperatorGroup(this.parentElement)"></div>';
    addOperatorBlockTitle = addOperatorBlockTitle.children[0];
    addOperatorBlockTitle.className = "addOperatorGroupTitle";
    addOperatorBlockTitle.innerText = "+ Add New Group";
    addOperatorBlock.appendChild(addOperatorBlockTitle);

    addOperatorBlockContents = document.createElement('div');
    addOperatorBlockContents.className = "addOperatorToGroupContents";
    addOperatorBlockContents.classList.add("hidden");
    addOperatorBlock.classList.add("collapsed");
    //insert contents here
    var input = document.createElement("input");
    input.type = "text";
    input.className = "addOperatorGroupName"; // set the CSS class
    input.placeholder = "Operator Group Name";
    addOperatorBlockContents.appendChild(input);
    input = document.createElement("input");
    input.type = "text";
    input.className = "addOperatorGroupMCCMNC_ID"; // set the CSS class
    input.placeholder = "MCCMNC (need at least one operator in group)";
    addOperatorBlockContents.appendChild(input);
    input = document.createElement("input");
    input.type = "text";
    input.className = "addOperatorGroupOperator_Name"; // set the CSS class
    input.placeholder = "Operator Name (need at least one operator in group)";
    addOperatorBlockContents.appendChild(input);
    input = document.createElement("input");
    input.type = "text";
    input.className = "addOperatorGroupCountry_ID"; // set the CSS class
    input.placeholder = "Country ID (need at least one operator in group)";
    addOperatorBlockContents.appendChild(input);
    input = document.createElement("div");
    input.innerHTML = '<input type="submit" onclick="submitNewOperatorGroup(this.parentElement);" />';
    input = input.children[0];
    input.className = "addOperatorGroupSubmit"; // set the CSS class
    addOperatorBlockContents.appendChild(input);
    //
    addOperatorBlock.appendChild(addOperatorBlockContents);
    operatorSection.appendChild(addOperatorBlock);


    var allCountries = document.createElement('div');
    allCountries.className = "settingsViewCountry";
    var allCountriesText = document.createElement('div');
    allCountriesText.className = "description";
    allCountriesText.innerText = "All Countries";
    var downArrow = document.createElement('div');
    downArrow.innerHTML = '<div onclick="toggleAllCountries(this.parentElement)" style="background-image: url(\'/images/arrow_drop_down.svg\'); background-repeat: no-repeat; background-size:100%;"></div>'
    downArrow = downArrow.children[0];
    downArrow.className = "rowImage";
    var allCountriesContents = document.createElement('div');


    for(var i = 0; i<countries.length; i++){
        var countryName = countries[i].name;
        var country = document.createElement('div');
        country.className = "settingsViewCountry";
        country.innerText = countryName;
        country.id = countries[i].Country_ID;
        allCountriesContents.appendChild(country);
    }
    allCountriesContents.className = "allCountriesContents";
    allCountriesContents.classList.add("hidden");
    allCountries.appendChild(allCountriesText);
    allCountries.appendChild(downArrow);
    allCountries.appendChild(allCountriesContents);
    countrySection.appendChild(allCountries);

    var addCountry = document.createElement('div');
    addCountry.className = "settingsViewAddCountry";
    var addCountryText = document.createElement('div');
    addCountryText.className = "description";
    addCountryText.innerText = "Add Country";
    downArrow = document.createElement('div');
    downArrow.innerHTML = '<div onclick="toggleAddCountry(this.parentElement)">+</div>'
    downArrow = downArrow.children[0];
    downArrow.className = "rowImage";
    var addCountryContents = document.createElement('div');
    addCountryContents.className = "addCountryContents";
    addCountryContents.classList.add("hidden");
    addCountry.classList.add("collapsed");

    var input = document.createElement("input");
    input.type = "text";
    input.className = "addCountryName"; // set the CSS class
    input.placeholder = "Country Name";
    addCountryContents.appendChild(input);
    input = document.createElement("input");
    input.type = "text";
    input.className = "addCountryID"; // set the CSS class
    input.placeholder = "Country ID";
    addCountryContents.appendChild(input);
    input = document.createElement("input");
    input.type = "text";
    input.className = "addCountryMCC"; // set the CSS class
    input.placeholder = "Country MCC";
    addCountryContents.appendChild(input);
    input = document.createElement("div");
    input.innerHTML = '<input type="submit" onclick="submitNewCountry(this.parentElement);" />';
    input = input.children[0];
    input.className = "addCountrySubmit"; // set the CSS class
    addCountryContents.appendChild(input);

    addCountry.appendChild(addCountryText);
    addCountry.appendChild(downArrow);
    addCountry.appendChild(addCountryContents);
    countrySection.appendChild(addCountry);

    settingsView.appendChild(operatorSection);
    settingsView.appendChild(countrySection);
    swapOutContainer.appendChild(settingsView);
}

function toggleAllCountries(allCountriesContainer) {
    allCountriesContainer.children[2].classList.toggle("hidden");
}
function toggleAddCountry(addCountryContainer) {
    addCountryContainer.children[2].classList.toggle("hidden");
    addCountryContainer.classList.toggle("collapsed");
}
function toggleOperatorGroup(operatorGroup){
    console.log("toggleOperatorGroup – operator group clicked:");
    console.log(operatorGroup);
    var operatorGroupName = operatorGroup.children[0].innerText;
    operatorGroup.classList.toggle('collapsed');
    if(operatorGroup.classList.contains('collapsed')){ //collapse app
        operatorGroup.children[2].classList.add("hidden");
    } else { //expand app
        operatorGroup.children[2].classList.remove("hidden");
    }
}

//finish these
function submitNewCountry(form){
    console.log("submitNewcountry – submitting new Country:");
    var countryName = form.children[0].value;
    var countryID = form.children[1].value;
    var countryMCC = form.children[2].value;
    console.log("submitNewcountry – " + countryName + " | " + countryID + " | " + countryMCC);
    var postRequestJSON = JSON.parse('{"functionToCall" : "submitNewCountry", "data" : {'
    + ' "Country_ID" : "'+ countryID + '",'
    + ' "Country_Name" : "'+ countryName + '",'
    + ' "Country_MCC" : "'+ countryMCC + '"'
    +'}}');
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(misc) {
        console.log("submitNewCountry – " + misc);
        if(misc === "success") {
            toggleSettingsView();
            toggleSettingsView();
        }
    });
}
function addOperatorToGroup(operator) {
    console.log("addOperatorToGroup – adding operator to group")
    var addOperatorContents = operator.children[1];
    addOperatorContents.classList.toggle("hidden");
    operator.classList.toggle("collapsed");
}
function submitNewOperator(operator) {
    var groupName = operator.parentElement.parentElement.parentElement.children[0].innerText;
    console.log(groupName);
    console.log("submitNewOperator – submitting operator to group " + groupName);

    var OperatorName = operator.children[0].value;
    var MCCMNC_ID =  operator.children[1].value;
    var Country_ID =  operator.children[2].value;
    console.log("submitNewOperator – " + OperatorName + " | "+ MCCMNC_ID + " | " + Country_ID);

    var postRequestJSON = JSON.parse('{"functionToCall" : "submitNewOperator", "data" : {'
    + ' "OperatorName" : "'+ OperatorName + '",'
    + ' "MCCMNC_ID" : "'+ MCCMNC_ID + '",'
    + ' "Country_ID" : "'+ Country_ID + '",'
    + ' "Operator_Group_Name" : "'+ groupName + '"'
    +'}}');
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(misc) {
        console.log("submitNewOperator – " + misc);
        if(misc === "success") {
            toggleSettingsView();
            toggleSettingsView();
        }
    });
}
function addOperatorGroup(operatorGroup) {
    console.log("addOperatorGroup – adding operator group")
    console.log(operatorGroup);
    var operatorGroupContents = operatorGroup.children[1];
    operatorGroupContents.classList.toggle("hidden");
    operatorGroup.classList.toggle("collapsed");
}
function submitNewOperatorGroup(form) {
    console.log("submitNewOperatorGroup – submitting operator group")
    console.log(form);
    var groupName = form.children[0].value;
    var MCCMNC_ID = form.children[1].value;
    var Operator_Name = form.children[2].value;
    var Country_ID = form.children[3].value;

    var postRequestJSON = JSON.parse('{"functionToCall" : "submitNewOperator", "data" : {'
    + ' "OperatorName" : "'+ Operator_Name + '",'
    + ' "MCCMNC_ID" : "'+ MCCMNC_ID + '",'
    + ' "Country_ID" : "'+ Country_ID + '",'
    + ' "Operator_Group_Name" : "'+ groupName + '"'
    +'}}');
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(misc) {
        console.log("submitNewOperatorGroup – " + misc);
        if(misc === "success") {
            toggleSettingsView();
            toggleSettingsView();
        }
    });
}

function deleteOperatorFromGroup(operator) {
    console.log("deleteOperatorFromGroup – deleteing operator from operatorgroups and operators")
    console.log(operator);
    var MCCMNC_ID =operator.id;
    var postRequestJSON = JSON.parse('{"functionToCall" : "deleteOperator", "data" : {'
    + ' "MCCMNC_ID" : "'+ MCCMNC_ID + '"'
    +'}}');
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(misc) {
        console.log("deleteOperatorFromGroup – " + misc);
        if(misc === "success") {
            toggleSettingsView();
            toggleSettingsView();
        }
    });
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
            appConfig.innerHTML = '<div onclick="redirectToConfigurationPage(this)"></div>';
            appConfig = appConfig.children[0];
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
            var appConfig2 = allAppConfigs.appConfigs[i];
            globalConfigArray.push(appConfig2);
        }
        console.log("showAppConfigOnHover – updated globalConfigArray:");
        console.log(globalConfigArray);

        console.log(appConfig);
        var appConfiguration;
        for(var j = 0; j < globalConfigArray.length; j++) {
            if(configNumber === globalConfigArray[j].Config_ID) {
                appConfiguration = globalConfigArray[j];
            }
        }
        var returnString = "";
        for (var element in appConfiguration) {
            var val = appConfiguration[element];
            returnString+=(element + " : " + val + "\n");
        }
        console.log("generateAppConfigHoverContents – Setting appconfig contents for config #" + configNumber);
        var appConfigHoverContents = document.createElement('div');
        appConfigHoverContents.className = 'appConfigHoverContent';
        appConfigHoverContents.innerText = returnString;


        console.log("showAppConfigOnHover – hoverContents innerHTML = " + appConfigHoverContents.innerText);
        var configFeaturedLocs = document.createElement('div');
        configFeaturedLocs.className = 'configproducts';
        configFeaturedLocs.innerHTML = '<div class=\'hidden\'></div>';
        configFeaturedLocs = configFeaturedLocs.children[0];
        appConfigHoverContents.appendChild(configFeaturedLocs);

        console.log("showAppConfigOnHover – Getting products for config " + configNumber)
        var postRequestText = '{"functionToCall" : "getproducts", "data" : {'
            + ' "Config_ID" : "'+ configNumber + '"'
            +'}}';
        var postRequestJSON = JSON.parse(postRequestText);
        console.log(postRequestJSON);
        server_post.post(post_url, postRequestJSON, function(products) {
            var newHTML = 'Products : ';
            for(i in products) {
                newHTML += (products[i] + ", ");
            }
            var parent = configFeaturedLocs.parentElement;
            if(appConfig.children[0]!=null){
                appConfig.children[0].lastChild.remove();
            }
            parent.innerHTML+= newHTML;
        });
        console.log("showAppConfigOnHover – Getting featureMappings for config " + configNumber)
        var postRequestText = '{"functionToCall" : "getFeatureMappings", "data" : {'
            + ' "Config_ID" : "'+ configNumber + '"'
            +'}}';
        var postRequestJSON = JSON.parse(postRequestText);
        console.log(postRequestJSON);
        server_post.post(post_url, postRequestJSON, function(featureMappings) {
            console.log("showAppConfigOnHover – server responded with:");
            console.log(featureMappings);
            var newHTML = 'Feature Mappings : \n';
            var oldFeatureType = "";
            var feature = document.createElement('div');
            feature.className = "hoverFeature";
            for(var i = 0; i < featureMappings.length; i++) {
                console.log(featureMappings[i]);
                newFeatureType = featureMappings[i].FeatureName;
                if(oldFeatureType === "") { //beginning of list
                    oldFeatureType = newFeatureType;
                    feature.id = newFeatureType;

                    var featureName = document.createElement('div');
                    feature.className = "hoverFeatureName"
                    feature.innerText
                    featureName.innerText = featureMappings[i].FeatureType;
                    feature.appendChild(featureName);
                } else if(oldFeatureType === newFeatureType) { //same feature as before, just add to feature element
                    var featureName = document.createElement('div');
                    feature.className = "hoverFeatureName"
                    featureName.innerText = featureMappings[i].FeatureType;
                    feature.appendChild(featureName);
                } else { //new Feature
                    appConfig.children[0].appendChild(feature);
                    oldFeatureType = newFeatureType;
                    feature = document.createElement('div');
                    feature.id = newFeatureType;
                    var featureName = document.createElement('div');
                    feature.className = "hoverFeatureName"
                    featureName.innerText = featureMappings[i].FeatureType;
                    feature.appendChild(featureName);
                }

            }
            appConfig.children[0].appendChild(feature);

        });


        appConfigHoverContents.classList.remove('hidden');
        if(appConfig.children[0] != null){
            appConfig.children[0].remove();
        }

        appConfig.prepend(appConfigHoverContents);
        appConfig.children[0].classList.remove('hidden');
    });
}


function hideAppConfigOnHover(appConfig) {
    appConfig.children[0].classList.add('hidden');
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
        var postRequestJSON = JSON.parse('{"functionToCall" : "globalView", "data" : {'
        + ' "App_OriginalName" : "'+ appOriginalName + '"'
        +'}}');
        server_post.post(post_url, postRequestJSON, function(appData) {
            console.log("toggleGlobalView – Success! Server returned:");
            console.log(appData);
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
                    appConfig.innerHTML = '<div onclick="redirectToConfigurationPage(this)"></div>';
                    appConfig = appConfig.children[0];
                    appConfig.className = "globalViewAppConfig";
                    appConfig.innerText = (appData.GlobalDataCountries[i].ConfigNumbers[o]);
                    if(appData.GlobalDataCountries[i].ActiveConfigs != null) {
                        for(var j = 0; j < appData.GlobalDataCountries[i].ActiveConfigs.length; j++){
                            if(configNumber === appData.GlobalDataCountries[i].ActiveConfigs[j]){
                                appConfig.classList.add("active");
                            }
                        }
                    }
                    setConfigHover(appConfig, configNumber);
                    appConfigs.appendChild(appConfig);
                }
                country.appendChild(countryTitle);
                country.appendChild(appConfigs);
                if(appData.GlobalDataCountries[i].ActiveConfigs === null ) {
                    country.appendChild(downArrow);
                } else if(appData.GlobalDataCountries[i].ActiveConfigs.length != appData.GlobalDataCountries[i].ConfigNumbers.length) {
                    country.appendChild(downArrow);
                }
                country.appendChild(countryContents);
                appElement.children[3].appendChild(country);
            }
        });
    }
}
function redirectToConfigurationPage(configElement) {
    console.log("redirectToConfigurationPage – \t\tElement clicked is");
    console.log(configElement);
    childBackup = configElement.children[0];
    configElement.children[0].remove();
    var text = (configElement.innerText)
    configElement.prepend(childBackup);
    console.log("/configs/" + text);
    window.location.href = ("/configs/" + text);
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
                    operatorConfig.innerHTML = '<div onclick="redirectToConfigurationPage(this)"></div>';
                    operatorConfig = operatorConfig.children[0];
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
    var searchfield_text = filterParams[1].value;
    var app_name = appID;

    var postRequestJSON = JSON.parse('{"functionToCall" : "appView", "data" : {'
    + ' "Selected_country" : "'+ selected_country + '",'
    + ' "Selected_operator" : "'+ selected_operator + '",'
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
    //add App to app tray//showWebapps_Old() again?
}
function addUltraApp(form)
{
    console.log("addUltraApp – Adding " + form.children[0].children[0].value + " Ultra for the current filter configuration. (Not implemented yet...)");
    var countriesList= "";
    var operatorsList = "";
    var operatorGroupList = "";
    var configMappings = form.children[1].children[0].children[1];
    var operatorGroupMappings = form.children[1].children[1].children[1];
    var existsEverywhere = false;
    if(configMappings.children[0]!=null) { //stuff exists inside of countrySearchResults
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
    }
    for(var i = 0; i < operatorGroupMappings.children.length; i++) {
        operatorGroupList += ("\""+operatorGroupMappings.children[i].children[0].innerText+"\"" + ", ")
    }
    countriesList = countriesList.replace(/,\s*$/, "");
    operatorsList = operatorsList.replace(/,\s*$/, "");
    operatorGroupList = operatorGroupList.replace(/,\s*$/, "");
    // debugger;
    var categoryName = "";
    console.log("PENIS");
    if(form.children[0].children[5].children[0].children[0].checked){
        categoryName = "default";
    } else if(form.children[0].children[5].children[1].children[0].checked){
        categoryName = "games"
    }
    var json = ('{"functionToCall" : "addNewConfig", "data" : {'
        + ' "App_ModifiableName" : "'+ form.children[0].children[1].value+ '",'
        + ' "App_OriginalName" : "'+ form.children[0].children[0].value + '",'
        + ' "App_Rank" : "'+ form.children[0].children[2].value + '",'
        + ' "App_HomeURL" : "'+ form.children[0].children[3].value + '",'
        + ' "App_IconURL" : "'+ form.children[0].children[4].value + '",'
        + ' "App_Category" : "'+ categoryName + '",'
        + ' "Packages" : ["'+ form.children[0].children[6].value + '"],'
        + ' "DefaultHiddenUI" : { '
            + ' "Splash" : '+form.children[0].children[7].children[0].children[0].checked+','
            + ' "Overlay" : '+form.children[0].children[7].children[1].children[0].checked+','
            + ' "FAB" : '+form.children[0].children[7].children[2].children[0].checked+','
            + ' "Badges" : '+form.children[0].children[7].children[3].children[0].checked+','
            + ' "Folder" : '+form.children[0].children[7].children[3].children[0].checked+''
        +'},'
        + ' "DefaultHiddenFeatures" : { '
            + ' "Savings" : '+form.children[0].children[8].children[0].children[0].checked+','
            + ' "Privacy" : '+form.children[0].children[8].children[1].children[0].checked+','
            + ' "Adblock" : '+form.children[0].children[8].children[2].children[0].checked+','
            + ' "NoImages" : '+form.children[0].children[8].children[3].children[0].checked+''
        +'},'
        + ' "DefaultEnabledFeatures" : { '
            + ' "Savings" : '+form.children[0].children[9].children[0].children[0].checked+','
            + ' "Privacy" : '+form.children[0].children[9].children[1].children[0].checked+','
            + ' "Adblock" : '+form.children[0].children[9].children[2].children[0].checked+','
            + ' "NoImages" : '+form.children[0].children[9].children[3].children[0].checked+''
        +'},'
        + ' "products" : { '
            + ' "MaxGlobal" : '+form.children[0].children[10].children[0].children[0].checked+','
            + ' "Max" : '+form.children[0].children[10].children[1].children[0].checked+','
            + ' "MaxGo" : '+form.children[0].children[10].children[2].children[0].checked+''
        +'},'
        + ' "App_ExistsEverywhere" : '+ existsEverywhere + ','
        + ' "App_ConfigurationMappings" : { '
            + ' "Countries" : ['
                + countriesList
            + '], "Operators" : ['
                + operatorsList
            + '], "OperatorGroups" : ['
                + operatorGroupList
            + ']'
        +'}'
    +'}}');
    console.log(json);
    json = JSON.parse(json);
    console.log("addUltraApp – Sending post request with the following JSON:");
    console.log(json);
    server_post.post(post_url, json, function(message) {
        console.log("addUltraApp – POST REQUEST SUCCESS!!!");
        if(message.result === "SUCCESS") {
            applyFilters();
            console.log("SUBMIT_NEW_APP – Closing popup window...")
            closeAddAppPopup();
        } else {
            console.log(message.result);
            var errorDiv = document.createElement("div");
            errorDiv.className = "errorDiv";
            errorDiv.textContent = message.result;
            form.appendChild(errorDiv);
            var element = document.getElementsByClassName("contents")[0];
            element.scrollTop = element.scrollHeight - element.clientHeight;
        }

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
    //find countries search field and set it to autofill
    var countrySearch = document.getElementById('countrySearch');
    var operatorSearch = document.getElementById('operatorSearch');
    autocomplete(countrySearch, all_countries);
    autocomplete(operatorSearch, all_operators);
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
                webAppHTML += app.productName;
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
function imageUploaded(newAppForm){
    document.getElementById("imageUploaded").classList.remove("hidden");
    console.log(newAppForm);
    console.log(newAppForm.children[3]);
    var filePath = newAppForm.children[1].children[0].files[0].name;
    newAppForm.children[3].children[0].children[4].value = ("ultra_apps/"+ filePath);
}
function generateAddAppPopupInputFields(){ //AddApp Popup window helper function
    var addAppViewHTML = '<form onsubmit="imageUploaded(this.parentElement);" id = "addAppIconForm" enctype="multipart/form-data" target="invisible" action="/upload" method="post"><input id="fileUpload" type="file" name="uploadfile" /><input type="submit" value="Upload Image" /><div id="imageUploaded" class="hidden"></div></form><iframe name="invisible" style="display:none;"></iframe>';
    addAppViewHTML += '<form id="addAppForm" onsubmit="submitNewApp(this); return false"><div id="appConfig">';
    addAppViewHTML += '<input type="text" placeholder="Ultra App ID (Original Name)" name="originalName">';
    addAppViewHTML += '<input type="text" placeholder="Ultra App Name" name="name">';
    addAppViewHTML += '<input type="text" placeholder="Ultra App Rank" name="rank">';
    addAppViewHTML += '<input type="text" placeholder="Webapp Link" name="homeUrl">';
    addAppViewHTML += '<input type="text" placeholder="Icon URL Link" name="iconUrl">';
    addAppViewHTML += '<div id ="addAppCategories">';
        addAppViewHTML += 'Category ';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="radio" name="category" value="default" checked />';
            addAppViewHTML += '<label for="default">Default</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="radio" name="category" value="games" />';
            addAppViewHTML += '<label for="games">Games</label>';
        addAppViewHTML += '</div>';
    addAppViewHTML += '</div>';
    addAppViewHTML += '<input type="text" placeholder="Native App Link(s)" name="nativeApps">';
    addAppViewHTML += '<div id ="addAppHiddenUI">';
        addAppViewHTML += 'Hidden UI ';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenUI" value="splash" />';
            addAppViewHTML += '<label for="splash">Splash</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenUI" value="overlay" />';
            addAppViewHTML += '<label for="overlay">Overlay</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenUI" value="ab" />';
            addAppViewHTML += '<label for="fab">FAB</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenUI" value="badges" />';
            addAppViewHTML += '<label for="badges">Badges</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenUI" value="folder" />';
            addAppViewHTML += '<label for="folder">Folder</label>';
        addAppViewHTML += '</div>';
    addAppViewHTML += '</div>';
    addAppViewHTML += '<div id ="addAppHiddenFeatures">';
        addAppViewHTML += 'Hidden Features ';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenFeatures" value="savings" />';
            addAppViewHTML += '<label for="savings">Savings</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenFeatures" value="privacy" />';
            addAppViewHTML += '<label for="privacy">Privacy</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenFeatures" value="adBlock" />';
            addAppViewHTML += '<label for="adBlock">Adblock</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="hiddenFeatures" value="noImages" />';
            addAppViewHTML += '<label for="noImages">No Images</label>';
        addAppViewHTML += '</div>';
    addAppViewHTML += '</div>';
    addAppViewHTML += '<div id ="addAppEnabledFeatures">';
        addAppViewHTML += 'Default Enabled Features ';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="savings" />';
            addAppViewHTML += '<label for="savings">Savings</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="privacy" />';
            addAppViewHTML += '<label for="privacy">Privacy</label>';
        addAppViewHTML += '</div>'
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="adBlock" />';
            addAppViewHTML += '<label for="adBlock">Adblock</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="noImages" />';
            addAppViewHTML += '<label for="noImages">No Images</label>';
        addAppViewHTML += '</div>';
    addAppViewHTML += '</div>';
    addAppViewHTML += '<div id ="products">';
        addAppViewHTML += 'Products';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="products" value="maxGlobal" checked />';
            addAppViewHTML += '<label for="maxGlobal">MaxGlobal</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="products" value="max" checked />';
            addAppViewHTML += '<label for="max">Max</label>';
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="addAppCheckboxContainer">';
            addAppViewHTML += '<input type="checkbox" name="products" value="maxGo" checked />';
            addAppViewHTML += '<label for="maxGo">MaxGo</label>';
        addAppViewHTML += '</div>';
    addAppViewHTML += '</div>';
    addAppViewHTML += '</div>';
    addAppViewHTML += '<div id="configurationMapping">';
        addAppViewHTML += '<div id="countryMapping">';
            addAppViewHTML += '<div class="autocomplete" style="width:448px; margin-left:14px;">'
                addAppViewHTML += '<input class="search" id="countrySearch" type="text" placeholder="Search for Country..." autocomplete="nope">'
            addAppViewHTML += '</div>'
            addAppViewHTML += '<div class="countrySearchResults"><div class="rowValue">ALL COUNTRIES</div></div>'
        addAppViewHTML += '</div>';
        addAppViewHTML += '<div id="operatorMapping">';
            addAppViewHTML += '<div class="autocomplete" style="width:448px; margin-left:14px;">'
                addAppViewHTML += '<input class="search" id="operatorSearch" type="text" placeholder="Search for Operator Group.." autocomplete="nope">'
            addAppViewHTML += '</div>'
            addAppViewHTML += '<div class="operatorSearchResults"></div>'
        addAppViewHTML += '</div>';
    addAppViewHTML += '</div>';
    addAppViewHTML += '<input type="submit" value="Submit"></form>';
    return addAppViewHTML;
}

function imgUpload(file){
    var uploadRequest = new XMLHttpRequest();
    uploadRequest.open('post', 'demo-url');
    uploadRequest.file = file;
    uploadRequest.onreadystatechange = function() {
    if (this.readyState == 4 && this.status == 200) {
      document.getElementById("demo").innerHTML =
      this.responseText;
      //or
      document.getElementById("demo").innerHTML =
      "Response code is 200 i.e successful image upload";
    }
  };
    uploadRequest.send();
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
                if(countrySearchResults.children[0]!=null){
                    if(countrySearchResults.children[0].textContent === "ALL COUNTRIES")
                    {
                        console.log("displayCountrySearchResults – 'All countries' bubble detected, deleting it")
                        console.log("displayCountrySearchResults – Country returned from search is not null, adding bubble")
                        countrySearchResults.innerHTML = ""; //get rid of all countries if valid country
                        console.log("Contents:");
                        console.log(country);
                    }
                }
                var countryBubbleHTML = getCountryBubbleHTML(country);

                console.log("displayCountrySearchResults – Adding to the countrySearchResults html the following bubble:");
                console.log(countryBubbleHTML);
                countrySearchResults.innerHTML += countryBubbleHTML;
            }
        });
    }
}
function displayOperatorSearchResults(operatorSearchFieldText){
    console.log("displayOperatorSearchResults – User input: " + operatorSearchFieldText);
    var operatorSearchResults = document.getElementsByClassName("operatorSearchResults")[0];
    var countrySearchResults = document.getElementsByClassName("countrySearchResults")[0];
    operatorBubbleExists = false;
    for(var i = 0; i < operatorSearchResults.children.length; i++) {
        if(operatorSearchResults.children[i].textContent===operatorSearchFieldText) {
            operatorBubbleExists = true;
            console.log("displayOperatorSearchResults – Operator bubble for " +operatorSearchFieldText +" already exits!");
            break;
        }
    }
    if(operatorSearchFieldText != ""  && !operatorBubbleExists)//only want to do things if textfield isn't empty, and bubble doens't already exist
    {
        console.log("displayOperatorSearchResults – getting operator by GroupName: "+ operatorSearchFieldText);
        getOperatorGroupByName(operatorSearchFieldText, function(operatorRows){
            var operator_group;

            if(operatorRows.operatorRows!=null) {
                console.log("displayOperatorSearchResults – GOT Operators");
                console.log(operatorRows);
                console.log(operatorRows);
                operator_group = operatorRows.operatorRows[0].Operator_Group_Name;
            }
            if(operator_group != "") {
                var operatorBubbleHTML = getOperatorBubbleHTML(operator_group);

                console.log("displayOperatorSearchResults – Adding to the operatorSearchResults html the following bubble:");
                console.log(operatorBubbleHTML);
                operatorSearchResults.innerHTML += operatorBubbleHTML;
                countrySearchResults.innerHTML = "";
            }
        });
    }
}

function getCountryBubbleHTML(country){
    var returnHTML = "";
    returnHTML += '<div class="rowValue" id="'+country.Country_ID+'"><div class="countryBubbleTitle">'+country.name+'</div><div class="rowImage" onClick="toggleCountryBubble(this.parentElement)" style="background-image: url(\'/images/arrow_drop_down.svg\'); background-repeat: no-repeat; background-size:100%;"></div></div>'
    return returnHTML;
}
function getOperatorBubbleHTML(operator_group) {
    var returnHTML = "";
    returnHTML += '<div class="rowValue" id="'+operator_group+'"><div class="operatorBubbleTitle">'+operator_group+'</div><div class="rowImage" onClick="toggleOperatorBubble(this.parentElement)" style="background-image: url(\'/images/arrow_drop_down.svg\'); background-repeat: no-repeat; background-size:100%;"></div></div>'
    return returnHTML;
}
function toggleOperatorBubble(operatorBubble) {
    console.log("toggleOperatorBubble – NOT IMPLEMENTED YET");
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
        if(operatorElement != null) {
            operatorElement.remove();
        }
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
                if(operators.operatorRows != null) {
                    for(var i = 0; i < operators.operatorRows.length; i++){
                        var operator = operators.operatorRows[i];
                        // if(operator.Operator_Name!=""){ //should probably get rid of this check, and get rid of null entries in DB
                            html+= ("<div value='"+operator.Operator_Name+"' id ='"+operator.MCCMNC_ID+"' class='operator checked' onClick='toggleOperator(this);'>"+operator.Operator_Name + " (" +operator.MCCMNC_ID+") </div>");
                        // }
                    }
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
function getOperatorGroupByName(operatorName, functionUsingOperator){
    var operator = {"hi" : "hello"};

    var postRequestJSON = JSON.parse('{"functionToCall" : "getOperatorGroupByName", "data" : {'
    + ' "Operator_Group_Name" : "'+ operatorName + '"'
    +'}}');

    server_post.post(post_url, postRequestJSON, function(operatorRows) {
        console.log("getOperatorGroupByName – Recieved the following JSON: ");
        console.log(operatorRows);
        functionUsingOperator(operatorRows);
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
