package models

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	AIChatSessionStatusActive  = "active"
	AIChatSessionStatusDeleted = "deleted"
)

type AIChatSession struct {
	SessionID      string     `gorm:"primaryKey;type:varchar(100)" json:"session"`
	UserID         string     `gorm:"type:varchar(100);not null;index:idx_ai_chat_sessions_user_status,priority:1;index:idx_ai_chat_sessions_user_activity,priority:1" json:"user_id"`
	Title          string     `gorm:"type:varchar(120);not null" json:"title"`
	Status         string     `gorm:"type:varchar(20);not null;default:'active';index:idx_ai_chat_sessions_user_status,priority:2" json:"status"`
	LastActivityAt time.Time  `gorm:"not null;default:now();index:idx_ai_chat_sessions_user_activity,priority:2" json:"last_activity_at"`
	CreatedAt      time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"type:timestamp with time zone" json:"deleted_at,omitempty"`
}

func (AIChatSession) TableName() string {
	return "ai_chat_sessions"
}

func CreateAIChatSession(ctx context.Context, session *AIChatSession) error {
	if session == nil {
		return fmt.Errorf("ai chat session is required")
	}
	if DBManager == nil {
		return fmt.Errorf("manager db unavailable")
	}
	session.SessionID = strings.TrimSpace(session.SessionID)
	session.UserID = strings.TrimSpace(session.UserID)
	session.Title = strings.TrimSpace(session.Title)
	if session.Status == "" {
		session.Status = AIChatSessionStatusActive
	}
	if session.LastActivityAt.IsZero() {
		session.LastActivityAt = time.Now()
	}
	return DBManager.WithContext(ctx).Create(session).Error
}

func GetAIChatSession(ctx context.Context, sessionID string, userID string) (AIChatSession, bool, error) {
	if DBManager == nil || strings.TrimSpace(sessionID) == "" || strings.TrimSpace(userID) == "" {
		return AIChatSession{}, false, nil
	}
	var session AIChatSession
	err := DBManager.WithContext(ctx).
		Where("session_id = ? AND user_id = ?", strings.TrimSpace(sessionID), strings.TrimSpace(userID)).
		Take(&session).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AIChatSession{}, false, nil
	}
	if err != nil {
		return AIChatSession{}, false, err
	}
	return session, true, nil
}

func ListAIChatSessions(ctx context.Context, userID string) ([]AIChatSession, error) {
	if DBManager == nil || strings.TrimSpace(userID) == "" {
		return nil, nil
	}
	var sessions []AIChatSession
	err := DBManager.WithContext(ctx).
		Where("user_id = ? AND status = ?", strings.TrimSpace(userID), AIChatSessionStatusActive).
		Order("last_activity_at DESC, created_at DESC, session_id DESC").
		Find(&sessions).
		Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func ListAllAIChatSessions(ctx context.Context, userID string) ([]AIChatSession, error) {
	if DBManager == nil || strings.TrimSpace(userID) == "" {
		return nil, nil
	}
	var sessions []AIChatSession
	err := DBManager.WithContext(ctx).
		Where("user_id = ?", strings.TrimSpace(userID)).
		Find(&sessions).
		Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func UpdateAIChatSessionTitle(ctx context.Context, sessionID string, userID string, title string) (AIChatSession, bool, error) {
	session, found, err := GetAIChatSession(ctx, sessionID, userID)
	if err != nil || !found {
		return session, found, err
	}
	updatedAt := time.Now()
	title = strings.TrimSpace(title)
	err = DBManager.WithContext(ctx).
		Model(&AIChatSession{}).
		Where("session_id = ? AND user_id = ?", session.SessionID, session.UserID).
		Updates(map[string]interface{}{"title": title, "updated_at": updatedAt}).
		Error
	if err != nil {
		return AIChatSession{}, true, err
	}
	session.Title = title
	session.UpdatedAt = updatedAt
	return session, true, nil
}

func TouchAIChatSessionActivity(ctx context.Context, sessionID string, userID string, activityAt time.Time) error {
	if DBManager == nil {
		return fmt.Errorf("manager db unavailable")
	}
	if activityAt.IsZero() {
		activityAt = time.Now()
	}
	db := DBManager.WithContext(ctx).
		Model(&AIChatSession{}).
		Where("session_id = ? AND user_id = ? AND status = ?", strings.TrimSpace(sessionID), strings.TrimSpace(userID), AIChatSessionStatusActive).
		Updates(map[string]interface{}{"last_activity_at": activityAt, "updated_at": activityAt})
	if db.Error != nil {
		return db.Error
	}
	if db.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func SoftDeleteAIChatSession(ctx context.Context, sessionID string, userID string, deletedAt time.Time) (AIChatSession, bool, error) {
	session, found, err := GetAIChatSession(ctx, sessionID, userID)
	if err != nil || !found {
		return session, found, err
	}
	if session.Status == AIChatSessionStatusDeleted {
		return session, true, nil
	}
	if deletedAt.IsZero() {
		deletedAt = time.Now()
	}
	err = DBManager.WithContext(ctx).
		Model(&AIChatSession{}).
		Where("session_id = ? AND user_id = ?", session.SessionID, session.UserID).
		Updates(map[string]interface{}{
			"status":     AIChatSessionStatusDeleted,
			"deleted_at": deletedAt,
			"updated_at": deletedAt,
		}).
		Error
	if err != nil {
		return AIChatSession{}, true, err
	}
	session.Status = AIChatSessionStatusDeleted
	session.DeletedAt = &deletedAt
	session.UpdatedAt = deletedAt
	return session, true, nil
}
