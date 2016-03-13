package main

import (
	"bytes"
  "io"
  "encoding/json"
)

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
  store := t.server.NewTenantStore(id)
  value, err := store.Get(id)
	if err != nil {
		return &TenantConfiguration{ID: id}, err
	}
	r := bytes.NewReader([]byte(value))
	return decode(r)
}

// Set adds a TenantConfiguration by id
func (t *TenantConfigurations) Set(configuration *TenantConfiguration) error {
	w := &bytes.Buffer{}
	err := configuration.encode(w)
	if err != nil {
		return err
	}
  store := t.server.NewTenantStore(configuration.ID)
	return store.Set(configuration.ID, w.Bytes())
}

// Del removes a Tenant by id string
func (t *TenantConfigurations) Del(id string) error {
  store := t.server.NewTenantStore(id)
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
