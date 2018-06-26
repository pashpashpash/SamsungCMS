package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"time"
    "fmt"
    "io"
    // "strings"     // fmt.Fprint(w, strings.Join(appNames, ", \n"))
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "encoding/json"
    "io/ioutil"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

const NegroniLogFmt = "{{.StartTime}} | {{.Status}} | {{.Duration}} \n          {{.Method}} {{.Path}}\n"
const NegroniDateFmt = time.Stamp
var cms_db (CMS_DB)

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

    cms_db = getDB() //THIS IS THE PARSED DATABASE OBJECT
    db, err := sql.Open("mysql", "root:MaxGo99!@tcp(localhost:3306)/")
    if err != nil {
        panic(err)
    }
    defer db.Close()
    create("cms", db, err)

    // =======

    result, err := db.Exec(
    	"INSERT INTO example (id, data) VALUES (1, 'SWAGGGG')",
    )
    log.Println(result)

    var (
    	id int
    	name string
    )
    rows, err := db.Query("select id, data from example where id = ?", 1)
    if err != nil {
    	log.Fatal(err)
    }
    defer rows.Close()
    for rows.Next() {
    	err := rows.Scan(&id, &name)
    	if err != nil {
    		log.Fatal(err)
    	}
    	log.Println(id, name)
    }
    err = rows.Err()
    if err != nil {
    	log.Fatal(err)
    }


    // =======

	server := NewServer()
	server.Run(":" + port)
}

func create(name string, db *sql.DB, err error) {


   _,err = db.Exec("CREATE DATABASE IF NOT EXISTS "+name)
   if err != nil {
       panic(err)
   }

   _,err = db.Exec("USE "+name)
   if err != nil {
       panic(err)
   }

   _,err = db.Exec("CREATE TABLE IF NOT EXISTS example ( id integer, data varchar(32) )")
   if err != nil {
       panic(err)
   }
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
    mx.HandleFunc("/rest/ultra/{appName}", restAppViewHandler)                          //handles all restAPI GET requests
    mx.HandleFunc("/rest/ultra/", restAppViewDocumentationHandler)
    mx.HandleFunc("/rest/{category}", restHandler)                          //handles all restAPI GET requests
    mx.HandleFunc("/rest/", restDocumentationHandler)                       //if someone types in /rest/ with no category
	mx.PathPrefix("/").Handler(FileServer(http.Dir(root + "/static/")))     //for all other urls, serve from /static/

    n.UseHandler(mx)
	return n
}


//ALL URL HANDLERS
func appViewHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    log.Println("App View Handler – App Name: ", vars["appName"])

    w.Header().Set("Content-Type", "text/html; charset=utf-8")

    myHtml := `
           <!DOCTYPE html>
           <html>
           <head>
                 <title>` + vars["appName"] +`</title>
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
    io.WriteString(w, myHtml)
}

func restAppViewHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    log.Println("Rest App View Handler – App Name: ", vars["appName"])

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
                log.Println("Rest App View Handler – Successfully marshalled JSON")
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
    log.Println("Rest Handler – Category: ", vars["category"])

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
            log.Println("Rest Handler – Successfully marshalled JSON")
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
    			ID                     string   `json:"id"`
    			Rank                   int      `json:"rank"`
    			Name                   string   `json:"name"`
    			HomeURL                string   `json:"homeUrl"`
    			DefaultEnabledFeatures []string `json:"defaultEnabledFeatures"`
    			HiddenUI               []string `json:"hiddenUI,omitempty"`
    			HiddenFeatures         []string `json:"hiddenFeatures"`
    			NativeApps             []string `json:"nativeApps,omitempty"`
    			IconURL                string   `json:"iconUrl"`
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
			ID                     string   `json:"id"`
			Rank                   int      `json:"rank"`
			Name                   string   `json:"name"`
			HomeURL                string   `json:"homeUrl"`
			DefaultEnabledFeatures []string `json:"defaultEnabledFeatures"`
			HiddenUI               []string `json:"hiddenUI,omitempty"`
			HiddenFeatures         []string `json:"hiddenFeatures"`
			NativeApps             []string `json:"nativeApps,omitempty"`
			IconURL                string   `json:"iconUrl"`
		} `json:"webapps"`
	} `json:"cms_database"`
}

type Webapp struct {
    ID                     string   `json:"id"`
    Rank                   int      `json:"rank"`
    Name                   string   `json:"name"`
    HomeURL                string   `json:"homeUrl"`
    DefaultEnabledFeatures []string `json:"defaultEnabledFeatures"`
    HiddenUI               []string `json:"hiddenUI,omitempty"`
    HiddenFeatures         []string `json:"hiddenFeatures"`
    NativeApps             []string `json:"nativeApps,omitempty"`
    IconURL                string   `json:"iconUrl"`
}
