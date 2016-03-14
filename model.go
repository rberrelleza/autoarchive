package main

import (
	"bitbucket.org/rbergman/go-hipchat-connect/web"
	"github.com/chakrit/go-bunyan"
)

type Server struct {
	web.Server
}

type Worker struct {
	ID          int
	Work        chan WorkRequest
	WorkerQueue chan chan WorkRequest
	QuitChan    chan bool
	Log         bunyan.Log
}

type WorkRequest struct {
	TenantID string
}
