//initialization
var selects = document.getElementsByTagName('select');
var searchField = document.getElementsByClassName('search')[0];
var filters = document.getElementById('filters');
//time to initialize (load in filter values)


function updateFilterValues() {
    var selected_country = filterParams[0][0].options[filterParams[0][0].selectedIndex].value;
    var selected_operator = filterParams[0][1].options[filterParams[0][1].selectedIndex].value;
    var selected_version = filterParams[0][2].options[filterParams[0][2].selectedIndex].value;
    var postRequestJSON = JSON.parse('{"functionToCall" : "updateFilterValues", "data" : {'
    + ' "Selected_country" : "'+ selected_country + '",'
    + ' "Selected_operator" : "'+ selected_operator + '",'
    + ' "Selected_version" : "'+ selected_version + '"'
    +'}}');
    console.log("updateFilterValues – Sending post request to call updateFilterValues method");
    server_post.post(post_url, postRequestJSON, function(filterData) {
        console.log("updateFilterValues – Post Request success. Calling loadFilters...");
        loadFilters(filterData);
    });
}
function loadFilters(filterData){
    var selected_country = filterParams[0][0].options[filterParams[0][0].selectedIndex];
    var selected_operator = filterParams[0][1].options[filterParams[0][1].selectedIndex];
    var selected_version = filterParams[0][2].options[filterParams[0][2].selectedIndex]
    console.log("loadFilters – Loading in filter data...");
    console.log("loadFilters – New filter values: ");
    console.log(selected_country, selected_operator, selected_version);
    console.log("loadFilters – Data to be loaded into filters: ");
    console.log(filterData);
    if(filterData.countryFilterRows != null)
    {
        filterParams[0][0].options.length = 0;
        if(filterData.countryFilterRows.length != 1) {
            filterParams[0][0].options.add(new Option("🔯", "star", true, true));
        }
        for (var i = 0; i < filterData.countryFilterRows.length; i++) {
            filterParams[0][0].options.add(new Option(filterData.countryFilterRows[i].Country_ID, filterData.countryFilterRows[i].name, false, false));
        }
        if(filterData.countryFilterRows.length === 1){
            filterParams[0][0].options.add(new Option("🔯", "star", false, false));
        }
    }
    if(filterData.operatorFilterRows != null)
    {
        filterParams[0][1].options.length = 0;
        filterParams[0][1].options.add(new Option("🔯", "star", true, true));
        for (var i = 0; i < filterData.operatorFilterRows.length; i++) {
            filterData.operatorFilterRows[i];
            filterParams[0][1].options.add(new Option(filterData.operatorFilterRows[i].Operator_Name, filterData.operatorFilterRows[i].MCCMNC_ID, false, false));
        }
    } else {
        filterParams[0][1].options.length = 0;
        filterParams[0][1].options.add(new Option("🔯", "star", true, true));
    }
    if(filterData.versionNumberRows != null)
    {
        if(filterData.versionNumberRows.length != (filterParams[0][2].options.length-1))
        {
            filterParams[0][2].options.length = 0;
            filterParams[0][2].options.add(new Option("🔯", "star", true, true));
            for (var i = 0; i < filterData.versionNumberRows.length; i++) {
                filterData.versionNumberRows[i];
                filterParams[0][2].options.add(new Option(filterData.versionNumberRows[i].versionNumber, filterData.versionNumberRows[i].versionNumber, false, false));
            }
        }
    }
}

//FILTER STUFF
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
    if(selectObject.value!="star")
    {
        starOFF();
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
    selects = document.getElementsByTagName('select');
    for(var z=0; z<selects.length; z++){
        if(selects[z].value!="star")
        {
            selects[z].value="star";
        }
    }
    console.log("SET_ALL_SELECTS_TOSTAR – All filters have stars");
}

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
    applyFilters()
};

searchField.oninput = function(){
    console.log("SEARCH_FIELD – Search field changed. New Value:");
    console.log(searchField.value);
    searchApplyFilters(searchField.value);
};
