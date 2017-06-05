package web

import (
	"net/http"

	"bitbucket.org/rbergman/go-hipchat-connect/util"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
)

// MountConfigurable mounts the following routes:
// * GET  /configurable -> get func
// * POST /configurable -> post func
// It accepts an optional path param to override the route path.
func (s *Server) MountConfigurable(get http.HandlerFunc, post http.HandlerFunc, path ...string) {
	p := util.FirstOrDefault(path, "/configurable")
	s.MountGetConfigurable(get, p)
	s.MountPostConfigurable(post, p)
}

// MountGetConfigurable mounts the following routes:
// * GET  /configurable -> get func
// It accepts an optional path param to override the route path.
func (s *Server) MountGetConfigurable(get http.HandlerFunc, path ...string) {
	n := negroni.New(
		NewAuthenticate(s),
		negroni.Wrap(context.ClearHandler(get)),
	)
	s.Router.Get(util.FirstOrDefault(path, "/configurable"), n)
}

// MountPostConfigurable mounts the following routes:
// * POST /configurable -> post func
// It accepts an optional path param to override the route path.
func (s *Server) MountPostConfigurable(post http.HandlerFunc, path ...string) {
	n := negroni.New(
		NewAuthenticate(s),
		negroni.Wrap(context.ClearHandler(post)),
	)
	s.Router.Post(util.FirstOrDefault(path, "/configurable"), n)
}
