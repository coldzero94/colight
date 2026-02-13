package testutil

import (
	"context"
	"os"
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

// CleanAllTables truncates all tables. Call from TestMain before tests run.
func CleanAllTables(client *ent.Client) {
	ctx := context.Background()
	client.ExperienceUsage.Delete().ExecX(ctx)
	client.ExperienceWeapon.Delete().ExecX(ctx)
	client.ExperienceTag.Delete().ExecX(ctx)
	client.CoachingSession.Delete().ExecX(ctx)
	client.CoverLetterVersion.Delete().ExecX(ctx)
	client.CoverLetter.Delete().ExecX(ctx)
	client.CompanyAnalysisCache.Delete().ExecX(ctx)
	client.CompanyAnalysis.Delete().ExecX(ctx)
	client.Application.Delete().ExecX(ctx)
	client.Experience.Delete().ExecX(ctx)
	client.QuestionPattern.Delete().ExecX(ctx)
	client.PromptTemplate.Delete().ExecX(ctx)
	client.TalentProfile.Delete().ExecX(ctx)
	client.UserProfile.Delete().ExecX(ctx)
	client.WeaponCategory.Delete().ExecX(ctx)
}
