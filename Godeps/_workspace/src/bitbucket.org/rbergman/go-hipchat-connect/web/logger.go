package web

import (
	"os"

	"bitbucket.org/rbergman/go-hipchat-connect/util"

	"github.com/chakrit/go-bunyan"
)

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
