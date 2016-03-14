package web

import (
	"net/http"
	"regexp"
)

var varsRe = regexp.MustCompile(`\$\{\w+\}`)

// HandleDescriptor responds with the add-on descriptor JSON.
func (s *Server) HandleDescriptor(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(s.Descriptor))
}
