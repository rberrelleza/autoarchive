package web

import (
	"time"

	"bitbucket.org/rbergman/go-hipchat-connect/store"
	"bitbucket.org/rbergman/go-hipchat-connect/tenant"
	"bitbucket.org/rbergman/go-hipchat-connect/util"

	"github.com/garyburd/redigo/redis"
)

// NewTenants creates a new instance of the Tenants tenant manager service
// backed by a RedisStore using a connection from the Server's Redis pool.
func (s *Server) NewTenants() *tenant.Tenants {
	conn := s.RedisPool.Get()
	rs := store.NewDefaultRedisStore(conn).Sub("tenants")
	return tenant.NewTenants(rs)
}

// NewTenantStore creates a Redis-backed Store scoped to the given tenantID.
func (s *Server) NewTenantStore(tenantID string) store.Store {
	conn := s.RedisPool.Get()
	return store.NewDefaultRedisStore(conn).Sub(tenantID)
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
