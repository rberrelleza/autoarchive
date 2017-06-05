package store

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type MemoryStore struct {
	Scope string
	data  *cache.Cache
}

func NewMemoryStore(scope string) *MemoryStore {
	data := cache.New(cache.NoExpiration, time.Minute)
	return &MemoryStore{
		Scope: scope,
		data:  data,
	}
}

func NewDefaultMemoryStore() *MemoryStore {
	return NewMemoryStore("")
}

func (s *MemoryStore) Key(k string) string {
	if s.Scope == "" {
		return k
	}
	return s.Scope + ":" + k
}

func (s *MemoryStore) Del(k string) error {
	s.data.Delete(k)
	return nil
}

func (s *MemoryStore) Get(k string) ([]byte, error) {
	v, _ := s.data.Get(s.Key(k))
	return v.([]byte), nil
}

func (s *MemoryStore) Set(k string, v []byte) error {
	if v == nil {
		return s.Del(k)
	}
	s.data.Set(s.Key(k), v, cache.NoExpiration)
	return nil
}

func (s *MemoryStore) SetEx(k string, v []byte, sec int) error {
	if v == nil {
		return s.Del(k)
	}
	s.data.Set(s.Key(k), v, time.Duration(sec)*time.Second)
	return nil
}

func (s *MemoryStore) Sub(scope string) Store {
	if s.Scope != "" {
		scope = s.Scope + ":" + scope
	}
	return NewMemoryStore(scope)
}
