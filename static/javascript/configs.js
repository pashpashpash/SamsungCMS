var server_post = new postRequest();
var post_url = "/post/";
var Config_ID = imported_config_id; //imported from static script on configs.html file
var body = document.getElementsByTagName('body')[0];
console.log(imported_config_id);
console.log(Config_ID);

var configData = document.createElement('div');
configData.className = "configData";
body.appendChild(configData);

var featuredLocations = document.createElement('div');
featuredLocations.className = "featuredLocations";
body.appendChild(featuredLocations);

var features = document.createElement('div');
features.className = "features";
body.appendChild(features);

var configMappings = document.createElement('div');
configMappings.className = "configMappings";
body.appendChild(configMappings);



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
            rowValue.className = "rowValue";
            rowDescription.innerText = key;

            rowValue.innerText = value;


            row.appendChild(rowDescription);
            row.appendChild(rowValue);
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



postRequestText = '{"functionToCall" : "getFeaturedLocations", "data" : {'
    + ' "Config_ID" : "'+ Config_ID + '"'
    +'}}';
postRequestJSON = JSON.parse(postRequestText);
console.log(postRequestJSON);
server_post.post(post_url, postRequestJSON, function(featuredLocations_data) {
    console.log("main – server returned with featured location data");
    console.log(featuredLocations);

    var featuredLocationString = document.createElement("div");
    featuredLocationString.className = "featuredLocationString";
    var newHTML = 'Featured Locations : ';
    for(i in featuredLocations_data) {
        newHTML += (featuredLocations_data[i] + ", ");
    }

    featuredLocationString.innerText = newHTML;

    featuredLocations.appendChild(featuredLocationString);

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
            feature.innerText
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
