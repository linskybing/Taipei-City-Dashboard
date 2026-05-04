package aitokenconsuming_test

import (
	"os"
	"strings"
	"testing"

	"TaipeiCityDashboardBE/app/controllers"
	"TaipeiCityDashboardBE/app/models"
	"TaipeiCityDashboardBE/global"

	"github.com/gin-gonic/gin"
)

func requireLiveTokenRun(t *testing.T) {
	t.Helper()
	if os.Getenv(liveTokenFlag) != "1" {
		t.Skipf("skipping TWCC token-consuming test; set %s=1 to run", liveTokenFlag)
	}
	if strings.TrimSpace(os.Getenv("TWCC_API_KEY")) == "" || strings.TrimSpace(global.TWCC.ApiKey) == "" {
		t.Fatal("TWCC_API_KEY must be set before running live AI token tests")
	}
}

func assertTWCCConfig(t *testing.T) {
	t.Helper()
	if global.TWCC.ApiUrl != allowedTWCCURL {
		t.Fatalf("TWCC_API_URL = %q, want %q", global.TWCC.ApiUrl, allowedTWCCURL)
	}
	if global.TWCC.Model != allowedTWCCModel {
		t.Fatalf("TWCC_MODEL = %q, want %q", global.TWCC.Model, allowedTWCCModel)
	}
	if strings.Contains(global.TWCC.Model, "32k") {
		t.Fatalf("TWCC_MODEL must not use non-compliant 32k model: %q", global.TWCC.Model)
	}
}

func connectLiveDatabases(t *testing.T) {
	t.Helper()
	managerDBName := createTemporaryDatabase(t, global.PostgresManager, "manager")
	dashboardDBName := createTemporaryDatabase(t, global.PostgresDashboard, "dashboard")
	originalManagerConfig := global.PostgresManager
	originalDashboardConfig := global.PostgresDashboard
	global.PostgresManager.DBName = managerDBName
	global.PostgresDashboard.DBName = dashboardDBName
	t.Cleanup(func() {
		global.PostgresManager = originalManagerConfig
		global.PostgresDashboard = originalDashboardConfig
	})
	t.Cleanup(closeModelDatabases)
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("connect live databases: %v", recovered)
		}
	}()
	models.ConnectToDatabases("MANAGER", "DASHBOARD")
	if models.DBManager == nil {
		t.Fatal("temporary manager DB connection is required for ai_chatlog audit verification")
	}
	if models.DBDashboard == nil {
		t.Fatal("temporary dashboard DB connection is required for assistant tool integration")
	}
	models.MigrateManagerSchema()
}

func newLiveRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST(chatRoute, func(c *gin.Context) {
		c.Set("loginType", "Email")
		c.Set("accountID", 1)
		c.Set("isAdmin", false)
		c.Set("permissions", []models.Permission{})
		controllers.ChatWithTWCC(c)
	})
	return router
}
