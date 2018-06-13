package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"time"
    "fmt"

    "encoding/json"
    "io/ioutil"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

const NegroniLogFmt = "{{.StartTime}} | {{.Status}} | {{.Duration}} \n          {{.Method}} {{.Path}}\n"
const NegroniDateFmt = time.Stamp



func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}


    cms_db := getDB()
    fmt.Println("JSON->Object Results:", cms_db)

	server := NewServer()
	server.Run(":" + port)
}


func getDB() []CMS_DB {
    raw, err := ioutil.ReadFile("./ultra_apps_db.json") //it does actually find this file
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }
    log.Println("Successfully found JSON")

    var c []CMS_DB
    json.Unmarshal(raw, &c)

    log.Println("Made CMS_DB object out of provided JSON")

    return c
}

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

func (cms_db CMS_DB) String() string {
    return cms_db.CmsDatabase[0].ConfigName
}


func GetProjectRoot() string {
	root, err := os.Getwd()
	if err != nil {
		panic("Could not retrieve working directory")
	}
	return root
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
    // mx.PathPrefix("/").Handler(HelloHandler())
	mx.PathPrefix("/").Handler(FileServer(http.Dir(root + "/static/")))
	n.UseHandler(mx)

	return n
}
func HelloHandler() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
        fmt.Fprintf(w, "Unfortunately, this site doesn't exist :(")
    })
}

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
