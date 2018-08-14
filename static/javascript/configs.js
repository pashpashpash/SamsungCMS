var server_post = new postRequest();
var post_url = "/post/";
var Config_ID = imported_config_id; //imported from static script on configs.html file
var body = document.getElementsByTagName('body')[0];
console.log(imported_config_id);
console.log(Config_ID);

var configData = document.createElement('div');
configData.className = "configData";
body.appendChild(configData);

var products = document.createElement('div');
products.innerHTML = "<div onClick=\"editProducts(this)\"></div>"
products = products.children[0];
products.className = "products";
body.appendChild(products);

var features = document.createElement('div');
features.className = "features";
body.appendChild(features);

var configMappings = document.createElement('div');
configMappings.className = "configMappings";
body.appendChild(configMappings);



window.addEventListener('keydown',function(e){if(e.keyIdentifier=='U+000A'||e.keyIdentifier=='Enter'||e.keyCode==13){if(e.target.nodeName=='INPUT'&&e.target.type=='text'){
    e.preventDefault();
    if(e.srcElement===document.getElementById('originalName'))
    {
        if(e.srcElement.value != "") {
            submitEdit(Config_ID, "originalName", e.srcElement.value, e.srcElement.parentElement);
        }
    } else if(e.srcElement===document.getElementById('modifiableName')) {
        if(e.srcElement.value != "") {
            submitEdit(Config_ID, "modifiableName", e.srcElement.value, e.srcElement.parentElement);
        }
    } else if(e.srcElement===document.getElementById('homeURL')) {
        if(e.srcElement.value != "") {
            submitEdit(Config_ID, "homeURL", e.srcElement.value, e.srcElement.parentElement);
        }
    } else if (e.srcElement===document.getElementById('rank')) {
        if(e.srcElement.value != "") {
            submitEdit(Config_ID, "rank", e.srcElement.value, e.srcElement.parentElement);
        }
    }
    return false;
}}},true);



var postRequestText = '{"functionToCall" : "getAppConfig", "data" : {'
    + ' "Config_ID" : "'+ Config_ID + '"'
    +'}}';
var postRequestJSON = JSON.parse(postRequestText);
console.log(postRequestJSON);
server_post.post(post_url, postRequestJSON, function(appConfig) {
    console.log("main – server returned with app config data");
    console.log(appConfig);

    Object.keys(appConfig).forEach(function(key) {
        var value = appConfig[key].toString();
        if((value != "" && key != "IconUrl")) {
            console.log(key);
            console.log(value);

            var row = document.createElement("div");
            row.className = "row";
            var rowDescription = document.createElement("div");
            rowDescription.className = "rowDescription";
            var rowValue = document.createElement("div");
            if(key != "Config_ID") { //make all fields editable except config_id
                rowValue.innerHTML = "<div onClick=\"editValue(this)\"></div>"
                rowValue = rowValue.children[0];
            }
            rowValue.className = "rowValue";
            rowValue.id = key;
            rowDescription.innerText = key;

            rowValue.innerText = value;



            row.appendChild(rowValue);
            row.appendChild(rowDescription);
            configData.appendChild(row)
        }
        if(key === "IconUrl") {
            var configIcon = document.createElement('div');
            configIcon.innerHTML = '<div style="background-image: url(\'/'+value+'\'); background-repeat: no-repeat; background-size:100%;"></div>'
            configIcon = configIcon.children[0];
            configIcon.className = "configIcon";
            var configIconURL = document.createElement('div');
            configIconURL.className = "configIconURL";
            configIconURL.innerText = value;
            var deleteButton = document.createElement('div');
            deleteButton.innerHTML = ("<div onClick = deleteConfig("+Config_ID+")></div>");
            deleteButton = deleteButton.children[0];
            deleteButton.className = "deleteConfig";
            deleteButton.innerText = "Delete Configuration (" + Config_ID + ")";
            var configIconContainer = document.createElement('div');
            configIconContainer.className = "configIconContainer";

            configIconContainer.appendChild(configIcon);
            configIconContainer.appendChild(configIconURL);
            configIconContainer.appendChild(deleteButton);
            configData.prepend(configIconContainer);
        }
    });

});



postRequestText = '{"functionToCall" : "getProducts", "data" : {'
    + ' "Config_ID" : "'+ Config_ID + '"'
    +'}}';
postRequestJSON = JSON.parse(postRequestText);
console.log(postRequestJSON);
server_post.post(post_url, postRequestJSON, function(products_data) {
    console.log("main – server returned with product data");
    console.log(products);

    var productString = document.createElement("div");
    productString.className = "productString";
    var newHTML = 'Products : ';
    productString.innerText = newHTML;
    for(i in products_data) {
        var child = document.createElement("div");
        child.className = "singleProduct";
        child.innerText = products_data[i];
        productString.appendChild(child);
    }

    products.appendChild(productString);

});


console.log("showAppConfigOnHover – Getting featureMappings for config " + Config_ID)
 postRequestText = '{"functionToCall" : "getFeatureMappings", "data" : {'
    + ' "Config_ID" : "'+ Config_ID + '"'
    +'}}';
postRequestJSON = JSON.parse(postRequestText);
console.log(postRequestJSON);
server_post.post(post_url, postRequestJSON, function(featureMappings) {
    console.log("main – server returned with feature mapping data");
    console.log(featureMappings);


    var oldFeatureType = "";
    var feature = document.createElement('div');
    feature.innerHTML = "<div onClick=\"editFeature(this)\"></div>"
    feature = feature.children[0];
    feature.className = "hoverFeature";
    for(var i = 0; i < featureMappings.length; i++) {
        console.log(featureMappings[i]);
        newFeatureType = featureMappings[i].FeatureName;
        if(oldFeatureType === "") { //beginning of list
            oldFeatureType = newFeatureType;
            feature.id = newFeatureType;
            feature.innerText = newFeatureType;
            var featureName = document.createElement('div');
            feature.className = "hoverFeatureName"
            featureName.innerText = featureMappings[i].FeatureType;
            feature.appendChild(featureName);
        } else if(oldFeatureType === newFeatureType) { //same feature as before, just add to feature element
            var featureName = document.createElement('div');
            feature.className = "hoverFeatureName"
            featureName.innerText = featureMappings[i].FeatureType;
            feature.appendChild(featureName);
        } else { //new Feature
            features.appendChild(feature);
            oldFeatureType = newFeatureType;
            feature = document.createElement('div');
            feature.innerHTML = "<div onClick=\"editFeature(this)\"></div>"
            feature = feature.children[0];
            feature.id = newFeatureType;
            feature.innerText = newFeatureType;
            var featureName = document.createElement('div');
            feature.className = "hoverFeatureName"
            featureName.innerText = featureMappings[i].FeatureType;
            feature.appendChild(featureName);
        }

    }
    features.appendChild(feature);

});


var postRequestText = '{"functionToCall" : "getConfigurationMappings", "data" : {'
    + ' "Config_ID" : "'+ Config_ID + '"'
    +'}}';
var postRequestJSON = JSON.parse(postRequestText);
console.log(postRequestJSON);
server_post.post(post_url, postRequestJSON, function(configurationMappings) {
    console.log("main – server returned with configuration mapping data");
    console.log(configurationMappings);
    var configMappingsContainer = document.createElement("div");
    configMappingsContainer.className = "configMappingsContainer";

    var countries = document.createElement("div");
    countries.className = "configMappingsCountries";
    var operators = document.createElement("div");
    operators.className = "configMappingsOperators";

    if(configurationMappings.countryFilterRows != null) {
        for(var i =0; i<configurationMappings.countryFilterRows.length;i++) {
            var country = document.createElement("div");
            country.className = "mappingsRow";
            var countryID = document.createElement("div");
            countryID.innerText = configurationMappings.countryFilterRows[i].Country_ID;
            var countryName = document.createElement("div");
            countryName.innerText = configurationMappings.countryFilterRows[i].name;
            if(configurationMappings.countryFilterRows[i].Country_ID === "*") {
                countryID.innerText = "";
                countryName.innerText = "All Countries Mapped";
            }
            country.appendChild(countryName);
            country.appendChild(countryID);
            countries.appendChild(country);
        }
    }
    if(configurationMappings.operatorFilterRows != null) {
        for(var i =0; i<configurationMappings.operatorFilterRows.length;i++) {

            var operator = document.createElement("div");
            operator.className = "mappingsRow";
            var operatorID = document.createElement("div");
            operatorID.innerText = configurationMappings.operatorFilterRows[i].MCCMNC_ID;
            var operatorName = document.createElement("div");
            operatorName.innerText = configurationMappings.operatorFilterRows[i].Operator_Name;
            var operatorGroup = document.createElement("div");
            operatorGroup.innerText = configurationMappings.operatorFilterRows[i].Operator_Group_Name;
            operator.appendChild(operatorGroup);
            operator.appendChild(operatorName);
            operator.appendChild(operatorID);
            operators.appendChild(operator);
        }
    }
    var countriesTitle = document.createElement("div");
    countriesTitle.className = "countriesTitle";
    var operatorsTitle = document.createElement("div");
    operatorsTitle.className = "operatorsTitle";
    countriesTitle.innerText = "Country Mappings";
    operatorsTitle.innerText = "Operator Mappings";
    countries.prepend(countriesTitle);
    operators.prepend(operatorsTitle);
    configMappingsContainer.appendChild(countries);
    configMappingsContainer.appendChild(operators);

    var configMappingsTitle = document.createElement("div");
    configMappingsTitle.className = "configMappingsTitle";
    configMappingsTitle.innerText = "Configuration Mappings";
    configMappings.appendChild(configMappingsContainer);
    configMappings.prepend(configMappingsTitle);
});


function deleteConfig(configNumber){
    var postRequestText = '{"functionToCall" : "deleteConfiguration", "data" : {'
        + ' "Config_ID" : "'+ configNumber + '"'
        +'}}';
    var postRequestJSON = JSON.parse(postRequestText);
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(result) {
        console.log(result);
        if(result==="success"){
            window.location.href = "/";
        }
    });
}

function editValue(rowValueElement) {
    if(rowValueElement.children.length === 0) { //only do something if field doesn't already exist
        console.log(rowValueElement);
        console.log(rowValueElement.id);
        var id = "";
        if(rowValueElement.id === "OriginalName") {
            id = "originalName";
        } else if(rowValueElement.id === "ModifiableName" ) {
            id = "modifiableName";
        } else if(rowValueElement.id === "HomeUrl" ) {
            id = "homeURL";
        } else if(rowValueElement.id === "Category" ) {
            id = "category";
        } else if(rowValueElement.id === "Rank" ) {
            id = "rank";
        }

        if(id != "" && id != "category") {
            var inputElementHTML = "<input id = \""+id+"\"type = \"text\" value = \""+rowValueElement.innerText+"\"></input>";
            rowValueElement.innerHTML = inputElementHTML;
        } else if(id = "category") {
            var categoryInputHTML = '<div id ="addAppCategories">';
                categoryInputHTML += 'Category ';
                if(rowValueElement.innerText === "Default") {
                    categoryInputHTML += '<div id="addAppCheckboxContainer">';
                        categoryInputHTML += '<input type="radio" name="category" value="default" checked />';
                        categoryInputHTML += '<label for="default">Default</label>';
                    categoryInputHTML += '</div>';
                    categoryInputHTML += '<div id="addAppCheckboxContainer">';
                        categoryInputHTML += '<input type="radio" name="category" value="games" />';
                        categoryInputHTML += '<label for="games">Games</label>';
                    categoryInputHTML += '</div>';
                } else if (rowValueElement.innerText === "Games"){
                    categoryInputHTML += '<div id="addAppCheckboxContainer">';
                        categoryInputHTML += '<input type="radio" name="category" value="default" />';
                        categoryInputHTML += '<label for="default">Default</label>';
                    categoryInputHTML += '</div>';
                    categoryInputHTML += '<div id="addAppCheckboxContainer">';
                        categoryInputHTML += '<input type="radio" name="category" value="games" checked/>';
                        categoryInputHTML += '<label for="games">Games</label>';
                    categoryInputHTML += '</div>';
                }
            categoryInputHTML += '<div id ="editFormSubmit" onClick = "submitEditForm(this.parentElement)">Submit</div>';
            categoryInputHTML += '</div>';
            rowValueElement.innerHTML = categoryInputHTML;
        }
    }
}

function editFeature(featureElement) {
    if(featureElement.children[0].id === "")
    {

        var featureElementHTML = "";
        if(featureElement.id==="hiddenUI") {
            var remainingFeatures = ["splash", "overlay", "fab", "badges", "folder"];
            featureElementHTML += '<div id ="addAppHiddenUI">';
                featureElementHTML += 'Hidden UI ';
                for(var i = 0; i < featureElement.children.length; i++) {
                    var featureName = featureElement.children[i].innerText;
                    if(featureName==="splash") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="hiddenUI" value="splash" checked/>';
                            featureElementHTML += '<label for="splash">Splash</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "splash"); //removes existing feature from list
                    } else if (featureName==="overlay") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="hiddenUI" value="overlay" checked />';
                            featureElementHTML += '<label for="overlay">Overlay</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "overlay"); //removes existing feature from list
                    } else if (featureName==="fab") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="hiddenUI" value="fab" checked />';
                            featureElementHTML += '<label for="fab">FAB</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "fab"); //removes existing feature from list
                    } else if (featureName==="badges") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="hiddenUI" value="badges" checked />';
                            featureElementHTML += '<label for="badges">Badges</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "badges"); //removes existing feature from list
                    } else if (featureName==="folder") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="hiddenUI" value="folder" checked />';
                            featureElementHTML += '<label for="folder">Folder</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "folder"); //removes existing feature from list
                    }
                }
                for(var i = 0; i < remainingFeatures.length; i++) {
                    featureElementHTML += '<div id="addAppCheckboxContainer">';
                        featureElementHTML += '<input type="checkbox" name="hiddenUI" value="'+remainingFeatures[i]+'" />';
                        featureElementHTML += '<label for="'+remainingFeatures[i]+'">'+remainingFeatures[i]+'</label>';
                    featureElementHTML += '</div>';
                }
            featureElementHTML += '<div id ="submitFeatureButton" onClick = "submitFeatureEdit(this.parentElement)">Submit</div>';
            featureElementHTML += '</div>';
        } else if(featureElement.id==="hiddenFeatures") {
            var remainingFeatures = ["savings", "privacy", "adBlock", "noImages"];
            featureElementHTML += '<div id ="addAppHiddenFeatures">';
                featureElementHTML += 'Hidden Features ';
                for(var i = 0; i < featureElement.children.length; i++) {
                    if(featureElement.children[i].innerText.toLowerCase()==="savings") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="hiddenFeatures" value="savings" checked />';
                            featureElementHTML += '<label for="savings">Savings</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "savings"); //removes existing feature from list
                    } else if (featureElement.children[i].innerText.toLowerCase()==="privacy") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="hiddenFeatures" value="privacy" checked />';
                            featureElementHTML += '<label for="privacy">Privacy</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "privacy"); //removes existing feature from list
                    } else if (featureElement.children[i].innerText.toLowerCase()==="adblock") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="hiddenFeatures" value="adBlock" checked />';
                            featureElementHTML += '<label for="adBlock">adBlock</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "adBlock"); //removes existing feature from list
                    } else if (featureElement.children[i].innerText.toLowerCase()==="noimages") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="hiddenFeatures" value="noImages" checked />';
                            featureElementHTML += '<label for="noImages">No Images</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "noImages"); //removes existing feature from list
                    }
                }
                for(var i =0; i< remainingFeatures.length; i++) {
                    featureElementHTML += '<div id="addAppCheckboxContainer">';
                        featureElementHTML += '<input type="checkbox" name="hiddenFeatures" value="'+remainingFeatures[i]+'" />';
                        featureElementHTML += '<label for="'+remainingFeatures[i]+'">'+remainingFeatures[i]+'</label>';
                    featureElementHTML += '</div>';
                }
            featureElementHTML += '<div id ="submitFeatureButton" onClick = "submitFeatureEdit(this.parentElement)">Submit</div>';
            featureElementHTML += '</div>';
        } else if (featureElement.id==="defaultEnabledFeatures") {
            var remainingFeatures = ["savings", "privacy", "adBlock", "noImages"];
            featureElementHTML += '<div id ="addAppEnabledFeatures">';
                featureElementHTML += 'Default Enabled Features ';
                for(var i = 0; i < featureElement.children.length; i++) {
                    if(featureElement.children[i].innerText.toLowerCase()==="savings") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="savings" checked />';
                            featureElementHTML += '<label for="savings">Savings</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "savings"); //removes existing feature from list
                    } else if (featureElement.children[i].innerText.toLowerCase()==="privacy") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="privacy" checked />';
                            featureElementHTML += '<label for="privacy">Privacy</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "privacy"); //removes existing feature from list
                    } else if (featureElement.children[i].innerText.toLowerCase()==="adblock") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="adBlock" checked />';
                            featureElementHTML += '<label for="adBlock">Adblock</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "adBlock"); //removes existing feature from list
                    } else if (featureElement.children[i].innerText.toLowerCase()==="noimages") {
                        featureElementHTML += '<div id="addAppCheckboxContainer">';
                            featureElementHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="noImages" checked />';
                            featureElementHTML += '<label for="noImages">No Images</label>';
                        featureElementHTML += '</div>';
                        remainingFeatures = remainingFeatures.filter(e => e !== "noImages"); //removes existing feature from list
                    }
                }
                for(var i =0; i< remainingFeatures.length; i++) {
                    featureElementHTML += '<div id="addAppCheckboxContainer">';
                        featureElementHTML += '<input type="checkbox" name="defaultEnabledFeatures" value="'+remainingFeatures[i]+'" />';
                        featureElementHTML += '<label for="'+remainingFeatures[i]+'">'+remainingFeatures[i]+'</label>';
                    featureElementHTML += '</div>';
                }
            featureElementHTML += '<div id ="submitFeatureButton" onClick = "submitFeatureEdit(this.parentElement)">Submit</div>';
            featureElementHTML += '</div>';
        }
        featureElement.innerHTML = featureElementHTML;
    }
}

function submitFeatureEdit(featureEdit) {
    var postRequestText = '{"functionToCall" : "editFeature", "data" : {'
    + ' "Config_ID" : "'+ Config_ID + '",';

    if(featureEdit.id === "addAppEnabledFeatures") {
        postRequestText += (' "FieldName" : "'+ "defaultEnabledFeatures" + '",'
        + ' "DefaultEnabledFeatures" : { '
            + ' "Savings" : '+featureEdit.children[0].children[0].checked+','
            + ' "Privacy" : '+featureEdit.children[1].children[0].checked+','
            + ' "Adblock" : '+featureEdit.children[2].children[0].checked+','
            + ' "NoImages" : '+featureEdit.children[3].children[0].checked+''
        +'}');
    } else if (featureEdit.id === "addAppHiddenFeatures") {
        postRequestText += (' "FieldName" : "'+ "hiddenFeatures" + '",'
        + ' "DefaultHiddenFeatures" : { '
            + ' "Savings" : '+featureEdit.children[0].children[0].checked+','
            + ' "Privacy" : '+featureEdit.children[1].children[0].checked+','
            + ' "Adblock" : '+featureEdit.children[2].children[0].checked+','
            + ' "NoImages" : '+featureEdit.children[3].children[0].checked+''
        +'}');
    } else if (featureEdit.id === "addAppHiddenUI") {
        postRequestText += (' "FieldName" : "'+ "hiddenUI" + '",'
        +' "DefaultHiddenUI" : { '
            + ' "Splash" : '+featureEdit.children[0].children[0].checked+','
            + ' "Overlay" : '+featureEdit.children[1].children[0].checked+','
            + ' "FAB" : '+featureEdit.children[2].children[0].checked+','
            + ' "Badges" : '+featureEdit.children[3].children[0].checked+','
            + ' "Folder" : '+featureEdit.children[4].children[0].checked+''
        +'}');
    }
    postRequestText+='}}';
    console.log(postRequestText);

    var postRequestJSON = JSON.parse(postRequestText);
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(result) {
        console.log(result);
        var listOfFeatures = [];
        if(result==="success"){
            var featuresExist = false;
            console.log(featureEdit);
            for(var i = 0; i<featureEdit.children.length; i++) {
                if(featureEdit.children[i].id === "addAppCheckboxContainer") {
                    if(featureEdit.children[i].children[0].checked) {
                        featuresExist = true;
                        listOfFeatures.push(featureEdit.children[i].children[0].value);
                    }
                }
            }
            if(!featuresExist) {
                featureEdit.parentElement.remove()
            } else {
                var parent = featureEdit.parentElement;
                parent.innerHTML = "";
                parentHTML = featureEdit.id;
                for(var o = 0; o < listOfFeatures.length; o++) {
                    parentHTML += ("<div>" + listOfFeatures[o] +"</div>")
                }
                parent.innerHTML = parentHTML;
            }
        }
    });
}





function submitEditForm(editFormElement) {
    console.log(editFormElement);
    console.log(editFormElement.children[0].children[0].checked);
    console.log(editFormElement.children[1]);
    if(editFormElement.children[0].children[0].checked) {
        console.log(editFormElement.children[0].children[0].value);
        submitEdit(Config_ID, "category", "Default", editFormElement.parentElement);
    } else if (editFormElement.children[1].children[0].checked) {
        console.log(editFormElement.children[1].children[0].value);
        submitEdit(Config_ID, "category", "Games", editFormElement.parentElement);
    }
}
function submitEdit(Config_ID, fieldName, newValue, parentElement) {
    console.log("submitEdit – Submitting edit with the following info:");
    console.log(fieldName);
    console.log(newValue);
    var postRequestText = '{"functionToCall" : "editAppConfigField", "data" : {'
        + ' "Config_ID" : "'+ Config_ID + '",'
        + ' "FieldName" : "'+ fieldName + '",'
        + ' "NewValue" : "'+ newValue + '"'
        +'}}';
    var postRequestJSON = JSON.parse(postRequestText);
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(result) {
        console.log(result);
        if(result==="success"){
            parentElement.innerHTML = "";
            parentElement.innerText = newValue;
        } else {
            var errorDiv = document.createElement("div");
            errorDiv.className = "errorDiv";
            errorDiv.textContent = result;
            parentElement.appendChild(errorDiv);
        }
    });

}
function editProducts(productsElement) {
    if(productsElement.children.length === 1) {
        console.log("editProducts – productsElement:");
        console.log(productsElement);
        console.log(productsElement.children[0].children.length);
        var remainingProductNames = ["maxGlobal", "max", "maxGo"];
        var productsElementHTML = "";
        productsElementHTML += 'Products';
        for(var i =0; i<productsElement.children[0].children.length; i++) {
            var existingProductName = productsElement.children[0].children[i].innerText;
            productsElementHTML += '<div id="addAppCheckboxContainer">';
                productsElementHTML += '<input type="checkbox" name="products" value="'+existingProductName+'" checked />';
                productsElementHTML += '<label for="'+existingProductName+'">'+existingProductName+'</label>';
            productsElementHTML += '</div>';
            remainingProductNames = remainingProductNames.filter(e => e !== existingProductName); //removes existing procut from list
        }
        for(var i =0; i<remainingProductNames.length; i++) {
            var remainingProductName = remainingProductNames[i];
            productsElementHTML += '<div id="addAppCheckboxContainer">';
                productsElementHTML += '<input type="checkbox" name="products" value="'+remainingProductName+'"/>';
                productsElementHTML += '<label for="'+remainingProductName+'">'+remainingProductName+'</label>';
            productsElementHTML += '</div>';
        }
        productsElementHTML += '<div id ="editFormSubmit" onClick = "submitProductEdit(this.parentElement)">Submit</div>';
        productsElementHTML += '</div>';
        productsElement.innerHTML = productsElementHTML;
    }
}

function submitProductEdit(productElement) {
    console.log(productElement);
    var maxGlobal = false;
    var max = false;
    var maxGo = false;
    for(var i = 0; i< productElement.children.length; i++) {
        if(productElement.children[i].id==="addAppCheckboxContainer") {
            if(productElement.children[i].children[0].checked) {
                if(productElement.children[i].children[0].value === "maxGlobal") {
                    maxGlobal = true;
                } else if (productElement.children[i].children[0].value === "max") {
                    max = true;
                } else  if (productElement.children[i].children[0].value === "maxGo") {
                    maxGo = true;
                }
            }
        }
    }

    var postRequestText = '{"functionToCall" : "editProducts", "data" : {'
        + ' "Config_ID" : "'+ Config_ID + '",'
        + ' "products" : { '
            + ' "MaxGlobal" : '+maxGlobal+','
            + ' "Max" : '+max+','
            + ' "MaxGo" : '+maxGo+''
        +'}'
        +'}}';
    var postRequestJSON = JSON.parse(postRequestText);
    console.log(postRequestJSON);

    server_post.post(post_url, postRequestJSON, function(result) {
        console.log(result);
        if(result==="success"){
            var productString = document.createElement("div");
            productString.className = "productString";
            var newHTML = 'Products : ';
            productString.innerText = newHTML;

            products.appendChild(productString);
            if(maxGlobal) {
                var child = document.createElement("div");
                child.className = "singleProduct";
                child.innerText = "maxGlobal";
                productString.appendChild(child);
            }
            if(max) {
                var child = document.createElement("div");
                child.className = "singleProduct";
                child.innerText = "max";
                productString.appendChild(child);
            }
            if(maxGo) {
                var child = document.createElement("div");
                child.className = "singleProduct";
                child.innerText = "maxGo";
                productString.appendChild(child);
            }
            productElement.innerHTML = "";
            productElement.appendChild(productString);
        } else {
            var errorDiv = document.createElement("div");
            errorDiv.className = "errorDiv";
            errorDiv.textContent = result;
            productElement.appendChild(errorDiv);
        }
    });

}
