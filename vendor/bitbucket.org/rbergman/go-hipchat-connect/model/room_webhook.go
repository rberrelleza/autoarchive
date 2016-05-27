package model

import (
	"encoding/json"
	"io"
)

type RoomWebhook struct {
	Event         string          `json:"event"`
	WebhookID     int             `json:"webhook_id"`
	OAuthClientID string          `json:"oauth_client_id"`
	Item          RoomWebhookItem `json:"item"`
}

type RoomWebhookItem struct {
	Room    RoomWebhookRoom    `json:"room"`
	Message RoomWebhookMessage `json:"message"`
}

type RoomWebhookRoom struct {
	ID         int              `json:"id"`
	IsArchived bool             `json:"is_archived"`
	Version    string           `json:"version"`
	Name       string           `json:"name"`
	Privacy    string           `json:"private"`
	Links      RoomWebhookLinks `json:"links"`
}

type RoomWebhookLinks struct {
	Self         string `json:"self"`
	Members      string `json:"members"`
	Participants string `json:"participants"`
	Webhooks     string `json:"webhooks"`
}

type RoomWebhookMessage struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Date     string            `json:"date"`
	Message  string            `json:"message"`
	From     RoomWebhookUser   `json:"from"`
	Mentions []RoomWebhookUser `json:"mentions"`
}

type RoomWebhookUser struct {
	ID          int                  `json:"id"`
	Name        string               `json:"name"`
	MentionName string               `json:"mention_name"`
	Version     string               `json:"version"`
	Links       RoomWebhookUserLinks `json:"links"`
}

type RoomWebhookUserLinks struct {
	Self string `json:"self"`
}

// DecodeInstallable decodes JSON from a Reader into a new instance.
func DecodeRoomWebhook(r io.Reader) (*RoomWebhook, error) {
	var rm RoomWebhook
	dec := json.NewDecoder(r)
	err := dec.Decode(&rm)
	return &rm, err
}
