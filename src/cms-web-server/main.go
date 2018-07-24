package main

import (
	"log"
	"net/http"
	"path"
	"time"
    "fmt"
    "io"
    "os"
    "text/tabwriter"
        // "database/sql"
    // "strings"     // fmt.Fprint(w, strings.Join(appNames, ", \n"))
     _ "github.com/mattn/go-sqlite3"
    "strconv"
    "encoding/json"
    // "io/ioutil"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

const NegroniLogFmt = "{{.StartTime}} | {{.Status}} | {{.Duration}} \n          {{.Method}} {{.Path}}\n"
const NegroniDateFmt = time.Stamp

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

    mx.HandleFunc("/ultra/{appName}", appViewHandler)
    mx.HandleFunc("/rest/ultra/{appName}", restAppViewHandler)      //handles all restAPI GET requests
    mx.HandleFunc("/rest/ultra/", restAppViewDocumentationHandler)
    mx.HandleFunc("/rest/{category}", restHandler)      //handles all restAPI GET requests
    mx.HandleFunc("/rest/", restDocumentationHandler)   //if someone types in /rest/ with no category
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
    Selected_version string `json:Selected_version, string, omitempty`
    Searchfield_text string `json:Searchfield_text, string, omitempty`
    App_name string `json:App_name, string, omitempty`
    Country_Name string `json:name, string, omitempty`
    Country_ID string `json:Country_ID, string, omitempty`
    Config_ID string `json:"Config_ID"`
    AppModifiableName        string `json:"App_ModifiableName"`
    AppOriginalName          string `json:"App_OriginalName"`
    AppRank                  string `json:"App_Rank"`
    AppHomeURL               string `json:"App_HomeURL"`
    AppNativeURL             string `json:"App_NativeURL"`
    AppIconURL               string `json:"App_IconURL"`
    AppVersionNumber         string `json:"App_VersionNumber"`
    AppExistsEverywhere      bool   `json:"App_ExistsEverywhere"`
    AppConfigurationMappings struct {
        Countries           []string `json:"Countries, omitempty"`
        Operators           []string `json:"Operators, omitempty"`
        FeaturedLocations   []string `json:"FeaturedLocations, omitempty"`
    } `json:"App_ConfigurationMappings"`
}
func postHandler(w http.ResponseWriter, r *http.Request) {
    log.Printf("postHandler –\t\tIncoming post request:")

    requestData := requestData{}

    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
      fmt.Println(err)
    }

    w.Header().Set("Content-Type", "application/json")

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
    } else if (requestData.FunctionToCall=="getAllAppConfigs") {
        log.Println("postHandler –\t\tgetAllAppConfigs method request detected")
        jsonResponse := getAllAppConfigs(requestData.Data)
        w.Write([]byte(jsonResponse))
    } else if (requestData.FunctionToCall=="getFeaturedLocations") {
        log.Println("postHandler –\t\tgetFeaturedLocations method request detected")
        jsonResponse := getFeaturedLocations(requestData.Data)
        w.Write([]byte(jsonResponse))
    }
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
}
type GlobalOperatorRow struct {
    MCCMNC_ID string `json: "MCCMNC_ID"`
    Operator_Name string `json: "Operator_Name"`
    App_Config_ID string `json:"Config_ID" db: "Config_ID"`
    ConfigNumbers []string `json: "configNumbers"`
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
            globalViewQuery := `SELECT Country_ID, name from countries WHERE Country_ID in (SELECT DISTINCT Country_ID from operators WHERE MCCMNC_ID in
            (SELECT MCCMNC_ID from configurationMappings WHERE Config_ID in
            (SELECT Config_ID from AppConfigs WHERE originalName = '`+Data.AppOriginalName+`')))`
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
                AND  MCCMNC_ID in (SELECT MCCMNC_ID from operators WHERE Country_ID = '`+Country_ID+`')`
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
                configList, err = db.Query(globalViewQuery)
                checkErr(err)
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





func addNewConfig(Config data) ([]byte) {
    log.Println(Config.AppConfigurationMappings.Countries)
    log.Println(Config.AppConfigurationMappings.Operators)
    log.Println("addNewConfig –\t\tRecieved request to add " + Config.AppOriginalName)

    statement := string(`INSERT INTO appConfigs (originalName, modifiableName, iconURL, homeURL, rank, versionNumber) VALUES (` + `"` + Config.AppOriginalName + `", "` + Config.AppModifiableName + `", "` + Config.AppIconURL + `", "` + Config.AppHomeURL + `", "` + Config.AppRank + `", "` + Config.AppVersionNumber + `"` + `)`)
    res, err := db.Exec(statement)
    checkErr(err)
    id, err := res.LastInsertId()
    checkErr(err)
    log.Println("addNewConfig –\t\tLast insert id: ", id)
    var New_App_Config_ID = id
    New_App_Config_ID_string := strconv.Itoa(int(New_App_Config_ID))


    for _, country := range Config.AppConfigurationMappings.Countries {
        for _, location := range Config.AppConfigurationMappings.FeaturedLocations {
            mappingstatement := string(`INSERT INTO configurationMappings (Config_ID, MCCMNC_ID, FeaturedLocationName)
            SELECT `+New_App_Config_ID_string+`, MCCMNC_ID, "`+location+`"  FROM operators WHERE Country_ID = '`+country+`';`)
            log.Println("addNewConfig –\t"+mappingstatement)
            res, err = db.Exec(mappingstatement)
            checkErr(err)
        }
    }
    for _, operator := range Config.AppConfigurationMappings.Operators {
        for _, location := range Config.AppConfigurationMappings.FeaturedLocations {
            log.Println("addNewConfig –\t\t" +operator +" | "+ location)
            mappingstatement := string(`INSERT INTO configurationMappings (Config_ID, MCCMNC_ID, FeaturedLocationName) VALUES (`+New_App_Config_ID_string+`, "`+operator+`", "`+location+`")`)
            log.Println("addNewConfig –\t\t"+mappingstatement)
            res, err = db.Exec(mappingstatement)
            checkErr(err)
        }
    }
    if(Config.AppExistsEverywhere) {
        for _, location := range Config.AppConfigurationMappings.FeaturedLocations {
            mappingstatement := string(`INSERT INTO configurationMappings (Config_ID, MCCMNC_ID, FeaturedLocationName)
            SELECT `+New_App_Config_ID_string+`, MCCMNC_ID, "`+location+`"  FROM operators`)
            log.Println("addNewConfig –\t\tApp EXISTS EVERYWHERE "+mappingstatement)
            res, err = db.Exec(mappingstatement)
            checkErr(err)
        }
    }


    var returnResult = ResultMessage{"SUCCESS"}
    jsonResponse, err := json.Marshal(returnResult)
    checkErr(err)
    return jsonResponse
}
type ResultMessage struct {
    Result string `json:"result"`
}

type OperatorRows struct {
    OperatorRows []OperatorRow `json:"operatorRows"`
}
type OperatorRow struct {
    MCCMNC_ID string `json: MCCMNC_ID`
    Operator_Name string `json: Operator_Name`
    Country_ID string `json: Country_ID`
}
func getOperatorsByCountryID(Country data) ([]byte) {
    log.Println("getOperatorsByCountryID –\tRecieved request to get operators in " + Country.Country_ID)

    var operatorRows = OperatorRows{}
    full_query := string(`
    SELECT MCCMNC_ID, Operator_Name, Country_ID from operators
    WHERE Country_ID="`+Country.Country_ID+`"
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
    VersionNumber string `json: "VersionNumber" db:"versionNumber"`
}
func getAllAppConfigs(notUsed data) ([]byte) {
    log.Println("getAllAppConfigs –\tRecieved request to get all app configs")
    var appConfigs = AppConfigs{}

    full_query := string(`
    SELECT DISTINCT Config_ID, originalName, modifiableName,iconURL, homeURL, rank, versionNumber  from appConfigs`)
    allAppConfigs, err := db.Query(full_query)
    checkErr(err)
    for allAppConfigs.Next() {
        var AppConfig = AppConfig{}
        allAppConfigs.Scan(&AppConfig.Config_ID, &AppConfig.OriginalName, &AppConfig.ModifiableName, &AppConfig.IconURL, &AppConfig.HomeURL, &AppConfig.Rank, &AppConfig.VersionNumber)
        appConfigs.AppConfigs = append(appConfigs.AppConfigs, AppConfig)
    }
    allAppConfigs.Close()
    jsonResponse, err := json.Marshal(appConfigs)
    checkErr(err)
    return jsonResponse
}
type CountryRow struct {
    Country_ID string `json:"Country_ID" db:"Country_ID"`
    CountryName string `json:"name" db:"name"`
    MCC_ID string `json:"MCC_ID" db:"MCC_ID"`
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

func getFeaturedLocations(Config data) ([]byte) {
    log.Println("getFeaturedLocations –\tRecieved request to get featured locations for " + Config.Config_ID)

    var featuredLocations = []string{}
    full_query := string(`
    SELECT DISTINCT FeaturedLocationName from configurationMappings
    WHERE Config_ID="`+Config.Config_ID+`"
    `)
    rows, err := db.Query(full_query)
    checkErr(err)
    for rows.Next() {
        featuredLocation := ""
        rows.Scan(&featuredLocation)
        featuredLocations = append(featuredLocations, featuredLocation)
    }
    rows.Close()

    jsonResponse, err := json.Marshal(featuredLocations)
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
type VersionNumberRow struct {
    Value string `json:"versionNumber" db:"versionNumber"`
}
type FilterRows struct{
    CountryFilterRows []CountryFilterRow `json:"countryFilterRows"`
    OperatorFilterRows []OperatorFilterRow `json:"operatorFilterRows"`
    VersionNumberRows []VersionNumberRow `json:"versionNumberRows"`
}
func updateFilterValues(Filters data) ([]byte) {
    log.Println("updateFilterValues –\tRecieved request to update filter items based off existing filter selection.")

    var filterRows = FilterRows{}

    if(Filters.Selected_operator != "star") { //if operator is not star, I don't need to update country dropdown
        full_query := string(`
        SELECT countries.Country_ID, countries.name from operators
        JOIN     countries USING (Country_ID)
        WHERE MCCMNC_ID="`+Filters.Selected_operator+`"
        `)
        rows, err := db.Query(full_query)
        checkErr(err)
        for rows.Next() {
            var countryFilterRow = CountryFilterRow{}
            rows.Scan(&countryFilterRow.Value, &countryFilterRow.Text)
            filterRows.CountryFilterRows = append(filterRows.CountryFilterRows, countryFilterRow)
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

        full_query := string(`SELECT DISTINCT Country_ID, name from countries`) //country query -- all distinct countries by value and name
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

        full_query = string(`SELECT DISTINCT versionNumber from versions`) //version query -- all distinct versions by value
        rows, err = db.Query(full_query)
        checkErr(err)
        for rows.Next() {
            var versionNumberRow = VersionNumberRow{}
            rows.Scan(&versionNumberRow.Value)
            filterRows.VersionNumberRows = append(filterRows.VersionNumberRows, versionNumberRow)
        }
        rows.Close()
    }
    if(Filters.Selected_version != "star") {
        var versionNumberRow = VersionNumberRow{"3.1"}
        filterRows.VersionNumberRows = append(filterRows.VersionNumberRows, versionNumberRow)
    }

    jsonResponse, err := json.Marshal(filterRows)
    checkErr(err)
    return jsonResponse
}

type AppsContainer struct {
    Apps []App
}

type App struct {
    Config_ID string `json:"Config_ID" db:"Config_ID" `
    OriginalName string `json: "originalName" db:"originalName" `
    ModifiableName string `json: "modifiableName" db:"modifiableName" `
    IconUrl string `json: "iconUrl" db:"iconURL" `
    HomeUrl string `json:"homeURL" db:"homeURL"`
    Rank string `json: "rank" db:"rank" `
    FeaturedLocationName string `json:"featuredLocationName" db:"featuredLocationName"`
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
    iconURL, homeURL, rank, configurationMappings.featuredLocationName FROM appConfigs
    JOIN     configurationMappings USING (Config_ID)
    WHERE originalName = "` + Data.App_name +`"
    AND Config_ID in (SELECT Config_ID from configurationMappings
    WHERE MCCMNC_ID IN (SELECT MCCMNC_ID FROM operators WHERE Country_ID like "%`+country_code+`%" ` + operator_query + `))` + ` LIMIT 1`
    rows, err := db.Query(appViewQuery)
    checkErr(err)
    log.Println("appView –\t\t" +  appViewQuery)

    var app = App{}
    for rows.Next() {

        rows.Scan(&app.Config_ID,&app.OriginalName, &app.ModifiableName, &app.IconUrl, &app.HomeUrl, &app.Rank, &app.FeaturedLocationName)
        log.Println("appView –\t\t" + app.Rank + " | " + app.OriginalName + " | " + app.FeaturedLocationName)
    }
    defer rows.Close()

    jsonResponse, err := json.Marshal(app)
    jsonString := string(jsonResponse)
    checkErr(err)
    log.Println("appView –\t\tReturning the following JSON string:")
    log.Println(jsonString)
    return jsonResponse
}

func loadAppTray(Filters data) ([]byte) {
    log.Println("loadAppTray –\t\tquerying db")

    searchfield_query := string("")
    country_code := string("")
    operator_query := string("")
    version_query := string("")


    if(Filters.Selected_operator != "star") { //more specific than country
        operator_query = `AND MCCMNC_ID like "%` + Filters.Selected_operator +`%"`
    } else if (Filters.Selected_country != "star") {
        country_code = Filters.Selected_country
    }
    if(Filters.Searchfield_text != ""){ //search field is NOT empty
        searchfield_query = `AND originalName like "%` + Filters.Searchfield_text +`%"`
    }
    if(Filters.Selected_version != "star") {
        version_query = `AND versionNumber >= ` + Filters.Selected_version +``
    }
    full_query := string(`
    SELECT DISTINCT appConfigs.Config_ID, originalName, modifiableName, iconURL,
    homeUrl, rank, configurationMappings.featuredLocationName FROM appConfigs
    JOIN     configurationMappings USING (Config_ID)
    WHERE Config_ID in (SELECT DISTINCT configurationMappings.Config_ID FROM configurationMappings WHERE
    MCCMNC_ID IN (SELECT MCCMNC_ID FROM operators WHERE Country_ID like "%`+country_code+`%"` + operator_query + `)) `+ searchfield_query + " " + version_query + `
    AND originalName like "%`+Filters.Searchfield_text+`%"`+`
    GROUP BY rank
    `)
    log.Println("loadAppTray –\t\tQuery looks like : " + full_query)
    rows, err := db.Query(full_query)
    checkErr(err)
    //
    //     if(Filters.Selected_country != "star")
    // if(Filters.Selected_country != "star"){
    //     var queryString = `SELECT DISTINCT appConfigs.Config_ID, originalName, modifiableName, iconURL, homeUrl, rank, configurationMappings.featuredLocationName FROM appConfigs
    //     JOIN     configurationMappings USING (Config_ID)
    //     WHERE Config_ID in (SELECT DISTINCT configurationMappings.Config_ID FROM configurationMappings WHERE
    //     MCCMNC_ID IN (SELECT MCCMNC_ID FROM operators WHERE Country_ID = "`+ Filters.Selected_country +`" )) GROUP BY rank`
    //
    //     rows, err = db.Query(queryString)
    // } else {
    //     rows, err = db.Query(`SELECT DISTINCT appConfigs.Config_ID, originalName, modifiableName, iconURL, homeUrl, rank, configurationMappings.featuredLocationName FROM appConfigs
    //     JOIN     configurationMappings USING (Config_ID)
    //     WHERE Config_ID in (SELECT DISTINCT configurationMappings.Config_ID FROM configurationMappings WHERE
    //     MCCMNC_ID IN (SELECT MCCMNC_ID FROM operators WHERE Country_ID like "%" )) GROUP BY rank`)
    //     checkErr(err)
    // }

    var appsContainer = AppsContainer{}

    for rows.Next() {
        var app = App{}
        rows.Scan(&app.Config_ID,&app.OriginalName, &app.ModifiableName, &app.IconUrl, &app.HomeUrl, &app.Rank, &app.FeaturedLocationName)
        log.Println("loadAppTray –\t\t" + app.Rank + " | " + app.OriginalName + " | " + app.FeaturedLocationName)
        appsContainer.Apps = append(appsContainer.Apps, app)
    }
    defer rows.Close()


    jsonResponse, err := json.Marshal(appsContainer.Apps)
    checkErr(err)
    return jsonResponse
}
//ALL URL HANDLERS
func appViewHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    tw := new(tabwriter.Writer)
    tw.Init(os.Stderr, 0, 8, 0, '\t', 0)
    log.Println("App View Handler –\tApp Name: ", vars["appName"])

    w.Header().Set("Content-Type", "text/html; charset=utf-8")

    myHtml := appViewHTML(vars["appName"])
    io.WriteString(w, myHtml)
}

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

func AppendIfMissing(slice []Webapp, app Webapp) []Webapp {
    for _, ele := range slice {
        if ele.ID == app.ID {
  return slice
        }
    }
    return append(slice, app)
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


type Webapp struct {
    ID string   `json:"id"`
    Rank         int      `json:"rank"`
    Name         string   `json:"name"`
    HomeURL      string   `json:"homeUrl"`
    DefaultEnabledFeatures []string `json:"defaultEnabledFeatures"`
    HiddenUI     []string `json:"hiddenUI,omitempty"`
    HiddenFeatures         []string `json:"hiddenFeatures"`
    NativeApps   []string `json:"nativeApps,omitempty"`
    IconURL      string   `json:"iconUrl"`
}
