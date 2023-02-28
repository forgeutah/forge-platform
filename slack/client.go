package slack

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"encore.dev/rlog"
	"github.com/kollalabs/sdk-go/kc"
)

const BaseURL = "https://slack.com/api/"

func HttpRequest(ctx context.Context, method string, path string, body []byte) ([]byte, *http.Response, error) {
	respBody := []byte{}
	resp := &http.Response{}

	kolla, err := kc.New(secrets.KollaAPIKey)
	if err != nil {
		rlog.Error("unable to load kolla connect client", "error", err)
		return respBody, resp, err
	}
	// Get consumer token
	creds, err := kolla.Credentials(ctx, "internal-slack", "internal") // Use consumer ID set in consumer token
	if err != nil {
		log.Fatalf("unable to load consumer credentials: %s\n", err)
		return respBody, resp, err
	}

	rlog.Debug("Slack API Request", "method", method, "url", BaseURL+path, "body", string(body))

	r := bytes.NewReader(body)
	req, err := http.NewRequest(method, BaseURL+path, r)
	if err != nil {
		rlog.Error("Error creating slack api request", "err", err)
		return respBody, resp, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+creds.Token)

	client := &http.Client{}
	resp, err = client.Do(req)

	if err != nil {
		rlog.Error("Error sending request to slack", "err", err)
		return respBody, resp, err
	}

	// Get response body and parse to json
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		rlog.Error("Error reading slack api response", "err", err)
		return respBody, resp, err
	}
	rlog.Debug("Data API Response", "body", string(respBody))
	// Check for error code
	if resp.StatusCode != http.StatusOK {
		rlog.Error("Error response from slack api", "status", resp.StatusCode, "body", string(respBody))
		return respBody, resp, fmt.Errorf("Error response from slack api: %s", string(respBody))
	}

	return respBody, resp, nil
}
