package model

import (
	"encoding/json"
	"io"
)

// NotificationColor represents any of the possible color constants available
// for API-generated room notification messages.
type NotificationColor string

const (
	// NotificationGray is the gray room notification message color.
	NotificationGray NotificationColor = "gray"
	// NotificationGreen is the green room notification message color.
	NotificationGreen NotificationColor = "green"
	// NotificationPurple is the purple room notification message color.
	NotificationPurple NotificationColor = "purple"
	// NotificationRandom indicates the server sshould choose a random room
	// notification message color.
	NotificationRandom NotificationColor = "random"
	// NotificationRed is the red room notification message color.
	NotificationRed NotificationColor = "red"
	// NotificationYellow is the yellow room notification message color.
	NotificationYellow NotificationColor = "yellow"
)

// NotificationFormat represents any of the possible format constants available
// for API-generated room notification messages.
type NotificationFormat string

const (
	// NotificationHTML is the html room notification format.
	NotificationHTML = "html"
	// NotificationText is the text room notification format.
	NotificationText = "text"
)

// Notification contains the data necessary for sending a room notification
// message.
// TODO: Support the card notification message sub-schema.
type Notification struct {
	Color   NotificationColor  `json:"color"`
	Format  NotificationFormat `json:"message_format"`
	Message string             `json:"message"`
	Notify  bool               `json:"notify"`
}

// Encode encodes a Notification as a stream of JSON bytes to the given Writer.
func (n *Notification) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(n)
}
