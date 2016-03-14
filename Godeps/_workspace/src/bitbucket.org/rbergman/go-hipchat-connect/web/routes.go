package web

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"net/http"
)

func (s *Server) addBaseRoutes() {
	s.Router.GetFunc("/descriptor", s.HandleDescriptor)
	s.Router.GetFunc("/healthcheck", s.HandleHealthcheck)
	s.Router.PostFunc("/installable", s.HandleInstall)
	s.Router.DeleteFunc("/installable/:tenantID", s.HandleUninstall)
}

func (s *Server) AddGetConfigurable(f http.HandlerFunc) {
	n := negroni.New(
		NewAuthenticate(s),
		negroni.Wrap(context.ClearHandler(f)))
	s.Router.Get("/configurable", n)
}

func (s *Server) AddPostConfigurable(f http.HandlerFunc) {
	n := negroni.New(
		NewAuthenticate(s),
		negroni.Wrap(context.ClearHandler(f)))
	s.Router.Post("/configurable", n)
}
