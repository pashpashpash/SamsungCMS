# SamsungCMS
Ultra Apps Content Management Service

## To do

- [ ] Get adding to work on client -> user provides the info contained in a webApp json
     - A nice feature to have on top of adding would be autofill (based off DB)
- [ ] Get deleting to work on client
- [ ] Get editing to work on client
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
