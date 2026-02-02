package openrouter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client wraps OpenRouter API config
type Client struct {
	APIKey       string
	Endpoint     string
	Model        string
	Timeout      time.Duration
	SystemPrompt string
}

// New creates a new OpenRouter client.
func New(apiKey, endpoint, model string, timeout time.Duration, systemPrompt string) *Client {
	return &Client{APIKey: apiKey, Endpoint: endpoint, Model: model, Timeout: timeout, SystemPrompt: systemPrompt}
}

// Ask sends a prompt to OpenRouter and returns the response text.
func (c *Client) Ask(prompt string) (string, error) {
	fmt.Printf("[openrouter] Calling model %q with system prompt: %q\n", c.Model, c.SystemPrompt)

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	messages := []map[string]interface{}{
		{"role": "user", "content": prompt},
	}
	if c.SystemPrompt != "" {
		messages = []map[string]interface{}{
			{"role": "system", "content": c.SystemPrompt},
			{"role": "user", "content": prompt},
		}
	}

	reqBody := map[string]interface{}{
		"model":    c.Model,
		"messages": messages,
	}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", c.Endpoint, strings.NewReader(string(b)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openrouter error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse OpenAI-compatible response format
	var parsed map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &parsed); err == nil {
		// Standard OpenAI format: choices[0].message.content
		if choices, ok := parsed["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						return strings.TrimSpace(content), nil
					}
				}
			}
		}
	}

	// If parsing didn't find the expected format, return the raw response body
	if len(bodyBytes) == 0 {
		return "", errors.New("no response from openrouter")
	}
	return strings.TrimSpace(string(bodyBytes)), nil
}
