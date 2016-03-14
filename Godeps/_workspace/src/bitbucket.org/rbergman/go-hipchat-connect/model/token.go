package model

import (
	"encoding/json"
	"io"
)

const (
	TokenTypeBearer            = "bearer"
	GrantTypeClientCredentials = "client_credentials"
	ScopeImportData            = "import_data"
	ScopeViewMessages          = "view_messages"
	ScopeSendNotification      = "send_notification"
	ScopeSendMessage           = "send_message"
	ScopeAdminRoom             = "admin_room"
	ScopeViewRoom              = "view_room"
	ScopeViewGroup             = "view_group"
	ScopeAdminGroup            = "admin_group"
	ScopeManageRoom            = "manage_rooms"
)

type Token struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scopes"`
	GroupID      int    `json:"group_id"`
	GroupName    string `json:"group_name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenRequest struct {
	GrantType string `json:"grant_type"`
	Scope     string `json:"scope"`
}

// DecodeToken decodes JSON from a Reader into a new instance.
func DecodeToken(r io.Reader) (*Token, error) {
	var t Token
	dec := json.NewDecoder(r)
	err := dec.Decode(&t)
	return &t, err
}

func (tr *TokenRequest) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(tr)
}
