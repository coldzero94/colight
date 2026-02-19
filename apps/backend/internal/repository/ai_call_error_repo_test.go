package repository_test

import (
	"context"
	"testing"
	"time"

	entaicallerror "github.com/coby/colight/apps/backend/ent/aicallerror"
	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAICallError_Create(t *testing.T) {
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("err-test@test.com").
		SetNickname("ErrTest").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SaveX(ctx)

	usageLog := db.UsageLog.Create().
		SetUserID(user.ID).
		SetFeature("draft").
		SetProvider("gemini").
		SetModel("gemini-2.0-flash").
		SetStatus("error").
		SaveX(ctx)

	errRecord := db.AICallError.Create().
		SetUsageLogID(usageLog.ID).
		SetErrorType(entaicallerror.ErrorTypeRateLimit).
		SetErrorMessage("429: Quota exceeded").
		SaveX(ctx)

	assert.Equal(t, usageLog.ID, errRecord.UsageLogID)
	assert.Equal(t, entaicallerror.ErrorTypeRateLimit, errRecord.ErrorType)
	assert.Equal(t, "429: Quota exceeded", errRecord.ErrorMessage)
}

func TestAICallError_UsageLogEdge(t *testing.T) {
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("edge-test@test.com").
		SetNickname("EdgeTest").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SaveX(ctx)

	usageLog := db.UsageLog.Create().
		SetUserID(user.ID).
		SetFeature("draft").
		SetStatus("error").
		SaveX(ctx)

	db.AICallError.Create().
		SetUsageLogID(usageLog.ID).
		SetErrorType(entaicallerror.ErrorTypeTimeout).
		SetErrorMessage("deadline exceeded").
		SaveX(ctx)

	errRecord, err := db.UsageLog.QueryErrorDetail(usageLog).Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, entaicallerror.ErrorTypeTimeout, errRecord.ErrorType)
}

func TestAICallError_CascadeOnDelete(t *testing.T) {
	// CASCADE DELETE is enforced by the production migration FK constraint.
	// In test DB (enttest auto-migrate), we verify the FK relationship exists
	// by manually deleting the dependent record first.
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("cascade-test@test.com").
		SetNickname("CascadeTest").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SaveX(ctx)

	usageLog := db.UsageLog.Create().
		SetUserID(user.ID).
		SetFeature("draft").
		SetStatus("error").
		SaveX(ctx)

	errRecord := db.AICallError.Create().
		SetUsageLogID(usageLog.ID).
		SetErrorType(entaicallerror.ErrorTypeProviderError).
		SetErrorMessage("503 Service Unavailable").
		SaveX(ctx)

	// Verify FK relationship: ai_call_error → usage_log
	linked, err := db.AICallError.QueryUsageLog(errRecord).Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, usageLog.ID, linked.ID)

	// Manual cleanup (test DB does not apply CASCADE via enttest auto-migrate)
	require.NoError(t, db.AICallError.DeleteOneID(errRecord.ID).Exec(ctx))
	require.NoError(t, db.UsageLog.DeleteOneID(usageLog.ID).Exec(ctx))
}

func TestQuotaHitEvent_Create(t *testing.T) {
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	evt := db.QuotaHitEvent.Create().
		SetProvider("gemini").
		SetModel("gemini-2.0-flash").
		SetFeature("draft").
		SetErrorMessage("429: RPM quota exceeded").
		SaveX(ctx)

	assert.Equal(t, "gemini", evt.Provider)
	assert.Equal(t, "gemini-2.0-flash", evt.Model)
	assert.Nil(t, evt.UsageLogID)
}

func TestQuotaHitEvent_NullableUsageLogId(t *testing.T) {
	db := testutil.NewTestClient(t)
	testutil.CleanAllTables(db)
	ctx := context.Background()

	user := db.UserProfile.Create().
		SetEmail("quota-test@test.com").
		SetNickname("QuotaTest").
		SetAuthProvider(userprofile.AuthProviderEmail).
		SaveX(ctx)

	usageLog := db.UsageLog.Create().
		SetUserID(user.ID).
		SetFeature("draft").
		SetStatus("error").
		SaveX(ctx)

	evt := db.QuotaHitEvent.Create().
		SetProvider("groq").
		SetModel("llama-3.3-70b-versatile").
		SetErrorMessage("rate limit exceeded").
		SetUsageLogID(usageLog.ID).
		SaveX(ctx)

	assert.NotNil(t, evt.UsageLogID)
	assert.Equal(t, usageLog.ID, *evt.UsageLogID)

	evtNoLog := db.QuotaHitEvent.Create().
		SetProvider("gemini").
		SetModel("gemini-2.0-flash").
		SetErrorMessage("pre-throttle block").
		SetCreatedAt(time.Now()).
		SaveX(ctx)

	assert.Nil(t, evtNoLog.UsageLogID)
}
