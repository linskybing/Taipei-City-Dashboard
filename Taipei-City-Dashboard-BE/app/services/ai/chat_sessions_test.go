package ai

import (
	"TaipeiCityDashboardBE/app/models"
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestEnsureOrCreateAIChatSessionCreatesDefaultSession(t *testing.T) {
	originalGet := getAIChatSessionRecord
	originalCreate := createAIChatSessionRecord
	originalGenerate := generateAIChatSessionID
	originalListAll := listAllAIChatSessionRecords
	originalSnapshots := listAIChatSessionSnapshots
	originalNow := nowAIChatSession
	defer func() {
		getAIChatSessionRecord = originalGet
		createAIChatSessionRecord = originalCreate
		generateAIChatSessionID = originalGenerate
		listAllAIChatSessionRecords = originalListAll
		listAIChatSessionSnapshots = originalSnapshots
		nowAIChatSession = originalNow
	}()

	fixedNow := time.Date(2026, time.May, 1, 10, 30, 0, 0, time.UTC)
	var created models.AIChatSession
	getAIChatSessionRecord = func(context.Context, string, string) (models.AIChatSession, bool, error) {
		return models.AIChatSession{}, false, nil
	}
	createAIChatSessionRecord = func(_ context.Context, session *models.AIChatSession) error {
		created = *session
		return nil
	}
	listAllAIChatSessionRecords = func(context.Context, string) ([]models.AIChatSession, error) { return nil, nil }
	listAIChatSessionSnapshots = func(context.Context, string) ([]models.AIChatSessionSnapshot, error) { return nil, nil }
	generateAIChatSessionID = func() string { return "session_fixed" }
	nowAIChatSession = func() time.Time { return fixedNow }

	session, err := EnsureOrCreateAIChatSession(context.Background(), "", "7", "")
	if err != nil {
		t.Fatalf("EnsureOrCreateAIChatSession returned error: %v", err)
	}
	if session.SessionID != "session_fixed" || created.SessionID != "session_fixed" {
		t.Fatalf("session id = %q, created = %q", session.SessionID, created.SessionID)
	}
	if session.Title != "新對話 2026-05-01 10:30" {
		t.Fatalf("title = %q", session.Title)
	}
	if session.Status != models.AIChatSessionStatusActive {
		t.Fatalf("status = %q", session.Status)
	}
	if !session.LastActivityAt.Equal(fixedNow) {
		t.Fatalf("last activity = %v, want %v", session.LastActivityAt, fixedNow)
	}
}

func TestEnsureOrCreateAIChatSessionRejectsDeletedSession(t *testing.T) {
	originalGet := getAIChatSessionRecord
	defer func() { getAIChatSessionRecord = originalGet }()

	getAIChatSessionRecord = func(context.Context, string, string) (models.AIChatSession, bool, error) {
		return models.AIChatSession{SessionID: "session_1", Status: models.AIChatSessionStatusDeleted}, true, nil
	}

	_, err := EnsureOrCreateAIChatSession(context.Background(), "session_1", "7", "")
	if !errors.Is(err, ErrAIChatSessionDeleted) {
		t.Fatalf("error = %v, want ErrAIChatSessionDeleted", err)
	}
}

func TestGetAIChatSessionDetailReturnsLogsForActiveSession(t *testing.T) {
	originalGet := getAIChatSessionRecord
	originalListLogs := listAIChatLogRecordsBySession
	defer func() {
		getAIChatSessionRecord = originalGet
		listAIChatLogRecordsBySession = originalListLogs
	}()

	getAIChatSessionRecord = func(context.Context, string, string) (models.AIChatSession, bool, error) {
		return models.AIChatSession{SessionID: "session_1", UserID: "7", Title: "demo", Status: models.AIChatSessionStatusActive}, true, nil
	}
	listAIChatLogRecordsBySession = func(context.Context, string, string) ([]models.AIChatLog, error) {
		return []models.AIChatLog{{ID: 42, Question: "Q", Answer: "A"}}, nil
	}

	detail, err := GetAIChatSessionDetail(context.Background(), "session_1", "7")
	if err != nil {
		t.Fatalf("GetAIChatSessionDetail returned error: %v", err)
	}
	if detail.Session.SessionID != "session_1" || len(detail.Logs) != 1 || detail.Logs[0].ID != 42 {
		t.Fatalf("unexpected detail: %#v", detail)
	}
}

func TestRecordAIChatSessionActivityMapsNotFound(t *testing.T) {
	originalTouch := touchAIChatSessionRecordActivity
	defer func() { touchAIChatSessionRecordActivity = originalTouch }()

	touchAIChatSessionRecordActivity = func(context.Context, string, string, time.Time) error {
		return gorm.ErrRecordNotFound
	}

	err := RecordAIChatSessionActivity(context.Background(), "session_1", "7", time.Now())
	if !errors.Is(err, ErrAIChatSessionNotFound) {
		t.Fatalf("error = %v, want ErrAIChatSessionNotFound", err)
	}
}

func TestListAIChatSessionsBackfillsLegacySnapshots(t *testing.T) {
	originalListAll := listAllAIChatSessionRecords
	originalSnapshots := listAIChatSessionSnapshots
	originalCreate := createAIChatSessionRecord
	originalList := listAIChatSessionRecords
	defer func() {
		listAllAIChatSessionRecords = originalListAll
		listAIChatSessionSnapshots = originalSnapshots
		createAIChatSessionRecord = originalCreate
		listAIChatSessionRecords = originalList
	}()

	var created []models.AIChatSession
	listAllAIChatSessionRecords = func(context.Context, string) ([]models.AIChatSession, error) {
		return nil, nil
	}
	listAIChatSessionSnapshots = func(context.Context, string) ([]models.AIChatSessionSnapshot, error) {
		return []models.AIChatSessionSnapshot{{
			SessionID:       "legacy_session",
			FirstActivityAt: time.Date(2026, time.May, 1, 9, 0, 0, 0, time.UTC),
			LastActivityAt:  time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC),
		}}, nil
	}
	createAIChatSessionRecord = func(_ context.Context, session *models.AIChatSession) error {
		created = append(created, *session)
		return nil
	}
	listAIChatSessionRecords = func(context.Context, string) ([]models.AIChatSession, error) {
		return created, nil
	}

	sessions, err := ListAIChatSessions(context.Background(), "7")
	if err != nil {
		t.Fatalf("ListAIChatSessions returned error: %v", err)
	}
	if len(sessions) != 1 || sessions[0].SessionID != "legacy_session" {
		t.Fatalf("unexpected sessions: %#v", sessions)
	}
	if sessions[0].Title != "歷史對話 2026-05-01 10:00" {
		t.Fatalf("backfilled title = %q", sessions[0].Title)
	}
}
