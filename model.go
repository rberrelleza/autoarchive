package main

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
}

type Group struct {
	groupId     int
	oauthId     string
	oauthSecret string
	threshold   int
}
