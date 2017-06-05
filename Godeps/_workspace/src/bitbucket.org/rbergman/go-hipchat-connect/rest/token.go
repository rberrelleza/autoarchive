package rest

import (
	"encoding/base64"

	"bitbucket.org/rbergman/go-hipchat-connect/model"
)

// GenerateToken performs a post request to /oauth/token to create a new OAuth2
// access token using the specified credentials.
func (c *Client) GenerateToken(user, pass string) (*model.Token, error) {
	var t model.Token

	tr := &model.TokenRequest{
		GrantType: model.GrantTypeClientCredentials,
		Scope:     model.ScopeSendNotification,
	}
	r, err := model.NewReader(tr)
	if err != nil {
		return &t, err
	}

	auth64 := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	res, err := c.Post(c.BaseURL+"/oauth/token", map[string]string{
		"Content-Type":  string(JSONType),
		"Authorization": "Basic " + auth64,
	}, r)
	if err != nil {
		return &t, err
	}

	err = checkJSONResponse(res, 200)
	if err != nil {
		return &t, err
	}

	return model.DecodeToken(res.Body)
}
