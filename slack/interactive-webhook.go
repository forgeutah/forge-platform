package slack

import (
	"bytes"
	"context"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"encore.dev/rlog"
	"github.com/kollalabs/sdk-go/kc"
	"github.com/tidwall/gjson"
)

var secrets struct {
	ForgeDataAPIToken string
	KollaAPIKey       string
}

//encore:api public raw method=POST path=/slack/interactive
func InteractiveRouter(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	payload := r.PostFormValue("payload")
	rlog.Info("webhook", "payload", payload)

	rlog.Debug("slack", "valid", gjson.Valid(payload))

	//jsonStr, err := url.QueryUnescape(string(body)[8:])

	webhookType := gjson.Get(payload, "type")

	if webhookType.String() == "shortcut" {
		callbackID := gjson.Get(payload, "callback_id")
		// switch statement to handle different types of slack shorcuts
		switch callbackID.String() {
		case "job_post":
			rlog.Debug("Job Posting Shortcut Fired")
			err := JobPostForm(ctx, gjson.Get(payload, "trigger_id").String())
			if err != nil {
				rlog.Error("Error sending Job Post Form", "err", err)
				return
			}
		case "meetup_link":
			rlog.Debug("Meetup Link Shortcut Fired", "payload", payload)
			triggerID := gjson.Get(payload, "trigger_id").String()
			userID := gjson.Get(payload, "user.id").String()
			go InitiateLinkMeetup(ctx, userID, triggerID)
			return
		}
	} else if webhookType.String() == "view_submission" {
		// What form was submitted
		callbackID := gjson.Get(payload, "view.callback_id")
		// switch statement to handle different types of slack shorcuts
		switch callbackID.String() {
		case "job_post_submit":
			rlog.Debug("Job Posting Form Submitted")
			JobPostSubmit(payload)
			// Get the values from the form
			//company := gjson.Get(payload, "view.state.values.company.company.value")
		}
	}

	//rlog.Info("Slack Interactive Webhook", "body_payload", jsonStr)

}

// Send the Job Post form modal in slack to the person that ran the shortcut
func JobPostForm(ctx context.Context, triggerID string) error {
	// Open template file and parse it
	t, err := template.New("job_post_form").Parse(tmplJobPostForm)
	if err != nil {
		rlog.Error("Error parsing template", "err", err)
		return err
	}
	// Put variables in template
	data := struct {
		TriggerID string
	}{triggerID}
	// Execute template and write to buffer
	var tpl bytes.Buffer
	err = t.Execute(&tpl, data)
	if err != nil {
		rlog.Error("Error executing template", "err", err)
		return err
	}

	kolla, err := kc.New(secrets.KollaAPIKey)
	if err != nil {
		rlog.Error("unable to load kolla connect client", "error", err)
		return err
	}
	// Get consumer token
	creds, err := kolla.Credentials(ctx, "internal-slack", "internal") // Use consumer ID set in consumer token
	if err != nil {
		log.Fatalf("unable to load consumer credentials: %s\n", err)
		return err
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/views.open", &tpl)
	if err != nil {
		rlog.Error("Error creating Job Post Form request", "err", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+creds.Token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		rlog.Error("Error sending Job Post Form", "err", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		rlog.Error("Error sending Job Post Form", "status", resp.StatusCode)
		return err
	}
	// Get response body
	body, err := ioutil.ReadAll(resp.Body)
	// decode body
	stringbody := string(body)
	rlog.Debug("Job Post Form Sent", "status", resp.StatusCode)
	rlog.Debug("Job Post Form Sent", "body", stringbody)

	if err != nil {
		rlog.Error("Error sending Job Post Form", "err", err)
		return err
	}

	return nil
}

func JobPostSubmit(payload string) error {
	// Get the values from the form
	company := gjson.Get(payload, "view.state.values.company.company.value")
	url := gjson.Get(payload, "view.state.values.url.url.value")
	contactEmail := gjson.Get(payload, "view.state.values.email.contact_email.value")
	description := gjson.Get(payload, "view.state.values.description.description.value")
	userID := gjson.Get(payload, "user.id")

	rlog.Debug("Job Post Form Submitted", "company", company.Str)
	rlog.Debug("Job Post Form Submitted", "url", url.Str)
	rlog.Debug("Job Post Form Submitted", "contact", contactEmail.Str)
	rlog.Debug("Job Post Form Submitted", "description", description.Str)

	// Load user from Forge Data API
	// Create a new http client and make request
	req, err := http.NewRequest("GET", "https://forge-api.fly.dev/api/people?filters[slack_id][$eq]="+userID.Str, nil)
	if err != nil {
		rlog.Error("Error creating person data api request", "err", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+secrets.ForgeDataAPIToken)

	client := &http.Client{}
	resp, err := client.Do(req)

	// Check for error code
	if resp.StatusCode != http.StatusOK {
		rlog.Error("Error sending person search to forge data api", "status", resp.StatusCode)
		return err
	}

	if err != nil {
		rlog.Error("Error sending person search to forge data api", "err", err)
		return err
	}
	// Get response body and parse to json
	body, err := ioutil.ReadAll(resp.Body)
	// decode body
	rlog.Debug("Person Search Response", "body", string(body))

	return nil
}

// JobPosting Slack Modal/Form
const tmplJobPostForm = `{
    "trigger_id": "{{.TriggerID}}",
    "view": {
        "type": "modal",
        "callback_id": "job_post_submit",
        "title": {
            "type": "plain_text",
            "text": "Add Job Post",
            "emoji": true
        },
        "submit": {
            "type": "plain_text",
            "text": "Submit",
            "emoji": true
        },
        "close": {
            "type": "plain_text",
            "text": "Cancel",
            "emoji": true
        },
        "blocks": [
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": "Fill out the information to add a job post to Forge Utah"
                }
            },
            {
                "type": "divider"
            },
            {
                "type": "input",
				"block_id": "company",
                "element": {
                    "type": "plain_text_input",
                    "action_id": "company"
                },
                "label": {
                    "type": "plain_text",
                    "text": "Company Name",
                    "emoji": true
                }
            },
            {
                "type": "input",
				"block_id": "description",
                "element": {
                    "type": "plain_text_input",
                    "multiline": true,
                    "action_id": "description"
                },
                "label": {
                    "type": "plain_text",
                    "text": "Description",
                    "emoji": true
                }
            },
			{
                "type": "input",
				"block_id": "url",
                "element": {
                    "type": "url_text_input",
                    "action_id": "url"
                },
                "label": {
                    "type": "plain_text",
                    "text": "URL of Official Job Posting",
                    "emoji": true
                }
            },
            {
                "type": "input",
				"block_id": "email",
				"optional": true,
                "element": {
                    "type": "email_text_input",
                    "action_id": "contact_emil"
                },
                "label": {
                    "type": "plain_text",
                    "text": "Contact Email",
                    "emoji": true
                }
            }
        ]
    }
}`
