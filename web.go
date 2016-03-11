package main

import (
	"bitbucket.org/atlassianlabs/hipchat-golang-base/util"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func (c *Context) RunWeb() {
	port := "8080"
	log.Infof("HipChat autoarchiver web - running on port:%v", port)
	r := c.routes()
	http.Handle("/", r)
	log.Infof("Starting the webserver")
	http.ListenAndServe(":"+port, nil)
}

func (context *Context) authenticate(r *http.Request) (*jwt.Token, error) {
	// var signedRequestParam = this.query.signed_request;
	log.Debugf("authenticate init")
	authorizationHeader := r.Header.Get("authorization")
	signedRequestParam := r.URL.Query().Get("signed_request")

	requestToken := ""
	if authorizationHeader != "" {
		requestToken = authorizationHeader[len("JWT "):]
	} else if signedRequestParam != "" {
		requestToken = signedRequestParam
	} else {
		return nil, fmt.Errorf("Request is missing an authorization header")
	}

	verifiedToken, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Debugf("invalid token: %s", token)
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		issuer, ok := token.Claims["iss"].(string)
		if (!ok) {
			return nil, fmt.Errorf("JWT claim did not contain the issuer (iss) claim")
		}
		group, err := GetGroupByOauthId(context, issuer)

		if (err != nil){
			log.Debugf("Couldn't find group with oauthId-%s", issuer)
			return nil, fmt.Errorf("Request can't be verified without a valid OAuth secret")
		}

		if (token.Header["alg"] != "HS256") {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(group.oauthSecret), nil
	})

	if err == nil && verifiedToken.Valid {
		return verifiedToken, err
	} else {
		return nil, err
	}
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
	checkErr(err)
	log.Infof("Added group gid-%d", groupId)

	//util.PrintDump(w, r, false)
	json.NewEncoder(w).Encode([]string{"OK"})
}

func (c *Context) configurable(w http.ResponseWriter, r *http.Request) {
	log.Debugf("configurable init")
	token, auth_error := c.authenticate(r)
	if auth_error != nil {
		log.Errorf("configurable authentication error %s", auth_error)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

  oauthId, _ := token.Claims["iss"].(string);
	group, err := GetGroupByOauthId(c, oauthId)
	if (err != nil) {
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
	log.Debugf("updateConfigurable init")
	token, auth_error := c.authenticate(r)
	if auth_error != nil {
		log.Debugf("postConfigurable authentication error %s", auth_error)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	oauthId, _ := token.Claims["iss"].(string)
	strThreshold := r.FormValue("threshold")

	if (strThreshold == "") {
		log.Debugf("postConfigurable bad values")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	threshold, err := strconv.Atoi(strThreshold)
	if (err != nil) {
		log.Debugf("postConfigurable threshold wasn't an integer")
		w.WriteHeader(http.StatusBadRequest)
	}

	group, err := UpdateThreshold(c, oauthId, threshold)
	if (err != nil){
		log.Errorf("postConfigurable failed to update threshold: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lp := path.Join("./static", "configurable.hbs")
	vals := map[string]string{
		"LocalBaseUrl": c.baseURL,
		"Threshold": strconv.Itoa(group.threshold),
	}

	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		log.Fatalf("%v", err)
	}
	tmpl.ExecuteTemplate(w, "config", vals)
}

func (c *Context) removeInstallable(w http.ResponseWriter, r *http.Request) {
	log.Debugf("removeInstallable init")
	_, auth_error := c.authenticate(r)
	if auth_error != nil {
		log.Debugf("removeInstallable authentication error %s", auth_error)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

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

// routes all URL routes for app add-on
func (c *Context) routes() *mux.Router {
	r := mux.NewRouter()
	r.Path("/healthcheck").Methods("GET").HandlerFunc(c.healthcheck)

	//descriptor for Atlassian Connect
	r.Path("/").Methods("GET").HandlerFunc(c.atlassianConnect)
	r.Path("/atlassian-connect.json").Methods("GET").HandlerFunc(c.atlassianConnect)

	// HipChat specific API routes
	r.Path("/installable").Methods("POST").HandlerFunc(c.installable)
	r.Path("/installable/{oauthId}").Methods("DELETE").HandlerFunc(c.removeInstallable)
	r.Path("/configurable").Methods("GET").HandlerFunc(c.configurable)
	r.Path("/configurable").Methods("POST").HandlerFunc(c.postConfigurable)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	return r
}
