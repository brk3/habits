package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/brk3/habits/internal/server"
	"github.com/brk3/habits/pkg/habit"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func New(base string) *Client {
	return &Client{
		BaseURL: base,
		HTTP:    http.DefaultClient,
	}
}

func (c *Client) ListHabits(ctx context.Context) ([]string, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/habits", nil)
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("list habits: %s", res.Status)
	}
	var response server.HabitListResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.Habits, nil
}

func (c *Client) GetHabitSummary(ctx context.Context, name string) (*habit.HabitSummary, error) {
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
