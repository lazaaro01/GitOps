package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	apiURL  string
	client  *http.Client
}

func NewClient(apiURL string) *Client {
	return &Client{
		apiURL: apiURL,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

type EventPayload struct {
	DeployID  string `json:"deploy_id"`
	EventType string `json:"event_type"`
	Status    string `json:"status,omitempty"`
	Step      string `json:"step,omitempty"`
	Level     string `json:"level,omitempty"`
	Message   string `json:"message,omitempty"`
}

func (c *Client) Publish(payload EventPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	resp, err := c.client.Post(
		c.apiURL+"/api/internal/events",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func (c *Client) PublishUpdate(deployID, status string) {
	c.Publish(EventPayload{
		DeployID:  deployID,
		EventType: "deploy_update",
		Status:    status,
	})
}

func (c *Client) PublishLog(deployID, step, level, message string) {
	c.Publish(EventPayload{
		DeployID:  deployID,
		EventType: "deploy_log",
		Step:      step,
		Level:     level,
		Message:   message,
	})
}

func APIURL(host string, port int) string {
	return fmt.Sprintf("http://%s:%d", host, port)
}
