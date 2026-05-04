package controllers

import (
	"TaipeiCityDashboardBE/app/services/ai"
	"TaipeiCityDashboardBE/app/services/ai/assistant"
	"TaipeiCityDashboardBE/app/util"
	"TaipeiCityDashboardBE/logs"
	"context"
	"errors"
	"fmt"
	"html"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"
)

// ChatWithTWCC is the controller for POST /api/v1/ai/chat/twai
func ChatWithTWCC(c *gin.Context) {
	var input AIChatInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     "error",
			"error_code": "INVALID_REQUEST",
			"message":    err.Error(),
		})
		return
	}

	assistantContext, err := assistant.NewContext(input.Theme, input.City, input.Audience, input.DashboardIndex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     "error",
			"error_code": "INVALID_ASSISTANT_CONTEXT",
			"message":    err.Error(),
		})
		return
	}

	// 1. Prepare AI Request Context
	_, accountID, _, _, _ := util.GetUserInfoFromContext(c)
	userID := fmt.Sprintf("%d", accountID)
	session, err := ensureOrCreateAIChatSession(c.Request.Context(), html.EscapeString(input.SessionID), userID, "")
	if err != nil {
		if errors.Is(err, ai.ErrAIChatSessionDeleted) {
			c.JSON(http.StatusGone, gin.H{
				"status":     "error",
				"error_code": "AI_SESSION_DELETED",
				"message":    err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     "error",
			"error_code": "AI_SESSION_ERROR",
			"message":    err.Error(),
		})
		return
	}

	// 2. Prepare AI Request
	req := ai.AIChatRequest{
		SessionID: session.SessionID,
		UserID:    userID,
		IPAddress: c.ClientIP(),
		Messages:  assistant.ApplyContext(input.ToServiceMessages(), assistantContext),
		Params:    assistantContext.Metadata(),
	}

	// 3. Prepare Dynamic Options
	options := input.ToGenerationOptions()
	options = append(options, assistant.ToolOptions()...)

	// 4. Handle Streaming Response
	if input.Stream {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Connection", "keep-alive")

		// Add Streaming Callback
		options = append(options, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			if string(chunk) == ": heartbeat\n\n" {
				return nil
			}
			_, err := c.Writer.Write(chunk)
			if err != nil {
				return err
			}
			c.Writer.Flush()
			return nil
		}))

		logEntry, err := executeAIChatWithTWCC(c.Request.Context(), req, options...)
		if err != nil {
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":     "error",
					"error_code": "AI_SERVICE_STREAM_ERROR",
					"message":    err.Error(),
				})
			}
			return
		}
		if err := recordAIChatSessionActivity(c.Request.Context(), session.SessionID, userID, logEntry.CreatedAt); err != nil {
			logs.FError("AI session activity update error: %v", err)
		}
		return
	}

	// 5. Standard Non-Streaming Response
	logEntry, err := executeAIChatWithTWCC(c.Request.Context(), req, options...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":     "error",
			"error_code": "AI_SERVICE_ERROR",
			"message":    err.Error(),
		})
		return
	}
	if err := recordAIChatSessionActivity(c.Request.Context(), session.SessionID, userID, logEntry.CreatedAt); err != nil {
		logs.FError("AI session activity update error: %v", err)
	}
	extras := assistant.ResponseExtrasFromMetadata(logEntry.Metadata)
	auditRef := fmt.Sprintf("ai_chatlog:%d", logEntry.ID)
	visualizationRefs := assistant.BuildVisualizationRefs(extras, auditRef)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"session": logEntry.SessionID,
			"content": logEntry.Answer,
			"usage": gin.H{
				"input_tokens":  logEntry.InputTokens,
				"output_tokens": logEntry.OutputTokens,
				"total_tokens":  logEntry.TotalTokens,
			},
			"tool_used":           logEntry.ToolUsed,
			"latency_ms":          logEntry.LatencyMS,
			"model":               logEntry.Model,
			"provider":            logEntry.Provider,
			"sources":             extras.Sources,
			"related_components":  extras.RelatedComponents,
			"recommended_actions": extras.RecommendedActions,
			"confidence_notes":    extras.ConfidenceNotes,
			"guardrails":          extras.Guardrails,
			"analysis_cards":      extras.AnalysisCards,
			"visualization_refs":  visualizationRefs,
			"audit_ref":           auditRef,
		},
	})
}
