package service

import (
	"context"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/aicallerror"
	"github.com/coby/colight/apps/backend/ent/feedback"
	"github.com/coby/colight/apps/backend/ent/predicate"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/ent/usagelog"
	"github.com/coby/colight/apps/backend/ent/userprofile"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// AdminAnalyticsService handles usage analytics, dashboard aggregation, and model statistics.
// Reusable by payment service (Phase 9) for usage-based billing.
type AdminAnalyticsService struct {
	db *ent.Client
}

// NewAdminAnalyticsService creates a new admin analytics service.
func NewAdminAnalyticsService(db *ent.Client) *AdminAnalyticsService {
	return &AdminAnalyticsService{db: db}
}

// --- Types ---

// AdminStats contains basic system-level statistics.
type AdminStats struct {
	TotalUsers       int
	EmailAuthUsers   int
	NaverAuthUsers   int
	ActiveUsersToday int
	TotalExperiences int
}

// UsageSummary contains aggregated usage statistics for a period.
type UsageSummary struct {
	TotalCalls   int
	TotalTokens  int
	TotalCostKRW float64
	ErrorRate    float64
	ErrorCount   int
	Days         int
}

// UsageDaily contains daily usage breakdown.
type UsageDaily struct {
	Date        string
	TotalCalls  int
	TotalTokens int
	ErrorCount  int
}

// ProviderCost contains provider-level cost breakdown.
type ProviderCost struct {
	Provider    string
	TotalTokens int
	TotalCost   float64
	CallCount   int
}

// UserUsage contains per-user usage metrics.
type UserUsage struct {
	UserID      uuid.UUID
	TotalTokens int
	TotalCost   float64
	CallCount   int
}

// ModelStatEntry contains per-model statistics with rate limits.
type ModelStatEntry struct {
	Model        string
	Provider     string
	CallCount    int
	SuccessCount int
	ErrorCount   int
	ErrorRate    float64
	TotalTokens  int
	InputTokens  int
	OutputTokens int
	TotalCostKRW float64
	AvgCostKRW   float64
	AvgLatencyMs float64
	LastUsed     time.Time
	RPM          int
	TPM          int
	RPD          int
	ContextSize  int
	Note         string
}

// ModelQuota contains quota usage for a model.
type ModelQuota struct {
	Model      string
	Provider   string
	Limit      int
	Used       int
	Remaining  int
	Percentage float64
	RPM        int
	TPM        int
}

// ModelStatsResult contains model statistics, feature model mappings, and quotas.
type ModelStatsResult struct {
	Models        []ModelStatEntry
	FeatureModels map[string]string
	ModelQuotas   []ModelQuota
}

// DashboardUserStats contains user statistics for the dashboard.
type DashboardUserStats struct {
	TotalUsers       int
	NewUsersToday    int
	ActiveUsersToday int
}

// DashboardAIMetrics contains AI metrics for the dashboard.
type DashboardAIMetrics struct {
	Calls          int
	CallsDeltaPct  float64
	ErrorRate      float64
	ErrorRateDelta float64
	CostKRW        float64
}

// DashboardDailyMetric contains a single day's metrics.
type DashboardDailyMetric struct {
	Date       string
	Calls      int
	ErrorCount int
	CostKRW    float64
}

// DashboardError contains a recent error entry for the dashboard.
type DashboardError struct {
	ID           uuid.UUID
	ErrorType    string
	ErrorMessage string
	CreatedAt    time.Time
	Feature      string
	UserID       uuid.UUID
	InputTokens  int
	OutputTokens int
	Model        string
	Provider     string
}

// DashboardFeedback contains a recent feedback entry for the dashboard.
type DashboardFeedback struct {
	ID             uuid.UUID
	Category       string
	ContentPreview string
	CreatedAt      time.Time
}

// DashboardData contains all aggregated dashboard data.
type DashboardData struct {
	UserStats            DashboardUserStats
	AIMetrics            DashboardAIMetrics
	DailyMetrics         []DashboardDailyMetric
	QuotaAlerts          []ModelQuota
	RecentErrors         []DashboardError
	RecentFeedbacks      []DashboardFeedback
	PendingFeedbackCount int
}

// ErrorListParams contains parameters for listing errors.
type ErrorListParams struct {
	Days      int
	Limit     int
	Offset    int
	ErrorType string
	Provider  string
	Feature   string
}

// ErrorListResult contains paginated error results.
type ErrorListResult struct {
	Errors []*ent.AICallError
	Total  int
}

// --- Methods ---

// GetStats returns system-level statistics.
func (s *AdminAnalyticsService) GetStats(ctx context.Context) (*AdminStats, error) {
	totalUsers, _ := s.db.UserProfile.Query().Count(ctx)
	emailUsers, _ := s.db.UserProfile.Query().Where(userprofile.AuthProviderEQ(userprofile.AuthProviderEmail)).Count(ctx)
	naverUsers, _ := s.db.UserProfile.Query().Where(userprofile.AuthProviderEQ(userprofile.AuthProviderNaver)).Count(ctx)

	today := time.Now().Truncate(24 * time.Hour)
	activeToday, _ := s.db.UserProfile.Query().Where(userprofile.LastLoginAtGTE(today)).Count(ctx)
	totalExperiences, _ := s.db.Experience.Query().Count(ctx)

	return &AdminStats{
		TotalUsers:       totalUsers,
		EmailAuthUsers:   emailUsers,
		NaverAuthUsers:   naverUsers,
		ActiveUsersToday: activeToday,
		TotalExperiences: totalExperiences,
	}, nil
}

// GetUsageSummary returns aggregated usage statistics for the given period.
func (s *AdminAnalyticsService) GetUsageSummary(ctx context.Context, days int) (*UsageSummary, error) {
	since := time.Now().AddDate(0, 0, -days)
	logs, err := s.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(since)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	totalCalls := len(logs)
	var totalTokens int
	var totalCost float64
	var errorCount int
	for _, l := range logs {
		totalTokens += l.TotalTokens
		if l.EstimatedCostKrw != nil {
			totalCost += *l.EstimatedCostKrw
		}
		if l.Status == "error" {
			errorCount++
		}
	}

	var errorRate float64
	if totalCalls > 0 {
		errorRate = float64(errorCount) / float64(totalCalls) * 100
	}

	return &UsageSummary{
		TotalCalls:   totalCalls,
		TotalTokens:  totalTokens,
		TotalCostKRW: totalCost,
		ErrorRate:    errorRate,
		ErrorCount:   errorCount,
		Days:         days,
	}, nil
}

// GetUsageDaily returns daily usage breakdown for the given period.
func (s *AdminAnalyticsService) GetUsageDaily(ctx context.Context, days int) ([]UsageDaily, error) {
	since := time.Now().AddDate(0, 0, -days)
	logs, err := s.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(since)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	dailyMap := make(map[string]*UsageDaily)
	for _, l := range logs {
		date := l.CreatedAt.Format("2006-01-02")
		if _, ok := dailyMap[date]; !ok {
			dailyMap[date] = &UsageDaily{Date: date}
		}
		d := dailyMap[date]
		d.TotalCalls++
		d.TotalTokens += l.TotalTokens
		if l.Status == "error" {
			d.ErrorCount++
		}
	}

	items := make([]UsageDaily, 0, len(dailyMap))
	for _, v := range dailyMap {
		items = append(items, *v)
	}
	return items, nil
}

// GetUsageCosts returns provider-level cost breakdown for the given period.
func (s *AdminAnalyticsService) GetUsageCosts(ctx context.Context, days int) ([]ProviderCost, error) {
	since := time.Now().AddDate(0, 0, -days)
	logs, err := s.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(since)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	providerMap := make(map[string]*ProviderCost)
	for _, l := range logs {
		p := "unknown"
		if l.Provider != nil {
			p = *l.Provider
		}
		if _, ok := providerMap[p]; !ok {
			providerMap[p] = &ProviderCost{Provider: p}
		}
		pc := providerMap[p]
		pc.TotalTokens += l.TotalTokens
		if l.EstimatedCostKrw != nil {
			pc.TotalCost += *l.EstimatedCostKrw
		}
		pc.CallCount++
	}

	items := make([]ProviderCost, 0, len(providerMap))
	for _, pc := range providerMap {
		items = append(items, *pc)
	}
	return items, nil
}

// GetUsageTopUsers returns top users ranked by token usage.
func (s *AdminAnalyticsService) GetUsageTopUsers(ctx context.Context, days, limit int) ([]UserUsage, error) {
	since := time.Now().AddDate(0, 0, -days)
	logs, err := s.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(since)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	userMap := make(map[uuid.UUID]*UserUsage)
	for _, l := range logs {
		if _, ok := userMap[l.UserID]; !ok {
			userMap[l.UserID] = &UserUsage{UserID: l.UserID}
		}
		uu := userMap[l.UserID]
		uu.TotalTokens += l.TotalTokens
		if l.EstimatedCostKrw != nil {
			uu.TotalCost += *l.EstimatedCostKrw
		}
		uu.CallCount++
	}

	// Sort by total tokens descending (bubble sort for small N)
	sorted := make([]UserUsage, 0, len(userMap))
	for _, uu := range userMap {
		sorted = append(sorted, *uu)
	}
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].TotalTokens > sorted[i].TotalTokens {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	if limit > len(sorted) {
		limit = len(sorted)
	}
	return sorted[:limit], nil
}

// GetModelStats returns model-level statistics with quota estimates.
func (s *AdminAnalyticsService) GetModelStats(ctx context.Context, days int) (*ModelStatsResult, error) {
	since := time.Now().AddDate(0, 0, -days)
	logs, err := s.db.UsageLog.Query().
		Where(usagelog.CreatedAtGTE(since)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Aggregate by model
	type modelAgg struct {
		Model        string
		Provider     string
		CallCount    int
		SuccessCount int
		ErrorCount   int
		TotalTokens  int
		InputTokens  int
		OutputTokens int
		TotalCost    float64
		LatencySum   float64
		LastUsed     time.Time
	}

	modelMap := make(map[string]*modelAgg)
	for _, l := range logs {
		model := "unknown"
		if l.Model != nil {
			model = *l.Model
		}
		provider := "unknown"
		if l.Provider != nil {
			provider = *l.Provider
		}
		if _, ok := modelMap[model]; !ok {
			modelMap[model] = &modelAgg{Model: model, Provider: provider}
		}
		ms := modelMap[model]
		ms.CallCount++
		if l.Status == "success" {
			ms.SuccessCount++
		} else {
			ms.ErrorCount++
		}
		ms.TotalTokens += l.TotalTokens
		ms.InputTokens += l.InputTokens
		ms.OutputTokens += l.OutputTokens
		if l.EstimatedCostKrw != nil {
			ms.TotalCost += *l.EstimatedCostKrw
		}
		if l.LatencyMs != nil {
			ms.LatencySum += float64(*l.LatencyMs)
		}
		if l.CreatedAt.After(ms.LastUsed) {
			ms.LastUsed = l.CreatedAt
		}
	}

	// Build result entries
	models := make([]ModelStatEntry, 0, len(modelMap))
	for _, ms := range modelMap {
		avgLatency := 0.0
		if ms.CallCount > 0 {
			avgLatency = ms.LatencySum / float64(ms.CallCount)
		}
		limits := ai.GetModelLimits(ms.Model)

		models = append(models, ModelStatEntry{
			Model:        ms.Model,
			Provider:     ms.Provider,
			CallCount:    ms.CallCount,
			SuccessCount: ms.SuccessCount,
			ErrorCount:   ms.ErrorCount,
			ErrorRate:    float64(ms.ErrorCount) / float64(ms.CallCount) * 100,
			TotalTokens:  ms.TotalTokens,
			InputTokens:  ms.InputTokens,
			OutputTokens: ms.OutputTokens,
			TotalCostKRW: ms.TotalCost,
			AvgCostKRW:   ms.TotalCost / float64(ms.CallCount),
			AvgLatencyMs: avgLatency,
			LastUsed:     ms.LastUsed,
			RPM:          limits.RPM,
			TPM:          limits.TPM,
			RPD:          limits.RPD,
			ContextSize:  limits.Context,
			Note:         limits.Note,
		})
	}

	// Feature models from prompt templates
	prompts, _ := s.db.PromptTemplate.Query().
		Where(prompttemplate.IsActiveEQ(true)).
		All(ctx)
	featureModels := make(map[string]string)
	for _, pt := range prompts {
		featureModels[pt.Category+"/"+pt.SubCategory] = pt.Model
	}

	// Estimate quotas
	quotas := s.EstimateModelQuotas(logs)

	return &ModelStatsResult{
		Models:        models,
		FeatureModels: featureModels,
		ModelQuotas:   quotas,
	}, nil
}

// EstimateModelQuotas estimates remaining RPD quota for all models with daily limits.
// Gemini quotas reset at midnight Pacific Time, Groq at midnight UTC.
// Exported for reuse by Phase 9 payment service.
func (s *AdminAnalyticsService) EstimateModelQuotas(logs []*ent.UsageLog) []ModelQuota {
	selectableModels := ai.GetSelectableModels()

	loc, _ := time.LoadLocation("America/Los_Angeles")
	nowPT := time.Now().In(loc)
	geminiStartOfDay := time.Date(nowPT.Year(), nowPT.Month(), nowPT.Day(), 0, 0, 0, 0, loc)
	nowUTC := time.Now().UTC()
	groqStartOfDay := time.Date(nowUTC.Year(), nowUTC.Month(), nowUTC.Day(), 0, 0, 0, 0, time.UTC)

	geminiUsage := make(map[string]int)
	groqUsage := make(map[string]int)
	for _, l := range logs {
		if l.Model == nil {
			continue
		}
		if l.CreatedAt.After(geminiStartOfDay) {
			geminiUsage[*l.Model]++
		}
		if l.CreatedAt.After(groqStartOfDay) {
			groqUsage[*l.Model]++
		}
	}

	result := []ModelQuota{}
	for _, m := range selectableModels {
		if m.RPD == 0 {
			continue
		}
		used := 0
		if m.Provider == "gemini" {
			used = geminiUsage[m.ID]
		} else if m.Provider == "groq" {
			used = groqUsage[m.ID]
		}
		remaining := m.RPD - used
		if remaining < 0 {
			remaining = 0
		}
		result = append(result, ModelQuota{
			Model:      m.ID,
			Provider:   m.Provider,
			Limit:      m.RPD,
			Used:       used,
			Remaining:  remaining,
			Percentage: float64(used) / float64(m.RPD) * 100,
			RPM:        m.RPM,
			TPM:        m.TPM,
		})
	}
	return result
}

// GetDashboard returns aggregated data for the admin dashboard.
func (s *AdminAnalyticsService) GetDashboard(ctx context.Context) (*DashboardData, error) {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)
	sevenDaysAgo := today.AddDate(0, 0, -7)

	// User stats
	totalUsers, _ := s.db.UserProfile.Query().Count(ctx)
	newUsersToday, _ := s.db.UserProfile.Query().Where(userprofile.CreatedAtGTE(today)).Count(ctx)
	activeToday, _ := s.db.UserProfile.Query().Where(userprofile.LastLoginAtGTE(today)).Count(ctx)

	// AI metrics today & yesterday
	todayLogs, _ := s.db.UsageLog.Query().Where(usagelog.CreatedAtGTE(today)).All(ctx)
	yesterdayLogs, _ := s.db.UsageLog.Query().Where(usagelog.CreatedAtGTE(yesterday), usagelog.CreatedAtLT(today)).All(ctx)

	callsToday := len(todayLogs)
	callsYesterday := len(yesterdayLogs)
	var errorsToday, errorsYesterday int
	var costToday float64
	for _, l := range todayLogs {
		if l.Status == "error" {
			errorsToday++
		}
		if l.EstimatedCostKrw != nil {
			costToday += *l.EstimatedCostKrw
		}
	}
	for _, l := range yesterdayLogs {
		if l.Status == "error" {
			errorsYesterday++
		}
	}

	var callsDeltaPct, errorRateToday, errorRateDelta float64
	if callsYesterday > 0 {
		callsDeltaPct = float64(callsToday-callsYesterday) / float64(callsYesterday) * 100
	}
	if callsToday > 0 {
		errorRateToday = float64(errorsToday) / float64(callsToday) * 100
	}
	var errorRateYesterday float64
	if callsYesterday > 0 {
		errorRateYesterday = float64(errorsYesterday) / float64(callsYesterday) * 100
	}
	errorRateDelta = errorRateToday - errorRateYesterday

	// Daily metrics (last 7 days)
	allLogs7, _ := s.db.UsageLog.Query().Where(usagelog.CreatedAtGTE(sevenDaysAgo)).All(ctx)
	dailyMap := map[string]*DashboardDailyMetric{}
	for i := 0; i < 7; i++ {
		d := today.AddDate(0, 0, -i).Format("2006-01-02")
		dailyMap[d] = &DashboardDailyMetric{Date: d}
	}
	for _, l := range allLogs7 {
		d := l.CreatedAt.Format("2006-01-02")
		if _, ok := dailyMap[d]; !ok {
			continue
		}
		dailyMap[d].Calls++
		if l.Status == "error" {
			dailyMap[d].ErrorCount++
		}
		if l.EstimatedCostKrw != nil {
			dailyMap[d].CostKRW += *l.EstimatedCostKrw
		}
	}
	dailyMetrics := make([]DashboardDailyMetric, 0, 7)
	for i := 6; i >= 0; i-- {
		d := today.AddDate(0, 0, -i).Format("2006-01-02")
		dailyMetrics = append(dailyMetrics, *dailyMap[d])
	}

	// Quota alerts (models > 80% RPD)
	allQuotas := s.EstimateModelQuotas(allLogs7)
	quotaAlerts := make([]ModelQuota, 0)
	for _, q := range allQuotas {
		if q.Percentage >= 80 {
			quotaAlerts = append(quotaAlerts, q)
		}
	}

	// Recent errors (last 5 from ai_call_errors)
	recentErrRecords, _ := s.db.AICallError.Query().
		WithUsageLog().
		Order(ent.Desc("created_at")).
		Limit(5).
		All(ctx)
	recentErrors := make([]DashboardError, 0, len(recentErrRecords))
	for _, er := range recentErrRecords {
		item := DashboardError{
			ID:           er.ID,
			ErrorType:    string(er.ErrorType),
			ErrorMessage: er.ErrorMessage,
			CreatedAt:    er.CreatedAt,
		}
		if er.Edges.UsageLog != nil {
			ul := er.Edges.UsageLog
			item.Feature = ul.Feature
			item.UserID = ul.UserID
			item.InputTokens = ul.InputTokens
			item.OutputTokens = ul.OutputTokens
			if ul.Model != nil {
				item.Model = *ul.Model
			}
			if ul.Provider != nil {
				item.Provider = *ul.Provider
			}
		}
		recentErrors = append(recentErrors, item)
	}

	// Recent pending feedbacks (last 3)
	pendingFBCount, _ := s.db.Feedback.Query().
		Where(feedback.AdminStatusEQ(feedback.AdminStatusPending)).Count(ctx)
	recentFBRecords, _ := s.db.Feedback.Query().
		Where(feedback.AdminStatusEQ(feedback.AdminStatusPending)).
		Order(ent.Desc("created_at")).
		Limit(3).
		All(ctx)
	recentFeedbacks := make([]DashboardFeedback, 0, len(recentFBRecords))
	for _, f := range recentFBRecords {
		preview := f.Content
		if len(preview) > 100 {
			preview = preview[:100]
		}
		recentFeedbacks = append(recentFeedbacks, DashboardFeedback{
			ID:             f.ID,
			Category:       string(f.Category),
			ContentPreview: preview,
			CreatedAt:      f.CreatedAt,
		})
	}

	return &DashboardData{
		UserStats: DashboardUserStats{
			TotalUsers:       totalUsers,
			NewUsersToday:    newUsersToday,
			ActiveUsersToday: activeToday,
		},
		AIMetrics: DashboardAIMetrics{
			Calls:          callsToday,
			CallsDeltaPct:  callsDeltaPct,
			ErrorRate:      errorRateToday,
			ErrorRateDelta: errorRateDelta,
			CostKRW:        costToday,
		},
		DailyMetrics:         dailyMetrics,
		QuotaAlerts:          quotaAlerts,
		RecentErrors:         recentErrors,
		RecentFeedbacks:      recentFeedbacks,
		PendingFeedbackCount: pendingFBCount,
	}, nil
}

// ListErrors returns paginated AI call errors.
func (s *AdminAnalyticsService) ListErrors(ctx context.Context, params ErrorListParams) (*ErrorListResult, error) {
	since := time.Now().AddDate(0, 0, -params.Days)

	q := s.db.AICallError.Query().
		WithUsageLog().
		Where(aicallerror.CreatedAtGTE(since)).
		Order(ent.Desc("created_at"))

	if params.ErrorType != "" {
		q = q.Where(aicallerror.ErrorTypeEQ(aicallerror.ErrorType(params.ErrorType)))
	}

	var usageLogPreds []predicate.UsageLog
	if params.Provider != "" {
		usageLogPreds = append(usageLogPreds, usagelog.ProviderEQ(params.Provider))
	}
	if params.Feature != "" {
		usageLogPreds = append(usageLogPreds, usagelog.FeatureEQ(params.Feature))
	}
	if len(usageLogPreds) > 0 {
		q = q.Where(aicallerror.HasUsageLogWith(usageLogPreds...))
	}

	total, _ := q.Count(ctx)

	records, err := q.Limit(params.Limit).Offset(params.Offset).All(ctx)
	if err != nil {
		return nil, err
	}

	return &ErrorListResult{Errors: records, Total: total}, nil
}
