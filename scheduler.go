package main

import (
	"bitbucket.org/rbergman/go-hipchat-connect/store"
	"bitbucket.org/rbergman/go-hipchat-connect/util"
	"github.com/garyburd/redigo/redis"
	"github.com/jasonlvhit/gocron"
	"time"
)

var WorkQueue = make(chan WorkRequest, 100)

func StartScheduler() {
	b := NewBackendServer("hiparchiver.scheduler")

	b.Log.Infof("Starting the scheduler")
	schedule := util.Env.GetStringOr("SCHEDULE_ENV", "24h")
	duration, error := time.ParseDuration(schedule)
	checkErr(error)

	go func() {
		seconds := uint64(duration.Seconds())
		b.Log.Infof("Scheduler will run every %s", schedule)
		gocron.Every(seconds).Seconds().Do(b.autoArchive)
		<-gocron.Start()
	}()

}

func (s *Server) autoArchive() {
	s.Log.Infof("start autoArchive")
	conn := s.RedisPool.Get()

	// TODO: Find a better way to pull this info, this won't scale at all
	redisStore := store.NewDefaultRedisStore(conn)
	tenantScope := redisStore.Key("tenants")
	redisStore.Scope = tenantScope
	allTenantsKey := redisStore.Key("*")
	s.Log.Infof("Retrieving all the tenants: %s", allTenantsKey)
	keys, err := redis.Strings(redisStore.Conn.Do("KEYS", allTenantsKey))

	if err != nil {
		s.Log.Errorf("Error getting the tenants: %s", err)
		return
	}

	s.Log.Debugf("Found keys: %v", keys)

	for _, key := range keys {
		tenantID := key[len("hipchat:tenants:"):]
		s.Log.Debugf("Start archiving tid-%s", tenantID)
		work := WorkRequest{tenantID}
		WorkQueue <- work
	}
}
