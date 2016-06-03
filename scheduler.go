package main

import (
	"fmt"

	"bitbucket.org/rbergman/go-hipchat-connect/store"
	"bitbucket.org/rbergman/go-hipchat-connect/util"
	"github.com/RichardKnop/machinery/v1/signatures"
	"github.com/garyburd/redigo/redis"
)

// StartScheduler schedules one job per tenant registered
func StartScheduler() {
	b := NewBackendServer("hiparchiver.scheduler")

	b.Log.Infof("Starting the scheduler")
	b.scheduleTasks()
}

func (s *Server) scheduleTasks() {
	s.Log.Infof("start autoArchive")

	taskServer := NewTaskServer()

	tenant := util.Env.GetString("TENANT")
	var keys []string
	if tenant == "" {
		keys = s.getAllTenants()
	} else {
		keys = s.getTenant(tenant)
	}

	for _, key := range keys {
		tenantID := key[len("hipchat:tenants:"):]
		s.Log.Infof("Start archiving tid-%s", tenantID)
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

func (s *Server) getAllTenants() []string {
	return s.getKeys("*")
}

func (s *Server) getTenant(tenantID string) []string {
	return s.getKeys(tenantID)
}

func (s *Server) getKeys(key string) []string {
	// TODO: Find a better way to pull this info, this won't scale at all
	conn := s.RedisPool.Get()
	redisStore := store.NewDefaultRedisStore(conn)
	tenantScope := redisStore.Key("tenants")
	redisStore.Scope = tenantScope
	tenantKey := redisStore.Key(key)
	keys, err := redis.Strings(redisStore.Conn.Do("KEYS", tenantKey))
	if err != nil {
		panic(fmt.Sprintf("Error getting the tenants: %s", err))
	}

	return keys
}
