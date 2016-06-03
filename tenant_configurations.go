package main

import (
	"bytes"
	"encoding/json"
	"io"

	_ "github.com/garyburd/redigo/redis"
)

const storeKey = "configurations"

// Tenants manages a collection of known integration tenants
type TenantConfigurations struct {
	server *Server
}

type TenantConfiguration struct {
	ID        string
	Threshold int
}

func (s *Server) NewTenantConfigurations() *TenantConfigurations {
	return &TenantConfigurations{server: s}
}

// Get returns a TenantConfiguration by id string
func (t *TenantConfigurations) Get(id string) (*TenantConfiguration, error) {
	store := t.server.NewTenantStore(storeKey)
	value, err := store.Get(id)

	if err != nil {
		t.server.Log.Debugf("Error when getting configuration for tid-%s: %s", id, err)
		return &TenantConfiguration{ID: id}, err
	} else if len(value) == 0 {
		t.server.Log.Debugf("Didn't find getting configuration for tid-%s, returning default", id)
		return &TenantConfiguration{ID: id, Threshold: 90}, nil
	} else {
		r := bytes.NewReader([]byte(value))
		return decode(r)
	}
}

// Set adds a TenantConfiguration by id
func (t *TenantConfigurations) Set(configuration *TenantConfiguration) error {
	w := &bytes.Buffer{}
	err := configuration.encode(w)
	if err != nil {
		return err
	}

	store := t.server.NewTenantStore(storeKey)
	return store.Set(configuration.ID, w.Bytes())
}

// Del removes a Tenant by id string
func (t *TenantConfigurations) Del(id string) error {
	store := t.server.NewTenantStore(storeKey)
	return store.Del(id)
}

func decode(r io.Reader) (*TenantConfiguration, error) {
	var t TenantConfiguration
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&t)
	if err != nil {
		return &t, err
	}
	return &t, nil
}

func (t *TenantConfiguration) encode(w io.Writer) error {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(t)
	if err != nil {
		return err
	}
	return nil
}
