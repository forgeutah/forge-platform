package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"encore.dev/beta/errs"
	"encore.dev/rlog"
)

type PeopleResponse struct {
	Data []*Person `json:"data"`
}

// Person struct from this data: {\"display_name\":\"soypete\",\"bio\":null,\"github_user\":null,\"twitter_handle\":null,\"createdAt\":\"2023-01-25T23:30:39.795Z\",\"updatedAt\":\"2023-02-28T02:34:11.217Z\",\"publishedAt\":\"2023-02-17T19:50:42.535Z\",\"slack_id\":\"soypete\",\"meetup_id\":null}
type Person struct {
	ID         int `json:"id,omitempty"`
	Attributes struct {
		DisplayName   string `json:"display_name"`
		Bio           string `json:"bio"`
		GithubUser    string `json:"github_user"`
		TwitterHandle string `json:"twitter_handle"`
		CreatedAt     string `json:"createdAt,omitempty"`
		UpdatedAt     string `json:"updatedAt,omitempty"`
		PublishedAt   string `json:"publishedAt,omitempty"`
		SlackID       string `json:"slack_id"`
		MeetupID      string `json:"meetup_id"`
		Email         string `json:"email"`
	} `json:"attributes"`
}

// Load a user by slack handle. If user does not exist, create a new user and populate with info from slack
//encore:api public path=/data/users/:slackID
func LoadUserBySlackID(ctx context.Context, slackID string) (*Person, error) {
	ret := &Person{}

	// Load user from Forge Data API
	// Create a new http client and make request
	body, _, err := HttpRequest("GET", "people?filters[slack_id][$eq]="+slackID, nil)
	if err != nil {
		rlog.Error("Error creating person data api request", "err", err)
		return ret, err
	}
	rlog.Debug("Person Search Response", "body", string(body))

	// decode body to PeopleResponse struct
	pr := &PeopleResponse{}
	err = json.Unmarshal(body, &pr)
	if err != nil {
		rlog.Error("Error decoding person search response", "err", err)
		return ret, err
	}
	if len(pr.Data) == 0 {
		// gotta create a new user
		rlog.Debug("No person found", "slackID", slackID)
		return ret, &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("User not found: %s", slackID),
		}
	}

	rlog.Debug("Person Search Response", "body", string(body))

	return pr.Data[0], nil
}

// encore:api public method=POST path=/data/people
func CreatePerson(ctx context.Context, p *Person) (*Person, error) {
	ret := &Person{}

	personRequest := &CreatePersonRequest{}
	personRequest.Data.SlackID = p.Attributes.SlackID
	personRequest.Data.DisplayName = p.Attributes.DisplayName
	personRequest.Data.Email = p.Attributes.Email
	// Create new io reader from request
	jsonReq, err := json.Marshal(personRequest)
	if err != nil {
		rlog.Error("Error marshaling person data api request", "err", err)
		return ret, fmt.Errorf("Error marshaling person data api request: %s", err)
	}

	// Load user from Forge Data API
	body, _, err := HttpRequest("POST", "people", jsonReq)
	if err != nil {
		rlog.Error("Error creating person", "err", err)
		return ret, err
	}
	rlog.Debug("Person Create Response", "body", string(body))
	cpr := &CreatePersonResponse{}
	err = json.Unmarshal(body, &cpr)
	if err != nil {
		rlog.Error("Error decoding person create response", "err", err)
		return ret, fmt.Errorf("Error decoding person create response: %s", err)
	}

	return cpr.Data, nil
}

// encore:api private method=PUT path=/data/people/:id
func UpdatePerson(ctx context.Context, id int, p *Person) (*Person, error) {
	ret := &Person{}
	personRequest := &CreatePersonRequest{}
	personRequest.Data.SlackID = p.Attributes.SlackID
	personRequest.Data.DisplayName = p.Attributes.DisplayName
	personRequest.Data.Email = p.Attributes.Email
	// Create new io reader from request
	jsonReq, err := json.Marshal(personRequest)
	if err != nil {
		rlog.Error("Error marshaling person data api request", "err", err)
		return ret, fmt.Errorf("Error marshaling person data api request: %s", err)
	}

	stringID := strconv.Itoa(id)

	// Load user from Forge Data API
	body, _, err := HttpRequest("PUT", "people/"+stringID, jsonReq)
	if err != nil {
		rlog.Error("Error updating person", "err", err)
		return ret, err
	}
	rlog.Debug("Person Update Response", "body", string(body))
	cpr := &CreatePersonResponse{}
	err = json.Unmarshal(body, &cpr)
	if err != nil {
		rlog.Error("Error decoding person update response", "err", err)
		return ret, fmt.Errorf("Error decoding person update response: %s", err)
	}

	return cpr.Data, nil

}

type CreatePersonRequest struct {
	Data struct {
		DisplayName   string `json:"display_name"`
		Bio           string `json:"bio"`
		GithubUser    string `json:"github_user"`
		TwitterHandle string `json:"twitter_handle"`
		CreatedAt     string `json:"createdAt,omitempty"`
		UpdatedAt     string `json:"updatedAt,omitempty"`
		PublishedAt   string `json:"publishedAt,omitempty"`
		SlackID       string `json:"slack_id"`
		MeetupID      string `json:"meetup_id"`
		Email         string `json:"email"`
	} `json:"data"`
}

type CreatePersonResponse struct {
	Data *Person
}
