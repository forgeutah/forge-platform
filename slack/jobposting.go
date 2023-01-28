package slack

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"encore.dev/rlog"
	"github.com/slack-go/slack"
)

//encore:api public raw method=POST path=/slack/interactive
func InteractiveRouter(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jsonStr, err := url.QueryUnescape(string(body)[8:])
	rlog.Info("Slack Interactive Webhook", "body_payload", jsonStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message slack.InteractionCallback
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		rlog.Error("[ERROR] Failed to decode json message from slack", "jsonstr", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch message.Type {
	case slack.InteractionTypeShortcut:
		message
	default:
		rlog.Error("[WARN] unknown message type: ", "meesage_type", message.Type)
		w.WriteHeader(http.StatusInternalServerError)
	}

	//rlog.Debug("Slack Interactive Webhook", "body_payload", message.Message)
}
