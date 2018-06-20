# SamsungCMS
Ultra Apps Content Management Service

### TODO
1) Build out appView html+javascript that shows app based off a webapp JSON object

2) Hook up server to serve the following at /ultra/{app}

    a) {app} JSON object pre-calculated with the filter setting OR

    b) The filter settings JSON + the whole database JSON (and let client do the filtering)

3) Once you know how to filter the cms_db object (either locally or on server), hook up the home page reactive filtering

    – To do this, basically just implement the applyFilters() function in main.js
