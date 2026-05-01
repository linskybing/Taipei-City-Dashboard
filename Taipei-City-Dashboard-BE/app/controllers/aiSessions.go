package controllers

import (
	"TaipeiCityDashboardBE/app/models"
	serviceai "TaipeiCityDashboardBE/app/services/ai"
	"TaipeiCityDashboardBE/app/services/ai/assistant"
	"TaipeiCityDashboardBE/app/util"
	"errors"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateAIChatSession(c *gin.Context) {
	var input AIChatSessionCreateInput
	if err := c.ShouldBindJSON(&input); err != nil && !strings.Contains(err.Error(), "EOF") {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	userID, ok := currentAIChatUserID(c)
	if !ok {
		return
	}
	session, err := createAIChatSessionService(c.Request.Context(), userID, sanitizeAIChatSessionTitle(input.Title))
	if err != nil {
		handleAIChatSessionError(c, err, "AI_SESSION_CREATE_ERROR")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": buildAIChatSessionSummary(session)})
}

func GetAIChatSessions(c *gin.Context) {
	userID, ok := currentAIChatUserID(c)
	if !ok {
		return
	}
	sessions, err := listAIChatSessionsService(c.Request.Context(), userID)
	if err != nil {
		handleAIChatSessionError(c, err, "AI_SESSION_LIST_ERROR")
		return
	}
	data := make([]gin.H, 0, len(sessions))
	for _, session := range sessions {
		data = append(data, buildAIChatSessionSummary(session))
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": data})
}

func GetAIChatSession(c *gin.Context) {
	userID, ok := currentAIChatUserID(c)
	if !ok {
		return
	}
	detail, err := getAIChatSessionDetail(c.Request.Context(), sanitizeAIChatSessionID(c.Param("session")), userID)
	if err != nil {
		handleAIChatSessionError(c, err, "AI_SESSION_DETAIL_ERROR")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"session": buildAIChatSessionSummary(detail.Session), "messages": buildAIChatHistory(detail.Logs)}})
}

func RenameAIChatSession(c *gin.Context) {
	var input AIChatSessionRenameInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error_code": "INVALID_REQUEST", "message": err.Error()})
		return
	}
	userID, ok := currentAIChatUserID(c)
	if !ok {
		return
	}
	session, err := renameAIChatSessionService(c.Request.Context(), sanitizeAIChatSessionID(c.Param("session")), userID, sanitizeAIChatSessionTitle(input.Title))
	if err != nil {
		handleAIChatSessionError(c, err, "AI_SESSION_RENAME_ERROR")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": buildAIChatSessionSummary(session)})
}

func DeleteAIChatSession(c *gin.Context) {
	userID, ok := currentAIChatUserID(c)
	if !ok {
		return
	}
	session, err := deleteAIChatSessionService(c.Request.Context(), sanitizeAIChatSessionID(c.Param("session")), userID)
	if err != nil {
		handleAIChatSessionError(c, err, "AI_SESSION_DELETE_ERROR")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": buildAIChatSessionSummary(session)})
}

func currentAIChatUserID(c *gin.Context) (string, bool) {
	_, accountID, _, _, _ := util.GetUserInfoFromContext(c)
	if accountID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
		return "", false
	}
	return fmt.Sprintf("%d", accountID), true
}

func buildAIChatSessionSummary(session models.AIChatSession) gin.H {
	return gin.H{"session": session.SessionID, "title": session.Title, "status": session.Status, "created_at": session.CreatedAt, "updated_at": session.UpdatedAt, "last_activity_at": session.LastActivityAt, "deleted_at": session.DeletedAt}
}

func buildAIChatHistory(logs []models.AIChatLog) []gin.H {
	messages := make([]gin.H, 0, len(logs)*2)
	for _, logEntry := range logs {
		if question := strings.TrimSpace(logEntry.Question); question != "" {
			messages = append(messages, gin.H{"id": fmt.Sprintf("%d:user", logEntry.ID), "role": "user", "content": question, "created_at": logEntry.CreatedAt})
		}
		if answer := strings.TrimSpace(logEntry.Answer); answer != "" {
			extras := assistant.ResponseExtrasFromMetadata(logEntry.Metadata)
			auditRef := fmt.Sprintf("ai_chatlog:%d", logEntry.ID)
			messages = append(messages, gin.H{"id": fmt.Sprintf("%d:assistant", logEntry.ID), "role": "assistant", "content": answer, "tool_used": logEntry.ToolUsed, "sources": extras.Sources, "related_components": extras.RelatedComponents, "recommended_actions": extras.RecommendedActions, "confidence_notes": extras.ConfidenceNotes, "guardrails": extras.Guardrails, "analysis_cards": extras.AnalysisCards, "visualization_refs": assistant.BuildVisualizationRefs(extras, auditRef), "audit_ref": auditRef, "created_at": logEntry.CreatedAt})
		}
	}
	return messages
}

func sanitizeAIChatSessionID(raw string) string    { return strings.TrimSpace(html.EscapeString(raw)) }
func sanitizeAIChatSessionTitle(raw string) string { return strings.TrimSpace(html.EscapeString(raw)) }

func handleAIChatSessionError(c *gin.Context, err error, code string) {
	switch {
	case errors.Is(err, serviceai.ErrAIChatSessionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error_code": "AI_SESSION_NOT_FOUND", "message": err.Error()})
	case errors.Is(err, serviceai.ErrAIChatSessionDeleted):
		c.JSON(http.StatusGone, gin.H{"status": "error", "error_code": "AI_SESSION_DELETED", "message": err.Error()})
	case strings.Contains(err.Error(), "title is required"):
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error_code": "INVALID_REQUEST", "message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error_code": code, "message": err.Error()})
	}
}
