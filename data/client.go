package data

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"encore.dev/rlog"
)

const BaseURL = "https://forge-api.fly.dev/api/"

var secrets struct {
	ForgeDataAPIToken string
	KollaAPIKey       string
}

func HttpRequest(method string, path string, body []byte) ([]byte, *http.Response, error) {
	respBody := []byte{}
	resp := &http.Response{}

	rlog.Debug("Data API Request", "method", method, "url", BaseURL+path, "body", string(body))

	r := bytes.NewReader(body)
	req, err := http.NewRequest(method, BaseURL+path, r)
	if err != nil {
		rlog.Error("Error creating api request", "err", err)
		return respBody, resp, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+secrets.ForgeDataAPIToken)

	client := &http.Client{}
	resp, err = client.Do(req)

	if err != nil {
		rlog.Error("Error sending request forge data api", "err", err)
		return respBody, resp, err
	}

	// Get response body and parse to json
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		rlog.Error("Error reading data api response", "err", err)
		return respBody, resp, err
	}
	rlog.Debug("Data API Response", "body", string(respBody))
	// Check for error code
	if resp.StatusCode != http.StatusOK {
		rlog.Error("Error response from forge data api", "status", resp.StatusCode, "body", string(respBody))
		return respBody, resp, fmt.Errorf("Error response from forge data api: %s", string(respBody))
	}

	return respBody, resp, nil
}
