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


    db = initDB("cms_2")
    log.Println( "main –\t\tcreating a SQLite tables")
    createTables(db)

    log.Println( "main –\t\tInserting into countries table...")
    statement, _ := db.Prepare("INSERT INTO countries (Country_ID, Name, MCC_ID) VALUES (?, ?, ?)")
    statement.Exec("GE-AB", "Abkhazia", 289)
    statement.Exec("AF", "Afghanistan", 412)
    statement.Exec("AL", "Albania", 276)

    log.Println( "main –\t\tInserting into operators table...")
    statement, _ = db.Prepare("INSERT INTO operators ( MCCMNC_ID, Operator_Name, Country_ID) VALUES (?, ?, ?)")
    statement.Exec(28967, "Aquafon JSC", "GE-AB")
    statement.Exec(41201, "Afghan Wireless Communication Company", "AF")
    statement.Exec(27601, "Telekom Albania", "AL")
    //
    // statement, _ = db.Prepare("INSERT INTO versionTable (app, version) VALUES (?, ?)")
    // statement.Exec("Twitter", "1.0")
    // statement.Exec("Cricbuzz", "1.0")



    log.Println("main –\t\tquerying operators ( MCCMNC_ID, Operator_Name, Country_ID)")
    rows, err := db.Query("SELECT MCCMNC_ID, Operator_Name, Country_ID FROM operators")
    checkErr(err)

    var MCCMNC_ID string
    var Operator_Name string
    var Country_ID string
    for rows.Next() {
        rows.Scan(&MCCMNC_ID, &Operator_Name, &Country_ID)
        log.Println("main –\t\t" + MCCMNC_ID + "\t| " + Operator_Name + " | " + Country_ID)
    }
    defer rows.Close()

	server := NewServer()
	server.Run(":" + port)
}


// DATABASE HELPER FUNCTION
func initDB(name string) (*sql.DB) {
    tw := new(tabwriter.Writer)
    tw.Init(os.Stderr, 0, 8, 0, '\t', 0)
    log.Println("initDB –\t\tInitializing SQLite db with the name " + name)
    db, err := sql.Open("sqlite3", "./"+name+".db")
    checkErr(err)
    return db
}

func createTables(db *sql.DB) {
    log.Println( "createTables –\tCreating countries table...")
    stmt, _ := db.Prepare("CREATE TABLE IF NOT EXISTS countries ( Country_ID TEXT PRIMARY KEY, name TEXT NOT NULL, MCC_ID INTEGER NOT NULL)")
    _, err := stmt.Exec()
    checkErr(err)

    log.Println( "createTables –\tCreating operators table...")
    stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS operators ( MCCMNC_ID integer PRIMARY KEY, Operator_Name TEXT, Country_ID TEXT, FOREIGN KEY(Country_ID) REFERENCES countries(Country_ID) )")
    _, err = stmt.Exec()
    checkErr(err)

    log.Println( "createTables –\tCreating versions table...")
    stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS versions (versionNumber FLOAT PRIMARY KEY)")
    _, err = stmt.Exec()
    checkErr(err)

    log.Println( "createTables –\tCreating featuredLocations table...")
    stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS featuredLocations ( featuredLocationName TEXT PRIMARY KEY )")
    _, err = stmt.Exec()
    checkErr(err)


    log.Println( "createTables –\tCreating appConfigs table...")
    stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS appConfigs ( Config_ID INTEGER PRIMARY KEY AUTOINCREMENT, originalName TEXT, modifiableName TEXT, iconURL TEXT, homeURL TEXT, rank INTEGER, versionNumber FLOAT, FOREIGN KEY(versionNumber) REFERENCES versions(versionNumber))")
    _, err = stmt.Exec()
    checkErr(err)

    log.Println( "createTables –\tCreating configurationMappings table...")
    stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS configurationMappings ( id INTEGER PRIMARY KEY AUTOINCREMENT, Config_ID INTEGER, MCCMNC_ID integer, featuredLocationName TEXT,  FOREIGN KEY(Config_ID) REFERENCES appConfigs(Config_ID), FOREIGN KEY(MCCMNC_ID) REFERENCES operators(MCCMNC_ID), FOREIGN KEY(featuredLocationName) REFERENCES featuredLocations(featuredLocationName))")
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
             Country
         </div>
         <select name="country"  onchange="selectChange(this); this.oldvalue = this.value;">
             <option value="star">🔯</option>
           <option value="af">Afghanistan</option>

           <option value="ge">Germany</option>
           <option value="uk">United Kingdom</option>
         </select>
     </div>
     <div id = "filterContainer">
         <div id="filterText">
             Operator
         </div>
         <select name="operator" onchange="selectChange(this); this.oldvalue = this.value;">
               <option value="star">🔯</option>
           <option value="tmobile(333444)">T-Mobile (333444)</option>

           <option value="T-Mobile(333445)">T-Mobile (333445)</option>
           <option value="AT&T(456932)">AT&T (456932)</option>
         </select>
     </div>
     <div id = "filterContainer">
         <div id="filterText">
             Version
         </div>
         <select name="version"  onchange="selectChange(this); this.oldvalue = this.value;">
             <option value="star">🔯</option>
             <option value="2.4">2.4</option>
           <option value="2.5">2.5</option>
           <option value="2.6">2.6</option>
         </select>
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
