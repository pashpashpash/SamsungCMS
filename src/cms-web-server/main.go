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
}
func postHandler(w http.ResponseWriter, r *http.Request) {
    log.Printf("postHandler –\t\tIncoming post request:")

    requestData := requestData{}

    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
      fmt.Println(err)
    }


    w.Header().Set("Content-Type", "application/json")

    if (requestData.FunctionToCall=="loadAppTray") {
        log.Println("postHandler –\tAll apps method request detected – Data: ")
        log.Println(requestData.Data)
        jsonResponse := loadAppTray(requestData.Data)
        w.Write([]byte(jsonResponse))
    } else if (requestData.FunctionToCall=="appView") {
        log.Println("postHandler –\tAll view method request detected")
        jsonResponse := appView(requestData.Data)
        w.Write([]byte(jsonResponse))
    } else if (requestData.FunctionToCall=="updateFilterValues") {
        log.Println("postHandler –\tupdateFilterValues method request detected")
        jsonResponse := updateFilterValues(requestData.Data)
        w.Write([]byte(jsonResponse))
    }
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

    if(Filters.Selected_operator != "star") { //if operator is not star, I don't need to update country table
        full_query := string(`
        SELECT countries.Country_ID, countries.name from operators
        JOIN     countries USING (Country_ID)
        WHERE MCCMNC_ID="`+Filters.Selected_operator+`"
        `)
        rows, err := db.Query(full_query)
        checkErr(err)
        //country query here, returns rows
        for rows.Next() {
            var countryFilterRow = CountryFilterRow{}
            rows.Scan(&countryFilterRow.Value, &countryFilterRow.Text)
            log.Println("updateFilterValues –\t" + countryFilterRow.Value + " | " + countryFilterRow.Text)
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
            log.Println("updateFilterValues –\t" + operatorFilterRow.Value + " | " + operatorFilterRow.Text)
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
            log.Println("updateFilterValues –\t" + countryFilterRow.Value + " | " + countryFilterRow.Text)
            filterRows.CountryFilterRows = append(filterRows.CountryFilterRows, countryFilterRow)
        }
        rows.Close()

        full_query = string(`SELECT DISTINCT MCCMNC_ID, Operator_Name from operators`) //operator query -- all distinct operators
        rows, err = db.Query(full_query)
        checkErr(err)
        for rows.Next() {
            var operatorFilterRow = OperatorFilterRow{}
            rows.Scan(&operatorFilterRow.Value, &operatorFilterRow.Text)
            log.Println("updateFilterValues –\t" + operatorFilterRow.Value + " | " + operatorFilterRow.Text)
            filterRows.OperatorFilterRows = append(filterRows.OperatorFilterRows, operatorFilterRow)
        }
        rows.Close()

        full_query = string(`SELECT DISTINCT versionNumber from versions`) //version query -- all distinct versions by value
        rows, err = db.Query(full_query)
        checkErr(err)
        for rows.Next() {
            var versionNumberRow = VersionNumberRow{}
            rows.Scan(&versionNumberRow.Value)
            log.Println("updateFilterValues –\t" + versionNumberRow.Value)
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
    log.Println("appView –\t\tquerying db")

    rows, err := db.Query(`SELECT DISTINCT appConfigs.Config_ID, originalName, modifiableName, iconURL, homeURL, rank, configurationMappings.featuredLocationName FROM appConfigs
    JOIN     configurationMappings USING (Config_ID)
    WHERE originalName = "facebook" LIMIT 1`)
    checkErr(err)

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
