package notificationservices

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	utils "github.com/st-tech/gatling-operator/pkg/utils"
)

type SlackNotificationServiceProvider struct {
	providerName string
}

func (p *SlackNotificationServiceProvider) GetName() string {
	return p.providerName
}

func (p *SlackNotificationServiceProvider) Notify(gatlingName string, reportURL string, secretData map[string][]byte) error {

	webhookURL, exists := utils.GetMapValue("incoming-webhook-url", secretData)
	if !exists {
		return errors.New("Insufficient secret data for slack: incoming-webhook-url is missing")
	}
	payloadTextFormat := `
[%s] Gatling has completed successfully!
Report URL: %s
`
	payloadText := fmt.Sprintf(payloadTextFormat, gatlingName, reportURL)
	if err := slackNotify(webhookURL, payloadText); err != nil {
		return err
	}
	return nil
}

type payload struct {
	Text string `json:"text"`
}

func slackNotify(webhookURL string, payloadText string) error {
	p := payload{
		Text: payloadText,
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	resp, err := http.PostForm(webhookURL,
		url.Values{"payload": {string(s)}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
