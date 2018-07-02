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
    // "strings"     // fmt.Fprint(w, strings.Join(appNames, ", \n"))
     _ "github.com/mattn/go-sqlite3"

    "encoding/json"
    "io/ioutil"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

const NegroniLogFmt = "{{.StartTime}} | {{.Status}} | {{.Duration}} \n          {{.Method}} {{.Path}}\n"
const NegroniDateFmt = time.Stamp
var cms_db (CMS_DB)

func main() {
    tw := new(tabwriter.Writer)
    tw.Init(os.Stderr, 0, 8, 0, '\t', 0)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

    cms_db = getDB() //THIS IS THE PARSED DATABASE OBJECT


    db = initDB("cms")
    log.Println( "main –\t\tcreating SQLite tables")
    createTables(db)

    statement, _ := db.Prepare(`UPDATE mytable SET MCCMNC_ID = mcc||""||mnc`)
    _, err := statement.Exec()
    checkErr(err)


    log.Println("main –\t\tquerying mytable")
    // rows, err := db.Query("SELECT COALESCE(mcc, '') || COALESCE(mnc, '') FROM mytable") //interesting way of concatting
    rows, err := db.Query("SELECT Operator_Name, Country_ID, MCCMNC_ID FROM mytable")
    checkErr(err)

    var Operator_Name string
    var Country_ID string
    var MCCMNC_ID string
    for rows.Next() {
        rows.Scan(&Operator_Name, &Country_ID, &MCCMNC_ID)
        // log.Println("main –\t\t" + MCCMNC_ID + " | " + Operator_Name + " | " + Country_ID)
    }
    defer rows.Close()


    log.Println("main –\t\tInitializing operators table with temporary MCC table data...")
    statement, _ = db.Prepare(`INSERT or IGNORE  INTO operators (MCCMNC_ID, Operator_Name, Country_ID) SELECT CAST(MCCMNC_ID AS INTEGER), Operator_Name, Country_ID FROM mytable`)
    _, err = statement.Exec()
    checkErr(err)

    log.Println("main –\t\tInitializing countries table with temporary MCC table data...")
    statement, _ = db.Prepare(`INSERT or IGNORE  INTO countries (Country_ID, name, MCC_ID) SELECT Country_ID, country, mcc FROM mytable`)
    _, err = statement.Exec()
    checkErr(err)

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
