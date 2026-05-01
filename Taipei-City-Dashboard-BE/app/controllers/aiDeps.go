package controllers

import "TaipeiCityDashboardBE/app/services/ai"

var (
	executeAIChatWithTWCC       = ai.ChatWithTWCC
	ensureOrCreateAIChatSession = ai.EnsureOrCreateAIChatSession
	createAIChatSessionService  = ai.CreateAIChatSession
	listAIChatSessionsService   = ai.ListAIChatSessions
	getAIChatSessionDetail      = ai.GetAIChatSessionDetail
	renameAIChatSessionService  = ai.RenameAIChatSession
	deleteAIChatSessionService  = ai.DeleteAIChatSession
	recordAIChatSessionActivity = ai.RecordAIChatSessionActivity
)
