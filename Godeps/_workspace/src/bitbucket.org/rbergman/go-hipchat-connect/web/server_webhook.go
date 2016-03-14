package web

import (
	"net/http"

	"bitbucket.org/rbergman/go-hipchat-connect/model"
)

// VerifyWebhook returns a RoomWebhook instance if valid, and nil otherwise.  If nil, the
// response will be complete.
// TODO try to turn this into route-specific middleware
func (s *Server) VerifyWebhook(w http.ResponseWriter, req *http.Request) *model.RoomWebhook {
	// TODO: validate auth header

	code, err := s.VerifyJSONRequest(req)
	if err != nil {
		w.WriteHeader(code)
		w.Write([]byte(err.Error()))
		return nil
	}

	wh, err := model.DecodeRoomWebhook(req.Body)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return nil
	}

	tenantID := wh.OAuthClientID
	t, err := s.NewTenants().Get(tenantID)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return nil
	}
	if t == nil {
		w.WriteHeader(404)
		return nil
	}

	return wh
}

// RespondText returns a gray-colored, text-formatted notification message from
// a webhook.
func (s *Server) RespondText(msg string, w http.ResponseWriter) {
	s.Respond(msg, model.NotificationText, model.NotificationGray, false, w)
}

// RespondHTML returns a gray-colored, HTML-formatted notification message from
// a webhook.
func (s *Server) RespondHTML(msg string, w http.ResponseWriter) {
	s.Respond(msg, model.NotificationHTML, model.NotificationGray, false, w)
}

// RespondConfused returns a gray-colored, text-formatted notification message
// from with a canned "confused" response a webhook.
func (s *Server) RespondConfused(w http.ResponseWriter) {
	s.RespondText("Sorry, I didn't understand that.", w)
}

// RespondError returns a red-colored, text-formatted error notification message
// from a webhook.
func (s *Server) RespondError(msg string, w http.ResponseWriter) {
	s.Respond(msg, model.NotificationText, model.NotificationRed, false, w)
}

// RespondServerError returns a red-colored, text-formatted canned, error
// notification message from a webhook, and logs the error to the server's
// logger.
func (s *Server) RespondServerError(err error, w http.ResponseWriter) {
	s.Log.Errorf("Error handling webhook: " + err.Error())
	msg := "Oh, oh... do you smell something burning?"
	s.RespondError(msg, w)
}

// Respond returns a notification message with the specified format, color,
// and notify flag set.
func (s *Server) Respond(msg string, format model.NotificationFormat, color model.NotificationColor, notify bool, w http.ResponseWriter) {
	w.WriteHeader(200)
	n := model.Notification{
		Message: msg,
		Format:  format,
		Color:   color,
		Notify:  notify,
	}
	n.Encode(w)
}
