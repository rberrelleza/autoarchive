package tenant

import (
	"bytes"

	"bitbucket.org/rbergman/go-hipchat-connect/store"
)

// Tenants manages a collection of known integration tenants
type Tenants struct {
	store store.Store
}

// NewTenants creates a new Tenants instance with persistence provided by a
// given Store
func NewTenants(s store.Store) *Tenants {
	return &Tenants{store: s}
}

// Del removes a Tenant by id string
func (t *Tenants) Del(id string) error {
	return t.store.Del(id)
}

// Get returns a Tenant by id string
func (t *Tenants) Get(id string) (*Tenant, error) {
	value, err := t.store.Get(id)
	if err != nil {
		return &Tenant{}, err
	}
	r := bytes.NewReader([]byte(value))
	return Decode(r)
}

// Set adds a Tenant by id
func (t *Tenants) Set(tenant *Tenant) error {
	w := &bytes.Buffer{}
	err := tenant.Encode(w)
	if err != nil {
		return err
	}
	return t.store.Set(tenant.ID, w.Bytes())
}
