package aitokenconsuming_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"TaipeiCityDashboardBE/app/models"
	"TaipeiCityDashboardBE/global"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

const temporaryDBPrefix = "tcd_ai_"

func createTemporaryDatabase(t *testing.T, cfg global.DatabaseConfig, role string) string {
	t.Helper()
	dbName := fmt.Sprintf("%s%s_test_%d", temporaryDBPrefix, role, time.Now().UTC().UnixNano())
	if !isSafeTemporaryDBName(dbName) {
		t.Fatalf("unsafe temporary database name: %s", dbName)
	}

	adminDB := openMaintenanceDB(t, cfg)
	defer adminDB.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := adminDB.ExecContext(ctx, "CREATE DATABASE "+pq.QuoteIdentifier(dbName)); err != nil {
		t.Fatalf("create temporary %s database: %v", role, err)
	}
	t.Logf("created temporary %s database %s", role, dbName)
	t.Cleanup(func() {
		dropTemporaryDatabase(t, cfg, dbName)
	})
	return dbName
}

func dropTemporaryDatabase(t *testing.T, cfg global.DatabaseConfig, dbName string) {
	t.Helper()
	if !strings.HasPrefix(dbName, temporaryDBPrefix) || !isSafeTemporaryDBName(dbName) {
		t.Fatalf("refusing to drop non-temporary database: %s", dbName)
	}
	adminDB := openMaintenanceDB(t, cfg)
	defer adminDB.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, _ = adminDB.ExecContext(ctx,
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1",
		dbName,
	)
	if _, err := adminDB.ExecContext(ctx, "DROP DATABASE IF EXISTS "+pq.QuoteIdentifier(dbName)); err != nil {
		t.Errorf("drop temporary database %s: %v", dbName, err)
		return
	}
	t.Logf("dropped temporary database %s", dbName)
}

func openMaintenanceDB(t *testing.T, cfg global.DatabaseConfig) *sql.DB {
	t.Helper()
	var lastErr error
	for _, dbName := range maintenanceDBCandidates(cfg.DBName) {
		db, err := sql.Open("postgres", postgresURL(cfg, dbName))
		if err != nil {
			lastErr = err
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(ctx)
		cancel()
		if err == nil {
			return db
		}
		db.Close()
		lastErr = err
	}
	t.Fatalf("connect PostgreSQL maintenance database: %v", lastErr)
	return nil
}

func maintenanceDBCandidates(configuredDB string) []string {
	values := []string{
		os.Getenv("TEST_POSTGRES_MAINTENANCE_DB"),
		"postgres",
		configuredDB,
	}
	candidates := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		candidates = append(candidates, value)
	}
	return candidates
}

func postgresURL(cfg global.DatabaseConfig, dbName string) string {
	port := cfg.Port
	if strings.TrimSpace(port) == "" {
		port = "5432"
	}
	sslMode := cfg.SSLMode
	if strings.TrimSpace(sslMode) == "" {
		sslMode = "disable"
	}
	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, port),
		Path:   dbName,
	}
	query := dsn.Query()
	query.Set("sslmode", sslMode)
	dsn.RawQuery = query.Encode()
	return dsn.String()
}

func isSafeTemporaryDBName(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '_' {
			continue
		}
		return false
	}
	return true
}

func closeModelDatabases() {
	closeGormDB(models.DBDashboard)
	closeGormDB(models.DBManager)
	models.DBDashboard = nil
	models.DBManager = nil
}

func closeGormDB(db *gorm.DB) {
	if db == nil {
		return
	}
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
}
