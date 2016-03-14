package model

import (
	"encoding/json"
	"io"
)

// Installable represents the payload from an installation webhook.
type Installable struct {
	CapabilitiesURL string `json:"capabilitiesUrl"`
	OAuthID         string `json:"oauthId"`
	OAuthSecret     string `json:"oauthSecret"`
	GroupID         int    `json:"groupId"`
	RoomID          int    `json:"roomId"`
}

// DecodeInstallable decodes JSON from a Reader into a new instance.
func DecodeInstallable(r io.Reader) (*Installable, error) {
	var i Installable
	dec := json.NewDecoder(r)
	err := dec.Decode(&i)
	return &i, err
}
