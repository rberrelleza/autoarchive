package store

import "github.com/garyburd/redigo/redis"

type RedisStore struct {
	Conn  redis.Conn
	Scope string
}

func NewRedisStore(conn redis.Conn, scope string) *RedisStore {
	return &RedisStore{
		Conn:  conn,
		Scope: scope,
	}
}

func NewDefaultRedisStore(conn redis.Conn) *RedisStore {
	return NewRedisStore(conn, "hipchat")
}

func (s *RedisStore) Key(k string) string {
	if s.Scope == "" {
		return k
	}
	return s.Scope + ":" + k
}

func (s *RedisStore) Del(k string) error {
	_, err := s.Conn.Do("DEL", s.Key(k))
	return err
}

func (s *RedisStore) Get(k string) ([]byte, error) {
	result, err := s.Conn.Do("GET", s.Key(k))
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result.([]byte), nil
	}
	return nil, nil
}

func (s *RedisStore) Set(k string, v []byte) error {
	if v == nil {
		return s.Del(k)
	}
	_, err := s.Conn.Do("SET", s.Key(k), v)
	return err
}

func (s *RedisStore) SetEx(k string, v []byte, sec int) error {
	if v == nil || sec <= 0 {
		return s.Del(k)
	}
	_, err := s.Conn.Do("SET", s.Key(k), v, "EX", sec)
	return err
}

func (s *RedisStore) Sub(scope string) Store {
	scope = s.Key(scope)
	return &RedisStore{
		Conn:  s.Conn,
		Scope: scope,
	}
}
