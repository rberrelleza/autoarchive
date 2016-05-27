package web

import (
	"net/http"

	"bitbucket.org/rbergman/go-hipchat-connect/util"
)

// MountDescriptor mounts the following routes:
// * GET /descriptor -> s.HandleDescriptor
// It accepts an optional path param to override the route path.
func (s *Server) MountDescriptor(path ...string) {
	s.Router.GetFunc(util.FirstOrDefault(path, "/descriptor"), s.HandleDescriptor)
}

// HandleDescriptor responds with the add-on descriptor JSON.
func (s *Server) HandleDescriptor(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(s.Descriptor))
}
