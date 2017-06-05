package web

import (
	"net/http"

	"bitbucket.org/rbergman/go-hipchat-connect/util"
)

// MountHealthCheck mounts the following routes:
// * GET /healthcheck -> s.HandleHealthCheck
// It accepts an optional path param to override the route path.
func (s *Server) MountHealthCheck(path ...string) {
	s.Router.GetFunc(util.FirstOrDefault(path, "/healthcheck"), s.HandleHealthCheck)
}

// HandleHealthCheck returns a 200 response when requested.
func (s *Server) HandleHealthCheck(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
}
