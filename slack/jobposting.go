package slack

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"encore.dev/rlog"
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

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	rlog.Info("Slack Interactive Webhook", "body_payload", jsonStr)

}
