package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"time"
    "fmt"

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

	server := NewServer()
	server.Run(":" + port)
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
