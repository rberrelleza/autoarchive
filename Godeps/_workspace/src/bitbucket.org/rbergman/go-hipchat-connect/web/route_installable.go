package web

import (
	"net/http"

	"bitbucket.org/rbergman/go-hipchat-connect/model"
	"bitbucket.org/rbergman/go-hipchat-connect/rest"
	"bitbucket.org/rbergman/go-hipchat-connect/tenant"

	"github.com/go-zoo/bone"
)

// HandleInstall validates and registers request tenant installations.
func (s *Server) HandleInstall(w http.ResponseWriter, req *http.Request) {
	s.Log.Debugf("starting install")
	if status, err := s.VerifyJSONRequest(req); err != nil {
		s.Log.Debugf("error: %s", err.Error())
		http.Error(w, err.Error(), status)
		return
	}

	installable, err := model.DecodeInstallable(req.Body)
	if err != nil {
		s.Log.Debugf("error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.Log.Debugf("installable: %+v", installable)

	descriptor, err := rest.GetDescriptor(installable.CapabilitiesURL)
	if err != nil {
		s.Log.Debugf("error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.Log.Debugf("descriptor: %+v", descriptor)

	t, err := tenant.New(installable, descriptor)
	if err != nil {
		s.Log.Debugf("error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	s.Log.Debugf("tenant: %+v", t)

	client := rest.NewClient(t.Links.API, nil)
	user := t.ID
	pass := t.Secret
	token, err := client.GenerateToken(user, pass)
	if err != nil {
		s.Log.Debugf("error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	s.Log.Debugf("token: %+v", token)

	t.GroupName = token.GroupName
	s.Log.Debugf("tenant: %+v", t)

	tenants := s.NewTenants()
	err = tenants.Set(t)
	if err != nil {
		s.Log.Debugf("error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.Log.Debugf("installed: %s", t.ID)
}

func (s *Server) HandleUninstall(w http.ResponseWriter, req *http.Request) {
	tenantID := bone.GetValue(req, "tenantID")
	s.Log.Debugf("starting uninstall: %s", tenantID)

	tenants := s.NewTenants()
	err := tenants.Del(tenantID)
	if err != nil {
		s.Log.Debugf("error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.Log.Debugf("uninstalled: %s", tenantID)
}
