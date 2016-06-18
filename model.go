package main

import (
	"time"

	"bitbucket.org/rbergman/go-hipchat-connect/web"
	"github.com/chakrit/go-bunyan"
	"github.com/tbruyelle/hipchat-go/hipchat"
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

type Job struct {
	JobID      string
	TenantID   string
	Log        bunyan.Log
	Client     *hipchat.Client
	Clock      clock
	HipChatURL string
	DryRun     bool
}

// clock is used to be able to mock time.Now() for testing purposes
type clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }
