package models

import (
	"context"
	"strings"
	"time"
)

// AIChatLog defines the model for AI chat logs as specified in the system design.
type AIChatLog struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID    string    `gorm:"type:varchar(100);not null;index:idx_ai_chatlog_session" json:"session"`
	UserID       string    `gorm:"type:varchar(100);index:idx_ai_chatlog_user" json:"user_id"`
	Provider     string    `gorm:"type:varchar(50);not null;default:'twcc'" json:"provider"`
	Model        string    `gorm:"type:varchar(100)" json:"model"`
	Question     string    `gorm:"type:text;not null" json:"question"`
	Answer       string    `gorm:"type:text" json:"answer"`
	ToolUsed     bool      `gorm:"default:false" json:"tool_used"`
	Tools        string    `gorm:"type:jsonb" json:"tools"` // Stored as JSONB in DB
	InputTokens  int       `gorm:"default:0" json:"input_tokens"`
	OutputTokens int       `gorm:"default:0" json:"output_tokens"`
	TotalTokens  int       `gorm:"default:0" json:"total_tokens"`
	LatencyMS    int       `json:"latency_ms"`
	Status       string    `gorm:"type:varchar(30);not null;default:'success'" json:"status"`
	ErrorCode    string    `gorm:"type:varchar(100)" json:"error_code"`
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	Metadata     string    `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	IPAddress    string    `gorm:"type:varchar(45);not null" json:"ip_address"`
	CreatedAt    time.Time `gorm:"not null;default:now()" json:"created_at"`
}

type AIChatSessionSnapshot struct {
	SessionID       string    `json:"session"`
	FirstActivityAt time.Time `json:"first_activity_at"`
	LastActivityAt  time.Time `json:"last_activity_at"`
}

// TableName overrides the table name used by AIChatLog to `ai_chatlog`
func (AIChatLog) TableName() string {
	return "ai_chatlog"
}

// CreateAIChatLog inserts a new AI chat log into the database.
func CreateAIChatLog(ctx context.Context, log *AIChatLog) error {
	return DBManager.WithContext(ctx).Create(log).Error
}

func GetRecentAIChatLogs(ctx context.Context, sessionID string, userID string, limit int) ([]AIChatLog, error) {
	if DBManager == nil || sessionID == "" || userID == "" || limit <= 0 {
		return nil, nil
	}
	var logs []AIChatLog
	err := DBManager.
		WithContext(ctx).
		Where("session_id = ? AND user_id = ? AND status = ?", sessionID, userID, "success").
		Where("question <> '' OR answer <> ''").
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&logs).
		Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func ListAIChatLogsBySession(ctx context.Context, sessionID string, userID string) ([]AIChatLog, error) {
	if DBManager == nil || strings.TrimSpace(sessionID) == "" || strings.TrimSpace(userID) == "" {
		return nil, nil
	}
	var logs []AIChatLog
	err := DBManager.
		WithContext(ctx).
		Where("session_id = ? AND user_id = ? AND status = ?", strings.TrimSpace(sessionID), strings.TrimSpace(userID), "success").
		Where("question <> '' OR answer <> ''").
		Order("created_at ASC, id ASC").
		Find(&logs).
		Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func ListAIChatSessionSnapshots(ctx context.Context, userID string) ([]AIChatSessionSnapshot, error) {
	if DBManager == nil || strings.TrimSpace(userID) == "" {
		return nil, nil
	}
	var snapshots []AIChatSessionSnapshot
	err := DBManager.
		WithContext(ctx).
		Table("ai_chatlog").
		Select("session_id, MIN(created_at) AS first_activity_at, MAX(created_at) AS last_activity_at").
		Where("user_id = ? AND status = ?", strings.TrimSpace(userID), "success").
		Where("session_id <> ''").
		Group("session_id").
		Scan(&snapshots).
		Error
	if err != nil {
		return nil, err
	}
	return snapshots, nil
}
