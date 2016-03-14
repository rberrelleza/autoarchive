package web

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/chakrit/go-bunyan"
)

// Recovery is a Negroni middleware that recovers from any panics and writes a 500 if there was one.
type Recovery struct {
	Logger     bunyan.Log
	PrintStack bool
	StackAll   bool
	StackSize  int
}

// NewRecovery returns a new instance of Recovery.
func NewRecovery(appName string) *Recovery {
	return &Recovery{
		Logger:     NewStdLogger(appName),
		PrintStack: true,
		StackAll:   false,
		StackSize:  1024 * 8,
	}
}

// ServeHTTP satisfies the contract to serve as Negroni middleware.
func (rec *Recovery) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			stack := make([]byte, rec.StackSize)
			stack = stack[:runtime.Stack(stack, rec.StackAll)]
			sstack := string(stack)

			rec.Logger.
				Record("error", err).
				Record("stack", sstack).
				Errorf("panic")

			if rec.PrintStack {
				fmt.Fprintf(rw, "PANIC: %s\n%s", err, sstack)
			}
		}
	}()

	next(rw, r)
}
