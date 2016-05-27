package main

import (
	"flag"
	"time"

	"bitbucket.org/rbergman/go-hipchat-connect/util"
	"bitbucket.org/rbergman/go-hipchat-connect/web"
	"github.com/garyburd/redigo/redis"
)

func NewBackendServer(appName string) *Server {
	log := web.NewStdLogger(appName)

	ws := &web.Server{
		AppName:   appName,
		Log:       log,
		RedisPool: newRedisPool(),
	}

	s := &Server{*ws}
	return s
}

func main() {
	var role = flag.String("role", "web", "Which role to start: web|scheduler|worker")
	flag.Parse()

	switch *role {
	case "all":
		StartScheduler()
		StartWorker()
		startWeb()

	case "web":
		startWeb()

	case "scheduler":
		StartScheduler()

	case "worker":
		StartWorker()
	}
}

func startWeb() {
	s := &Server{*web.NewServer("./static/descriptor.json", "")}
	s.MountConfigurable(s.configurable, s.postConfigurable)
	s.Start()
}

func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			var redisEnv = util.Env.GetStringOr("REDIS_ENV", "REDIS_URL")
			var redisURL = util.Env.GetStringOr(redisEnv, "redis://127.0.0.1:6379")
			return redis.DialURL(redisURL)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
