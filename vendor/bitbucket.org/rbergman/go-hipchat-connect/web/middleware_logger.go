package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/chakrit/go-bunyan"
	"github.com/codegangsta/negroni"
)

// Logger is a middleware handler that logs the request as it goes in and the
// response as it goes out.
type Logger struct {
	// Logger inherits from bunyan.Log used to log messages with the Logger
	// middleware.
	bunyan.Log
}

// NewLogger returns a new Logger instance.
func NewLogger(appName string) *Logger {
	return &Logger{NewStdLogger(appName)}
}

// ServeHTTP satisfies the contract to serve as Negroni middleware.
func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	l.Record("method", r.Method).
		Record("path", r.URL.Path).
		Record("requestId", r.Header.Get("X-Request-Id")).
		Infof("started")

	next(rw, r)

	res := rw.(negroni.ResponseWriter)
	status := res.Status()
	statusText := http.StatusText(status)
	duration := time.Since(start)
	durationText := fmt.Sprintf("%v", duration)
	l.Record("status", status).
		Record("statusText", statusText).
		Record("duration", duration).
		Record("durationText", durationText).
		Record("requestId", r.Header.Get("X-Request-Id")).
		Infof("completed")
}
