package routes

import (
	"TaipeiCityDashboardBE/app/controllers"
	"TaipeiCityDashboardBE/app/middleware"
	"TaipeiCityDashboardBE/global"
)

func configureAIRoutes() {
	aiRoutes := RouterGroup.Group("/ai")
	aiRoutes.Use(middleware.LimitAPIRequests(global.AIChatLimitAPIRequestsTimes, global.LimitRequestsDuration))
	aiRoutes.Use(middleware.LimitTotalRequests(global.ComponentLimitTotalRequestsTimes, global.LimitRequestsDuration))
	aiRoutes.Use(middleware.IsLoggedIn())
	{
		aiRoutes.POST("/chat/twai", controllers.ChatWithTWCC)
	}
}
