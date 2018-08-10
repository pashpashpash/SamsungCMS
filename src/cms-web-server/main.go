package main

import (
	"log"
	"net/http"
    // "html/template"
    // "bytes"
    "github.com/mholt/archiver"
    "path/filepath"
	"path"
	"time"
    "fmt"
    "io"
    "os"
    "github.com/gorilla/securecookie"
    "text/tabwriter"
        // "database/sql"
    // "strings"     // fmt.Fprint(w, strings.Join(appNames, ", \n"))
     _ "github.com/mattn/go-sqlite3"
    "strconv"
    "encoding/json"
    "io/ioutil"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

const NegroniLogFmt = "{{.StartTime}} | {{.Status}} | {{.Duration}} \n          {{.Method}} {{.Path}}\n"
const NegroniDateFmt = time.Stamp
var cookieHandler = securecookie.New(securecookie.GenerateRandomKey(64),securecookie.GenerateRandomKey(32))
var loggedInIndex string
func main() {
    tw := new(tabwriter.Writer)
    tw.Init(os.Stderr, 0, 8, 0, '\t', 0)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

    log.Println("main –\t\tCalling initDB with schema name 'cms'...")
    db = initDB("cms")

	server := NewServer()
	server.Run(":" + port)
}


// NewServer configures and returns a Server.
func NewServer() *negroni.Negroni {
	n := negroni.New()
	n.Use(negroni.NewRecovery())
	l := negroni.NewLogger()
	l.SetFormat(NegroniLogFmt)
	l.SetDateFormat(NegroniDateFmt)
	n.Use(l)

	root := GetProjectRoot()
	mx := mux.NewRouter()

    mx.HandleFunc("/rest/ultra/{appName}", restAppViewHandler)      //handles all restAPI GET requests
    mx.HandleFunc("/rest/ultra/", restAppViewDocumentationHandler)
    mx.HandleFunc("/rest/{category}", restHandler)      //handles all restAPI GET requests
    mx.HandleFunc("/rest/", restDocumentationHandler)   //if someone types in /rest/ with no category
    mx.HandleFunc("/configs/{Config_ID}", configPageHandler)   //for config page
    mx.HandleFunc("/export", exportPageHandler)
    mx.HandleFunc("/post/login", loginAuthentication)   //handles all post requests
    mx.HandleFunc("/upload", UploadFile)   //handles all post requests
    mx.HandleFunc("/post/", postHandler)   //handles all post requests
	mx.PathPrefix("/").Handler(FileServer(http.Dir(root + "/static/")))     //for all other urls, serve from /static/

    n.UseHandler(mx)
	return n
}


type requestData struct {
    FunctionToCall string `json:functionToCall`
    Data data `json:data, string, omitempty`
}
type data struct {
    Selected_country string `json:Selected_country, string, omitempty`
    Selected_operator string `json:Selected_operator, string, omitempty`
    Searchfield_text string `json:Searchfield_text, string, omitempty`
    App_name string `json:App_name, string, omitempty`
    Country_Name string `json:name, string, omitempty`
    OperatorName string `json:Operator_Name, string, omitempty`
    MCCMNC_ID string  `json:MCCMNC_ID, string, omitempty`
    Operator_Group_Name string `json:Operator_Group_Name, string, omitempty`
    Country_ID string `json:Country_ID, string, omitempty`
    Country_MCC string `json:Country_MCC, string, omitempty`
    Config_ID string `json:"Config_ID"`
    AppModifiableName        string `json:"App_ModifiableName"`
    AppOriginalName          string `json:"App_OriginalName"`
    AppRank                  string `json:"App_Rank"`
    AppHomeURL               string `json:"App_HomeURL"`
    AppIconURL               string `json:"App_IconURL"`
    AppCategory         string `json:"App_Category"`
    AppExistsEverywhere      bool   `json:"App_ExistsEverywhere"`
    AppConfigurationMappings struct {
        Countries           []string `json:"Countries, omitempty"`
        Operators           []string `json:"Operators, omitempty"`
        OperatorGroups      []string `json:"OperatorGroups, omitempty"`
    } `json:"App_ConfigurationMappings"`
    DefaultEnabledFeatures struct {
        Savings           bool `json:"Savings, omitempty"`
        Privacy           bool `json:"Privacy, omitempty"`
        Adblock           bool `json:"Adblock, omitempty"`
        NoImages          bool `json:"NoImages, omitempty"`
    } `json:"DefaultEnabledFeatures"`
    DefaultHiddenFeatures struct {
        Savings           bool `json:"Savings, omitempty"`
        Privacy           bool `json:"Privacy, omitempty"`
        Adblock           bool `json:"Adblock, omitempty"`
        NoImages          bool `json:"NoImages, omitempty"`
    } `json:"DefaultHiddenFeatures"`
    DefaultHiddenUI struct {
        Splash           bool `json:"Splash, omitempty"`
        Overlay           bool `json:"Overlay, omitempty"`
        AB           bool `json:"AB, omitempty"`
        Badges          bool `json:"Badges, omitempty"`
        Folder          bool `json:"Folder, omitempty"`
    } `json:"DefaultHiddenUI"`
    Products struct {
        MaxGlobal           bool `json:"MaxGlobal, omitempty"`
        Max           bool `json:"Max, omitempty"`
        MaxGo          bool `json:"MaxGo, omitempty"`
    } `json:"Products"`
    Packages      []string   `json:"Packages, omitempty"`
}
func postHandler(w http.ResponseWriter, r *http.Request) {
    log.Printf("postHandler –\t\tIncoming post request:")

    requestData := requestData{}

    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
      fmt.Println(err)
    }

    w.Header().Set("Content-Type", "application/json")
    username := string("")
    username = getUserName(r)
    if(username!= "" || requestData.FunctionToCall=="checkIfLoggedIn") { //user is logged in or post request coming in to check
        if (requestData.FunctionToCall=="loadAppTray") {
            log.Println("postHandler –\t\tAll apps method request detected – Data: ")
            log.Println(requestData.Data)
            jsonResponse := loadAppTray(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="appView") {
            log.Println("postHandler –\t\tApp view method request detected")
            jsonResponse := appView(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="updateFilterValues") {
            log.Println("postHandler –\t\tupdateFilterValues method request detected")
            jsonResponse := updateFilterValues(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="getCountryByName") {
            log.Println("postHandler –\t\tgetCountryByName method request detected")
            jsonResponse := getCountryByName(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="getOperatorsByCountryID") {
            log.Println("postHandler –\t\tgetOperatorsByCountryID method request detected")
            jsonResponse := getOperatorsByCountryID(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="addNewConfig") {
            log.Println("postHandler –\t\taddNewConfig method request detected")
            jsonResponse := addNewConfig(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="globalView") {
            log.Println("postHandler –\t\tglobalView method request detected")
            jsonResponse := globalView(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="settingsView") {
            log.Println("postHandler –\t\tsettingsView method request detected")
            jsonResponse := settingsView(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="getAllAppConfigs") {
            log.Println("postHandler –\t\tgetAllAppConfigs method request detected")
            jsonResponse := getAllAppConfigs(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="getproducts") {
            log.Println("postHandler –\t\tgetproducts method request detected")
            jsonResponse := getproducts(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="getFeatureMappings") {
            log.Println("postHandler –\t\tgetFeatureMappings method request detected")
            jsonResponse := getFeatureMappings(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="getConfigurationMappings") {
            log.Println("postHandler –\t\tgetConfigurationMappings method request detected")
            jsonResponse := getConfigurationMappings(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="getOperatorGroupByName") {
            log.Println("postHandler –\t\tgetOperatorGroupByName method request detected")
            jsonResponse := getOperatorGroupByName(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="submitNewOperator") {
            log.Println("postHandler –\t\tsubmitNewOperator method request detected")
            jsonResponse := submitNewOperator(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="deleteOperator") {
            log.Println("postHandler –\t\tdeleteOperator method request detected")
            jsonResponse := deleteOperator(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="submitNewCountry") {
            log.Println("postHandler –\t\tsubmitNewCountry method request detected")
            jsonResponse := submitNewCountry(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="getAppConfig") {
            log.Println("postHandler –\t\tgetAppConfig method request detected")
            jsonResponse := getAppConfig(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="deleteConfiguration") {
            log.Println("postHandler –\t\tdeleteConfiguration method request detected")
            jsonResponse := deleteConfiguration(requestData.Data)
            w.Write([]byte(jsonResponse))
        } else if (requestData.FunctionToCall=="updateConfigurationINI") {
            log.Println("postHandler –\t\tupdateConfigurationINI method request detected")
            jsonResponse := updateConfigurationINI(requestData.Data)
            w.Write([]byte(jsonResponse))
        }  else if (requestData.FunctionToCall=="checkIfLoggedIn") {
            log.Println("postHandler –\t\tcheckIfLoggedIn method request detected")
            jsonResponse := checkIfLoggedIn(w,r)
            w.Write([]byte(jsonResponse))
        }
    }


}
func loginAuthentication(w http.ResponseWriter, r *http.Request) {

    name := r.FormValue("name")
    pass := r.FormValue("password")

    loginAuthQuery := `SELECT DISTINCT userID from users WHERE username="`+name+`" AND password="`+pass+`"`
    log.Println("loginAuthentication –\t\tQuery = " + loginAuthQuery)
    userID := string("")
    loginAuth, err := db.Query(loginAuthQuery)
    checkErr(err)
    for(loginAuth.Next()) {
        loginAuth.Scan(&userID)
    }
    if(userID != "") {
        //correct login
        setSession(name, w)
        http.Redirect(w, r, "/", 308)
    }

    jsonResponse, err := json.Marshal("Incorrect user credentials. Please try again.")
    checkErr(err)
    log.Println("globalView –\t\tReturning JSON string...")
    w.Write([]byte(jsonResponse))
    http.Redirect(w, r, "/", 308)
}

func checkIfLoggedIn(w http.ResponseWriter, r *http.Request) ([]byte){
    username := string("")
    username = getUserName(r)
    if(username!= "") { //user is logged in
        jsonResponse, err := json.Marshal(true)
        checkErr(err)
        log.Println("globalView –\t\tReturning JSON string...")
        return jsonResponse
    } else {
        jsonResponse, err := json.Marshal(false)
        checkErr(err)
        log.Println("globalView –\t\tReturning JSON string...")
        return jsonResponse
    }
}

func setSession(userName string, response http.ResponseWriter) {
    value := map[string]string{
        "name": userName,
    }
    if encoded, err := cookieHandler.Encode("session", value); err == nil {
        cookie := &http.Cookie{
            Name:  "session",
            Value: encoded,
            Path:  "/",
        }
        http.SetCookie(response, cookie)
    }
}

func getUserName(request *http.Request) (userName string) {
    if cookie, err := request.Cookie("session"); err == nil {
        cookieValue := make(map[string]string)
        if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
            userName = cookieValue["name"]
        }
    }
    return userName
}

func clearSession(response http.ResponseWriter) { //use for logout
    cookie := &http.Cookie{
        Name:   "session",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
     }
    http.SetCookie(response, cookie)
}

type GlobalData struct {
    GlobalDataApps []GlobalDataApp `json: "globalDataApps" `
    GlobalDataCountries []GlobalDataCountry `json: "globalDataCountries" `
    OperatorRows []GlobalOperatorRow `json:"operatorRows"`
}
type GlobalDataApp struct {
    OriginalName string `json: "originalName" db:"originalName"`
    ConfigNumbers []string `json: "configNumbers"`
}
type GlobalDataCountry struct {
    Country_ID string `json:"Country_ID" db:"Country_ID"`
    CountryName string `json:"name" db:"name"`
    App_Config_ID string `json:"Config_ID" db: "Config_ID"`
    ConfigNumbers []string `json: "configNumbers"`
    OperatorRows []GlobalOperatorRow `json:"operatorRows", omitempty`
    ActiveConfigs []string `json : "activeConfigs", omitempty`
}
type GlobalOperatorRow struct {
    MCCMNC_ID string `json: "MCCMNC_ID"`
    Operator_Name string `json: "Operator_Name"`
    App_Config_ID string `json:"Config_ID" db: "Config_ID"`
    ConfigNumbers []string `json: "configNumbers"`
    ActiveConfig string `json : "activeConfig", omitempty`
}
type SettingsViewData struct {
    OperatorGroups []OperatorRows `json:"OperatorGroups"`
    CountryRows    []CountryRow `json:"Countries"`
}
type OperatorRows struct {
    OperatorRows []OperatorRow `json:"operatorRows"`
}
type OperatorRow struct {
    MCCMNC_ID string `json: MCCMNC_ID`
    Operator_Name string `json: Operator_Name`
    Country_ID string `json: Country_ID`
    Operator_Group_Name string `json: "Country_ID, omitempty"`
}
type CountryRow struct {
    Country_ID string `json:"Country_ID" db:"Country_ID"`
    CountryName string `json:"name" db:"name"`
    MCC_ID string `json:"MCC_ID" db:"MCC_ID"`
}
func settingsView(Data data) ([]byte) {
    settingsViewQuery := `SELECT DISTINCT operatorGroups.Operator_Group_Name, operatorGroups.MCCMNC_ID, Operator_Name, Country_ID from operators
    JOIN operatorGroups USING (MCCMNC_ID) ORDER BY Operator_Group_Name`
    log.Println("settingsView –\t\tQuery = " + settingsViewQuery)
    operatorList, err := db.Query(settingsViewQuery)
    checkErr(err)
    settingsData := SettingsViewData{}
    operatorRows := OperatorRows{}
    oldGroup := string("")
    newGroup := string("")
    for(operatorList.Next()){
        operatorRow := OperatorRow{}
        operatorList.Scan(&operatorRow.Operator_Group_Name, &operatorRow.MCCMNC_ID, &operatorRow.Operator_Name, &operatorRow.Country_ID)
        newGroup = operatorRow.Operator_Group_Name
        if(oldGroup == "") { //start of list, append first entry to operatorRows
            oldGroup = newGroup
            operatorRows.OperatorRows = append(operatorRows.OperatorRows, operatorRow)
        } else if(oldGroup == newGroup) { //same group, add
            operatorRows.OperatorRows = append(operatorRows.OperatorRows, operatorRow)
        } else { //end of group, add old group to settingsData, append new row to new group
            settingsData.OperatorGroups = append(settingsData.OperatorGroups, operatorRows)
            operatorRows = OperatorRows{}
            operatorRows.OperatorRows = append(operatorRows.OperatorRows, operatorRow)
            oldGroup = newGroup
        }
    }

    settingsData.OperatorGroups = append(settingsData.OperatorGroups, operatorRows) //adds last group

    settingsViewQuery = `SELECT DISTINCT Country_ID, name, MCC_ID from countries ORDER BY name`
    log.Println("settingsView –\t\tQuery = " + settingsViewQuery)
    countryList, err := db.Query(settingsViewQuery)
    checkErr(err)
    for(countryList.Next()){
        countryRow := CountryRow{}
        countryList.Scan(&countryRow.Country_ID, &countryRow.CountryName, &countryRow.MCC_ID)
        settingsData.CountryRows = append(settingsData.CountryRows, countryRow)
    }
    jsonResponse, err := json.Marshal(settingsData)
    checkErr(err)
    log.Println("globalView –\t\tReturning JSON string...")
    return jsonResponse
}
func globalView(Data data) ([]byte) {
    if(Data.AppOriginalName != ""){
        if(Data.Country_ID != "") { //appName and country specified, show operators + configs
            log.Println("globalView –\t\tGlobalView | Operator Level")
            globalViewQuery := `SELECT DISTINCT operators.MCCMNC_ID, operators.Operator_Name, appConfigs.Config_ID
            from operators
            INNER JOIN configurationMappings ON operators.MCCMNC_ID = configurationMappings.MCCMNC_ID
            INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
            INNER JOIN countries ON operators.Country_ID = countries.Country_ID
            WHERE appConfigs.originalName ='`+Data.AppOriginalName+`'
            AND countries.Country_ID = '`+Data.Country_ID+`'
            ORDER BY operators.MCCMNC_ID`
            log.Println("globalView –\t\tQuery = " + globalViewQuery)
            operatorList, err := db.Query(globalViewQuery)
            checkErr(err)
            globalData := GlobalData{}
            oldMCCMNC_ID := ""
            operator := GlobalOperatorRow{}
            for(operatorList.Next()){
                var newMCCMNC_ID string
                var newOperator_Name string
                var newConfig_ID string
                operatorList.Scan(&newMCCMNC_ID, &newOperator_Name, &newConfig_ID)
                if(oldMCCMNC_ID == "") { //beginning of the list
                    operator = GlobalOperatorRow{} //initialize new operator
                    oldMCCMNC_ID = newMCCMNC_ID
                    operator.MCCMNC_ID = newMCCMNC_ID
                    operator.Operator_Name = newOperator_Name
                    operator.ConfigNumbers = append (operator.ConfigNumbers, newConfig_ID) //add configs to configList
                } else if(newMCCMNC_ID != oldMCCMNC_ID) { //newApp
                    globalData.OperatorRows = append(globalData.OperatorRows, operator) //add the previous operator
                    operator = GlobalOperatorRow{} //initialize new operator
                    operator.MCCMNC_ID = newMCCMNC_ID
                    operator.Operator_Name = newOperator_Name
                    operator.ConfigNumbers = append (operator.ConfigNumbers, newConfig_ID) //add configs to configList
                    oldMCCMNC_ID = operator.MCCMNC_ID //set oldMCCMNC_ID to new operator MCCMNC_ID
                } else if (newMCCMNC_ID == oldMCCMNC_ID){ //same operator
                    operator.ConfigNumbers = append (operator.ConfigNumbers, newConfig_ID) //add configs to configList
                }
            }
            globalData.OperatorRows = append(globalData.OperatorRows, operator) //add the last operator
            jsonResponse, err := json.Marshal(globalData)
            checkErr(err)
            log.Println("globalView –\t\tReturning JSON string...")
            return jsonResponse
        } else { //only appName specified, show countries + configs
            //get all Countries that the app is in, and for loop through each country, checking which operators are selected. If all operators are selected, make the operatorList empty. If only part of them are selected, make operatorList contain only the selected ones.
            globalData := GlobalData{}
            log.Println("globalView –\t\tGlobalView | Country Level")
            globalViewQuery := `SELECT DISTINCT Country_ID, name from countries
            WHERE Country_ID in (SELECT DISTINCT Country_ID
            from configurationMappings WHERE Config_ID in
            (SELECT Config_ID from appConfigs
            where originalName = '`+Data.AppOriginalName+`'))

            OR  Country_ID in (SELECT DISTINCT Country_ID
            from operators WHERE MCCMNC_ID in
            (SELECT MCCMNC_ID from configurationMappings
            WHERE Config_ID in (SELECT Config_ID from appConfigs
            where originalName = '`+Data.AppOriginalName+`')))`
            log.Println("globalView –\t\tQuery looks like...\n" + globalViewQuery)
            countryList, err := db.Query(globalViewQuery)
            checkErr(err)
            for countryList.Next() { //for each country this app exists in, fill the GlobalDataCountry inside GlobalDataApp
                var Country_ID string
                var CountryName string
                var App_Config_ID string //if all operators inside country mapped to the same config
                countryList.Scan(&Country_ID, &CountryName)
                globalDataCountry := GlobalDataCountry{}
                globalDataCountry.Country_ID = Country_ID
                globalDataCountry.CountryName = CountryName



                globalViewQuery = `SELECT DISTINCT Config_ID from ConfigurationMappings WHERE Config_ID in
                (SELECT Config_ID from AppConfigs WHERE originalName = '`+Data.AppOriginalName+`')
                AND  (MCCMNC_ID in (SELECT MCCMNC_ID from operators WHERE Country_ID = '`+Country_ID+`') OR Country_ID = '`+Country_ID+`')`
                configList, err := db.Query(globalViewQuery)
                checkErr(err)
                ConfigCount := 0
                for configList.Next() { //sets configList for app, counts number of configs
                    ConfigCount++
                    var configNumber string
                    configList.Scan(&configNumber)
                    globalDataCountry.ConfigNumbers = append(globalDataCountry.ConfigNumbers, configNumber)
                }
                configList.Close()

                globalViewQuery = `SELECT configurationMappings.Config_ID from ConfigurationMappings
                INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
                INNER JOIN products ON appConfigs.Config_ID = products.Config_ID
                WHERE configurationMappings.Config_ID in
                (SELECT Config_ID from AppConfigs WHERE originalName = '`+Data.AppOriginalName+`')
                AND Country_ID = '`+Country_ID+`' AND products.productName = "homescreen" ORDER BY configurationMappings.id DESC LIMIT 1`
                activeConfigs, err := db.Query(globalViewQuery)
                checkErr(err)
                for activeConfigs.Next() {
                    var configNumber string
                    activeConfigs.Scan(&configNumber)
                    if(configNumber != "") {
                        globalDataCountry.ActiveConfigs = AppendIfMissing(globalDataCountry.ActiveConfigs, configNumber)
                    }
                }
                activeConfigs.Close()
                globalViewQuery = `SELECT configurationMappings.Config_ID from ConfigurationMappings
                INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
                INNER JOIN products ON appConfigs.Config_ID = products.Config_ID
                WHERE configurationMappings.Config_ID in
                (SELECT Config_ID from AppConfigs WHERE originalName = '`+Data.AppOriginalName+`')
                AND Country_ID = '`+Country_ID+`' AND products.productName = "folder" ORDER BY configurationMappings.id DESC LIMIT 1`
                activeConfigs, err = db.Query(globalViewQuery)
                checkErr(err)
                for activeConfigs.Next() {
                    var configNumber string
                    activeConfigs.Scan(&configNumber)
                    if(configNumber != "") {
                        globalDataCountry.ActiveConfigs = AppendIfMissing(globalDataCountry.ActiveConfigs, configNumber)
                    }
                }
                activeConfigs.Close()
                globalViewQuery = `SELECT configurationMappings.Config_ID from ConfigurationMappings
                INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
                INNER JOIN products ON appConfigs.Config_ID = products.Config_ID
                WHERE configurationMappings.Config_ID in
                (SELECT Config_ID from AppConfigs WHERE originalName = '`+Data.AppOriginalName+`')
                AND Country_ID = '`+Country_ID+`' AND products.productName = "max" ORDER BY configurationMappings.id DESC LIMIT 1`
                activeConfigs, err = db.Query(globalViewQuery)
                checkErr(err)
                for activeConfigs.Next() {
                    var configNumber string
                    activeConfigs.Scan(&configNumber)
                    if(configNumber != "") {
                        globalDataCountry.ActiveConfigs = AppendIfMissing(globalDataCountry.ActiveConfigs, configNumber)
                    }
                }
                activeConfigs.Close()
                globalViewQuery = `SELECT configurationMappings.Config_ID from ConfigurationMappings
                INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
                INNER JOIN products ON appConfigs.Config_ID = products.Config_ID
                WHERE configurationMappings.Config_ID in
                (SELECT Config_ID from AppConfigs WHERE originalName = '`+Data.AppOriginalName+`')
                AND Country_ID = '`+Country_ID+`' AND products.productName = "maxGo" ORDER BY configurationMappings.id DESC LIMIT 1`
                activeConfigs, err = db.Query(globalViewQuery)
                checkErr(err)
                for activeConfigs.Next() {
                    var configNumber string
                    activeConfigs.Scan(&configNumber)
                    if(configNumber != "") {
                        globalDataCountry.ActiveConfigs = AppendIfMissing(globalDataCountry.ActiveConfigs, configNumber)
                    }
                }
                activeConfigs.Close()

                globalViewQuery = `
                SELECT * FROM (SELECT DISTINCT MCCMNC_ID from operators WHERE Country_ID="`+Country_ID+`")
                EXCEPT
                SELECT * FROM (SELECT DISTINCT MCCMNC_ID from operators WHERE MCCMNC_ID in
                (SELECT MCCMNC_ID from configurationMappings WHERE Config_ID in
                (SELECT Config_ID from AppConfigs WHERE originalName = '`+Data.AppOriginalName+`')
                AND MCCMNC_ID in (SELECT MCCMNC_ID from operators where Country_ID = "`+Country_ID+`")));`
                differenceList, err := db.Query(globalViewQuery)
                checkErr(err)
                differenceCount := 0
                for differenceList.Next() { //gets the amount of unmapped operators within a country
                    differenceCount++
                }

                if(ConfigCount == 1) { //only one config for whole country
                    for configList.Next() {
                        configList.Scan(&App_Config_ID)
                    }
                    if(differenceCount == 0) { //every operator is mapped
                        globalDataCountry.App_Config_ID = App_Config_ID //sets the single config for the whole country!
                    } else { //every operator IS NOT MAPPED, but there is still only one config
                        globalViewQuery = `SELECT DISTINCT MCCMNC_ID, operators.Operator_Name from ConfigurationMappings
                        JOIN     operators USING (MCCMNC_ID)
                        WHERE Config_ID = '`+App_Config_ID+`' AND MCCMNC_ID in (SELECT MCCMNC_ID from operators WHERE Country_ID = '`+Country_ID+`');`
                        mappedOperators, err := db.Query(globalViewQuery)
                        checkErr(err)
                        for mappedOperators.Next() {
                            globalOperatorRow := GlobalOperatorRow{}
                            mappedOperators.Scan(&globalOperatorRow.MCCMNC_ID, &globalOperatorRow.Operator_Name)
                            globalOperatorRow.App_Config_ID = App_Config_ID
                            globalDataCountry.OperatorRows = append(globalDataCountry.OperatorRows, globalOperatorRow)
                        }
                    }

                } else { //multiple App_Config_ID's found for this country
                    for configList.Next() { //for all possible config-id's, find the operator mapped to it, store it in globalOperatorRow
                        configList.Scan(&App_Config_ID)
                        globalViewQuery = `SELECT DISTINCT MCCMNC_ID, operators.Operator_Name from ConfigurationMappings
                        JOIN     operators USING (MCCMNC_ID)
                        WHERE Config_ID = '`+App_Config_ID+`'
                        AND  MCCMNC_ID in (SELECT MCCMNC_ID from operators WHERE Country_ID = '`+Country_ID+`');`

                        operatorRows, err := db.Query(globalViewQuery)
                        checkErr(err)
                        for operatorRows.Next() {
                            globalOperatorRow := GlobalOperatorRow{}
                            operatorRows.Scan(&globalOperatorRow.MCCMNC_ID, &globalOperatorRow.Operator_Name)
                            globalOperatorRow.App_Config_ID = App_Config_ID
                            globalDataCountry.OperatorRows = append(globalDataCountry.OperatorRows, globalOperatorRow)
                        }
                    }
                }
                globalData.GlobalDataCountries = append(globalData.GlobalDataCountries, globalDataCountry)
            }

            jsonResponse, err := json.Marshal(globalData)
            checkErr(err)
            log.Println("globalView –\t\tReturning JSON string...")
            return jsonResponse
        }
    } else { //appName null, show all appNames + configs
        globalViewQuery := `SELECT DISTINCT Config_ID, AppConfigs.originalName from configurationMappings
        JOIN AppConfigs USING (Config_ID)
        WHERE Config_ID in (SELECT DISTINCT Config_ID from AppConfigs)
    	ORDER BY AppConfigs.rank;`
        log.Println("globalView –\t\tGlobalView | App/Index Level")
        appListWithRedundancies, err := db.Query(globalViewQuery)
        checkErr(err)



        globalData := GlobalData{}
        app := GlobalDataApp{}
        oldAppName := ""

        for appListWithRedundancies.Next() {
            var newAppName string
            var newConfig_ID string

            appListWithRedundancies.Scan(&newConfig_ID, &newAppName)
            log.Println("globalView –\t\tAppList | " + newAppName + " | " +newConfig_ID)
            if(oldAppName == "") { //beginning of the list
                oldAppName = newAppName //set oldAppName
                app = GlobalDataApp{} //initialize new app
                app.OriginalName = newAppName
                app.ConfigNumbers = append(app.ConfigNumbers, newConfig_ID) //add configs to configList
            } else if(newAppName != oldAppName) { //newApp
                globalData.GlobalDataApps = append(globalData.GlobalDataApps, app) //add the previous app object to array
                oldAppName = newAppName //set old app name to newAppName
                app = GlobalDataApp{} //initialize new app
                app.OriginalName = newAppName
                app.ConfigNumbers = append (app.ConfigNumbers, newConfig_ID) //add configs to configList
            } else if (app.OriginalName == oldAppName){ //same App
                app.ConfigNumbers = append (app.ConfigNumbers, newConfig_ID) //add configs to configList
            }
        }
        appListWithRedundancies.Close()
        globalData.GlobalDataApps = append(globalData.GlobalDataApps, app) //add the last app object to array

        jsonResponse, err := json.Marshal(globalData)
        checkErr(err)
        log.Println("globalView –\t\tReturning JSON string...")
        return jsonResponse
    }
    return nil
}



func addNewFeaturesAndProducts(Config data, New_App_Config_ID_string string) () {
        statement := string("")
        //defaultenabledfeatures
        if(Config.DefaultEnabledFeatures.Savings ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"defaultEnabledFeatures", "savings"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        if(Config.DefaultEnabledFeatures.Privacy ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"defaultEnabledFeatures", "privacy"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        if(Config.DefaultEnabledFeatures.Adblock ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"defaultEnabledFeatures", "adBlock"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        if(Config.DefaultEnabledFeatures.NoImages ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"defaultEnabledFeatures", "noImages"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        //hiddenfeatures
        if(Config.DefaultHiddenFeatures.Savings ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"hiddenFeatures", "savings"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        if(Config.DefaultHiddenFeatures.Privacy ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"hiddenFeatures", "privacy"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        if(Config.DefaultHiddenFeatures.Adblock ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"hiddenFeatures", "adBlock"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        if(Config.DefaultHiddenFeatures.NoImages ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"hiddenFeatures", "noImages"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        //hiddenUI
        if(Config.DefaultHiddenUI.Splash ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"hiddenUI", "spash"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        if(Config.DefaultHiddenUI.Overlay ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"hiddenUI", "overlay"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        if(Config.DefaultHiddenUI.AB ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"hiddenUI", "ab"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        if(Config.DefaultHiddenUI.Badges ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"hiddenUI", "badges"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        if(Config.DefaultHiddenUI.Folder ==true) {
            statement = string(`INSERT INTO featureMappings (Config_ID, featureType, featureName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"hiddenUI", "folder"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }

        //products
        if(Config.Products.MaxGlobal ==true) {
            statement = string(`INSERT INTO products (Config_ID, productName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"maxGlobal"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        if(Config.Products.Max ==true) {
            statement = string(`INSERT INTO products (Config_ID, productName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"max"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }
        if(Config.Products.MaxGo ==true) {
            statement = string(`INSERT INTO products (Config_ID, productName) VALUES (` + `"` + New_App_Config_ID_string + `", ` + `"maxGo"` + `)`)
            _, err := db.Exec(statement)
            checkErr(err)
        }

        //packages
        for _, packageName := range Config.Packages {
            statement = string(`INSERT INTO packages (Config_ID, packageName) VALUES (` + `"` + New_App_Config_ID_string + `", "` + packageName + `")`)
            _, err := db.Exec(statement)
            checkErr(err)
        }

}

func addNewConfig(Config data) ([]byte) {
    log.Println("addNewConfig –\t\tRecieved request to add " + Config.AppOriginalName + " | Rank: " + Config.AppRank)
    log.Println(Config.DefaultHiddenUI)
    log.Println(Config.DefaultHiddenFeatures)
    log.Println(Config.DefaultEnabledFeatures)
    log.Println(Config.Products)

    statement := string(`INSERT INTO appConfigs (originalName, modifiableName, iconURL, homeURL, rank, category) VALUES (` + `"` + Config.AppOriginalName + `", "` + Config.AppModifiableName + `", "` + Config.AppIconURL + `", "` + Config.AppHomeURL + `", "` + Config.AppRank + `", "` + Config.AppCategory + `"` + `)`)
    log.Println("addNewConfig –\t\tInsert statement: " + statement)
    res, err := db.Exec(statement)
    checkErr(err)
    id, err := res.LastInsertId()
    checkErr(err)
    log.Println("addNewConfig –\t\tLast insert id: ", id)
    var New_App_Config_ID = id
    New_App_Config_ID_string := strconv.Itoa(int(New_App_Config_ID))

    addNewFeaturesAndProducts(Config, New_App_Config_ID_string) //handles products Table inserts and featureMappings table inserts

    for _, country := range Config.AppConfigurationMappings.Countries {
        mappingstatement := string(`INSERT INTO configurationMappings (Config_ID, Country_ID)
        VALUES (`+New_App_Config_ID_string+`, '`+country+`')`)
        log.Println("addNewConfig –\t"+mappingstatement)
        res, err = db.Exec(mappingstatement)
        checkErr(err)
    }
    for _, operator := range Config.AppConfigurationMappings.Operators {

        mappingstatement := string(`INSERT INTO configurationMappings (Config_ID, MCCMNC_ID) VALUES (`+New_App_Config_ID_string+`, "`+operator+`")`)
        log.Println("addNewConfig –\t\t"+mappingstatement)
        res, err = db.Exec(mappingstatement)
        checkErr(err)
    }
    if(Config.AppExistsEverywhere) {
        mappingstatement := string(`INSERT INTO configurationMappings (Config_ID, Country_ID)
        SELECT `+New_App_Config_ID_string+`, Country_ID  FROM countries`)
        log.Println("addNewConfig –\t\tApp EXISTS EVERYWHERE "+mappingstatement)
        res, err = db.Exec(mappingstatement)
        checkErr(err)
    }
    for _, operatorGroup := range Config.AppConfigurationMappings.OperatorGroups {
        mappingstatement := string(`INSERT INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT `+New_App_Config_ID_string+`, MCCMNC_ID FROM operatorGroups WHERE Operator_Group_Name = "`+operatorGroup+`"`)
        log.Println("addNewConfig –\t\t"+mappingstatement)
        res, err = db.Exec(mappingstatement)
        checkErr(err)
    }


    var returnResult = ResultMessage{"SUCCESS"}
    jsonResponse, err := json.Marshal(returnResult)
    checkErr(err)
    return jsonResponse
}
type ResultMessage struct {
    Result string `json:"result"`
}

func submitNewOperator(Operator data) ([]byte) {
    log.Println("submitNewOperator –\t\tRecieved request to add operator named " + Operator.OperatorName)
    statement := string(`SELECT Operator_Name, Country_ID FROM mytable WHERE MCCMNC_ID = "`+Operator.MCCMNC_ID+`"`)
    operator, err := db.Query(statement)
    checkErr(err)
    var operatorExists bool
    var operatorName string = ""
    var operatorCountry_ID string = ""
    for(operator.Next()){
        operator.Scan(&operatorName, &operatorCountry_ID)
        operatorExists = true
    }
    if(Operator.OperatorName != "") {
        operatorName = Operator.OperatorName
    }
    if(Operator.Country_ID != "") {
        operatorCountry_ID = Operator.Country_ID
    }
    var returnString string = ""
    if(operatorExists) {
        addOperatorStatement := string(`INSERT INTO operators (MCCMNC_ID, Operator_Name, Country_ID) VALUES ("`+Operator.MCCMNC_ID+`", "`+operatorName+`", "`+operatorCountry_ID+`")`)
        log.Println("submitNewOperator –\t\t"+addOperatorStatement)
        _, err := db.Exec(addOperatorStatement)
        checkErr(err)
        addOperatorStatement = string(`INSERT INTO operatorGroups (Operator_Group_Name, MCCMNC_ID) VALUES ("`+Operator.Operator_Group_Name+`", "`+Operator.MCCMNC_ID+`")`)
        log.Println("submitNewOperator –\t\t"+addOperatorStatement)
        _, err = db.Exec(addOperatorStatement)
        checkErr(err)
        returnString = "success"
    } else {
        returnString = "error"
    }

    jsonResponse, err := json.Marshal(returnString)
    checkErr(err)
    return jsonResponse
}
func deleteOperator(Operator data) ([]byte) {
    var returnString string = ""
    if(Operator.MCCMNC_ID != "") {
        deleteOperatorStatement := string(`DELETE FROM operators WHERE MCCMNC_ID = "`+Operator.MCCMNC_ID+`"`)
        log.Println("deleteOperator –\t\t"+deleteOperatorStatement)
        _, err := db.Exec(deleteOperatorStatement)
        checkErr(err)
        deleteOperatorStatement = string(`DELETE FROM operatorGroups WHERE MCCMNC_ID = "`+Operator.MCCMNC_ID+`"`)
        log.Println("deleteOperator –\t\t"+deleteOperatorStatement)
        _, err = db.Exec(deleteOperatorStatement)
        checkErr(err)
        returnString = "success"
    } else {
        returnString = "error"
    }

    jsonResponse, err := json.Marshal(returnString)
    checkErr(err)
    return jsonResponse
}
func deleteConfiguration(Data data) ([]byte) {
    var returnString string = ""
    if(Data.Config_ID != "") {
        deleteConfigurationstatement := string(`DELETE FROM appConfigs WHERE Config_ID = "`+Data.Config_ID+`"`)
        log.Println("deleteOperator –\t\t"+deleteConfigurationstatement)
        _, err := db.Exec(deleteConfigurationstatement)
        checkErr(err)
        deleteConfigurationstatement = string(`DELETE FROM configurationMappings WHERE Config_ID = "`+Data.Config_ID+`"`)
        log.Println("deleteOperator –\t\t"+deleteConfigurationstatement)
        _, err = db.Exec(deleteConfigurationstatement)
        checkErr(err)
        deleteConfigurationstatement = string(`DELETE FROM products WHERE Config_ID = "`+Data.Config_ID+`"`)
        log.Println("deleteOperator –\t\t"+deleteConfigurationstatement)
        _, err = db.Exec(deleteConfigurationstatement)
        checkErr(err)
        deleteConfigurationstatement = string(`DELETE FROM featureMappings WHERE Config_ID = "`+Data.Config_ID+`"`)
        log.Println("deleteOperator –\t\t"+deleteConfigurationstatement)
        _, err = db.Exec(deleteConfigurationstatement)
        checkErr(err)
        returnString = "success"
    } else {
        returnString = "error"
    }

    jsonResponse, err := json.Marshal(returnString)
    checkErr(err)
    return jsonResponse
}
func updateConfigurationINI(Data data) ([]byte) {
    _, err := ioutil.ReadFile("static/ultra_apps_configuration/configuration.ini")
    checkErr(err)
    log.Println("updateConfigurationINI –\t\tupdating configuration.ini...")

    //clear static/ultra_apps_json folder first
    RemoveContents("static/ultra_apps_json")
    output := generateConfigurationINI()
    err = ioutil.WriteFile("static/ultra_apps_configuration/configuration.ini", []byte(output), 0644)
    checkErr(err)
    log.Println("updateConfigurationINI –\t\t wrote to file")

    //package together static/configuration.ini, static/ultra_apps_json, and static/ultra_apps into static/configuration.zip
    files := []string{"static/ultra_apps_configuration/configuration.ini", "static/ultra_apps_json", "static/ultra_apps"}
    zipOutput := "static/configuration.zip"
    err = archiver.Zip.Make(zipOutput, files)

    log.Println("Zipped File: " + zipOutput)

    jsonResponse, err := json.Marshal("success")
    checkErr(err)
    return jsonResponse
}
func RemoveContents(dir string) error {
    d, err := os.Open(dir)
    if err != nil {
        return err
    }
    defer d.Close()
    names, err := d.Readdirnames(-1)
    if err != nil {
        return err
    }
    for _, name := range names {
        err = os.RemoveAll(filepath.Join(dir, name))
        if err != nil {
            return err
        }
    }
    return nil
}

func submitNewCountry(Country data) ([]byte) {
    var returnString string = ""
    if(Country.Country_ID != "") {
        addCountrystatement := string(`INSERT INTO countries (Country_ID, name, MCC_ID) VALUES ("`+Country.MCCMNC_ID+`", "`+Country.Country_Name+`", "`+Country.Country_MCC+`")`)
        log.Println("submitNewCountry –\t\t"+addCountrystatement)
        _, err := db.Exec(addCountrystatement)
        checkErr(err)
        returnString = "success"
    } else {
        returnString = "error"
    }

    jsonResponse, err := json.Marshal(returnString)
    checkErr(err)
    return jsonResponse
}

func getOperatorsByCountryID(Country data) ([]byte) {
    log.Println("getOperatorsByCountryID –\tRecieved request to get operators in " + Country.Country_ID)

    var operatorRows = OperatorRows{}
    full_query := string(`
    SELECT MCCMNC_ID, Operator_Name, Country_ID  from operators
    WHERE Country_ID='`+Country.Country_ID+`'
    ORDER BY Operator_Name`)
    rows, err := db.Query(full_query)
    checkErr(err)
    for rows.Next() {
        var operatorRow = OperatorRow{}
        rows.Scan(&operatorRow.MCCMNC_ID, &operatorRow.Operator_Name, &operatorRow.Country_ID)
        operatorRows.OperatorRows = append(operatorRows.OperatorRows, operatorRow)
    }
    rows.Close()

    jsonResponse, err := json.Marshal(operatorRows)
    checkErr(err)
    return jsonResponse
}
type AppConfigs struct {
    AppConfigs []AppConfig `json:"appConfigs, omitempty"`
}
type AppConfig struct {
    Config_ID string `json: "Config_ID" db:"Config_ID"`
    OriginalName string `json: "OriginalName" db:"originalName"`
    ModifiableName string `json: "ModifiableName" db:"modifiableName"`
    IconURL string `json: "IconURL" db:"iconURL"`
    HomeURL string `json: "HomeURL" db:"homeURL"`
    Rank string `json: "Rank" db:"rank"`
    Category string `json: "Category" db:"category"`
}
func getAllAppConfigs(notUsed data) ([]byte) {
    log.Println("getAllAppConfigs –\tRecieved request to get all app configs")
    var appConfigs = AppConfigs{}

    full_query := string(`
    SELECT DISTINCT Config_ID, originalName, modifiableName,iconURL, homeURL, rank, category  from appConfigs`)
    allAppConfigs, err := db.Query(full_query)
    checkErr(err)
    for allAppConfigs.Next() {
        var AppConfig = AppConfig{}
        allAppConfigs.Scan(&AppConfig.Config_ID, &AppConfig.OriginalName, &AppConfig.ModifiableName, &AppConfig.IconURL, &AppConfig.HomeURL, &AppConfig.Rank, &AppConfig.Category)
        appConfigs.AppConfigs = append(appConfigs.AppConfigs, AppConfig)
    }
    allAppConfigs.Close()
    jsonResponse, err := json.Marshal(appConfigs)
    checkErr(err)
    return jsonResponse
}
func getCountryByName(Country data) ([]byte) {
    log.Println("getCountryByName –\t\tRecieved request to get country by " + Country.Country_Name)

    var countryRow = CountryRow{}
    full_query := string(`
    SELECT Country_ID, name, MCC_ID from countries
    WHERE name="`+Country.Country_Name+`"
    `)
    rows, err := db.Query(full_query)
    checkErr(err)
    for rows.Next() {
        rows.Scan(&countryRow.Country_ID, &countryRow.CountryName, &countryRow.MCC_ID)
        log.Println("getCountryByName –\t\t" + countryRow.Country_ID + " | " + countryRow.CountryName + " | " + countryRow.MCC_ID)
    }
    rows.Close()

    jsonResponse, err := json.Marshal(countryRow)
    checkErr(err)
    return jsonResponse
}

func getOperatorGroupByName(Operator data) ([]byte) {
    log.Println("getOperatorGroupByName –\tRecieved request to get operator by " + Operator.Operator_Group_Name)

    var operatorRows = OperatorRows{}
    full_query := string(`
    SELECT MCCMNC_ID, Operator_Name, Country_ID, operatorGroups.Operator_Group_Name from operators
    JOIN operatorGroups USING (MCCMNC_ID)
    WHERE operatorGroups.Operator_Group_Name="`+Operator.Operator_Group_Name+`"
    ORDER BY Operator_Name`)
    rows, err := db.Query(full_query)
    checkErr(err)
    for rows.Next() {
        var operatorRow = OperatorRow{}
        rows.Scan(&operatorRow.MCCMNC_ID, &operatorRow.Operator_Name, &operatorRow.Country_ID, &operatorRow.Operator_Group_Name)
        operatorRows.OperatorRows = append(operatorRows.OperatorRows, operatorRow)
    }
    rows.Close()

    jsonResponse, err := json.Marshal(operatorRows)
    checkErr(err)
    return jsonResponse
}

func getproducts(Config data) ([]byte) {
    log.Println("getproducts –\tRecieved request to get featured locations for " + Config.Config_ID)

    var products = []string{}
    full_query := string(`
    SELECT DISTINCT productName from products
    WHERE Config_ID="`+Config.Config_ID+`"
    `)
    rows, err := db.Query(full_query)
    checkErr(err)
    for rows.Next() {
        featuredLocation := ""
        rows.Scan(&featuredLocation)
        products = append(products, featuredLocation)
    }
    rows.Close()

    jsonResponse, err := json.Marshal(products)
    checkErr(err)
    return jsonResponse
}
type FeatureMapping struct {
    FeatureType string `json:"FeatureName" db:"featureType"`
    FeatureName string `json:"FeatureType" db:"featureName"`
}
func getFeatureMappings(Config data) ([]byte) {
    log.Println("getFeatureMappings –\tRecieved request to get featured locations for " + Config.Config_ID)

    var featureMappings = []FeatureMapping{}
    full_query := string(`
    SELECT DISTINCT featureType, featureName from featureMappings
    WHERE Config_ID="`+Config.Config_ID+`"  ORDER BY featureType
    `)
    rows, err := db.Query(full_query)
    checkErr(err)
    log.Println("getFeatureMappings –\t"+full_query)
    for rows.Next() {
        featureMapping := FeatureMapping{}
        rows.Scan(&featureMapping.FeatureType, &featureMapping.FeatureName)
        featureMappings = append(featureMappings, featureMapping)
    }
    rows.Close()

    jsonResponse, err := json.Marshal(featureMappings)
    checkErr(err)
    return jsonResponse
}

type MappingsCountryRow struct {
    Value string `json:"name" db:"name"`
    Text string `json:"Country_ID" db:"Country_ID"`
}
type MappingsOperatorRow struct {
    Value string `json:"MCCMNC_ID" db:"MCCMNC_ID" `
    Text string `json:"Operator_Name" db:"Operator_Name"`
    Group string `json:"Operator_Group_Name" db:"Operator_Group_Name"`
}
type ConfigurationMappingsResult struct{
    CountryRows []MappingsCountryRow `json:"countryFilterRows"`
    OperatorRows []MappingsOperatorRow `json:"operatorFilterRows"`
}
func getConfigurationMappings(Config data) ([]byte) {
    log.Println("getConfigurationMappings –\tRecieved request to get configuration mappings for " + Config.Config_ID)

    full_query := string(`
    SELECT Country_ID FROM (SELECT DISTINCT Country_ID FROM countries)
    EXCEPT
    SELECT Country_ID FROM(SELECT Country_ID, countries.name from configurationMappings
    JOIN countries USING (Country_ID)
    WHERE Config_ID="`+Config.Config_ID+`" AND Country_ID IS NOT NULL)
    `)
    rows, err := db.Query(full_query)
    checkErr(err)
    log.Println("getConfigurationMappings –\t"+full_query)
    var allCountriesExist = true
    for rows.Next() {
        allCountriesExist = false
    }
    rows.Close()
    var result = ConfigurationMappingsResult{}

    if(!allCountriesExist)  {

        full_query = string(`
        SELECT Country_ID, countries.name from configurationMappings
        JOIN countries USING (Country_ID)
        WHERE Config_ID="`+Config.Config_ID+`" AND Country_ID IS NOT NULL
        `)
        countryRows, err := db.Query(full_query)
        checkErr(err)
        log.Println("getConfigurationMappings –\t"+full_query)
        for countryRows.Next() {
            countryRow := MappingsCountryRow{}
            countryRows.Scan(&countryRow.Text, &countryRow.Value)
            log.Println("getConfigurationMappings –\t"+countryRow.Text)
            result.CountryRows = append(result.CountryRows, countryRow)
        }
        countryRows.Close()
    } else {
        countryRow := MappingsCountryRow{}
        countryRow.Text = "*"
        countryRow.Value = "*"
        result.CountryRows = append(result.CountryRows, countryRow)
    }

    full_query = string(`
    SELECT MCCMNC_ID, operators.Operator_Name, operatorGroups.Operator_Group_Name from configurationMappings
    JOIN operators USING (MCCMNC_ID)
    JOIN operatorGroups USING (MCCMNC_ID)
    WHERE Config_ID="`+Config.Config_ID+`" AND MCCMNC_ID IS NOT NULL ORDER BY operatorGroups.Operator_Group_Name
    `)
    operatorRows, err := db.Query(full_query)
    checkErr(err)
    log.Println("getConfigurationMappings –\t"+full_query)
    for operatorRows.Next() {
        operatorRow := MappingsOperatorRow{}
        operatorRows.Scan(&operatorRow.Value, &operatorRow.Text, &operatorRow.Group)
        result.OperatorRows = append(result.OperatorRows, operatorRow)
    }
    operatorRows.Close()
    log.Println(result)
    jsonResponse, err := json.Marshal(result)
    checkErr(err)
    return jsonResponse
}


type CountryFilterRow struct {
    Value string `json:"name" db:"name"`
    Text string `json:"Country_ID" db:"Country_ID"`
}
type OperatorFilterRow struct {
    Value string `json:"MCCMNC_ID" db:"MCCMNC_ID" `
    Text string `json:"Operator_Name" db:"Operator_Name"`
}

type FilterRows struct{
    CountryFilterRows []CountryFilterRow `json:"countryFilterRows"`
    OperatorFilterRows []OperatorFilterRow `json:"operatorFilterRows"`
}
func updateFilterValues(Filters data) ([]byte) {
    log.Println("updateFilterValues –\tRecieved request to update filter items based off existing filter selection.")

    var filterRows = FilterRows{}

    if(Filters.Selected_operator != "star") { //if operator is not star, I don't need to update country dropdown

        full_query := string(`
        SELECT countries.Country_ID, countries.name from operators
        JOIN     countries USING (Country_ID)
        WHERE MCCMNC_ID="`+Filters.Selected_operator+`" LIMIT 1
        `)
        rows, err := db.Query(full_query)
        checkErr(err)
        country_ID := string("")
        for rows.Next() {
            var countryFilterRow = CountryFilterRow{}
            rows.Scan(&countryFilterRow.Value, &countryFilterRow.Text)
            country_ID = countryFilterRow.Value
            if(Filters.Selected_country == "star") {
                filterRows.CountryFilterRows = append(filterRows.CountryFilterRows, countryFilterRow)
            }
        }
        rows.Close()

        full_query = string(`
            SELECT MCCMNC_ID, Operator_Name from operators
            WHERE Country_ID = "`+country_ID+`"
        `)
        rows, err = db.Query(full_query)
        checkErr(err)

        //operator query here, returns rows
        for rows.Next() {
            var operatorFilterRow = OperatorFilterRow{}
            rows.Scan(&operatorFilterRow.Value, &operatorFilterRow.Text)
            filterRows.OperatorFilterRows = append(filterRows.OperatorFilterRows, operatorFilterRow)
        }
        rows.Close()

    } else if (Filters.Selected_country != "star") { //operator IS star, country is NOT star. Thus we need to update operator dropdown
        full_query := string(`
            SELECT MCCMNC_ID, Operator_Name from operators
            WHERE Country_ID = "`+Filters.Selected_country+`"
        `)
        rows, err := db.Query(full_query)
        checkErr(err)

        //operator query here, returns rows
        for rows.Next() {
            var operatorFilterRow = OperatorFilterRow{}
            rows.Scan(&operatorFilterRow.Value, &operatorFilterRow.Text)
            filterRows.OperatorFilterRows = append(filterRows.OperatorFilterRows, operatorFilterRow)
        }
        rows.Close()
    } else { //stars in both country and operator, load full tables

        full_query := string(`SELECT DISTINCT Country_ID, name from countries WHERE Country_ID in (SELECT Country_ID from favoriteCountries)`) //country query -- all distinct countries by value and name
        rows, err := db.Query(full_query)
        checkErr(err)
        for rows.Next() {
            var countryFilterRow = CountryFilterRow{}
            rows.Scan(&countryFilterRow.Value, &countryFilterRow.Text)
            filterRows.CountryFilterRows = append(filterRows.CountryFilterRows, countryFilterRow)
        }
        rows.Close()

        full_query = string(`SELECT DISTINCT MCCMNC_ID, Operator_Name from operators`) //operator query -- all distinct operators
        rows, err = db.Query(full_query)
        checkErr(err)
        for rows.Next() {
            var operatorFilterRow = OperatorFilterRow{}
            rows.Scan(&operatorFilterRow.Value, &operatorFilterRow.Text)
            filterRows.OperatorFilterRows = append(filterRows.OperatorFilterRows, operatorFilterRow)
        }
        rows.Close()
    }

    jsonResponse, err := json.Marshal(filterRows)
    checkErr(err)
    return jsonResponse
}

type AppsContainer struct {
    Apps []App
}

type App struct {
    Config_ID string
    OriginalName string
    ModifiableName string
    IconUrl string
    HomeUrl string
    Category string `db: "category, omitempty"`
    Rank int `db: "rank"`
    ProductName string `json: ", omitempty"`

}

func appView(Data data) ([]byte) {
    log.Println("appView –\t\tquerying db...")

    country_code := string("")
    operator_query := string("")
    if(Data.Selected_operator != "star") { //more specific than country
        operator_query = `AND MCCMNC_ID like "%` + Data.Selected_operator +`%"`
    } else if (Data.Selected_country != "star") {
        country_code = Data.Selected_country
    }

    appViewQuery := `SELECT DISTINCT appConfigs.Config_ID, originalName, modifiableName,
    iconURL, homeURL, rank, configurationMappings.productName FROM appConfigs
    JOIN     configurationMappings USING (Config_ID)
    WHERE originalName = "` + Data.App_name +`"
    AND Config_ID in (SELECT Config_ID from configurationMappings
    WHERE MCCMNC_ID IN (SELECT MCCMNC_ID FROM operators WHERE Country_ID like "%`+country_code+`%" ` + operator_query + `))` + ` LIMIT 1`
    rows, err := db.Query(appViewQuery)
    checkErr(err)
    log.Println("appView –\t\t" +  appViewQuery)

    var app = App{}
    for rows.Next() {

        rows.Scan(&app.Config_ID,&app.OriginalName, &app.ModifiableName, &app.IconUrl, &app.HomeUrl, &app.Rank, &app.ProductName)
    }
    defer rows.Close()

    jsonResponse, err := json.Marshal(app)
    jsonString := string(jsonResponse)
    checkErr(err)
    log.Println("appView –\t\tReturning the following JSON string:")
    log.Println(jsonString)
    return jsonResponse
}
func getAppConfig(Data data) ([]byte) {

    getAppConfigQuery := `SELECT Config_ID, originalName, modifiableName, iconURL, homeURL, rank, category FROM appConfigs WHERE Config_ID = "`+Data.Config_ID+`"`
    rows, err := db.Query(getAppConfigQuery)
    checkErr(err)

    log.Println("getAppConfig –\t\t" +  getAppConfigQuery)

    var app = App{}
    for (rows.Next()) {
        rows.Scan(&app.Config_ID, &app.OriginalName, &app.ModifiableName, &app.IconUrl, &app.HomeUrl,  &app.Rank, &app.Category)
    }
    jsonResponse, err := json.Marshal(app)
    jsonString := string(jsonResponse)
    checkErr(err)
    log.Println("getAppConfig –\t\tReturning the following JSON string:")
    log.Println(jsonString)
    return jsonResponse
}
func loadAppTray(Filters data) ([]byte) {
    log.Println("loadAppTray –\t\tquerying db")

    searchfield_query := string("")
    country_code := string("")
    operator_query := string("")
    country_query := string("")
    full_query := string("")
    if(Filters.Searchfield_text != ""){ //search field is NOT empty
        searchfield_query = ` AND originalName like "%` + Filters.Searchfield_text +`%"`
    }

    if(Filters.Selected_operator != "star") { //more specific than country
        operator_query = `Config_ID in (SELECT Config_ID FROM configurationMappings WHERE MCCMNC_ID = "`+Filters.Selected_operator+`") `

        full_query = string(`
        SELECT DISTINCT appConfigs.Config_ID, originalName, modifiableName, iconURL,
        homeUrl, rank, products.productName FROM appConfigs
        JOIN     configurationMappings USING (Config_ID)
        JOIN     products USING (Config_ID)
        WHERE `+operator_query+` `+searchfield_query+`
        ORDER BY rank ASC, configurationMappings.id DESC;
        `)
    } else if (Filters.Selected_country != "star") {
        country_code = Filters.Selected_country
        country_query = `Config_ID in (SELECT Config_ID FROM configurationMappings WHERE Country_ID LIKE "%`+country_code+`%") `
        full_query = string(`
        SELECT DISTINCT appConfigs.Config_ID, originalName, modifiableName, iconURL,
        homeUrl, rank, products.productName FROM appConfigs
        JOIN     configurationMappings USING (Config_ID)
        JOIN     products USING (Config_ID)
        WHERE `+country_query+` `+searchfield_query+`
        ORDER BY rank ASC, configurationMappings.id DESC;
        `)
    }
    if(Filters.Selected_country =="star" && Filters.Selected_operator =="star") {
        if(searchfield_query != "") {
            full_query = string(`
            SELECT DISTINCT appConfigs.Config_ID, originalName, modifiableName, iconURL,
            homeUrl, rank, products.productName FROM appConfigs
            JOIN     configurationMappings USING (Config_ID)
            JOIN     products USING (Config_ID)
            WHERE  originalName like "%` + Filters.Searchfield_text +`%"
            ORDER BY rank ASC, configurationMappings.id DESC;
            `)
        } else {
            full_query = string(`
            SELECT DISTINCT appConfigs.Config_ID, originalName, modifiableName, iconURL,
            homeUrl, rank, products.productName FROM appConfigs
            JOIN     configurationMappings USING (Config_ID)
            JOIN     products USING (Config_ID)
            ORDER BY rank ASC, configurationMappings.id DESC;
            `)
        }
    }

    log.Println("loadAppTray –\t\tQuery looks like : " + full_query)
    rows, err := db.Query(full_query)
    checkErr(err)
    //
    //     if(Filters.Selected_country != "star")
    // if(Filters.Selected_country != "star"){
    //     var queryString = `SELECT DISTINCT appConfigs.Config_ID, originalName, modifiableName, iconURL, homeUrl, rank, configurationMappings.productName FROM appConfigs
    //     JOIN     configurationMappings USING (Config_ID)
    //     WHERE Config_ID in (SELECT DISTINCT configurationMappings.Config_ID FROM configurationMappings WHERE
    //     MCCMNC_ID IN (SELECT MCCMNC_ID FROM operators WHERE Country_ID = "`+ Filters.Selected_country +`" )) GROUP BY rank`
    //
    //     rows, err = db.Query(queryString)
    // } else {
    //     rows, err = db.Query(`SELECT DISTINCT appConfigs.Config_ID, originalName, modifiableName, iconURL, homeUrl, rank, configurationMappings.productName FROM appConfigs
    //     JOIN     configurationMappings USING (Config_ID)
    //     WHERE Config_ID in (SELECT DISTINCT configurationMappings.Config_ID FROM configurationMappings WHERE
    //     MCCMNC_ID IN (SELECT MCCMNC_ID FROM operators WHERE Country_ID like "%" )) GROUP BY rank`)
    //     checkErr(err)
    // }

    var appsContainer = AppsContainer{}

    for rows.Next() {
        var app = App{}
        rows.Scan(&app.Config_ID,&app.OriginalName, &app.ModifiableName, &app.IconUrl, &app.HomeUrl, &app.Rank, &app.ProductName)
        appsContainer.Apps = append(appsContainer.Apps, app)
    }
    defer rows.Close()


    jsonResponse, err := json.Marshal(appsContainer.Apps)
    checkErr(err)
    return jsonResponse
}
//ALL URL HANDLERS
func restAppViewHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    tw := new(tabwriter.Writer)
    tw.Init(os.Stderr, 0, 8, 0, '\t', 0)
    log.Println("Rest App View Handler –\tApp Name: ", vars["appName"])

    w.WriteHeader(http.StatusOK)
    w.Header().Add("Content-Type", "text/html")

    var allApps []Webapp = getAllApps()
    for _, app := range allApps {
        if app.ID == vars["appName"] {
  appJSON, err := json.Marshal(app)
  if err != nil {
      fmt.Fprint(w, err)
  } else {
      fmt.Fprint(w, string(appJSON))
      log.Println("Rest App View Handler –\tSuccessfully marshalled JSON")
  }
        }
    }
}

func restAppViewDocumentationHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Header().Add("Content-Type", "text/html")

    fmt.Fprint(w, "Please specify appID such as: \n/ultra/facebook")
}

func restHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    tw := new(tabwriter.Writer)
    tw.Init(os.Stderr, 0, 8, 0, '\t', 0)
    log.Println("Rest Handler –\tCategory: ", vars["category"])

    w.WriteHeader(http.StatusOK)
    w.Header().Add("Content-Type", "text/html")

    if(vars["category"]=="allApps") {
        var allApps []Webapp = getAllApps()
        // var appNames = getAppNames(allApps)
        allAppsJSON, err := json.Marshal(allApps)
        if err != nil {
            fmt.Fprint(w, err)
                } else {
            fmt.Fprint(w, string(allAppsJSON))
            log.Println("Rest Handler –\tSuccessfully marshalled JSON")
        }
    }
}

func restDocumentationHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Header().Add("Content-Type", "text/html")

    fmt.Fprint(w, "If you'd like to access the CMS restAPI, please direct all requests in the following format: \n/rest/AllApps")
}
func configPageHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    log.Println("Config Page Handler –\tConfig_ID: ", vars["Config_ID"])

    // template := template.Must(template.ParseFiles("templates/index.html"))
    // checkErr(err)
    // template.Execute(w, "Hello World!")

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    myHtml := configPageHTML(vars["Config_ID"])
    io.WriteString(w, myHtml)

}
func exportPageHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)

    // template := template.Must(template.ParseFiles("templates/index.html"))
    // checkErr(err)
    // template.Execute(w, "Hello World!")
    username := string("")
    username = getUserName(r)
    if(username!= "") {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        myHtml := exportPageHTML(vars["Config_ID"])
        io.WriteString(w, myHtml)
    } else {
        io.WriteString(w, "Unauthenticated Connection. Please log in.")
    }

}


//HELPER FUNCTIONS

func FileServer(fs http.FileSystem) http.Handler {
	fsh := http.FileServer(fs)
    tw := new(tabwriter.Writer)
    tw.Init(os.Stderr, 0, 8, 0, '\t', 0)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fs.Open(path.Clean(r.URL.Path))

		if os.IsNotExist(err) {
            log.Println("Path doesnt exist")
			return
		}

		fsh.ServeHTTP(w, r)
	})
}

func GetProjectRoot() string {
	root, err := os.Getwd()
	if err != nil {
		panic("Could not retrieve working directory")
	}
	return root
}

func AppendIfMissing(slice []string, i string) []string {
    for _, ele := range slice {
        if ele == i {
            return slice
        }
    }
    return append(slice, i)
}

func getAllApps() []Webapp {
    var webApps []Webapp
    return webApps
}

func getAppNames(slice []Webapp) []string {
    var appNames []string
    for _, webApp := range slice {
        appNames = append(appNames, webApp.Name)
    }
    return appNames
}

func checkErr(err error) {
    tw := new(tabwriter.Writer)
    tw.Init(os.Stderr, 0, 8, 0, '\t', 0)
     if err != nil {
         log.Println("checkErr –\t\t" + "ERROR FOUND")
         panic(err)
     }
 }

 // upload logic
 func UploadFile(w http.ResponseWriter, r *http.Request) {
     file, handler, err := r.FormFile("uploadfile")
     defer file.Close()
      checkErr(err)
      log.Println("UploadFile –\t\tDetected uploaded file: " + handler.Filename)

      byteContainer, err := ioutil.ReadAll(file)
      checkErr(err)
    err = ioutil.WriteFile("static/ultra_apps/" +handler.Filename, byteContainer, 0644)
    checkErr(err)

    var returnResult = ResultMessage{"SUCCESS"}
    jsonResponse, err := json.Marshal(returnResult)
    checkErr(err)
    w.Write([]byte(jsonResponse))
 }

type Webapp struct {
    ID string   `json:"id"`
    Rank         int      `json:"rank"`
    Name         string   `json:"name"`
    HomeURL      string   `json:"homeUrl"`
    DefaultEnabledFeatures []string `json:"defaultEnabledFeatures,omitempty"`
    HiddenUI     []string `json:"hiddenUI,omitempty"`
    HiddenFeatures         []string `json:"hiddenFeatures,omitempty"`
    NativeApps   []string `json:"nativeApps,omitempty"`
    IconURL      string   `json:"iconUrl"`
}
type Webapps struct {
    WebappArray []Webapp `json:"webapps"`
}
