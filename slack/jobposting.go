package slack

import (
	"io/ioutil"
	"log"
	"net/http"

	"encore.dev/rlog"
)

//encore:api public raw method=POST path=/slack/interactive
func InteractiveRouter(w http.ResponseWriter, req *http.Request) {
	// Parse the request body
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	rlog.Debug("Slack Interactive Webhook", "body", body)
}
