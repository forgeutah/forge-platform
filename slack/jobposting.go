package slack

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"encore.dev/rlog"
)

type SlackPayload struct {
	Payload json.RawMessage `json:"payload"`
}

//encore:api public raw method=POST path=/slack/interactive
func InteractiveRouter(w http.ResponseWriter, req *http.Request) {
	// Parse the request body
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	// Create payload instance
	var payload SlackPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		rlog.Error("Error unmarshalling payload", "payload", err)
		http.Error(w, "can't unmarshal payload", http.StatusBadRequest)
		return
	}
	rlog.Debug("Slack Interactive Webhook", "body_payload", payload.Payload)
}
