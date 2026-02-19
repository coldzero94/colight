package testutil

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/enttest"

	_ "github.com/lib/pq"
)

const defaultTestDSN = "postgres://postgres:password@localhost:5532/colight_test?sslmode=disable"

// TestDSN returns the test database connection string.
// Override with TEST_DATABASE_URL env var (useful for CI).
func TestDSN() string {
	if v := os.Getenv("TEST_DATABASE_URL"); v != "" {
		return v
	}
	return defaultTestDSN
}

// NewTestClient creates an Ent client backed by PostgreSQL test DB.
// It auto-migrates the schema. Caller should close with t.Cleanup.
func NewTestClient(t *testing.T) *ent.Client {
	t.Helper()

	client := enttest.Open(t, "postgres", TestDSN(),
		enttest.WithOptions(ent.Log(t.Log)),
	)

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

// allTables lists all Ent-managed tables.
var allTables = []string{
	"quota_hit_events",
	"ai_call_errors",
	"deletion_requests",
	"admin_audit_logs",
	"system_configs",
	"feedbacks",
	"usage_logs",
	"experience_usages",
	"experience_weapons",
	"experience_tags",
	"coaching_sessions",
	"cover_letter_versions",
	"cover_letters",
	"company_analysis_caches",
	"company_analyses",
	"applications",
	"experiences",
	"question_patterns",
	"prompt_templates",
	"talent_profiles",
	"user_profiles",
	"weapon_categories",
}

// CleanAllTables truncates all tables atomically using TRUNCATE CASCADE.
// This avoids FK ordering issues and race conditions in parallel tests.
func CleanAllTables(_ *ent.Client) {
	ctx := context.Background()
	db, err := sql.Open("postgres", TestDSN())
	if err != nil {
		panic("CleanAllTables: " + err.Error())
	}
	defer db.Close()

	query := "TRUNCATE TABLE " + strings.Join(allTables, ", ") + " CASCADE"
	if _, err := db.ExecContext(ctx, query); err != nil {
		panic("CleanAllTables: " + err.Error())
	}
}
