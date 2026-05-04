package controllers

import (
	"TaipeiCityDashboardBE/app/models"
	serviceai "TaipeiCityDashboardBE/app/services/ai"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"
)

func TestChatWithTWCCRejectsDeletedSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalEnsure := ensureOrCreateAIChatSession
	originalExecute := executeAIChatWithTWCC
	defer func() {
		ensureOrCreateAIChatSession = originalEnsure
		executeAIChatWithTWCC = originalExecute
	}()

	ensureOrCreateAIChatSession = func(_ context.Context, sessionID string, userID string, title string) (models.AIChatSession, error) {
		return models.AIChatSession{}, serviceai.ErrAIChatSessionDeleted
	}
	executeAIChatWithTWCC = func(_ context.Context, _ serviceai.AIChatRequest, _ ...llms.CallOption) (*models.AIChatLog, error) {
		t.Fatal("executeAIChatWithTWCC should not be called for deleted sessions")
		return nil, nil
	}

	body := []byte(`{"messages":[{"role":"user","content":"hello"}],"stream":false}`)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("accountID", 7)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat/twai", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ChatWithTWCC(c)

	if w.Code != http.StatusGone {
		t.Fatalf("status = %d, want %d; body=%s", w.Code, http.StatusGone, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"AI_SESSION_DELETED"`)) {
		t.Fatalf("body missing deleted session error: %s", w.Body.String())
	}
}

func TestGetAIChatSessionBuildsHistoryMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalDetail := getAIChatSessionDetail
	defer func() { getAIChatSessionDetail = originalDetail }()

	getAIChatSessionDetail = func(_ context.Context, sessionID string, userID string) (serviceai.ChatSessionDetail, error) {
		return serviceai.ChatSessionDetail{
			Session: models.AIChatSession{SessionID: sessionID, UserID: userID, Title: "demo", Status: models.AIChatSessionStatusActive, CreatedAt: time.Unix(10, 0), UpdatedAt: time.Unix(10, 0), LastActivityAt: time.Unix(20, 0)},
			Logs:    []models.AIChatLog{{ID: 9, Question: "問題", Answer: "回答", Metadata: `{"recommended_actions":["下一步"]}`}},
		}, nil
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("accountID", 7)
	c.Params = gin.Params{{Key: "session", Value: "session_1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/ai/sessions/session_1", nil)

	GetAIChatSession(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data := payload["data"].(map[string]interface{})
	messages := data["messages"].([]interface{})
	if len(messages) != 2 {
		t.Fatalf("messages len = %d, want 2; body=%s", len(messages), w.Body.String())
	}
	assistantMessage := messages[1].(map[string]interface{})
	if assistantMessage["role"] != "assistant" {
		t.Fatalf("assistant role = %#v", assistantMessage["role"])
	}
	if assistantMessage["audit_ref"] != "ai_chatlog:9" {
		t.Fatalf("assistant audit_ref = %#v", assistantMessage["audit_ref"])
	}
	if len(assistantMessage["recommended_actions"].([]interface{})) != 1 {
		t.Fatalf("assistant recommended_actions = %#v", assistantMessage["recommended_actions"])
	}
}
