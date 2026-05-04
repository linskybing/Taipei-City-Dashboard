package aitokenconsuming_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"TaipeiCityDashboardBE/app/models"
)

const (
	liveTokenFlag    = "RUN_AI_TOKEN_TESTS"
	allowedTWCCURL   = "https://api-ams.twcc.ai/api"
	allowedTWCCModel = "llama3.3-ffm-70b-16k-chat"
	chatRoute        = "/api/v1/ai/chat/twai"
)

type chatResponse struct {
	Status string `json:"status"`
	Data   struct {
		Session string `json:"session"`
		Content string `json:"content"`
		Usage   struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
		Model    string `json:"model"`
		Provider string `json:"provider"`
		AuditRef string `json:"audit_ref"`
	} `json:"data"`
}

func TestLiveAIChatEndpointWritesAuditLog(t *testing.T) {
	requireLiveTokenRun(t)
	assertTWCCConfig(t)
	connectLiveDatabases(t)

	router := newLiveRouter()
	sessionID := fmt.Sprintf("live_ai_token_test_%d", time.Now().UTC().UnixNano())
	payload := map[string]interface{}{
		"session":        sessionID,
		"stream":         false,
		"theme":          "health",
		"city":           "metrotaipei",
		"audience":       "government",
		"max_new_tokens": 120,
		"temperature":    0.1,
		"messages": []map[string]string{{
			"role": "user",
			"content": strings.Join([]string{
				"Reply in Traditional Chinese.",
				"Use the required dashboard assistant sections.",
				"Confirm briefly that this live AI chat integration test can answer.",
			}, " "),
		}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	req := httptest.NewRequest(http.MethodPost, chatRoute, bytes.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", resp.Code, resp.Body.String())
	}
	parsed := decodeChatResponse(t, resp.Body.Bytes())
	assertSuccessfulChatResponse(t, parsed, sessionID)
	assertAuditLog(t, sessionID, parsed)
}

func decodeChatResponse(t *testing.T, body []byte) chatResponse {
	t.Helper()
	assertChatResponseShape(t, body)
	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("decode response: %v; body: %s", err, string(body))
	}
	return parsed
}

func assertSuccessfulChatResponse(t *testing.T, parsed chatResponse, sessionID string) {
	t.Helper()
	if parsed.Status != "success" {
		t.Fatalf("response status = %q, want success", parsed.Status)
	}
	if parsed.Data.Session != sessionID {
		t.Fatalf("session = %q, want %q", parsed.Data.Session, sessionID)
	}
	if strings.TrimSpace(parsed.Data.Content) == "" {
		t.Fatal("response content is empty")
	}
	if parsed.Data.Provider != "twcc" {
		t.Fatalf("provider = %q, want twcc", parsed.Data.Provider)
	}
	if parsed.Data.Model != allowedTWCCModel {
		t.Fatalf("model = %q, want %q", parsed.Data.Model, allowedTWCCModel)
	}
	if parsed.Data.Usage.TotalTokens <= 0 {
		t.Fatalf("total_tokens = %d, want positive token usage", parsed.Data.Usage.TotalTokens)
	}
}

func assertAuditLog(t *testing.T, sessionID string, parsed chatResponse) {
	t.Helper()
	var log models.AIChatLog
	if err := models.DBManager.
		Where("session_id = ?", sessionID).
		Order("id DESC").
		First(&log).Error; err != nil {
		t.Fatalf("find ai_chatlog for session %q: %v", sessionID, err)
	}
	if parsed.Data.AuditRef != fmt.Sprintf("ai_chatlog:%d", log.ID) {
		t.Fatalf("audit_ref = %q, want ai_chatlog:%d", parsed.Data.AuditRef, log.ID)
	}
	if log.Status != "success" {
		t.Fatalf("ai_chatlog status = %q, want success; error=%s", log.Status, log.ErrorMessage)
	}
	if log.Provider != "twcc" || log.Model != allowedTWCCModel {
		t.Fatalf("ai_chatlog provider/model = %q/%q", log.Provider, log.Model)
	}
	if strings.TrimSpace(log.Question) == "" || strings.TrimSpace(log.Answer) == "" {
		t.Fatal("ai_chatlog question and answer must be recorded")
	}
	if log.TotalTokens <= 0 {
		t.Fatalf("ai_chatlog total_tokens = %d, want positive token usage", log.TotalTokens)
	}
	assertMetadataContext(t, log.Metadata)
}

func assertMetadataContext(t *testing.T, raw string) {
	t.Helper()
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		t.Fatalf("decode ai_chatlog metadata: %v; raw=%s", err, raw)
	}
	expected := map[string]string{"theme": "health", "city": "metrotaipei", "audience": "government"}
	for key, want := range expected {
		if got, _ := metadata[key].(string); got != want {
			t.Fatalf("metadata[%s] = %q, want %q", key, got, want)
		}
	}
	if _, ok := metadata["tool_events"]; !ok {
		t.Fatal("metadata missing tool_events")
	}
}
