//initialization
var selects = document.getElementsByTagName('select');
var maxCheckbox = document.getElementById("maxCheck");
var folderCheckbox = document.getElementById("folderCheck");
var homescreenCheckbox = document.getElementById("homescreenCheck");
var searchField = document.getElementsByClassName('search')[0];


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
    var selects = document.getElementsByTagName('select');
    for(var z=0; z<selects.length; z++){
        if(selects[z].value!="star")
        {
            selects[z].value="star";
        }
    }
    console.log("SETALLSELECTSTOSTAR – All filters have stars");
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
