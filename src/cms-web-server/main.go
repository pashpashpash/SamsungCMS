package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"time"
    "fmt"
    // "strings"     // fmt.Fprint(w, strings.Join(appNames, ", \n"))

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

    mx.HandleFunc("/rest/{category}", restHandler)                          //handles all restAPI GET requests
    mx.HandleFunc("/rest/", restDocumentationHandler)                       //if someone types in /rest/ with no category
	mx.PathPrefix("/").Handler(FileServer(http.Dir(root + "/static/")))     //for all other urls, serve from /static/

    n.UseHandler(mx)
	return n
}


//ALL URL HANDLERS

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
