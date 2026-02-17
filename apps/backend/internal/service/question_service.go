package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/application"
	"github.com/coby/colight/apps/backend/ent/experience"
	"github.com/coby/colight/apps/backend/ent/experienceusage"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/google/uuid"
)

// QuestionService handles question analysis
type QuestionService struct {
	entClient  *ent.Client
	aiProvider *ai.AIProvider
}

// NewQuestionService creates a new question service
func NewQuestionService(entClient *ent.Client, aiProvider *ai.AIProvider) *QuestionService {
	return &QuestionService{
		entClient:  entClient,
		aiProvider: aiProvider,
	}
}

// QuestionAnalysisResult represents the analysis result for a question
type QuestionAnalysisResult struct {
	SurfaceQuestion       string            `json:"surface_question"`
	RealIntents           []RealIntent      `json:"real_intents"`
	RequiredWeapons       RequiredWeapons   `json:"required_weapons"`
	WritingStructure      WritingStructure  `json:"writing_structure"`
	KeyKeywords           []string          `json:"key_keywords"`
	AvoidList             []string          `json:"avoid_list"`
	GoodStructureExample  string            `json:"good_structure_example"`
}

// RealIntent represents a hidden intent behind the question
type RealIntent struct {
	Intent string `json:"intent"`
	Why    string `json:"why"`
}

// RequiredWeapons represents required competencies
type RequiredWeapons struct {
	Primary   WeaponInfo   `json:"primary"`
	Secondary []WeaponInfo `json:"secondary"`
}

// WeaponInfo represents weapon/competency information
type WeaponInfo struct {
	WeaponID   string `json:"weapon_id"`
	WeaponName string `json:"weapon_name"`
	Reason     string `json:"reason"`
}

// WritingStructure represents recommended writing structure
type WritingStructure struct {
	TotalChars int       `json:"total_chars"`
	Sections   []Section `json:"sections"`
}

// Section represents a writing section
type Section struct {
	Name      string  `json:"name"`
	CharRatio float64 `json:"char_ratio"`
	CharCount int     `json:"char_count"`
	Guide     string  `json:"guide"`
}

// ListApplications returns the user's applications for coaching page company select
func (s *QuestionService) ListApplications(ctx context.Context, userID uuid.UUID) ([]*ent.Application, error) {
	apps, err := s.entClient.Application.Query().
		Where(application.UserIDEQ(userID)).
		Order(ent.Desc(application.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	}
	return apps, nil
}

// AnalyzeQuestion analyzes a cover letter question using Claude.
// applicationID is optional — when nil, companyName is used directly (standalone coaching).
func (s *QuestionService) AnalyzeQuestion(ctx context.Context, userID uuid.UUID, applicationID *uuid.UUID, companyName string, questionText string, charLimit int) (*QuestionAnalysisResult, error) {
	// 1. Resolve company context from application or direct input
	resolvedCompanyName := companyName
	resolvedPosition := ""
	talentKeywords := ""
	valuesKeywords := ""

	if applicationID != nil {
		app, err := s.entClient.Application.Query().
			Where(application.IDEQ(*applicationID)).
			WithAnalysis().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, ErrApplicationNotFound
			}
			return nil, err
		}

		if app.UserID != userID {
			return nil, ErrApplicationForbidden
		}

		resolvedCompanyName = app.CompanyName
		resolvedPosition = app.Position

		if app.Edges.Analysis != nil && app.Edges.Analysis.AnalysisResult != nil {
			if coreValues, ok := app.Edges.Analysis.AnalysisResult["core_values"].([]any); ok && len(coreValues) > 0 {
				var keywords []string
				for _, v := range coreValues {
					if vm, ok := v.(map[string]any); ok {
						if keyword, ok := vm["keyword"].(string); ok {
							keywords = append(keywords, keyword)
						}
					}
				}
				valuesKeywords = strings.Join(keywords, ", ")
			}

			if talentTraits, ok := app.Edges.Analysis.AnalysisResult["talent_traits"].([]any); ok && len(talentTraits) > 0 {
				var traits []string
				for _, t := range talentTraits {
					if tm, ok := t.(map[string]any); ok {
						if trait, ok := tm["trait"].(string); ok {
							traits = append(traits, trait)
						}
					}
				}
				talentKeywords = strings.Join(traits, ", ")
			}
		}
	}

	// 2. Load weapon categories for context
	weapons, err := s.entClient.WeaponCategory.Query().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load weapon categories: %w", err)
	}

	weaponList := make([]string, len(weapons))
	for i, w := range weapons {
		weaponList[i] = fmt.Sprintf("- %s (%s): %s", w.Code, w.Name, strings.Join(w.Keywords, ", "))
	}
	weaponContext := strings.Join(weaponList, "\n")

	// 3. Load prompt template
	prompt, err := s.entClient.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("coaching"),
			prompttemplate.SubCategoryEQ("question_analysis"),
			prompttemplate.IsActiveEQ(true),
		).
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load prompt template: %w", err)
	}

	// 4. Build prompt
	userPrompt := strings.ReplaceAll(prompt.UserPromptTemplate, "{{company_name}}", resolvedCompanyName)
	userPrompt = strings.ReplaceAll(userPrompt, "{{position}}", resolvedPosition)
	userPrompt = strings.ReplaceAll(userPrompt, "{{talent_keywords}}", talentKeywords)
	userPrompt = strings.ReplaceAll(userPrompt, "{{values_keywords}}", valuesKeywords)
	userPrompt = strings.ReplaceAll(userPrompt, "{{weapon_categories}}", weaponContext)
	userPrompt = strings.ReplaceAll(userPrompt, "{{question_text}}", questionText)
	userPrompt = strings.ReplaceAll(userPrompt, "{{char_limit}}", fmt.Sprintf("%d", charLimit))

	// 5. Call Claude
	llmReq := ai.LLMRequest{
		SystemPrompt: prompt.SystemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  prompt.Temperature,
		MaxTokens:    prompt.MaxTokens,
	}

	resp, err := s.aiProvider.CallByModelName(ctx, prompt.Model, llmReq)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	// 6. Parse response
	var result QuestionAnalysisResult
	if err := ai.ExtractJSON(resp.Content, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 7. Validate result structure
	if len(result.RealIntents) != 3 {
		return nil, fmt.Errorf("real_intents must have exactly 3 items, got %d", len(result.RealIntents))
	}

	if result.RequiredWeapons.Primary.WeaponID == "" {
		return nil, fmt.Errorf("primary weapon must exist")
	}

	// Verify char_count sums to total_chars
	sum := 0
	for _, section := range result.WritingStructure.Sections {
		sum += section.CharCount
	}
	if sum != result.WritingStructure.TotalChars {
		return nil, fmt.Errorf("char_count sum (%d) does not match total_chars (%d)", sum, result.WritingStructure.TotalChars)
	}

	return &result, nil
}

// RecommendInput holds the input parameters for experience recommendation
type RecommendInput struct {
	RequiredWeapons RequiredWeapons `json:"required_weapons"`
	KeyKeywords     []string        `json:"key_keywords"`
	ApplicationID   *uuid.UUID      `json:"application_id"`
	Limit           int             `json:"limit"`
}

// ExperienceRecommendation represents a recommended experience with match score
type ExperienceRecommendation struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Category       string    `json:"category"`
	PeriodStart    string    `json:"period_start,omitempty"`
	PeriodEnd      string    `json:"period_end,omitempty"`
	StarSituation  string    `json:"star_situation"`
	Weapons        []string  `json:"weapons"`
	MatchScore     int       `json:"match_score"` // 0~100
	MatchReasons   []string  `json:"match_reasons"`
	IsUsed         bool      `json:"is_used"`
	KeywordMatches []string  `json:"keyword_matches"`
}

// recommendationScore holds intermediate scoring details for a single experience
type recommendationScore struct {
	exp            *ent.Experience
	score          int
	reasons        []string
	keywordMatches []string
	isUsed         bool
}

// RecommendExperiences recommends top N experiences using multi-factor scoring:
// weapon match (50pts), keyword overlap (30pts), usage penalty (-15pts), freshness bonus (+5pts)
func (s *QuestionService) RecommendExperiences(ctx context.Context, userID uuid.UUID, input RecommendInput) ([]ExperienceRecommendation, error) {
	// 1. Get all user experiences with weapons
	experiences, err := s.entClient.Experience.Query().
		Where(experience.UserIDEQ(userID)).
		WithWeapons().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query experiences: %w", err)
	}

	if len(experiences) == 0 {
		return []ExperienceRecommendation{}, nil
	}

	// 2. Query usage data: which experience IDs are already used in this application
	usedInAppSet := make(map[uuid.UUID]bool)
	if input.ApplicationID != nil {
		usedIDs, err := s.entClient.ExperienceUsage.Query().
			Where(experienceusage.HasApplicationWith(application.IDEQ(*input.ApplicationID))).
			QueryExperience().
			IDs(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query experience usage: %w", err)
		}
		for _, id := range usedIDs {
			usedInAppSet[id] = true
		}
	}

	// 3. Query global usage counts for freshness bonus
	globalUsedSet := make(map[uuid.UUID]bool)
	globalUsedIDs, err := s.entClient.ExperienceUsage.Query().
		Where(experienceusage.UserIDEQ(userID)).
		QueryExperience().
		IDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query global usage: %w", err)
	}
	for _, id := range globalUsedIDs {
		globalUsedSet[id] = true
	}

	// 4. Build lowercase keyword set from question analysis
	keywordSet := make(map[string]bool, len(input.KeyKeywords))
	for _, kw := range input.KeyKeywords {
		keywordSet[strings.ToLower(kw)] = true
	}

	// 5. Score each experience
	var scored []recommendationScore
	for _, exp := range experiences {
		rs := s.calculateRecommendationScore(exp, input.RequiredWeapons, keywordSet, len(input.KeyKeywords), usedInAppSet[exp.ID], globalUsedSet[exp.ID])
		scored = append(scored, rs)
	}

	// 6. Sort by score (descending)
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// 7. Take top N (include zero-score items so frontend can show all)
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if len(scored) < limit {
		limit = len(scored)
	}

	// 8. Build recommendations
	recommendations := make([]ExperienceRecommendation, 0, limit)
	for i := 0; i < limit; i++ {
		rs := scored[i]
		exp := rs.exp
		weaponNames := make([]string, len(exp.Edges.Weapons))
		for j, w := range exp.Edges.Weapons {
			weaponNames[j] = w.WeaponCode
		}

		rec := ExperienceRecommendation{
			ID:             exp.ID,
			Title:          exp.Title,
			Category:       exp.Category,
			StarSituation:  exp.StarSituation,
			Weapons:        weaponNames,
			MatchScore:     rs.score,
			MatchReasons:   rs.reasons,
			IsUsed:         rs.isUsed,
			KeywordMatches: rs.keywordMatches,
		}

		if exp.PeriodStart != nil {
			rec.PeriodStart = exp.PeriodStart.Format("2006-01")
		}
		if exp.PeriodEnd != nil {
			rec.PeriodEnd = exp.PeriodEnd.Format("2006-01")
		}

		recommendations = append(recommendations, rec)
	}

	return recommendations, nil
}

// calculateRecommendationScore computes multi-factor score for an experience.
// Weapon match: 50pts max, Keyword overlap: 30pts max, Usage penalty: -15pts, Freshness bonus: +5pts.
// Raw score is normalized to 0–100.
func (s *QuestionService) calculateRecommendationScore(
	exp *ent.Experience,
	requiredWeapons RequiredWeapons,
	keywordSet map[string]bool,
	totalKeywords int,
	isUsedInApp bool,
	isUsedGlobally bool,
) recommendationScore {
	rs := recommendationScore{exp: exp, isUsed: isUsedInApp}
	rawScore := 0.0

	// --- Weapon match (max 50 points) ---
	// Primary weapon: 35 points × confidence
	for _, w := range exp.Edges.Weapons {
		if w.WeaponCode == requiredWeapons.Primary.WeaponID {
			pts := 35.0 * w.Confidence
			rawScore += pts
			rs.reasons = append(rs.reasons, fmt.Sprintf("주 무기 '%s' 일치 (신뢰도 %d%%)", requiredWeapons.Primary.WeaponName, int(w.Confidence*100)))
			break
		}
	}

	// Secondary weapons: share 15 points equally
	numSecondary := len(requiredWeapons.Secondary)
	if numSecondary > 0 {
		ptsPerSecondary := 15.0 / float64(numSecondary)
		for _, secondary := range requiredWeapons.Secondary {
			for _, w := range exp.Edges.Weapons {
				if w.WeaponCode == secondary.WeaponID {
					pts := ptsPerSecondary * w.Confidence
					rawScore += pts
					rs.reasons = append(rs.reasons, fmt.Sprintf("부 무기 '%s' 일치", secondary.WeaponName))
					break
				}
			}
		}
	}

	// --- Keyword overlap (max 30 points) ---
	if totalKeywords > 0 && len(exp.Keywords) > 0 {
		var matched []string
		for _, kw := range exp.Keywords {
			if keywordSet[strings.ToLower(kw)] {
				matched = append(matched, kw)
			}
		}
		if len(matched) > 0 {
			overlapRatio := float64(len(matched)) / float64(totalKeywords)
			if overlapRatio > 1.0 {
				overlapRatio = 1.0
			}
			pts := overlapRatio * 30.0
			rawScore += pts
			rs.keywordMatches = matched
			rs.reasons = append(rs.reasons, fmt.Sprintf("키워드 %d개 매칭: %s", len(matched), strings.Join(matched, ", ")))
		}
	}

	// --- Usage penalty / freshness bonus ---
	if isUsedInApp {
		rawScore -= 15.0
		rs.reasons = append(rs.reasons, "이 지원서에서 이미 사용됨")
	} else if !isUsedGlobally {
		rawScore += 5.0
		rs.reasons = append(rs.reasons, "아직 사용되지 않은 경험")
	}

	// Normalize: max possible raw = 50 (weapon) + 30 (keyword) + 5 (freshness) = 85
	// Map to 0–100 scale
	normalized := (rawScore / 85.0) * 100.0
	normalized = math.Max(0, math.Min(100, normalized))
	rs.score = int(math.Round(normalized))

	return rs
}
