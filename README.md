# SamsungCMS
Currently there are a lot of inefficiencies with creating/editing Ultra App configurations. A relatively simple action such as deleting Facebook Ultra in India requires a lot of manual work from the team (i.e. our main server guy manually figuring out the ordering/inheritance of the country/carrier filter stack & deploying config updates by hand).

Considering our ultimate goal is managing hundreds if not thousands of ultra apps, it would be nice to have a scalable system in place that does all of this work automatically with an easy UI....

The purpose of the CMS project is to create a tool that allows anyone on the team to

1. Log in with their Samsung Opus Team credentials
2. Check what ultra apps are in production (with proper targeting, eg. features enabled/ disabled, etc)  
3. Add/remove/edit ultra apps in production
4. Create timeline of edits to the config files (with time stamps) and store them (with ability to backup at any point).

## To do
- [ ] Optimize globalview query to bring down load time from 7 seconds
     -  globalView function in main.go
- [ ] Add login functionality + server validation
- [ ] Delpoy site to a hosted domain provided by Sergey. Alternatively, hook up a GCP server myself.
- [ ] Write a translator in go that translates ultra app configs to config.ini sections
     -  [ ] Blocked atm: Once Michal gets back he will make a simplified version of the config (2 weeks?).
     -  [ ] While building translator, tweak db schema, initialization, addNewApp, queries. BIG refactor project...
- [ ] Write logic that pushes go-generated config.ini's to production (dev cluster)
-  Maybe figure out country mappings instead of just operators...
     -  [ ] Talk to Sergey about this.
##

- [x] Built out golang restAPI that provides various data at /rest/{category} and /rest/ultra/{appID} in response to GET requests
- [x] Built out Samsung Ultra CMS Index page (with "ultraApps appTray view", working filter mechanics, dynamic elements & mobile css compatibility)
- [x] Built out App Details page with javascript that generates HTML for the app based off a [webApp] JSON object request at /rest/ultra/{appID}
- [x] Get "appTray adding" to work on client-> user provides the info contained in a webApp json
     -  [ ] A nice feature to have on top of adding would be autofill (based off DB)
- [x] Get deleting to work on client
- [x] Designed DB schema for the CMS
- [x] Implemented SQLite implementation of DB schema for the CMS
- [x] Finish implementing DB initialization function – need to fill these tables during init:
     -  [x] Countries (All possible countries + MCC codes)
     -  [x] Operators (All possible operators + MCCMNC codes)
     -  [x] FeaturedLocations (All possible featured locations, total of 4)
     -  [x] Versions (All possible supported minVersions of Samsung Max, only 3.1 exists so far)
     -  [x] AppConfigs (Every existing ultra app state)
     -  [x] ConfigMappings (Maps every existing ultra app state to an operator and a featured Location)
- [x] Update existing site load mechanics to query the SQL DB instead of the current JSONDB that I made by hand.
     -  [x] Homepage (Apptray should query appConfigs table to display all apps with unique originalName's)
     -  [x] Appview Page (On icon click in appTray, JS should query appConfigs+Mappings tables to figure out correct appConfig to show on that page)
- [x] Finish building "smart" select filters.
     -  [x] Country+Operator select filters should work together, with one reacting based off a selection in the other.
     -  [x] Update appTray on change, query the appConfig+configMappings tables to figure out which apps to show.
          - To do this on client, just implement the applyFilters() function in main.js
-  [x] Finish "Add Ultra App" view.
   -  [x] Currently all it has is all of the fields to input the appConfig
   -  [x] But, it doesn't have a UX for inputting where to insert this appConfig.
        -  [x] By default, ALL countries is selected (if app already exists, this config will override the shit out of it everywhere)
        -  [x] Search-field -> Input a country, press enter -> "ALL" bubble gets replaced with inputted country.
             -  [x] Country bubble has a dropdown which on press shows all operators that are in the country, all selected by default.
        -  [x] Submit button, packages all data (appConfig + appMappings) into json, sends to server for insert.
             -  [ ]  Still need to finish Server validation – either rejects & sends error message or approves.
-  [x] Reuse finished "Add Ultra App" view components to create a new "ultraApps global view" which shows a breakdown of all apps and their states (different from "ultraApps appTray view")
-  [x] Generate configHTML for hover popups on Global View Page on site load.
     -  Query all unique app configs (with distinct originalNames), and generate a map of html elements to the app config number.
