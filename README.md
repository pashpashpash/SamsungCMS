# SamsungCMS
Currently there are a lot of inefficiencies with creating/editing Ultra App configurations. A relatively simple action such as deleting Facebook Ultra in India requires a lot of manual work from the team (i.e. our main server guy manually figuring out the ordering/inheritance of the country/carrier filter stack & deploying config updates by hand).

Considering our ultimate goal is managing hundreds if not thousands of ultra apps, it would be nice to have a scalable system in place that does all of this work automatically with an easy UI....

The purpose of the CMS project is to create a tool that allows anyone on the team to

1. Log in with their Samsung Opus Team credentials
2. Check what ultra apps are in production (with proper targeting, eg. features enabled/ disabled, etc)  
3. Add/remove/edit ultra apps in production
4. Create timeline of edits to the config files (with time stamps) and store them (with ability to backup at any point).

## To do

- [ ] Get editing to work on client
- [ ] Add login functionality + server validation
- [ ] Figure out NEW DB based off sergey's excel sheet.
- [ ] Add new view based off excel sheet.
- [ ] Implement DB add/edit/delete with half-done add/edit/delte functions in the client.
- [ ] Get filtering to work globally (server + client filtering logic)
     - To do this on client, just implement the applyFilters() function in main.js
     - Golang will need a custom implementation
##

- [x] Built out golang restAPI that provides various data at /rest/{category} and /rest/ultra/{appID} in response to GET requests
- [x] Built out Samsung Ultra CMS Index page (with app tray, working filter mechanics, dynamic elements & mobile css compatibility)
- [x] Built out App Details page with javascript that generates HTML for the app based off a [webApp] JSON object request at /rest/ultra/{appID}
- [x] Get adding to work on client -> user provides the info contained in a webApp json
     -  [ ] A nice feature to have on top of adding would be autofill (based off DB)
- [x] Get deleting to work on client
