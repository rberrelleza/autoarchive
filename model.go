package main

import (
	"github.com/dgrijalva/jwt-go"
)

type WorkRequest struct {
	gid int
}

// Context keep context of the running application
type Context struct {
	baseURL    string
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
}
