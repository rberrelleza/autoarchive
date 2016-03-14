package main

import (
	"bitbucket.org/rbergman/go-hipchat-connect/store"
	"bitbucket.org/rbergman/go-hipchat-connect/util"
	"github.com/garyburd/redigo/redis"
	"github.com/jasonlvhit/gocron"
	"time"
	"github.com/RichardKnop/machinery/v1/signatures"
)

func StartScheduler() {
	b := NewBackendServer("hiparchiver.scheduler")

	b.Log.Infof("Starting the scheduler")
	scheduleExternal := util.Env.GetInt("SCHEDULE_EXTERNAL")

	if scheduleExternal == 1 {
		// this is intended for environments like Heroku, where the scheduler is external
		b.autoArchive()
	} else {
		schedule := util.Env.GetStringOr("SCHEDULE_ENV", "24h")
		duration, err := time.ParseDuration(schedule)
		if err != nil {
			panic(err)
		}

		go func() {
			seconds := uint64(duration.Seconds())
			b.Log.Infof("Scheduler will run every %s", schedule)
			gocron.Every(seconds).Seconds().Do(b.autoArchive)
			<-gocron.Start()
		}()
	}
}

func (s *Server) autoArchive() {
	s.Log.Infof("start autoArchive")

	taskServer := NewTaskServer()
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
		task := signatures.TaskSignature{
	  	Name: "autoArchive",
	  	Args: []signatures.TaskArg{
	    	signatures.TaskArg{
	      	Type:  "string",
	      	Value: tenantID,
	    },
	  },
	}

	_, err := taskServer.SendTask(&task)
	if err != nil {
	  s.Log.Errorf("Failed to schedule task for tid-%s: %s", tenantID, err)
	}

	}
}
