package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"net/http"
	"os"
	"path"

	"bitbucket.org/atlassianlabs/hipchat-golang-base/util"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("main.go")
var format = logging.MustStringFormatter(
    `%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

// Context keep context of the running application
type Context struct {
	baseURL string
	static  string
}

func (c *Context) healthcheck(w http.ResponseWriter, r *http.Request) {
  work := WorkRequest{gid: 1}
  WorkQueue <- work

	json.NewEncoder(w).Encode([]string{"OK"})
}

func (c *Context) atlassianConnect(w http.ResponseWriter, r *http.Request) {
	lp := path.Join("./static", "atlassian-connect.json")
	vals := map[string]string{
		"LocalBaseUrl": c.baseURL,
	}
	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		log.Fatalf("%v", err)
	}
	tmpl.ExecuteTemplate(w, "config", vals)
}

func (c *Context) installable(w http.ResponseWriter, r *http.Request) {
	payload, err := util.DecodePostJSON(r, true)
	if err != nil {
		log.Errorf("Parsed auth data failed:%v\n", err)
		 w.WriteHeader(http.StatusBadRequest)
		 return
	}

	groupId := int(payload["groupId"].(float64))

	err = AddGroup(groupId, payload["oauthId"].(string), payload["oauthSecret"].(string))
	checkErr(err)
	log.Infof("Added group gid-%d", groupId)

	//util.PrintDump(w, r, false)
	json.NewEncoder(w).Encode([]string{"OK"})
}

// routes all URL routes for app add-on
func (c *Context) routes() *mux.Router {
	r := mux.NewRouter()
	r.Path("/healthcheck").Methods("GET").HandlerFunc(c.healthcheck)

  //descriptor for Atlassian Connect
	r.Path("/").Methods("GET").HandlerFunc(c.atlassianConnect)
	r.Path("/atlassian-connect.json").Methods("GET").HandlerFunc(c.atlassianConnect)

	// HipChat specific API routes
	r.Path("/installable").Methods("POST").HandlerFunc(c.installable)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(c.static)))
	return r
}

func main() {
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	backendLeveled.SetLevel(logging.INFO, "")

	logging.SetBackend(backendLeveled)



	var (
		port    = flag.String("port", "8080", "web server port")
		static  = flag.String("static", "./static/", "static folder")
		baseURL = flag.String("baseurl", os.Getenv("BASE_URL"), "local base url")
    nWorkers = flag.Int("n", 4, "The number of workers to start")
	)
	flag.Parse()

	c := &Context{ baseURL: *baseURL, static:  *static }

	log.Infof("HipChat autoarchiver v0.10 - running on port:%v", *port)

  log.Infof("Starting the cronner")
  StartCron()

  log.Infof("Starting the dispatcher")
  StartDispatcher(*nWorkers)

  r := c.routes()
	http.Handle("/", r)
  log.Infof("Starting the webserver")
	http.ListenAndServe(":"+*port, nil)
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
