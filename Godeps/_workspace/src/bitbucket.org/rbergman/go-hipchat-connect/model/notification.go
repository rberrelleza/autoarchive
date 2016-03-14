package model

import (
	"encoding/json"
	"io"
)

type NotificationColor string
type NotificationFormat string

const (
	NotificationGray   NotificationColor = "gray"
	NotificationGreen  NotificationColor = "green"
	NotificationPurple NotificationColor = "purple"
	NotificationRandom NotificationColor = "random"
	NotificationRed    NotificationColor = "red"
	NotificationYellow NotificationColor = "yellow"

	NotificationHTML = "html"
	NotificationText = "text"
)

type Notification struct {
	Color   NotificationColor  `json:"color"`
	Format  NotificationFormat `json:"message_format"`
	Message string             `json:"message"`
	Notify  bool               `json:"notify"`
}

func (n *Notification) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(n)
}
