package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/brk3/habits/internal/logger"
	"github.com/brk3/habits/internal/server"
	"github.com/brk3/habits/pkg/habit"
)

type APIClient struct {
	BaseURL   string
	HTTP      *http.Client
	AuthToken string
}

func New(base, authToken string) *APIClient {
	return &APIClient{
		BaseURL:   base,
		HTTP:      http.DefaultClient,
		AuthToken: authToken,
	}
}

func (c *APIClient) ListHabits(ctx context.Context) ([]string, error) {
	logger.Debug("Listing habits via API", "base_url", c.BaseURL)
	req, _ := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/habits", nil)
	req.Header.Add("Authorization", `Bearer `+c.AuthToken)
	res, err := c.HTTP.Do(req)
	if err != nil {
		logger.Error("Failed to list habits", "error", err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		logger.Warn("List habits request failed", "status", res.Status)
		return nil, fmt.Errorf("list habits: %s", res.Status)
	}
	var response server.HabitListResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		logger.Error("Failed to decode habits response", "error", err)
		return nil, err
	}
	logger.Debug("Listed habits successfully", "count", len(response.Habits))
	return response.Habits, nil
}

func (c *APIClient) GetHabitSummary(ctx context.Context, name string) (*habit.HabitSummary, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/habits/"+name+"/summary", nil)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("summary %s: %s", name, res.Status)
	}
	var out habit.HabitSummary
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *APIClient) PutHabit(ctx context.Context, h *habit.Habit) error {
	logger.Debug("Putting habit via API", "habit_name", h.Name, "base_url", c.BaseURL)
	habitJson, _ := json.Marshal(h)
	req, _ := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/habits", nil)
	req.Header.Add("Authorization", `Bearer `+c.AuthToken)
	req.Header.Add("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(habitJson))

	res, err := c.HTTP.Do(req)
	if err != nil {
		logger.Error("Failed to put habit", "habit_name", h.Name, "error", err)
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		logger.Warn("Put habit request failed", "habit_name", h.Name, "status", res.Status)
		return fmt.Errorf("put habit failed: %s", res.Status)
	}
	logger.Debug("Put habit successful", "habit_name", h.Name)
	return nil
}
