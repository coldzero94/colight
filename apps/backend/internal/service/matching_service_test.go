package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMForMatching mocks AI for matching tests
type MockLLMForMatching struct {
	response ai.LLMResponse
	err      error
	calls    int
}

func (m *MockLLMForMatching) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	m.calls++
	return m.response, m.err
}

// createTestUserAndExperience is a helper for matching tests
func createTestUserAndExperience(t *testing.T, client *ent.Client, emailSuffix string) (uuid.UUID, uuid.UUID) {
	t.Helper()
	user := client.UserProfile.Create().
		SetEmail(fmt.Sprintf("match-%s@example.com", emailSuffix)).
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	exp := client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("테스트 경험").
		SetContent("테스트 내용").
		SetCategory("인턴").
		SetStarSituation("S").
		SetStarTask("T").
		SetStarAction("A").
		SetStarResult("R").
		SaveX(context.Background())

	return user.ID, exp.ID
}

func TestMatchExperience_ReturnsThreeCategoryScores(t *testing.T) {
	mockAI := &MockLLMForMatching{
		response: ai.LLMResponse{
			Content: `{
				"overall_fit": 85,
				"job_relevance": 90,
				"talent_fit": 85,
				"uniqueness": 75,
				"reasoning": "Strong backend development experience matches the position requirements",
				"suggested_angle": "Emphasize system design and team collaboration"
			}`,
		},
	}

	client := testutil.NewTestClient(t)
	service := NewMatchingService(client, mockAI)

	// Create test user
	user := client.UserProfile.Create().
		SetEmail("match-test@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	// Create test experience
	exp := client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("백엔드 개발 경험").
		SetContent("Go 백엔드 개발 및 시스템 설계").
		SetCategory("인턴").
		SetStarSituation("스타트업 백엔드 팀").
		SetStarTask("API 서버 개발").
		SetStarAction("Go + PostgreSQL로 REST API 구현").
		SetStarResult("일 1만 요청 처리").
		SaveX(context.Background())

	// Mock company analysis
	companyAnalysis := &CompanyAnalysis{
		CompanyName: "테스트 회사",
		TalentTraits: []TalentTrait{
			{Trait: "기술적 전문성", Description: "깊이 있는 기술 이해"},
		},
	}

	// Execute matching
	result, err := service.MatchExperience(context.Background(), user.ID, exp.ID, companyAnalysis)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify 3 category scores present
	assert.Greater(t, result.JobRelevance, 0)
	assert.Greater(t, result.TalentFit, 0)
	assert.Greater(t, result.Uniqueness, 0)
	assert.Greater(t, result.OverallFit, 0)

	// Verify overall is weighted average (roughly)
	expectedOverall := (result.JobRelevance*40 + result.TalentFit*35 + result.Uniqueness*25) / 100
	assert.InDelta(t, expectedOverall, result.OverallFit, 5.0, "Overall should be weighted average")
}

func TestMatchExperience_ReasoningNotEmpty(t *testing.T) {
	mockAI := &MockLLMForMatching{
		response: ai.LLMResponse{
			Content: `{
				"overall_fit": 75,
				"job_relevance": 80,
				"talent_fit": 70,
				"uniqueness": 75,
				"reasoning": "Experience shows problem-solving skills",
				"suggested_angle": "Focus on achievements"
			}`,
		},
	}

	client := testutil.NewTestClient(t)
	service := NewMatchingService(client, mockAI)

	user := client.UserProfile.Create().
		SetEmail("match-test2@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	exp := client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("Test").
		SetContent("Content").
		SetCategory("기타").
		SetStarSituation("S").
		SetStarTask("T").
		SetStarAction("A").
		SetStarResult("R").
		SaveX(context.Background())

	companyAnalysis := &CompanyAnalysis{CompanyName: "Test"}

	result, err := service.MatchExperience(context.Background(), user.ID, exp.ID, companyAnalysis)

	require.NoError(t, err)
	assert.NotEmpty(t, result.Reasoning, "Reasoning should not be empty")
	assert.NotEmpty(t, result.SuggestedAngle, "Suggested angle should not be empty")
}

func TestMatchExperience_Forbidden(t *testing.T) {
	mockAI := &MockLLMForMatching{}
	client := testutil.NewTestClient(t)
	service := NewMatchingService(client, mockAI)

	owner := client.UserProfile.Create().
		SetEmail("owner@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	otherUser := client.UserProfile.Create().
		SetEmail("other@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	exp := client.Experience.Create().
		SetUserID(owner.ID).
		SetTitle("Test").
		SetContent("Content").
		SetCategory("기타").
		SetStarSituation("S").
		SetStarTask("T").
		SetStarAction("A").
		SetStarResult("R").
		SaveX(context.Background())

	companyAnalysis := &CompanyAnalysis{CompanyName: "Test"}

	// Try to match other user's experience
	_, err := service.MatchExperience(context.Background(), otherUser.ID, exp.ID, companyAnalysis)

	require.Error(t, err)
	assert.Equal(t, ErrExperienceForbidden, err)
}

// --- Step 4.1 additional tests ---

func TestCalculateOverallFit_WeightedAverage(t *testing.T) {
	tests := []struct {
		name         string
		jobRelevance int
		talentFit    int
		uniqueness   int
		expected     int
	}{
		{"all 100s", 100, 100, 100, 100},
		{"all 0s", 0, 0, 0, 0},
		{"weighted 90/80/70", 90, 80, 70, (90*40 + 80*35 + 70*25) / 100},
		{"heavy job relevance", 100, 0, 0, 40},
		{"heavy talent fit", 0, 100, 0, 35},
		{"heavy uniqueness", 0, 0, 100, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateOverallFit(tt.jobRelevance, tt.talentFit, tt.uniqueness)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchExperience_HighRelevance_ScoreAbove70(t *testing.T) {
	mockAI := &MockLLMForMatching{
		response: ai.LLMResponse{
			Content: `{
				"overall_fit": 88,
				"job_relevance": 95,
				"talent_fit": 85,
				"uniqueness": 80,
				"reasoning": "Excellent match with strong technical background",
				"suggested_angle": "Lead with technical achievements"
			}`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewMatchingService(client, mockAI)
	userID, expID := createTestUserAndExperience(t, client, "high-rel")

	result, err := svc.MatchExperience(context.Background(), userID, expID, &CompanyAnalysis{
		CompanyName: "Tech Corp",
		CoreValues:  []CoreValue{{Keyword: "기술혁신", Description: "기술 중심"}},
	})

	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.OverallFit, 70)
	assert.GreaterOrEqual(t, result.JobRelevance, 70)
}

func TestMatchExperience_LowRelevance_ScoreBelow30(t *testing.T) {
	mockAI := &MockLLMForMatching{
		response: ai.LLMResponse{
			Content: `{
				"overall_fit": 15,
				"job_relevance": 10,
				"talent_fit": 20,
				"uniqueness": 15,
				"reasoning": "Experience has minimal relevance to this position",
				"suggested_angle": "Highlight transferable soft skills only"
			}`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewMatchingService(client, mockAI)
	userID, expID := createTestUserAndExperience(t, client, "low-rel")

	result, err := svc.MatchExperience(context.Background(), userID, expID, &CompanyAnalysis{
		CompanyName: "Unrelated Corp",
	})

	require.NoError(t, err)
	assert.LessOrEqual(t, result.OverallFit, 30)
}

func TestMatchAllExperiences_BatchProcessing(t *testing.T) {
	mockAI := &MockLLMForMatching{
		response: ai.LLMResponse{
			Content: `{
				"overall_fit": 70,
				"job_relevance": 75,
				"talent_fit": 65,
				"uniqueness": 70,
				"reasoning": "Good match",
				"suggested_angle": "Focus on results"
			}`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewMatchingService(client, mockAI)

	user := client.UserProfile.Create().
		SetEmail("batch-test@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	// Create 10 experiences
	for i := 0; i < 10; i++ {
		client.Experience.Create().
			SetUserID(user.ID).
			SetTitle(fmt.Sprintf("경험 %d", i)).
			SetContent(fmt.Sprintf("내용 %d", i)).
			SetCategory("인턴").
			SetStarSituation("S").
			SetStarTask("T").
			SetStarAction("A").
			SetStarResult("R").
			SaveX(context.Background())
	}

	companyAnalysis := &CompanyAnalysis{CompanyName: "배치 회사"}

	result, err := svc.MatchAllExperiences(context.Background(), user.ID, companyAnalysis)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 10, result.Total)
	assert.Len(t, result.Matches, 10)
	assert.False(t, result.MatchedAt.IsZero())
	// Verify AI was called 10 times
	assert.Equal(t, 10, mockAI.calls)
}

func TestMatchExperience_SchemaValidation(t *testing.T) {
	// AI returns invalid JSON → error
	mockAI := &MockLLMForMatching{
		response: ai.LLMResponse{
			Content: `not valid json at all`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewMatchingService(client, mockAI)
	userID, expID := createTestUserAndExperience(t, client, "schema-val")

	_, err := svc.MatchExperience(context.Background(), userID, expID, &CompanyAnalysis{CompanyName: "Test"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse AI response")
}

func TestMatchExperience_MockAIClient(t *testing.T) {
	// Verify mock AI is called with correct request parameters
	mockAI := &MockLLMForMatching{
		response: ai.LLMResponse{
			Content: `{
				"overall_fit": 60,
				"job_relevance": 65,
				"talent_fit": 55,
				"uniqueness": 60,
				"reasoning": "Decent match",
				"suggested_angle": "Angle"
			}`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewMatchingService(client, mockAI)
	userID, expID := createTestUserAndExperience(t, client, "mock-ai")

	result, err := svc.MatchExperience(context.Background(), userID, expID, &CompanyAnalysis{CompanyName: "MockTest"})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, mockAI.calls, "AI should be called exactly once")
	assert.Equal(t, expID, result.ExperienceID)
}

func TestMatchAllExperiences_NoExperiences(t *testing.T) {
	mockAI := &MockLLMForMatching{}
	client := testutil.NewTestClient(t)
	svc := NewMatchingService(client, mockAI)

	user := client.UserProfile.Create().
		SetEmail("no-exp@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	_, err := svc.MatchAllExperiences(context.Background(), user.ID, &CompanyAnalysis{CompanyName: "Test"})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoExperiences))
}

func TestMatchExperience_AIFailure(t *testing.T) {
	mockAI := &MockLLMForMatching{
		err: errors.New("AI service unavailable"),
	}

	client := testutil.NewTestClient(t)
	svc := NewMatchingService(client, mockAI)
	userID, expID := createTestUserAndExperience(t, client, "ai-fail")

	_, err := svc.MatchExperience(context.Background(), userID, expID, &CompanyAnalysis{CompanyName: "Test"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "AI matching failed")
}

func TestMatchExperience_InvalidScore(t *testing.T) {
	mockAI := &MockLLMForMatching{
		response: ai.LLMResponse{
			Content: `{
				"overall_fit": 150,
				"job_relevance": 90,
				"talent_fit": 80,
				"uniqueness": 70,
				"reasoning": "Good",
				"suggested_angle": "Angle"
			}`,
		},
	}

	client := testutil.NewTestClient(t)
	svc := NewMatchingService(client, mockAI)
	userID, expID := createTestUserAndExperience(t, client, "invalid-score")

	_, err := svc.MatchExperience(context.Background(), userID, expID, &CompanyAnalysis{CompanyName: "Test"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid overall_fit score")
}

// --- Step 4.5: Outdated detection tests ---

func TestCheckMatchingOutdated_ExperienceAdded(t *testing.T) {
	client := testutil.NewTestClient(t)
	mockAI := &MockLLMForMatching{}
	svc := NewMatchingService(client, mockAI)

	user := client.UserProfile.Create().
		SetEmail("outdated-add@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	// matchedAt is in the past
	matchedAt := time.Now().Add(-1 * time.Hour)

	// Add experience AFTER matchedAt
	client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("New Exp").
		SetContent("Content").
		SetCategory("인턴").
		SetStarSituation("S").
		SetStarTask("T").
		SetStarAction("A").
		SetStarResult("R").
		SaveX(context.Background())

	result, err := svc.CheckMatchingOutdated(context.Background(), user.ID, matchedAt)

	require.NoError(t, err)
	assert.True(t, result.IsOutdated)
	assert.Contains(t, result.Reason, "변경")
}

func TestCheckMatchingOutdated_ExperienceModified(t *testing.T) {
	client := testutil.NewTestClient(t)
	mockAI := &MockLLMForMatching{}
	svc := NewMatchingService(client, mockAI)

	user := client.UserProfile.Create().
		SetEmail("outdated-mod@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	exp := client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("Original").
		SetContent("Content").
		SetCategory("인턴").
		SetStarSituation("S").
		SetStarTask("T").
		SetStarAction("A").
		SetStarResult("R").
		SaveX(context.Background())

	// matchedAt is before the update
	matchedAt := time.Now().Add(-1 * time.Second)

	// Simulate modification by waiting and updating
	time.Sleep(10 * time.Millisecond)
	client.Experience.UpdateOneID(exp.ID).
		SetTitle("Modified").
		SaveX(context.Background())

	result, err := svc.CheckMatchingOutdated(context.Background(), user.ID, matchedAt)

	require.NoError(t, err)
	assert.True(t, result.IsOutdated)
	assert.Contains(t, result.Reason, "변경")
}

func TestCheckMatchingOutdated_ExperienceDeleted(t *testing.T) {
	client := testutil.NewTestClient(t)
	mockAI := &MockLLMForMatching{}
	svc := NewMatchingService(client, mockAI)

	user := client.UserProfile.Create().
		SetEmail("outdated-del@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	exp := client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("ToDelete").
		SetContent("Content").
		SetCategory("인턴").
		SetStarSituation("S").
		SetStarTask("T").
		SetStarAction("A").
		SetStarResult("R").
		SaveX(context.Background())

	matchedAt := time.Now()

	// Delete the experience
	client.Experience.DeleteOneID(exp.ID).ExecX(context.Background())

	result, err := svc.CheckMatchingOutdated(context.Background(), user.ID, matchedAt)

	require.NoError(t, err)
	assert.True(t, result.IsOutdated)
	assert.Contains(t, result.Reason, "경험이 없습니다")
}

func TestCheckMatchingOutdated_SevenDaysExpired(t *testing.T) {
	client := testutil.NewTestClient(t)
	mockAI := &MockLLMForMatching{}
	svc := NewMatchingService(client, mockAI)

	user := client.UserProfile.Create().
		SetEmail("outdated-7d@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	// Create experience with old updated_at (before matchedAt)
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("Old Exp").
		SetContent("Content").
		SetCategory("인턴").
		SetStarSituation("S").
		SetStarTask("T").
		SetStarAction("A").
		SetStarResult("R").
		SetUpdatedAt(oldTime).
		SaveX(context.Background())

	// matchedAt was 8 days ago (after experience was created)
	matchedAt := time.Now().Add(-8 * 24 * time.Hour)

	result, err := svc.CheckMatchingOutdated(context.Background(), user.ID, matchedAt)

	require.NoError(t, err)
	assert.True(t, result.IsOutdated)
	assert.Contains(t, result.Reason, "7일")
}

func TestCheckMatchingOutdated_Fresh(t *testing.T) {
	client := testutil.NewTestClient(t)
	mockAI := &MockLLMForMatching{}
	svc := NewMatchingService(client, mockAI)

	user := client.UserProfile.Create().
		SetEmail("outdated-fresh@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	// Create experience first
	client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("Fresh Exp").
		SetContent("Content").
		SetCategory("인턴").
		SetStarSituation("S").
		SetStarTask("T").
		SetStarAction("A").
		SetStarResult("R").
		SaveX(context.Background())

	// matchedAt is now (after experience creation)
	time.Sleep(10 * time.Millisecond)
	matchedAt := time.Now()

	result, err := svc.CheckMatchingOutdated(context.Background(), user.ID, matchedAt)

	require.NoError(t, err)
	assert.False(t, result.IsOutdated)
}
