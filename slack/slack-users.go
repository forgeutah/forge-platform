package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"encore.app/data"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"github.com/kollalabs/sdk-go/kc"
)

// encore:api private  method=GET path=/slack/users/:id/sync
func SyncSlackUserToDataApi(ctx context.Context, id string) (*data.Person, error) {
	var person *data.Person

	slackUser, err := GetSlackUserByID(ctx, id)
	if err != nil {
		return person, fmt.Errorf("couldn't load slack user: %w", err)
	}

	p, derr := data.LoadUserBySlackID(ctx, id)
	if derr != nil {
		e, ok := derr.(*errs.Error)
		if !ok {
			rlog.Debug("Error loading user by slack id", "error", derr.Error())
			return person, derr
		}
		if e.Code == errs.NotFound {
			// Create a new user
			pReq := &data.Person{}
			pReq.Attributes.SlackID = slackUser.ID
			pReq.Attributes.DisplayName = slackUser.Name
			pReq.Attributes.Email = slackUser.Profile.Email

			p, err = data.CreatePerson(ctx, pReq)
			rlog.Debug("Person created", p)
			if err != nil {
				return person, err
			}
		}
	}
	rlog.Debug("Person found or credated", "person", p)
	cpr := &data.Person{}
	cpr.Attributes.SlackID = slackUser.ID
	cpr.Attributes.DisplayName = slackUser.Name
	cpr.Attributes.Email = slackUser.Profile.Email

	newp, err := data.UpdatePerson(ctx, p.ID, cpr)
	return newp, nil
}

// GetSlackUserByID returns a slack user by id
// encore:api private path=/slack/users/:id
func GetSlackUserByID(ctx context.Context, id string) (SlackUser, error) {
	var user SlackUser
	kolla, err := kc.New(secrets.KollaAPIKey)
	if err != nil {
		rlog.Error("unable to load kolla connect client", "error", err)
		return user, err
	}
	// Get consumer token
	creds, err := kolla.Credentials(ctx, "internal-slack", "internal") // Use consumer ID set in consumer token
	if err != nil {
		log.Fatalf("unable to load consumer credentials: %s\n", err)
		return user, err
	}

	req, err := http.NewRequest("GET", "https://slack.com/api/users.info?user="+id, nil)
	if err != nil {
		rlog.Error("Error creating Job Post Form request", "err", err)
		return user, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+creds.Token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		rlog.Error("Error sending Job Post Form", "err", err)
		return user, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rlog.Error("Error sending Job Post Form", "status", resp.StatusCode)
		return user, err
	}
	// Get response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		rlog.Error("Error reading body from slack api", "err", err)
		return user, err
	}
	// decode body to UserProfileResponse
	userResp := UserProfileResponse{}
	err = json.Unmarshal(body, &userResp)
	if err != nil {
		rlog.Error("Error decoding user profile response", "err", err)
		return user, err
	}
	if !userResp.Ok {
		rlog.Error("Error getting user profile", "err", userResp.Error)
	}

	return userResp.User, err
}

type UserProfileResponse struct {
	Ok    bool      `json:"ok"`
	Error string    `json:"error"`
	User  SlackUser `json:"user"`
}

type SlackUser struct {
	ID       string `json:"id"`
	TeamID   string `json:"team_id"`
	Name     string `json:"name"`
	Deleted  bool   `json:"deleted"`
	Color    string `json:"color"`
	RealName string `json:"real_name"`
	Tz       string `json:"tz"`
	TzLabel  string `json:"tz_label"`
	TzOffset int    `json:"tz_offset"`
	Profile  struct {
		AvatarHash            string `json:"avatar_hash"`
		StatusText            string `json:"status_text"`
		StatusEmoji           string `json:"status_emoji"`
		RealName              string `json:"real_name"`
		DisplayName           string `json:"display_name"`
		RealNameNormalized    string `json:"real_name_normalized"`
		DisplayNameNormalized string `json:"display_name_normalized"`
		Email                 string `json:"email"`
		ImageOriginal         string `json:"image_original"`
		Image24               string `json:"image_24"`
		Image32               string `json:"image_32"`
		Image48               string `json:"image_48"`
		Image72               string `json:"image_72"`
		Image192              string `json:"image_192"`
		Image512              string `json:"image_512"`
		Team                  string `json:"team"`
	} `json:"profile"`
	IsAdmin           bool `json:"is_admin"`
	IsOwner           bool `json:"is_owner"`
	IsPrimaryOwner    bool `json:"is_primary_owner"`
	IsRestricted      bool `json:"is_restricted"`
	IsUltraRestricted bool `json:"is_ultra_restricted"`
	IsBot             bool `json:"is_bot"`
	Updated           int  `json:"updated"`
	IsAppUser         bool `json:"is_app_user"`
	Has2Fa            bool `json:"has_2fa"`
}
