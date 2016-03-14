package web

import "net/http"

// HandleHealthcheck returns a 200 response when requested.
func (s *Server) HandleHealthcheck(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
}
