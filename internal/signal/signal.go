package signal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type SignalClient struct {
	APIURL string
	Number string
}

func NewSignalClient(apiURL, number string) *SignalClient {
	return &SignalClient{
		APIURL: apiURL,
		Number: number,
	}
}

// ReceiveEvents fetches new events from the Signal REST API for the given number.
func (c *SignalClient) ReceiveEvents() ([]Envelope, error) {
	fmt.Printf("[signal] Fetching events for %s from %s\n", c.Number, c.APIURL)
	url := fmt.Sprintf("%s/v1/receive/%s", strings.TrimRight(c.APIURL, "/"), c.Number)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("receive returned %d: %s", resp.StatusCode, string(body))
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	if len(bodyBytes) == 0 {
		return nil, nil
	}
	// Unmarshal as array of EnvelopeWrapper
	var wrappers []EnvelopeWrapper
	if err := json.Unmarshal(bodyBytes, &wrappers); err != nil {
		// Try single object
		var single EnvelopeWrapper
		if err2 := json.Unmarshal(bodyBytes, &single); err2 == nil {
			wrappers = append(wrappers, single)
		} else {
			return nil, fmt.Errorf("decode receive response: %v", err)
		}
	}
	log.Printf("[signal] Received events for %s: %s\n", c.Number, string(bodyBytes))
	// Extract Envelopes
	var events []Envelope
	for _, w := range wrappers {
		events = append(events, w.Envelope)
	}
	return events, nil
}

// GetGroupPublicID fetches the public group ID for a given internal group ID
func (c *SignalClient) GetGroupPublicID(internalGroupID string) (string, error) {
	fmt.Printf("[signal] Looking up public group ID for internal ID: %s\n", internalGroupID)
	url := fmt.Sprintf("%s/v1/groups/%s", strings.TrimRight(c.APIURL, "/"), c.Number)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groups returned %d: %s", resp.StatusCode, string(body))
	}
	var groups []struct {
		ID         string   `json:"id"`
		InternalID string   `json:"internal_id"`
		Name       string   `json:"name"`
		Members    []string `json:"members"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return "", fmt.Errorf("failed to decode groups: %w", err)
	}
	for _, g := range groups {
		if g.InternalID == internalGroupID && g.ID != "" {
			return g.ID, nil
		}
	}
	return "", fmt.Errorf("public group id not found for internal id: %s", internalGroupID)
}

// SendMessage posts a message to /v2/send to the specified recipient
func (c *SignalClient) SendMessage(to, message string) error {
	fmt.Printf("[signal] Sending message to %s: %q\n", to, message)
	payload := map[string]interface{}{
		"message": message,
		"number":  c.Number,
	}
	payload["recipients"] = []string{to}
	b, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/v2/send", strings.TrimRight(c.APIURL, "/"))
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("send non-2xx: %d - %s", resp.StatusCode, string(body))
	}
	return nil
}
