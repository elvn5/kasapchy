// Package usersapi is an HTTP client for the users service internal API.
package usersapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/your-org/app/bot/internal/domain"
)

// Client implements usecase.UserRegistry over HTTP.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient creates a client for the users service at baseURL.
func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 5 * time.Second}}
}

type registerRequest struct {
	TelegramID int64   `json:"telegramId"`
	Username   *string `json:"username"`
	FirstName  *string `json:"firstName"`
	LastName   *string `json:"lastName"`
}

// Register implements usecase.UserRegistry.
func (c *Client) Register(ctx context.Context, user domain.TelegramUser) error {
	body, err := json.Marshal(registerRequest{
		TelegramID: user.ID,
		Username:   nullable(user.Username),
		FirstName:  nullable(user.FirstName),
		LastName:   nullable(user.LastName),
	})
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/internal/users/telegram", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call users service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("users service responded with %s", resp.Status)
	}
	return nil
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
