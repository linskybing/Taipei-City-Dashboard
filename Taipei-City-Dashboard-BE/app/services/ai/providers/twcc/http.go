package twcc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

func (m *TWCC) doRequest(ctx context.Context, body []byte, isStreaming bool) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s/models/conversation", m.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", m.APIKey)

	client := m.HTTPClient
	if isStreaming {
		client = &http.Client{Timeout: 0}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to TWCC: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("TWCC API returned error status %d: %s", resp.StatusCode, string(raw))
	}
	return resp, nil
}
