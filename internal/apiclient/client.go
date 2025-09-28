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
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/habits", nil)
	if err != nil {
		logger.Error("Failed to create list habits request", "base_url", c.BaseURL, "error", err)
		return nil, fmt.Errorf("failed to create request for %s/habits: %w", c.BaseURL, err)
	}
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
	url := c.BaseURL + "/habits/" + name + "/summary"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}
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
	habitJson, err := json.Marshal(h)
	if err != nil {
		logger.Error("Failed to marshal habit", "habit_name", h.Name, "error", err)
		return fmt.Errorf("failed to marshal habit %s: %w", h.Name, err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/habits", nil)
	if err != nil {
		logger.Error("Failed to create put habit request", "base_url", c.BaseURL, "error", err)
		return fmt.Errorf("failed to create request for %s/habits: %w", c.BaseURL, err)
	}
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
