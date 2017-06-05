package web

import (
	"os"

	"bitbucket.org/rbergman/go-hipchat-connect/util"

	"github.com/chakrit/go-bunyan"
)

// NewStdLogger returns a Bunyan logger that writes to stdout, whose format is
// configured for the current server run mode, as follows:
// * `GO_ENV=development`: human-friendly record formatting
// * `GO_ENV=production`:  machine-friendly (JSON) record formatting
func NewStdLogger(appName string) bunyan.Log {
	var level bunyan.Level
	if util.Env.Debug {
		level = bunyan.DEBUG
	} else {
		level = bunyan.INFO
	}
	var sink bunyan.Sink
	if util.Env.IsProd() {
		sink = bunyan.StdoutSink()
	} else {
		sink = bunyan.NewFormatSink(os.Stdout)
	}
	filter := bunyan.FilterSink(level, sink)
	return bunyan.NewStdLogger(appName, filter)
}
