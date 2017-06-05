package tenant

import (
	"encoding/json"
	"errors"
	"io"

	"bitbucket.org/rbergman/go-hipchat-connect/model"
)

// A Tenant represents a remote integration tenant.
type Tenant struct {
	ID        string
	Secret    string
	GroupID   int
	GroupName string
	RoomID    int
	Links     Links
}

// Links is a list of URLs to key resources for the tenant.
type Links struct {
	API          string
	Base         string
	Capabilities string
	Token        string
}

// New creates a new Tenant by combining information from an Installable
// payload and the host Descriptor.
func New(i *model.Installable, d *model.Descriptor) (*Tenant, error) {
	if i.CapabilitiesURL != d.Links.Self {
		return nil, errors.New("Capabilities URL mismatch: " + i.CapabilitiesURL + " != " + d.Links.Self)
	}
	return &Tenant{
		ID:      i.OAuthID,
		Secret:  i.OAuthSecret,
		GroupID: i.GroupID,
		RoomID:  i.RoomID,
		Links: Links{
			API:          d.Links.API,
			Base:         d.Links.Homepage,
			Capabilities: i.CapabilitiesURL,
			Token:        d.Capabilities.OAuth2Provider.TokenURL,
		},
	}, nil
}

// Decode decodes a Tenant from a stream of JSON bytes from the given Reader.
func Decode(r io.Reader) (*Tenant, error) {
	var t Tenant
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&t); err != nil {
		return &t, err
	}
	return &t, nil
}

// Encode encodes a Tenant as a stream of JSON bytes to the given Writer.
func (t *Tenant) Encode(w io.Writer) error {
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(t); err != nil {
		return err
	}
	return nil
}
