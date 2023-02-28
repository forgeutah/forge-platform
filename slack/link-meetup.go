package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"encore.dev/rlog"
)

type ConnectorLinkRequest struct {
	ConsumerID       string `json:"consumer_id"`
	ConsumerMetadata struct {
		Title string `json:"title"`
		Email string `json:"email"`
	} `json:"consumer_metadata"`
}

type ConnectorLinkResponse struct {
	Name             string `json:"name"`
	URI              string `json:"uri"`
	Connector        string `json:"connector"`
	ConsumerID       string `json:"consumer_id"`
	ConsumerMetadata struct {
		Title string `json:"title"`
		Email string `json:"email"`
	} `json:"consumer_metadata"`
	CreateTime time.Time `json:"create_time"`
	ExpireTime time.Time `json:"expire_time"`
}

func InitiateLinkMeetup(ctx context.Context, slackID string, triggerID string) error {

	p, err := SyncSlackUserToDataApi(ctx, slackID)
	if err != nil {
		return err
	}
	// Get connect link from kolla
	// Create http client and submit request with bearer token
	connectorLinkRequest := ConnectorLinkRequest{}
	connectorLinkRequest.ConsumerID = p.Attributes.SlackID
	connectorLinkRequest.ConsumerMetadata.Title = p.Attributes.DisplayName
	connectorLinkRequest.ConsumerMetadata.Email = p.Attributes.Email

	body, err := json.Marshal(connectorLinkRequest)

	r := bytes.NewReader(body)
	req, err := http.NewRequest("POST", "https://connect.getkolla.com/v1/connectors/meetup-kolla/links", r)
	if err != nil {
		rlog.Error("Error creating api request", "err", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+secrets.KollaAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		rlog.Error("Error sending request forge data api", "err", err)
		return err
	}

	// Get response body and parse to json
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		rlog.Error("Error reading data api response", "err", err)
		return err
	}
	rlog.Debug("Data API Response", "body", string(respBody))
	// Check for error code
	if resp.StatusCode != http.StatusOK {
		rlog.Error("Error response from forge data api", "status", resp.StatusCode, "body", string(respBody))
		return fmt.Errorf("Error response from forge data api: %s", string(respBody))
	}

	connectorLinkResponse := ConnectorLinkResponse{}
	err = json.Unmarshal(respBody, &connectorLinkResponse)
	if err != nil {
		rlog.Error("Error parsing response from forge data api", "err", err)
		return err
	}

	reqBody := `{ "channel":"` + slackID + `", "text":"<` + connectorLinkResponse.URI + `|Click here to link your meetup account>" }`
	_, _, err = HttpRequest(ctx, "POST", "chat.postMessage", []byte(reqBody))
	if err != nil {
		rlog.Error("Error sending slack message", "err", err)
		return err
	}

	return nil
}
