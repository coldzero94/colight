package service

import (
	"context"
	"testing"

	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Removed unused imports: ent, uuid

// MockLLMForMatching mocks AI for matching tests
type MockLLMForMatching struct {
	response ai.LLMResponse
	err      error
}

func (m *MockLLMForMatching) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	return m.response, m.err
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
