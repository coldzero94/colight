package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	entaicallerror "github.com/coby/colight/apps/backend/ent/aicallerror"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// AICallParams carries all fields needed to record an AI call in usage_logs.
type AICallParams struct {
	UserID       uuid.UUID
	Feature      string
	Model        string   // e.g. "gemini-2.0-flash"
	Provider     string   // inferred from Model if empty
	InputTokens  int
	OutputTokens int
	LatencyMs    int      // 0 if not measured
	CostKRW      float64  // 0 if unknown
	Err          error    // non-nil → status=error
}

// LogAICall records an AI call in usage_logs. On error, it also creates an
// ai_call_errors record. Rate limit errors additionally create a quota_hit_events
// record for persistent monitoring.
//
// LogAICall is a best-effort operation: errors are logged but not returned so
// they don't interrupt the caller's main flow.
func LogAICall(ctx context.Context, db *ent.Client, p AICallParams) error {
	provider := p.Provider
	if provider == "" {
		provider = inferProvider(p.Model)
	}

	status := "success"
	if p.Err != nil {
		status = "error"
	}

	builder := db.UsageLog.Create().
		SetUserID(p.UserID).
		SetFeature(p.Feature).
		SetStatus(status).
		SetInputTokens(p.InputTokens).
		SetOutputTokens(p.OutputTokens).
		SetTotalTokens(p.InputTokens + p.OutputTokens)

	if provider != "" {
		builder = builder.SetProvider(provider)
	}
	if p.Model != "" {
		builder = builder.SetModel(p.Model)
	}
	if p.LatencyMs > 0 {
		builder = builder.SetLatencyMs(p.LatencyMs)
	}
	if p.CostKRW > 0 {
		builder = builder.SetEstimatedCostKrw(p.CostKRW)
	}

	usageLog, err := builder.Save(ctx)
	if err != nil {
		slog.Error("LogAICall: failed to save usage_log", "error", err)
		return err
	}

	if p.Err == nil {
		return nil
	}

	// Classify and persist error
	errorType := ai.ClassifyError(p.Err)
	errRecord, createErr := db.AICallError.Create().
		SetUsageLogID(usageLog.ID).
		SetErrorType(entaicallerror.ErrorType(errorType)).
		SetErrorMessage(p.Err.Error()).
		Save(ctx)
	if createErr != nil {
		slog.Error("LogAICall: failed to save ai_call_error",
			"usage_log_id", usageLog.ID, "error", createErr)
	}

	// Rate limit events also go to quota_hit_events for dedicated monitoring
	if errorType == "rate_limit" && errRecord != nil {
		qb := db.QuotaHitEvent.Create().
			SetProvider(provider).
			SetModel(p.Model).
			SetErrorMessage(p.Err.Error()).
			SetUsageLogID(usageLog.ID).
			SetCreatedAt(time.Now())
		if p.Feature != "" {
			qb = qb.SetFeature(p.Feature)
		}
		if _, qErr := qb.Save(ctx); qErr != nil {
			slog.Error("LogAICall: failed to save quota_hit_event", "error", qErr)
		}
	}

	return nil
}

// inferProvider derives the AI provider name from a model name.
func inferProvider(model string) string {
	m := strings.ToLower(model)
	if strings.HasPrefix(m, "gemini") {
		return "gemini"
	}
	if strings.HasPrefix(m, "llama") || strings.HasPrefix(m, "groq") ||
		strings.HasPrefix(m, "mixtral") || strings.HasPrefix(m, "deepseek") {
		return "groq"
	}
	return ""
}
