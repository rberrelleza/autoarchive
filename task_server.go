package main

import (
	"bitbucket.org/rbergman/go-hipchat-connect/util"
	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
)

func NewTaskServer() *machinery.Server {
	var redisEnv = util.Env.GetStringOr("REDIS_WORKER_ENV", "REDIS_URL")
	var redisURL = util.Env.GetStringOr(redisEnv, "redis://127.0.0.1:6379")

	var cnf = config.Config{
		Broker:        redisURL,
		ResultBackend: redisURL,
		DefaultQueue:  "machinery_tasks",
	}

	server, err := machinery.NewServer(&cnf)
	if err != nil {
		panic(err)
	}

	return server
}
