//initialization
var server = new restRequest();
var filterParams = [selects, maxCheckbox, folderCheckbox, homescreenCheckbox, searchField]; //from filters.js
var appContainer = document.getElementsByClassName("webApp")[0];


// Sending rest request for a specific ultra app
var appID = document.getElementsByTagName("title")[0].innerHTML;
var url = "/rest/ultra/" + appID;
server.get(url, function(app) {
    showWebapp(appContainer, app);
});

//=====================FUNCTION DECLARATIONS============================//
//input an app container + a webApp json, and this func will display it in the container with proper details
function showWebapp(appContainer, app) {
    var webAppHTML = "";  //set webAppHTML string to null, so we can += to it later
    console.log("showWebapp – Adding "+app.name+" app to the HTML");
    webAppHTML += "<div id='appName'>";
    webAppHTML += app.name;
    webAppHTML += ("</div>");

    appContainer.innerHTML = (webAppHTML);
}
