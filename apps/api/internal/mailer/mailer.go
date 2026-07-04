// Package mailer sends transactional email through the Resend HTTP API
// (https://resend.com/docs/api-reference/emails/send-email).
package mailer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ErrNotConfigured is returned by Send when no Resend API key is configured.
// Callers should treat this as "email sending is disabled" (e.g. local dev
// without an API key) rather than a hard failure.
var ErrNotConfigured = errors.New("mailer: resend api key not configured")

const sendEndpoint = "https://api.resend.com/emails"

type Config struct {
	APIKey string
	From   string
}

type Mailer struct {
	cfg    Config
	client *http.Client
}

func New(cfg Config) *Mailer {
	return &Mailer{cfg: cfg, client: &http.Client{Timeout: 10 * time.Second}}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

// Send delivers a plain-text email via Resend. It returns ErrNotConfigured
// if no API key is set.
func (m *Mailer) Send(to, subject, body string) error {
	if m.cfg.APIKey == "" {
		return ErrNotConfigured
	}

	payload, err := json.Marshal(sendRequest{
		From:    m.cfg.From,
		To:      []string{to},
		Subject: subject,
		Text:    body,
	})
	if err != nil {
		return fmt.Errorf("mailer: encode request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, sendEndpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("mailer: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("mailer: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailer: resend api error (status %d): %s", resp.StatusCode, respBody)
	}

	return nil
}
