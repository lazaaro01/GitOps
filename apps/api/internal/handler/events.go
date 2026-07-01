package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sync"

	"gitops-lite/pkg/model"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

var (
	clients   = make(map[string][]chan []byte)
	clientsMu sync.RWMutex
)

func NotifyDeployEvent(deployID string, eventType string, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal SSE event")
		return
	}

	event := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(payload))

	clientsMu.RLock()
	chans, ok := clients[deployID]
	clientsMu.RUnlock()

	if !ok {
		return
	}

	for _, ch := range chans {
		select {
		case ch <- []byte(event):
		default:
		}
	}
}

func PublishDeployLog(deployID, step string, level model.LogLevel, message string) {
	NotifyDeployEvent(deployID, "deploy_log", map[string]interface{}{
		"deploy_id": deployID,
		"step":      step,
		"level":     string(level),
		"message":   message,
	})
}

func PublishDeployUpdate(deployID, status string) {
	NotifyDeployEvent(deployID, "deploy_update", map[string]interface{}{
		"deploy_id": deployID,
		"status":    status,
	})

	if status == "success" || status == "failed" {
		NotifyDeployEvent(deployID, "deploy_completed", map[string]interface{}{
			"deploy_id": deployID,
			"status":    status,
		})
	}
}

type EventsHandler struct{}

func NewEventsHandler() *EventsHandler {
	return &EventsHandler{}
}

func (h *EventsHandler) Stream(c *fiber.Ctx) error {
	deployID := c.Query("deploy_id")
	if deployID == "" {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Error:   "deploy_id query parameter is required",
		})
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	ch := make(chan []byte, 64)

	clientsMu.Lock()
	clients[deployID] = append(clients[deployID], ch)
	clientsMu.Unlock()

	removeClient := func() {
		clientsMu.Lock()
		chans := clients[deployID]
		for i, c := range chans {
			if c == ch {
				clients[deployID] = append(chans[:i], chans[i+1:]...)
				break
			}
		}
		if len(clients[deployID]) == 0 {
			delete(clients, deployID)
		}
		clientsMu.Unlock()
	}

	ctx := c.Context()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer removeClient()

		initial, _ := json.Marshal(map[string]string{
			"status":    "connected",
			"deploy_id": deployID,
		})
		fmt.Fprintf(w, "event: connected\ndata: %s\n\n", initial)
		w.Flush()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				w.Write(msg)
				w.Flush()
			}
		}
	})

	return nil
}

func (h *EventsHandler) InternalPublish(c *fiber.Ctx) error {
	var req struct {
		DeployID  string `json:"deploy_id"`
		EventType string `json:"event_type"`
		Status    string `json:"status,omitempty"`
		Step      string `json:"step,omitempty"`
		Level     string `json:"level,omitempty"`
		Message   string `json:"message,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.ErrorResponse{
			Success: false,
			Error:   "invalid body",
		})
	}

	switch req.EventType {
	case "deploy_update":
		PublishDeployUpdate(req.DeployID, req.Status)
	case "deploy_log":
		PublishDeployLog(req.DeployID, req.Step, model.LogLevel(req.Level), req.Message)
	case "deploy_completed":
		NotifyDeployEvent(req.DeployID, "deploy_completed", map[string]interface{}{
			"deploy_id": req.DeployID,
			"status":    req.Status,
		})
	}

	return c.JSON(fiber.Map{"success": true})
}
