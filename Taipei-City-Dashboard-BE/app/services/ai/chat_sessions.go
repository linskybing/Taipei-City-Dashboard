package ai

import (
	"TaipeiCityDashboardBE/app/models"
	"TaipeiCityDashboardBE/app/util"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	createAIChatSessionRecord        = models.CreateAIChatSession
	getAIChatSessionRecord           = models.GetAIChatSession
	listAllAIChatSessionRecords      = models.ListAllAIChatSessions
	listAIChatSessionRecords         = models.ListAIChatSessions
	updateAIChatSessionRecordTitle   = models.UpdateAIChatSessionTitle
	softDeleteAIChatSessionRecord    = models.SoftDeleteAIChatSession
	touchAIChatSessionRecordActivity = models.TouchAIChatSessionActivity
	listAIChatLogRecordsBySession    = models.ListAIChatLogsBySession
	listAIChatSessionSnapshots       = models.ListAIChatSessionSnapshots
	generateAIChatSessionID          = func() string { return "session_" + util.GenerateRandomString(10) }
	nowAIChatSession                 = time.Now
)

var (
	ErrAIChatSessionNotFound = errors.New("ai chat session not found")
	ErrAIChatSessionDeleted  = errors.New("ai chat session deleted")
)

type ChatSessionDetail struct {
	Session models.AIChatSession
	Logs    []models.AIChatLog
}

func EnsureOrCreateAIChatSession(ctx context.Context, sessionID string, userID string, title string) (models.AIChatSession, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return models.AIChatSession{}, fmt.Errorf("user id is required")
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		sessionID = generateAIChatSessionID()
	}
	session, found, err := getAIChatSessionRecord(ctx, sessionID, userID)
	if err != nil {
		return models.AIChatSession{}, fmt.Errorf("load ai chat session: %w", err)
	}
	if found {
		if session.Status == models.AIChatSessionStatusDeleted {
			return models.AIChatSession{}, ErrAIChatSessionDeleted
		}
		return session, nil
	}
	created := newAIChatSession(sessionID, userID, title)
	if err := createAIChatSessionRecord(ctx, &created); err != nil {
		retry, retryFound, retryErr := getAIChatSessionRecord(ctx, sessionID, userID)
		if retryErr == nil && retryFound {
			if retry.Status == models.AIChatSessionStatusDeleted {
				return models.AIChatSession{}, ErrAIChatSessionDeleted
			}
			return retry, nil
		}
		return models.AIChatSession{}, fmt.Errorf("create ai chat session: %w", err)
	}
	return created, nil
}

func CreateAIChatSession(ctx context.Context, userID string, title string) (models.AIChatSession, error) {
	return EnsureOrCreateAIChatSession(ctx, "", userID, title)
}

func ListAIChatSessions(ctx context.Context, userID string) ([]models.AIChatSession, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if err := backfillAIChatSessions(ctx, userID); err != nil {
		return nil, err
	}
	return listAIChatSessionRecords(ctx, userID)
}

func GetAIChatSessionDetail(ctx context.Context, sessionID string, userID string) (ChatSessionDetail, error) {
	session, err := loadActiveAIChatSession(ctx, sessionID, userID)
	if err != nil {
		return ChatSessionDetail{}, err
	}
	logs, err := listAIChatLogRecordsBySession(ctx, session.SessionID, session.UserID)
	if err != nil {
		return ChatSessionDetail{}, fmt.Errorf("list ai chat logs: %w", err)
	}
	return ChatSessionDetail{Session: session, Logs: logs}, nil
}

func RenameAIChatSession(ctx context.Context, sessionID string, userID string, title string) (models.AIChatSession, error) {
	session, err := loadActiveAIChatSession(ctx, sessionID, userID)
	if err != nil {
		return models.AIChatSession{}, err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return models.AIChatSession{}, fmt.Errorf("title is required")
	}
	updated, _, err := updateAIChatSessionRecordTitle(ctx, session.SessionID, session.UserID, title)
	if err != nil {
		return models.AIChatSession{}, fmt.Errorf("rename ai chat session: %w", err)
	}
	return updated, nil
}

func DeleteAIChatSession(ctx context.Context, sessionID string, userID string) (models.AIChatSession, error) {
	session, err := loadActiveAIChatSession(ctx, sessionID, userID)
	if err != nil {
		return models.AIChatSession{}, err
	}
	deleted, _, err := softDeleteAIChatSessionRecord(ctx, session.SessionID, session.UserID, nowAIChatSession())
	if err != nil {
		return models.AIChatSession{}, fmt.Errorf("delete ai chat session: %w", err)
	}
	return deleted, nil
}

func RecordAIChatSessionActivity(ctx context.Context, sessionID string, userID string, activityAt time.Time) error {
	if strings.TrimSpace(sessionID) == "" || strings.TrimSpace(userID) == "" {
		return nil
	}
	if activityAt.IsZero() {
		activityAt = nowAIChatSession()
	}
	err := touchAIChatSessionRecordActivity(ctx, sessionID, userID, activityAt)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrAIChatSessionNotFound
	}
	if err != nil {
		return fmt.Errorf("touch ai chat session: %w", err)
	}
	return nil
}

func loadActiveAIChatSession(ctx context.Context, sessionID string, userID string) (models.AIChatSession, error) {
	sessionID = strings.TrimSpace(sessionID)
	userID = strings.TrimSpace(userID)
	session, found, err := getAIChatSessionRecord(ctx, sessionID, userID)
	if err != nil {
		return models.AIChatSession{}, fmt.Errorf("load ai chat session: %w", err)
	}
	if !found {
		return models.AIChatSession{}, ErrAIChatSessionNotFound
	}
	if session.Status == models.AIChatSessionStatusDeleted {
		return models.AIChatSession{}, ErrAIChatSessionDeleted
	}
	return session, nil
}

func newAIChatSession(sessionID string, userID string, title string) models.AIChatSession {
	now := nowAIChatSession()
	return models.AIChatSession{
		SessionID:      sessionID,
		UserID:         userID,
		Title:          normalizeAIChatSessionTitle(title, now),
		Status:         models.AIChatSessionStatusActive,
		LastActivityAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func normalizeAIChatSessionTitle(title string, now time.Time) string {
	title = strings.TrimSpace(title)
	if title != "" {
		return title
	}
	return "新對話 " + now.Format("2006-01-02 15:04")
}

func backfillAIChatSessions(ctx context.Context, userID string) error {
	existing, err := listAllAIChatSessionRecords(ctx, userID)
	if err != nil {
		return fmt.Errorf("list ai chat sessions: %w", err)
	}
	existingByID := make(map[string]bool, len(existing))
	for _, session := range existing {
		existingByID[session.SessionID] = true
	}
	snapshots, err := listAIChatSessionSnapshots(ctx, userID)
	if err != nil {
		return fmt.Errorf("list ai chat session snapshots: %w", err)
	}
	for _, snapshot := range snapshots {
		if existingByID[snapshot.SessionID] {
			continue
		}
		created := models.AIChatSession{
			SessionID:      snapshot.SessionID,
			UserID:         userID,
			Title:          "歷史對話 " + snapshot.LastActivityAt.Format("2006-01-02 15:04"),
			Status:         models.AIChatSessionStatusActive,
			LastActivityAt: snapshot.LastActivityAt,
			CreatedAt:      snapshot.FirstActivityAt,
			UpdatedAt:      snapshot.LastActivityAt,
		}
		if err := createAIChatSessionRecord(ctx, &created); err != nil {
			if retry, found, retryErr := getAIChatSessionRecord(ctx, snapshot.SessionID, userID); retryErr == nil && found {
				existingByID[retry.SessionID] = true
				continue
			}
			return fmt.Errorf("backfill ai chat session %s: %w", snapshot.SessionID, err)
		}
		existingByID[created.SessionID] = true
	}
	return nil
}
