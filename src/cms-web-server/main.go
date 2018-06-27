package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"time"
    "fmt"
    "io"
    "text/tabwriter"
    // "strings"     // fmt.Fprint(w, strings.Join(appNames, ", \n"))
    "database/sql"
     _ "github.com/mattn/go-sqlite3"

    "encoding/json"
    "io/ioutil"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

const NegroniLogFmt = "{{.StartTime}} | {{.Status}} | {{.Duration}} \n          {{.Method}} {{.Path}}\n"
const NegroniDateFmt = time.Stamp
var cms_db (CMS_DB)
var db (*sql.DB)


func main() {

    tw := new(tabwriter.Writer)
    tw.Init(os.Stderr, 0, 8, 0, '\t', 0)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

    cms_db = getDB() //THIS IS THE PARSED DATABASE OBJECT


    db = initDB("cms")
    log.Println( "main –\tcreating a SQLite table named platformTable")
    createTables(db)

    statement, _ := db.Prepare("INSERT INTO countryTable (app, country) VALUES (?, ?)")
    statement.Exec("Twitter", "IN")
    statement.Exec("Cricbuzz", "IN")
    statement.Exec("Freebasics", "PL")

    statement, _ = db.Prepare("INSERT INTO platformTable (app, platform) VALUES (?, ?)")
    statement.Exec("Twitter", "Samsung J2")
    statement.Exec("Cricbuzz", "Samsung J1")

    statement, _ = db.Prepare("INSERT INTO operatorTable (app, operator) VALUES (?, ?)")
    statement.Exec("Twitter", "T-Mobile")
    statement.Exec("Cricbuzz", "Verizon")

    statement, _ = db.Prepare("INSERT INTO versionTable (app, version) VALUES (?, ?)")
    statement.Exec("Twitter", "1.0")
    statement.Exec("Cricbuzz", "1.0")



    log.Println("main –\tquerying platformTable (app, platform)")
    rows, err := db.Query("SELECT app, platform FROM platformTable")
    checkErr(err)

    var app string
    var platform string
    for rows.Next() {
        rows.Scan(&app, &platform)
        log.Println("main –\t" + app + " | " + platform)
    }
    defer rows.Close()

	server := NewServer()
	server.Run(":" + port)
}


// DATABASE HELPER FUNCTION
func initDB(name string) (*sql.DB) {
    tw := new(tabwriter.Writer)
    tw.Init(os.Stderr, 0, 8, 0, '\t', 0)
    log.Println("initDB –\tInitializing SQLite db with the name " + name)
    db, err := sql.Open("sqlite3", "./"+name+".db")
    checkErr(err)
    return db
}

func createTables(db *sql.DB) {
    stmt, _ := db.Prepare("CREATE TABLE IF NOT EXISTS platformTable ( app TEXT PRIMARY KEY, platform TEXT )")
    _, err := stmt.Exec()
    checkErr(err)

    stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS operatorTable ( app TEXT PRIMARY KEY, operator TEXT )")
    _, err = stmt.Exec()
    checkErr(err)

    stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS countryTable ( app TEXT PRIMARY KEY, country TEXT )")
    _, err = stmt.Exec()
    checkErr(err)

    stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS versionTable ( app TEXT PRIMARY KEY, version TEXT )")
    _, err = stmt.Exec()
    checkErr(err)
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
	mx.PathPrefix("/").Handler(FileServer(http.Dir(root + "/static/")))     //for all other urls, serve from /static/

    n.UseHandler(mx)
	return n
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

func getDB() CMS_DB { //gets JSON from hard-coded filepath & parses it into an OBJECT
    tw := new(tabwriter.Writer)
    tw.Init(os.Stderr, 0, 8, 0, '\t', 0)
    raw, err := ioutil.ReadFile("./static/cms-database.json") //JSON CONFIG FILE
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }
    log.Println("Successfully found JSON")

    c := struct {
        CmsDatabase []struct {
    		ConfigName string   `json:"config_name"`
    		Order      int      `json:"order"`
    		Inherit    []string `json:"inherit,omitempty"`
    		Filter     []struct {
    			Product  []string `json:"product,omitempty"`
    			Operator []string `json:"operator,omitempty"`
    		} `json:"filter,omitempty"`
    		Webapps []struct {
    			ID string   `json:"id"`
    			Rank         int      `json:"rank"`
    			Name         string   `json:"name"`
    			HomeURL      string   `json:"homeUrl"`
    			DefaultEnabledFeatures []string `json:"defaultEnabledFeatures"`
    			HiddenUI     []string `json:"hiddenUI,omitempty"`
    			HiddenFeatures         []string `json:"hiddenFeatures"`
    			NativeApps   []string `json:"nativeApps,omitempty"`
    			IconURL      string   `json:"iconUrl"`
    		} `json:"webapps"`
    	} `json:"cms_database"`
    }{}

    json.Unmarshal(raw, &c)

    log.Println("Generated CMS_DB")
    // log.Println(c.CmsDatabase[0].Webapps[0].Name)

    return CMS_DB(c)
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
    for _, CmsDatabase := range cms_db.CmsDatabase {
        for _, webApp := range CmsDatabase.Webapps {
  webApps = AppendIfMissing(webApps, Webapp(webApp))
        }
    }
    // var webApps = []Webapp {
    //     Webapp {
    //         ID:"facebook",
    //         Rank: 2,
    //         Name:"swag",
    //         HomeURL:"swag",
    //         DefaultEnabledFeatures: []string{"Penn", "Teller"},
    //         HiddenUI: []string{"Penn", "Teller"},
    //         HiddenFeatures: []string{"Penn", "Teller"},
    //         NativeApps: []string{"Penn", "Teller"},
    //         IconURL:"swag",
    //     },
    //     Webapp {
    //         ID:"facebook",
    //         Rank: 2,
    //         Name:"swag",
    //         HomeURL:"swag",
    //         DefaultEnabledFeatures: []string{"Penn", "Teller"},
    //         HiddenUI: []string{"Penn", "Teller"},
    //         HiddenFeatures: []string{"Penn", "Teller"},
    //         NativeApps: []string{"Penn", "Teller"},
    //         IconURL:"swag",
    //     },
    // }
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
         log.Println(err)
         panic(err)
     }
 }


// TYPE DECLARATIONS

type CMS_DB struct {
	CmsDatabase []struct {
		ConfigName string   `json:"config_name"`
		Order      int      `json:"order"`
		Inherit    []string `json:"inherit,omitempty"`
		Filter     []struct {
			Product  []string `json:"product,omitempty"`
			Operator []string `json:"operator,omitempty"`
		} `json:"filter,omitempty"`
		Webapps []struct {
			ID string   `json:"id"`
			Rank         int      `json:"rank"`
			Name         string   `json:"name"`
			HomeURL      string   `json:"homeUrl"`
			DefaultEnabledFeatures []string `json:"defaultEnabledFeatures"`
			HiddenUI     []string `json:"hiddenUI,omitempty"`
			HiddenFeatures         []string `json:"hiddenFeatures"`
			NativeApps   []string `json:"nativeApps,omitempty"`
			IconURL      string   `json:"iconUrl"`
		} `json:"webapps"`
	} `json:"cms_database"`
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
func appViewHTML(appName string) (string){
    return `
 <!DOCTYPE html>
 <html>
 <head>
       <title>` + appName +`</title>
       <link rel="stylesheet" type="text/css" href="../stylesheets/main.css">
 </head>
 <body>
     <div id = "header">
         <div id="headerIcon" onClick="location.reload();location.href='../index.html'"></div><div id="headerText" onClick="location.reload();location.href='../index.html'"> Ultra Apps <span id="smallerText">CMS</span></div>
     </div>
     <div id = "filters">
         <div id = "filterContainer">
   <div id="filterText">
       Platform
   </div>
   <select name="platform" onchange="selectChange(this); this.oldvalue = this.value;">
        <option value="star">🔯</option>
     <option value="samsungJ2">Samsung J2</option>
     <option value="mercedes">Mercedes</option>
     <option value="audi">Audi</option>
   </select>
         </div>
         <div id = "filterContainer">
   <div id="filterText">
       Operator
   </div>
   <select name="operator" onchange="selectChange(this); this.oldvalue = this.value;">
         <option value="star">🔯</option>
     <option value="tmobile">T-Mobile</option>

     <option value="mercedes">Mercedes</option>
     <option value="audi">Audi</option>
   </select>
         </div>
         <div id = "filterContainer">
   <div id="filterText">
       Country
   </div>
   <select name="country"  onchange="selectChange(this); this.oldvalue = this.value;">
       <option value="star">🔯</option>
     <option value="afghanistan">Afghanistan</option>

     <option value="mercedes">Mercedes</option>
     <option value="audi">Audi</option>
   </select>
         </div>
         <div id = "filterContainer">
   <div id="filterText">
       Version
   </div>
   <select name="version"  onchange="selectChange(this); this.oldvalue = this.value;">
       <option value="star">🔯</option>
       <option value="2.4">2.4</option>
     <option value="mercedes">Mercedes</option>
     <option value="audi">Audi</option>
   </select>
         </div>
         <div id = "filterContainer" style="margin-top:12px;">
   <div id="checkboxFilterText">
       Featured Location
   </div>
   <div class="checkboxContainer" id="maxCheck">
       Max
   </div>
   <div class="checkboxContainer" id="folderCheck">
       Folder
   </div>
   <div class="checkboxContainer" id="homescreenCheck">
       Homescreen
   </div>
         </div>
         <div id = "starandsearch">
   <div class="clicked" id="star">
   </div>
    <input class="search" type="text" placeholder="Search..">
         </div>
     </div>
     <div class ="webApp">
     </div>
     </main>
     <script type="text/javascript" src="../javascript/filters.js"></script>
     <script type="text/javascript" src="../javascript/rest.js"></script>
     <script type="text/javascript" src="../javascript/appView.js"></script>
 </body>
 `
}
