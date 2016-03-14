package web

import (
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/lonnblad/negroni-etag/etag"
	"github.com/pilu/xrequestid"
)

// NewMiddleware returns a Negroni middleware stack configured for a typical
// HipChat Connect add-on application.
func NewMiddleware(appName, staticDir string) *negroni.Negroni {
	return negroni.New(
		NewRecovery(appName),
		xrequestid.New(8),
		NewLogger(appName),
		negroni.NewStatic(http.Dir(staticDir)),
		etag.Etag(),
	)
}
