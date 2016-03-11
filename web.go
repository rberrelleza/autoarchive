package main

import (
	"bitbucket.org/atlassianlabs/hipchat-golang-base/util"
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"path"
	"strconv"
)

func (c *Context) RunWeb() {
	port := "8080"
	log.Infof("HipChat autoarchiver web - running on port:%v", port)
	n := negroni.Classic()

	r := mux.NewRouter()
	r.HandleFunc("/", c.atlassianConnect).Methods("GET")
	r.HandleFunc("/atlassian-connect.json", c.atlassianConnect).Methods("GET")
	r.HandleFunc("/healthcheck", c.healthcheck).Methods("GET")
	r.HandleFunc("/installable", c.installable).Methods("POST")

	r.Handle("/configurable", negroni.New(
		negroni.HandlerFunc(c.AuthenticateMiddleware),
		negroni.Wrap(http.HandlerFunc(c.configurable)),
	)).Methods("GET")

	r.Handle("/installable/{oauthId}", negroni.New(
		negroni.HandlerFunc(c.AuthenticateMiddleware),
		negroni.Wrap(http.HandlerFunc(c.removeInstallable)),
	)).Methods("DELETE")

	r.Handle("/configurable", negroni.New(
		negroni.HandlerFunc(c.AuthenticateMiddleware),
		negroni.Wrap(http.HandlerFunc(c.postConfigurable)),
	)).Methods("POST")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	n.UseHandler(r)
	log.Fatal(http.ListenAndServe(":8080", n))
}

func (c *Context) healthcheck(w http.ResponseWriter, r *http.Request) {
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
	tmpl.ExecuteTemplate(w, "connect", vals)
}

func (c *Context) installable(w http.ResponseWriter, r *http.Request) {
	log.Debug("init installable")

	payload, err := util.DecodePostJSON(r, true)
	if err != nil {
		log.Errorf("Parsed auth data failed:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	groupId := int(payload["groupId"].(float64))

	err = AddGroup(c, groupId, payload["oauthId"].(string), payload["oauthSecret"].(string), 90)
	if err != nil {
		log.Error("Failed to add gid-%d: %s", groupId, err)
	}

	log.Infof("Added group gid-%d", groupId)

	json.NewEncoder(w).Encode([]string{"OK"})
}

func (c *Context) configurable(w http.ResponseWriter, r *http.Request) {
	log.Debugf("configurable init")
	oauthId, _ := c.token.Claims["iss"].(string)
	group, err := GetGroupByOauthId(c, oauthId)
	if err != nil {
		log.Errorf("Couldn't find group with oauthId %s", oauthId)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	lp := path.Join("./static", "configurable.hbs")
	vals := map[string]string{
		"Threshold": strconv.Itoa(group.threshold),
	}

	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		log.Fatalf("%v", err)
	}
	tmpl.ExecuteTemplate(w, "config", vals)
}

func (c *Context) postConfigurable(w http.ResponseWriter, r *http.Request) {
	oauthId, _ := c.token.Claims["iss"].(string)
	strThreshold := r.FormValue("threshold")

	if strThreshold == "" {
		log.Debugf("postConfigurable bad values")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	threshold, err := strconv.Atoi(strThreshold)
	if err != nil {
		log.Debugf("postConfigurable threshold wasn't an integer")
		w.WriteHeader(http.StatusBadRequest)
	}

	group, err := UpdateThreshold(c, oauthId, threshold)
	if err != nil {
		log.Errorf("postConfigurable failed to update threshold: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lp := path.Join("./static", "configurable.hbs")
	vals := map[string]string{
		"LocalBaseUrl": c.baseURL,
		"Threshold":    strconv.Itoa(group.threshold),
	}

	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		log.Fatalf("%v", err)
	}
	tmpl.ExecuteTemplate(w, "config", vals)
}

func (c *Context) removeInstallable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	oauthId := vars["oauthId"]
	log.Infof("Removing addon for oauthId %s", oauthId)

	deleted, err := DeleteGroup(c, oauthId)
	if err != nil {
		log.Errorf("Failed to remove addon :%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Debugf("Successfully deleted gid-%d", deleted.groupId)
		json.NewEncoder(w).Encode([]string{"OK"})
	}
}
