package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"

	"bitbucket.org/rbergman/go-hipchat-connect/web"
)

func (s *Server) configurable(w http.ResponseWriter, r *http.Request) {

	tenant, error := web.GetTenant(r)
	s.Log.Debugf("tenant: %v", tenant)

	if error != nil {
		err := fmt.Errorf("Internal Server Error: tenant wasn't in the context")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tenantConfigurations := s.NewTenantConfigurations()
	tenantConfiguration, err := tenantConfigurations.Get(tenant.ID)

	if err != nil {
		err := fmt.Errorf("Couldn't get a configuration for %v: %s", tenant.ID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.buildConfigTemplate(w, tenantConfiguration)
}

func (s *Server) postConfigurable(w http.ResponseWriter, r *http.Request) {
	tenant, error := web.GetTenant(r)
	if error != nil {
		err := fmt.Errorf("Internal Server Error: tenant wasn't in the context")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	strThreshold := r.FormValue("threshold")

	if strThreshold == "" {
		s.Log.Debugf("postConfigurable bad values")
		err := fmt.Errorf("Threshold't value is missing")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	threshold, err := strconv.Atoi(strThreshold)
	if err != nil {
		err := fmt.Errorf("Threshold't value wasn't an integer: %s", strThreshold)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tenantConfigurations := s.NewTenantConfigurations()
	tenantConfiguration := &TenantConfiguration{tenant.ID, threshold}
	err = tenantConfigurations.Set(tenantConfiguration)

	if err != nil {
		s.Log.Errorf("postConfigurable failed to update threshold: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		err := fmt.Errorf("Internal Server Error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.buildConfigTemplate(w, tenantConfiguration)
}

func (s *Server) buildConfigTemplate(w http.ResponseWriter, tenantConfiguration *TenantConfiguration) {
	lp := path.Join("./static", "configurable.hbs")
	vals := map[string]string{
		"Threshold": strconv.Itoa(tenantConfiguration.Threshold),
	}

	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		s.Log.Fatalf("%v", err)
	}

	err = tmpl.Execute(w, vals)
	if err != nil {
		s.Log.Errorf("Error when rendering template config: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return
}
