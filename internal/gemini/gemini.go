package gemini

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

// Client wraps Gemini API config
type Client struct {
	APIKey       string
	Model        string
	Timeout      time.Duration
	SystemPrompt string
}

// New creates a new Gemini client
func New(apiKey, model string, timeout time.Duration, systemPrompt string) *Client {
	return &Client{APIKey: apiKey, Model: model, Timeout: timeout, SystemPrompt: systemPrompt}
}

// Ask sends a prompt to Gemini and returns the response text, prepending the system prompt if set
func (c *Client) Ask(prompt string) (string, error) {
	fmt.Printf("[gemini] Making API call with system prompt: %q\n", c.SystemPrompt)
	fmt.Printf("[gemini] User prompt: %q\n", prompt)
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.Model, c.APIKey)

	fullPrompt := prompt
	if c.SystemPrompt != "" {
		fullPrompt = c.SystemPrompt + "\n" + prompt
	}

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": fullPrompt},
				},
			},
		},
	}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(b)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes := make([]byte, 0)
	bodyBytes, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("gemini error %d: %s", resp.StatusCode, string(bodyBytes))
	}
	var out struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(bodyBytes, &out); err != nil {
		return "", err
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no response from gemini")
	}
	fmt.Printf("[gemini] Generated response: %q\n", out.Candidates[0].Content.Parts[0].Text)
	return strings.TrimSpace(out.Candidates[0].Content.Parts[0].Text), nil
}
