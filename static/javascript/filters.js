//initialization
var selects = document.getElementsByTagName('select');
var searchField = document.getElementsByClassName('search')[0];
var filters = document.getElementById('filters');
//time to initialize (load in filter values)
var last_operator;
var last_country;
var all_countries = ["Abkhazia","Afghanistan","Albania","Algeria","American Samoa","Andorra","Angola","Anguilla","Antigua and Barbuda","Argentina Republic","Armenia","Aruba","Australia","Austria","Azerbaijan","Bahamas","Bahrain","Bangladesh","Barbados","Belarus","Belgium","Belize","Benin","Bermuda","Bhutan","Bolivia","Bosnia & Herzegov.","Botswana","Brazil","British Virgin Islands","Brunei Darussalam","Bulgaria","Burkina Faso","Burundi","Cambodia","Cameroon","Canada","Cape Verde","Cayman Islands","Central African Rep.","Chad","Chile","China","Colombia","Comoros","Congo, Dem. Rep.","Congo, Republic","Cook Islands","Costa Rica","Croatia","Cuba","Curacao","Cyprus","Czech Rep.","Denmark","Djibouti","Dominica","Dominican Republic","Ecuador","Egypt","El Salvador","Equatorial Guinea","Eritrea","Estonia","Ethiopia","Falkland Islands (Malvinas)","Faroe Islands","Fiji","Finland","France","French Guiana","French Polynesia","Gabon","Gambia","Germany","Ghana","Gibraltar","Greece","Greenland","Grenada","Guadeloupe","Guam","Guatemala","Guinea","Guinea-Bissau","Guyana","Haiti","Honduras","Hongkong, China","Hungary","Iceland","India","Indonesia","International Networks","Iran","Iraq","Ireland","Israel","Italy","Ivory Coast","Jamaica","Japan","Jordan","Kazakhstan","Kenya","Kiribati","Korea N., Dem. People's Rep.","Korea S, Republic of","Kuwait","Kyrgyzstan","Laos P.D.R.","Latvia","Lebanon","Lesotho","Liberia","Libya","Liechtenstein","Lithuania","Luxembourg","Macao, China","Macedonia","Madagascar","Malawi","Malaysia","Maldives","Mali","Malta","Martinique (French Department of)","Mauritania","Mauritius","Mexico","Micronesia","Moldova","Monaco","Mongolia","Montenegro","Montserrat","Morocco","Mozambique","Myanmar (Burma)","Namibia","Nepal","Netherlands","Netherlands Antilles","New Caledonia","New Zealand","Nicaragua","Niger","Nigeria","Niue","Norway","Oman","Pakistan","Palau (Republic of)","Palestinian Territory","Panama","Papua New Guinea","Paraguay","Peru","Philippines","Poland","Portugal","Puerto Rico","Qatar","Reunion","Romania","Russian Federation","Rwanda","Saint Kitts and Nevis","Saint Lucia","Samoa","San Marino","Sao Tome & Principe","Saudi Arabia","Senegal","Serbia","Seychelles","Sierra Leone","Singapore","Slovakia","Slovenia","Solomon Islands","Somalia","South Africa","South Sudan (Republic of)","Spain","Sri Lanka","St. Pierre & Miquelon","St. Vincent & Gren.","Sudan","Suriname","Swaziland","Sweden","Switzerland","Syrian Arab Republic","Taiwan","Tajikistan","Tanzania","Thailand","Timor-Leste","Togo","Tonga","Trinidad and Tobago","Tunisia","Turkey","Turkmenistan","Turks and Caicos Islands","Tuvalu","Uganda","Ukraine","United Arab Emirates","United Kingdom","United States","Uruguay","Uzbekistan","Vanuatu","Venezuela","Viet Nam","Virgin Islands, U.S.","Yemen","Zambia","Zimbabwe"];
var all_operators = ["freebasics","viva-bo","tigo-co","claro-co","alegro-ec","movistar-ec","claro-ec","movistar-gt","att_mx","telcel","ncell-np","cwp-pa","claro-pa","digicel-pa","movistar-pa","claro-pe","movistar-pe","entel-pe","bitel-pe","claro-pr","beeline","mobifone","sfone","vietnamobile","viettel","vinaphone"];


function updateFilterValues() {
    var selected_country = filterParams[0][0].options[filterParams[0][0].selectedIndex].value;
    var selected_operator = filterParams[0][1].options[filterParams[0][1].selectedIndex].value;
    var last_operator = selected_operator;
    var last_country = selected_country;
    var postRequestJSON = JSON.parse('{"functionToCall" : "updateFilterValues", "data" : {'
    + ' "Selected_country" : "'+ selected_country + '",'
    + ' "Selected_operator" : "'+ selected_operator + '"'
    +'}}');
    console.log("updateFilterValues – Sending post request to call updateFilterValues method");
    server_post.post(post_url, postRequestJSON, function(filterData) {
        console.log("updateFilterValues – Post Request success. Calling loadFilters...");
        loadFilters(filterData, last_country, last_operator);
    });
}
function loadFilters(filterData, lastcountry, last_operator){
    var selected_country = filterParams[0][0].options[filterParams[0][0].selectedIndex];
    var selected_operator = filterParams[0][1].options[filterParams[0][1].selectedIndex];
    console.log("loadFilters – Loading in filter data...");
    console.log("loadFilters – New filter values: ");
    console.log(selected_country, selected_operator);
    console.log("loadFilters – Data to be loaded into filters: ");
    console.log(filterData);
    if(filterData.countryFilterRows != null)
    {
        filterParams[0][0].options.length = 0;
        if(filterData.countryFilterRows.length != 1) {
            filterParams[0][0].options.add(new Option("All Countries  ▾", "star", true, true));
        }
        for (var i = 0; i < filterData.countryFilterRows.length; i++) {
            filterParams[0][0].options.add(new Option(filterData.countryFilterRows[i].Country_ID, filterData.countryFilterRows[i].name, false, false));

        }
        if(filterData.countryFilterRows.length === 1){
            filterParams[0][0].options.add(new Option("All Countries  ▾", "star", false, false));

        }
    }
    if(filterData.operatorFilterRows != null)
    {
        if(filterData.countryFilterRows===null){
            filterParams[0][1].options.length = 0;
            var previousoperatorExists = false;

            for (var i = 0; i < filterData.operatorFilterRows.length; i++) {
                if(last_operator != filterData.operatorFilterRows[i].MCCMNC_ID) {
                    filterParams[0][1].options.add(new Option(filterData.operatorFilterRows[i].Operator_Name, filterData.operatorFilterRows[i].MCCMNC_ID, false, false));
                } else {
                        previousoperatorExists = true;
                        filterParams[0][1].options.add(new Option(filterData.operatorFilterRows[i].Operator_Name, filterData.operatorFilterRows[i].MCCMNC_ID, true, true));
                }
            }
            if(!previousoperatorExists) {
                filterParams[0][1].options.add(new Option("All Operators  ▾", "star", true, true));
            } else {
                filterParams[0][1].options.add(new Option("All Operators  ▾", "star", false, false));
            }
        } else { //if country also not null, that means operator was pressed first, so just update operator list while keeping same selection
            if(filterData.countryFilterRows.length===1) {
                filterParams[0][1].options.length = 0;
                var previousoperatorExists = false;
                for (var i = 0; i < filterData.operatorFilterRows.length; i++) {
                    if(last_operator != filterData.operatorFilterRows[i].MCCMNC_ID) {
                        filterParams[0][1].options.add(new Option(filterData.operatorFilterRows[i].Operator_Name, filterData.operatorFilterRows[i].MCCMNC_ID, false, false));
                    } else {
                            previousoperatorExists = true;
                            filterParams[0][1].options.add(new Option(filterData.operatorFilterRows[i].Operator_Name, filterData.operatorFilterRows[i].MCCMNC_ID, true, true));
                    }
                }
                if(!previousoperatorExists) {
                    filterParams[0][1].options.add(new Option("All Operators  ▾", "star", true, true));
                } else {
                    filterParams[0][1].options.add(new Option("All Operators  ▾", "star", false, false));
                }
            } else {
                var previousoperatorExists = false;
                filterParams[0][1].options.length = 0;
                for (var i = 0; i < filterData.operatorFilterRows.length; i++) {
                    if(last_operator != filterData.operatorFilterRows[i].MCCMNC_ID) {
                        filterParams[0][1].options.add(new Option(filterData.operatorFilterRows[i].Operator_Name, filterData.operatorFilterRows[i].MCCMNC_ID, false, false));
                    } else {
                            previousoperatorExists = true;
                            filterParams[0][1].options.add(new Option(filterData.operatorFilterRows[i].Operator_Name, filterData.operatorFilterRows[i].MCCMNC_ID, true, true));
                    }
                }
                if(!previousoperatorExists) {
                    filterParams[0][1].options.add(new Option("All Operators  ▾", "star", true, true));
                } else {
                    filterParams[0][1].options.add(new Option("All Operators  ▾", "star", false, false));
                }
            }
        }


    } else {
        filterParams[0][1].options.length = 0;
        filterParams[0][1].options.add(new Option("Operators ▾", "star", true, true));
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


function autocomplete(inp, arr) {
  /*the autocomplete function takes two arguments,
  the text field element and an array of possible autocompleted values:*/
  var currentFocus;
  /*execute a function when someone writes in the text field:*/
  inp.addEventListener("input", function(e) {
      var a, b, i, val = this.value;
      /*close any already open lists of autocompleted values*/
      closeAllLists();
      if (!val) { return false;}
      currentFocus = -1;
      /*create a DIV element that will contain the items (values):*/
      a = document.createElement("DIV");
      a.setAttribute("id", this.id + "autocomplete-list");
      a.setAttribute("class", "autocomplete-items");
      /*append the DIV element as a child of the autocomplete container:*/
      this.parentNode.appendChild(a);
      /*for each item in the array...*/
      for (i = 0; i < arr.length; i++) {
        /*check if the item starts with the same letters as the text field value:*/
        if (arr[i].substr(0, val.length).toUpperCase() == val.toUpperCase()) {
          /*create a DIV element for each matching element:*/
          b = document.createElement("DIV");
          /*make the matching letters bold:*/
          b.innerHTML = "<strong>" + arr[i].substr(0, val.length) + "</strong>";
          b.innerHTML += arr[i].substr(val.length);
          /*insert a input field that will hold the current array item's value:*/
          b.innerHTML += "<input type='hidden' value='" + arr[i] + "'>";
          /*execute a function when someone clicks on the item value (DIV element):*/
              b.addEventListener("click", function(e) {
              /*insert the value for the autocomplete text field:*/
              inp.value = this.getElementsByTagName("input")[0].value;
              /*close the list of autocompleted values,
              (or any other open lists of autocompleted values:*/
              closeAllLists();
          });
          a.appendChild(b);
        }
      }
  });
  /*execute a function presses a key on the keyboard:*/
  inp.addEventListener("keydown", function(e) {
      var x = document.getElementById(this.id + "autocomplete-list");
      if (x) x = x.getElementsByTagName("div");
      if (e.keyCode == 40) {
        /*If the arrow DOWN key is pressed,
        increase the currentFocus variable:*/
        currentFocus++;
        /*and and make the current item more visible:*/
        addActive(x);
      } else if (e.keyCode == 38) { //up
        /*If the arrow UP key is pressed,
        decrease the currentFocus variable:*/
        currentFocus--;
        /*and and make the current item more visible:*/
        addActive(x);
      } else if (e.keyCode == 13) {
        /*If the ENTER key is pressed, prevent the form from being submitted,*/
        e.preventDefault();
        if (currentFocus > -1) {
          /*and simulate a click on the "active" item:*/
          if (x) x[currentFocus].click();
        }
      }
  });
  function addActive(x) {
    /*a function to classify an item as "active":*/
    if (!x) return false;
    /*start by removing the "active" class on all items:*/
    removeActive(x);
    if (currentFocus >= x.length) currentFocus = 0;
    if (currentFocus < 0) currentFocus = (x.length - 1);
    /*add class "autocomplete-active":*/
    x[currentFocus].classList.add("autocomplete-active");
  }
  function removeActive(x) {
    /*a function to remove the "active" class from all autocomplete items:*/
    for (var i = 0; i < x.length; i++) {
      x[i].classList.remove("autocomplete-active");
    }
  }
  function closeAllLists(elmnt) {
    /*close all autocomplete lists in the document,
    except the one passed as an argument:*/
    var x = document.getElementsByClassName("autocomplete-items");
    for (var i = 0; i < x.length; i++) {
      if (elmnt != x[i] && elmnt != inp) {
      x[i].parentNode.removeChild(x[i]);
    }
  }
}
/*execute a function when someone clicks in the document:*/
document.addEventListener("click", function (e) {
    closeAllLists(e.target);
});
}
