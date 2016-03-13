package main

import (
  "fmt"
  "path"
  "strconv"
  "html/template"
  "net/http"
  "bitbucket.org/rbergman/go-hipchat-connect/web"
)

func (s *Server) configurable(w http.ResponseWriter, r *http.Request) {

	tenant := web.GetTenant(r)
  s.Log.Debugf("tenant: %v", tenant)

  if tenant.ID == "" {
    err := fmt.Errorf("Internal Server Error: tenant wasn't in the context")
    http.Error(w, err.Error(), http.StatusInternalServerError)
		return
  }

  tenantConfigurations := s.NewTenantConfigurations()
  tenantConfiguration, err := tenantConfigurations.Get(tenant.ID)

  if err != nil {
    s.Log.Debugf("Couldn't get a configuration for %v", tenant)
    w.WriteHeader(http.StatusBadRequest)
  }

  s.buildConfigTemplate(w, tenantConfiguration)
}

func (s *Server) postConfigurable(w http.ResponseWriter, r *http.Request) {
	tenant := web.GetTenant(r)
	strThreshold := r.FormValue("threshold")

	if strThreshold == "" {
		s.Log.Debugf("postConfigurable bad values")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	threshold, err := strconv.Atoi(strThreshold)
	if err != nil {
		s.Log.Debugf("postConfigurable threshold wasn't an integer")
		w.WriteHeader(http.StatusBadRequest)
	}

	tenantConfigurations := s.NewTenantConfigurations()
  tenantConfiguration := &TenantConfiguration{tenant.ID, threshold}
  err = tenantConfigurations.Set(tenantConfiguration)

	if err != nil {
		s.Log.Errorf("postConfigurable failed to update threshold: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
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

	tmpl.ExecuteTemplate(w, "config", vals)
}
