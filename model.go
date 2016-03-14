package main

import (
	"bitbucket.org/rbergman/go-hipchat-connect/web"
	"github.com/chakrit/go-bunyan"
	"github.com/dgrijalva/jwt-go"
)

type Server struct {
	web.Server
}

// Context keep context of the running application
type Context struct {
	pghost     string
	pguser     string
	pgpass     string
	pgdatabase string
	nworkers   int
	token      *jwt.Token
}

type Group struct {
	groupId     int
	oauthId     string
	oauthSecret string
	threshold   int
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
