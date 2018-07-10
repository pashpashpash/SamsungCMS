# SamsungCMS
Currently there are a lot of inefficiencies with creating/editing Ultra App configurations. A relatively simple action such as deleting Facebook Ultra in India requires a lot of manual work from the team (i.e. our main server guy manually figuring out the ordering/inheritance of the country/carrier filter stack & deploying config updates by hand).

Considering our ultimate goal is managing hundreds if not thousands of ultra apps, it would be nice to have a scalable system in place that does all of this work automatically with an easy UI....

The purpose of the CMS project is to create a tool that allows anyone on the team to

1. Log in with their Samsung Opus Team credentials
2. Check what ultra apps are in production (with proper targeting, eg. features enabled/ disabled, etc)  
3. Add/remove/edit ultra apps in production
4. Create timeline of edits to the config files (with time stamps) and store them (with ability to backup at any point).

## To do
- [ ] Finish building "smart" select filters.
     -  [ ] Country+Operator select filters should work together, with one reacting based off a selection in the other.
     -  [ ] Update appTray on change, query the appConfig+configMappings tables to figure out which apps to show.
          - To do this on client, just implement the applyFilters() function in main.js
- [ ] Implement add/edit/delete RPC's + UX (Button press on client -> javascript function specifying what method the server should call -> server validates method, client, and method arguments -> server either runs methods and returns result, or returns error message)
- [ ] Add login functionality + server validation
- [ ] Create new global "Add Ultra App" view which allows user to check all locations/versions they want ultra app to have (different from appTray "Add Ultra App" view)
     -  [ ] Reuse "Add Ultra App" view components to create a new "ultraApps global view" which shows a breakdown of all apps and their states (different from "ultraApps appTray view")
- [ ] Delpoy site to a hosted domain provided by Sergey. Alternatively, hook up a GCP server myself.
- [ ] Write a translator in go that translates ultra app configs to config.ini sections
- [ ] Write logic that pushes go-generated config.ini's to production (dev cluster)
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
