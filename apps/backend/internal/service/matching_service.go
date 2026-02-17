package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/experience"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// MatchingService handles experience-company matching
type MatchingService struct {
	entClient    *ent.Client
	aiProvider   *ai.AIProvider
	resultCache  map[string]*MatchResult // Key: experienceID_companyName
	cacheExpiry  map[string]time.Time
}

// NewMatchingService creates a new matching service
func NewMatchingService(entClient *ent.Client, aiProvider *ai.AIProvider) *MatchingService {
	return &MatchingService{
		entClient:   entClient,
		aiProvider:  aiProvider,
		resultCache: make(map[string]*MatchResult),
		cacheExpiry: make(map[string]time.Time),
	}
}

// MatchResult represents matching result for an experience
type MatchResult struct {
	ExperienceID   uuid.UUID `json:"experience_id"`
	OverallFit     int       `json:"overall_fit"`      // 0-100
	JobRelevance   int       `json:"job_relevance"`    // 0-100 (40% weight)
	TalentFit      int       `json:"talent_fit"`       // 0-100 (35% weight)
	Uniqueness     int       `json:"uniqueness"`       // 0-100 (25% weight)
	Reasoning      string    `json:"reasoning"`
	SuggestedAngle string    `json:"suggested_angle"`
}

// aiMatchResponse matches expected AI response format
type aiMatchResponse struct {
	OverallFit     int    `json:"overall_fit"`
	JobRelevance   int    `json:"job_relevance"`
	TalentFit      int    `json:"talent_fit"`
	Uniqueness     int    `json:"uniqueness"`
	Reasoning      string `json:"reasoning"`
	SuggestedAngle string `json:"suggested_angle"`
}

// MatchExperience matches a single experience against company analysis
func (s *MatchingService) MatchExperience(ctx context.Context, userID uuid.UUID, experienceID uuid.UUID, companyAnalysis *CompanyAnalysis) (*MatchResult, error) {
	// Step 4.5: Check cache first (5-minute TTL for matching results)
	cacheKey := experienceID.String() + "_" + companyAnalysis.CompanyName
	if cached, ok := s.resultCache[cacheKey]; ok {
		if expiry, exists := s.cacheExpiry[cacheKey]; exists && time.Now().Before(expiry) {
			return cached, nil
		}
		// Cache expired, remove
		delete(s.resultCache, cacheKey)
		delete(s.cacheExpiry, cacheKey)
	}

	// 1. Verify experience exists and user owns it
	exp, err := s.entClient.Experience.Query().
		Where(experience.IDEQ(experienceID)).
		WithWeapons().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrExperienceNotFound
		}
		return nil, err
	}

	if exp.UserID != userID {
		return nil, ErrExperienceForbidden
	}

	// 2. Build experience text
	experienceText := fmt.Sprintf(`제목: %s
상황: %s
과제: %s
행동: %s
결과: %s
카테고리: %s`, exp.Title, exp.StarSituation, exp.StarTask, exp.StarAction, exp.StarResult, exp.Category)

	// 3. Build company context
	companyContext := fmt.Sprintf(`회사명: %s
핵심가치: %v
인재상: %v`, companyAnalysis.CompanyName, formatCoreValues(companyAnalysis.CoreValues), formatTalentTraits(companyAnalysis.TalentTraits))

	// 4. Call AI for matching (load prompt from DB)
	pt, ptErr := s.loadMatchingPrompt(ctx)
	if ptErr != nil {
		return nil, fmt.Errorf("failed to load matching prompt: %w", ptErr)
	}

	userPrompt := pt.UserPromptTemplate
	for k, v := range map[string]string{
		"experience_text":  experienceText,
		"company_context":  companyContext,
	} {
		userPrompt = strings.ReplaceAll(userPrompt, "{{"+k+"}}", v)
	}

	startTime := time.Now()
	resp, err := ai.CallByModelNameWithRetry(ctx, s.aiProvider, pt.Model, ai.LLMRequest{
		SystemPrompt: pt.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  pt.Temperature,
		MaxTokens:    pt.MaxTokens,
		JSONMode:     true,
	}, ai.DefaultRetryConfig())
	if err != nil {
		return nil, fmt.Errorf("AI matching failed: %w", err)
	}

	s.updateMatchingPromptStats(ctx, pt, time.Since(startTime))

	// 5. Parse AI response
	var aiResult aiMatchResponse
	if err := ai.ExtractJSON(resp.Content, &aiResult); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 6. Validate scores
	if aiResult.OverallFit < 0 || aiResult.OverallFit > 100 {
		return nil, fmt.Errorf("invalid overall_fit score: %d", aiResult.OverallFit)
	}

	// 7. Build result
	result := &MatchResult{
		ExperienceID:   exp.ID,
		OverallFit:     aiResult.OverallFit,
		JobRelevance:   aiResult.JobRelevance,
		TalentFit:      aiResult.TalentFit,
		Uniqueness:     aiResult.Uniqueness,
		Reasoning:      aiResult.Reasoning,
		SuggestedAngle: aiResult.SuggestedAngle,
	}

	// Step 4.5: Save to cache (5-minute TTL)
	s.resultCache[cacheKey] = result
	s.cacheExpiry[cacheKey] = time.Now().Add(5 * time.Minute)

	return result, nil
}

func formatCoreValues(values []CoreValue) string {
	result := ""
	for _, v := range values {
		result += v.Keyword + ", "
	}
	return result
}

func formatTalentTraits(traits []TalentTrait) string {
	result := ""
	for _, t := range traits {
		result += t.Trait + ", "
	}
	return result
}
// CalculateOverallFit computes the weighted average overall fit score
// Weights: job_relevance 40%, talent_fit 35%, uniqueness 25%
func CalculateOverallFit(jobRelevance, talentFit, uniqueness int) int {
	return (jobRelevance*40 + talentFit*35 + uniqueness*25) / 100
}

// BatchMatchResult holds all matching results for a user's experiences
type BatchMatchResult struct {
	Matches   []MatchResult `json:"matches"`
	Total     int           `json:"total"`
	MatchedAt time.Time     `json:"matched_at"`
}

// MatchAllExperiences matches all user experiences against company analysis
func (s *MatchingService) MatchAllExperiences(ctx context.Context, userID uuid.UUID, companyAnalysis *CompanyAnalysis) (*BatchMatchResult, error) {
	// 1. Get all user experiences
	experiences, err := s.entClient.Experience.Query().
		Where(experience.UserIDEQ(userID)).
		WithWeapons().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query experiences: %w", err)
	}

	if len(experiences) == 0 {
		return nil, ErrNoExperiences
	}

	// 2. Match each experience
	var matches []MatchResult
	for _, exp := range experiences {
		result, err := s.MatchExperience(ctx, userID, exp.ID, companyAnalysis)
		if err != nil {
			continue // skip failed matches
		}
		matches = append(matches, *result)
	}

	return &BatchMatchResult{
		Matches:   matches,
		Total:     len(matches),
		MatchedAt: time.Now(),
	}, nil
}

// OutdatedCheckResult indicates whether a matching result is stale
type OutdatedCheckResult struct {
	IsOutdated bool   `json:"is_outdated"`
	Reason     string `json:"reason,omitempty"`
}

// CheckMatchingOutdated checks if a matching result is outdated
// Conditions: experience added/modified/deleted after matchedAt, or 7+ days elapsed
func (s *MatchingService) CheckMatchingOutdated(ctx context.Context, userID uuid.UUID, matchedAt time.Time) (*OutdatedCheckResult, error) {
	// 1. Find latest experience update time
	latestExp, err := s.entClient.Experience.Query().
		Where(experience.UserIDEQ(userID)).
		Order(ent.Desc(experience.FieldUpdatedAt)).
		First(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("failed to query latest experience: %w", err)
	}

	// Check if any experience was modified after matching
	if latestExp != nil && latestExp.UpdatedAt.After(matchedAt) {
		return &OutdatedCheckResult{
			IsOutdated: true,
			Reason:     "매칭 이후 경험이 변경되었습니다",
		}, nil
	}

	// Check experience count changed (deletion detection via create_time after match)
	// If no experiences exist but matching was done, experiences were deleted
	if latestExp == nil {
		return &OutdatedCheckResult{
			IsOutdated: true,
			Reason:     "등록된 경험이 없습니다",
		}, nil
	}

	// Check 7-day expiry
	if time.Since(matchedAt) > 7*24*time.Hour {
		return &OutdatedCheckResult{
			IsOutdated: true,
			Reason:     "매칭 결과가 7일 이상 지났습니다",
		}, nil
	}

	return &OutdatedCheckResult{IsOutdated: false}, nil
}

// GetClient returns the Ent client
func (s *MatchingService) GetClient() *ent.Client {
	return s.entClient
}

// loadMatchingPrompt loads the matching prompt template from DB.
// Falls back to hardcoded defaults when template is not seeded.
func (s *MatchingService) loadMatchingPrompt(ctx context.Context) (*ent.PromptTemplate, error) {
	if s.entClient != nil {
		pt, err := s.entClient.PromptTemplate.Query().
			Where(
				prompttemplate.CategoryEQ("matching"),
				prompttemplate.SubCategoryEQ("match_experience"),
				prompttemplate.IsActiveEQ(true),
			).
			Order(prompttemplate.ByVersion(sql.OrderDesc())).
			First(ctx)
		if err == nil {
			return pt, nil
		}
		slog.Warn("matching_prompt_fallback", "error", err)
	}
	return &ent.PromptTemplate{
		Model:              "groq/compound",
		SystemPrompt:       "You are an experience-company matching analyst. Provide objective fit scores.",
		UserPromptTemplate: "Experience:\n{{experience_text}}\n\nCompany Requirements:\n{{company_context}}\n\nReturn JSON with scores (0-100): overall_fit, job_relevance, talent_fit, uniqueness, reasoning, suggested_angle.",
		Temperature:        0.2,
		MaxTokens:          1000,
	}, nil
}

// updateMatchingPromptStats updates usage count and avg latency
func (s *MatchingService) updateMatchingPromptStats(ctx context.Context, pt *ent.PromptTemplate, latency time.Duration) {
	if s.entClient == nil || pt.ID.String() == "00000000-0000-0000-0000-000000000000" {
		return
	}
	latencyMs := int(latency.Milliseconds())
	_ = s.entClient.PromptTemplate.UpdateOneID(pt.ID).
		SetUsageCount(pt.UsageCount + 1).
		SetAvgLatencyMs((pt.AvgLatencyMs*pt.UsageCount + latencyMs) / (pt.UsageCount + 1)).
		Exec(ctx)
}
