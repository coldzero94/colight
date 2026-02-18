package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/ent/experience"
	"github.com/coby/colight/apps/backend/ent/experienceweapon"
	"github.com/coby/colight/apps/backend/ent/prompttemplate"
	"github.com/coby/colight/apps/backend/ent/weaponcategory"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMProvider is a mock implementation of ai.LLMProvider for testing
type MockLLMProvider struct {
	response ai.LLMResponse
	err      error
}

func (m *MockLLMProvider) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	return m.response, m.err
}

func newTestWeaponTaggingService(t *testing.T, mockAI *MockLLMProvider) (*WeaponTaggingService, *ent.Client) {
	t.Helper()
	client := testutil.NewTestClient(t)
	aiProvider := ai.NewAIProviderForTest(mockAI)
	svc := NewWeaponTaggingService(client, aiProvider)
	return svc, client
}

func createTestUserForWeapon(t *testing.T, client *ent.Client) uuid.UUID {
	t.Helper()
	user := client.UserProfile.Create().
		SetEmail("test-" + uuid.New().String()[:8] + "@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())
	return user.ID
}

func ensureWeaponCategories(t *testing.T, client *ent.Client) {
	t.Helper()
	ctx := context.Background()

	weapons := []struct {
		code string
		name string
		icon string
	}{
		{"W01", "위기극복", "🔥"},
		{"W02", "리더십", "👑"},
	}

	for _, w := range weapons {
		// Check if exists
		exists, _ := client.WeaponCategory.Query().
			Where(weaponcategory.CodeEQ(w.code)).
			Exist(ctx)

		if !exists {
			client.WeaponCategory.Create().
				SetCode(w.code).
				SetParentCode(w.code).
				SetName(w.name).
				SetKeywords([]string{"test"}).
				SetIcon(w.icon).
				SaveX(ctx)
		}
	}

	// Ensure prompt template exists
	promptExists, _ := client.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("experience_classify"),
			prompttemplate.SubCategoryEQ("weapon_tagging"),
		).
		Exist(ctx)

	if !promptExists {
		client.PromptTemplate.Create().
			SetCategory("experience_classify").
			SetSubCategory("weapon_tagging").
			SetName("Test Weapon Tagging").
			SetSystemPrompt("Analyze and classify weapons").
			SetUserPromptTemplate("Experience: {{experience_text}}\n\nWeapons: {{weapon_categories}}").
			SetModel("gemini-2.0-flash").
			SetTemperature(0.2).
			SetMaxTokens(2000).
			SetVersion(1).
			SetIsActive(true).
			SaveX(ctx)
	}
}

func TestTagExperience_Success(t *testing.T) {
	// Mock AI response
	mockAI := &MockLLMProvider{
		response: ai.LLMResponse{
			Content: `{
				"primary_weapon": {
					"code": "W01",
					"confidence": 0.9,
					"reasoning": "위기 상황을 극복한 사례"
				},
				"secondary_weapons": [
					{
						"code": "W02",
						"confidence": 0.7,
						"reasoning": "팀을 이끈 리더십 발휘"
					}
				]
			}`,
			InputTokens:  100,
			OutputTokens: 50,
			Model:        "gemini-2.0-flash",
		},
	}

	svc, client := newTestWeaponTaggingService(t, mockAI)
	ctx := context.Background()
	userID := createTestUserForWeapon(t, client)
	ensureWeaponCategories(t, client)

	// Create test experience
	exp := client.Experience.Create().
		SetUserID(userID).
		SetTitle("Test Experience").
		SetCategory("프로젝트").
		SetContent("경험 내용").
		SetStarSituation("상황").
		SetStarTask("과제").
		SetStarAction("행동").
		SetStarResult("결과").
		SaveX(ctx)

	// Execute tagging
	result, err := svc.TagExperience(ctx, exp.ID, userID)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify result structure
	assert.Equal(t, "W01", result.PrimaryWeapon.Code)
	assert.Equal(t, 0.9, result.PrimaryWeapon.Confidence)
	assert.Len(t, result.SecondaryWeapons, 1)
	assert.Equal(t, "W02", result.SecondaryWeapons[0].Code)

	// Verify DB records
	weapons, err := client.ExperienceWeapon.Query().
		Where(experienceweapon.HasExperienceWith(experience.IDEQ(exp.ID))).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, weapons, 2)

	// Find primary weapon
	var primary *ent.ExperienceWeapon
	for _, w := range weapons {
		if w.IsPrimary {
			primary = w
			break
		}
	}
	require.NotNil(t, primary)
	assert.Equal(t, "W01", primary.WeaponCode)
	assert.Equal(t, 0.9, primary.Confidence)
	assert.False(t, primary.UserConfirmed)
	assert.False(t, primary.UserModified)
}

func TestTagExperience_ShortExperience(t *testing.T) {
	mockAI := &MockLLMProvider{}
	svc, client := newTestWeaponTaggingService(t, mockAI)
	ctx := context.Background()
	userID := createTestUserForWeapon(t, client)

	// Create experience with very short content (< 50 chars)
	exp := client.Experience.Create().
		SetUserID(userID).
		SetTitle("짧음").
		SetCategory("기타").
		SetContent("짧음").
		SetStarSituation("짧음").
		SetStarTask("짧음").
		SetStarAction("짧음").
		SetStarResult("짧음").
		SaveX(ctx)

	// Should return error for short experience
	_, err := svc.TagExperience(ctx, exp.ID, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

func TestTagExperience_AIError(t *testing.T) {
	// Mock AI error (rate limit)
	mockAI := &MockLLMProvider{
		err: fmt.Errorf("rate limit exceeded"),
	}

	svc, client := newTestWeaponTaggingService(t, mockAI)
	ctx := context.Background()
	userID := createTestUserForWeapon(t, client)

	exp := client.Experience.Create().
		SetUserID(userID).
		SetTitle("Test Experience").
		SetCategory("프로젝트").
		SetContent("경험 내용").
		SetStarSituation("상황입니다. 이것은 충분히 긴 텍스트입니다.").
		SetStarTask("과제").
		SetStarAction("행동").
		SetStarResult("결과").
		SaveX(ctx)

	_, err := svc.TagExperience(ctx, exp.ID, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit")
}

func TestTagExperience_InvalidWeaponCode(t *testing.T) {
	// Mock AI response with invalid weapon code
	mockAI := &MockLLMProvider{
		response: ai.LLMResponse{
			Content: `{
				"primary_weapon": {
					"code": "W99",
					"confidence": 0.9,
					"reasoning": "Invalid code"
				},
				"secondary_weapons": []
			}`,
		},
	}

	svc, client := newTestWeaponTaggingService(t, mockAI)
	ctx := context.Background()
	userID := createTestUserForWeapon(t, client)

	exp := client.Experience.Create().
		SetUserID(userID).
		SetTitle("Test Experience").
		SetCategory("프로젝트").
		SetContent("경험 내용").
		SetStarSituation("상황입니다. 이것은 충분히 긴 텍스트입니다.").
		SetStarTask("과제").
		SetStarAction("행동").
		SetStarResult("결과").
		SaveX(ctx)

	_, err := svc.TagExperience(ctx, exp.ID, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid weapon code")
}

func TestTagExperience_Forbidden(t *testing.T) {
	mockAI := &MockLLMProvider{}
	svc, client := newTestWeaponTaggingService(t, mockAI)
	ctx := context.Background()

	ownerID := createTestUserForWeapon(t, client)
	otherUserID := createTestUserForWeapon(t, client)

	exp := client.Experience.Create().
		SetUserID(ownerID).
		SetTitle("Owner's Experience").
		SetCategory("프로젝트").
		SetContent("경험 내용").
		SetStarSituation("상황").
		SetStarTask("과제").
		SetStarAction("행동").
		SetStarResult("결과").
		SaveX(ctx)

	// Try to tag other user's experience
	_, err := svc.TagExperience(ctx, exp.ID, otherUserID)
	require.Error(t, err)
	assert.Equal(t, ErrExperienceForbidden, err)
}

func TestRetagExperience_SkipUserConfirmed(t *testing.T) {
	mockAI := &MockLLMProvider{
		response: ai.LLMResponse{
			Content: `{
				"primary_weapon": {
					"code": "W02",
					"confidence": 0.8,
					"reasoning": "New reasoning"
				},
				"secondary_weapons": []
			}`,
		},
	}

	svc, client := newTestWeaponTaggingService(t, mockAI)
	ctx := context.Background()
	userID := createTestUserForWeapon(t, client)
	ensureWeaponCategories(t, client)

	// Create experience
	exp := client.Experience.Create().
		SetUserID(userID).
		SetTitle("Test Experience").
		SetCategory("프로젝트").
		SetContent("경험 내용").
		SetStarSituation("상황").
		SetStarTask("과제").
		SetStarAction("행동").
		SetStarResult("결과").
		SaveX(ctx)

	// Create user-confirmed weapon (should NOT be deleted on retag)
	confirmedWeapon := client.ExperienceWeapon.Create().
		SetExperienceID(exp.ID).
		SetWeaponCode("W01").
		SetConfidence(1.0).
		SetIsPrimary(true).
		SetUserConfirmed(true).
		SetUserModified(false).
		SaveX(ctx)

	// Re-tag the experience
	result, err := svc.TagExperience(ctx, exp.ID, userID)
	require.NoError(t, err)

	// AI suggested W02, but user-confirmed W01 should be preserved
	weapons, err := client.ExperienceWeapon.Query().
		Where(experienceweapon.HasExperienceWith(experience.IDEQ(exp.ID))).
		All(ctx)
	require.NoError(t, err)

	// Should have user-confirmed weapon preserved
	found := false
	for _, w := range weapons {
		if w.ID == confirmedWeapon.ID {
			found = true
			assert.True(t, w.UserConfirmed)
		}
	}
	assert.True(t, found, "User-confirmed weapon should be preserved")

	// Result should show the preserved user-confirmed weapon (AI call skipped)
	assert.Equal(t, "W01", result.PrimaryWeapon.Code)
}

func TestRetagExperience_DeleteOnlyAITags(t *testing.T) {
	mockAI := &MockLLMProvider{
		response: ai.LLMResponse{
			Content: `{
				"primary_weapon": {
					"code": "W02",
					"confidence": 0.9,
					"reasoning": "New AI reasoning"
				},
				"secondary_weapons": []
			}`,
		},
	}

	svc, client := newTestWeaponTaggingService(t, mockAI)
	ctx := context.Background()
	userID := createTestUserForWeapon(t, client)
	ensureWeaponCategories(t, client)

	// Create experience
	exp := client.Experience.Create().
		SetUserID(userID).
		SetTitle("Test Experience").
		SetCategory("프로젝트").
		SetContent("경험 내용").
		SetStarSituation("상황").
		SetStarTask("과제").
		SetStarAction("행동").
		SetStarResult("결과").
		SaveX(ctx)

	// Create AI-tagged weapon (user_modified=false, should be deleted)
	aiWeapon := client.ExperienceWeapon.Create().
		SetExperienceID(exp.ID).
		SetWeaponCode("W01").
		SetConfidence(0.8).
		SetIsPrimary(true).
		SetUserConfirmed(false).
		SetUserModified(false).
		SaveX(ctx)

	// Create user-modified weapon (should be preserved)
	userWeapon := client.ExperienceWeapon.Create().
		SetExperienceID(exp.ID).
		SetWeaponCode("W01").
		SetConfidence(1.0).
		SetIsPrimary(false).
		SetUserConfirmed(false).
		SetUserModified(true).
		SaveX(ctx)

	// Simulate content change by touching the experience after weapons were created
	time.Sleep(10 * time.Millisecond)
	client.Experience.UpdateOneID(exp.ID).SetContent("경험 내용이 수정됨").SaveX(ctx)

	// Re-tag
	_, err := svc.TagExperience(ctx, exp.ID, userID)
	require.NoError(t, err)

	// Check weapons
	weapons, err := client.ExperienceWeapon.Query().
		Where(experienceweapon.HasExperienceWith(experience.IDEQ(exp.ID))).
		All(ctx)
	require.NoError(t, err)

	// AI weapon should be deleted, user weapon preserved, new AI weapon added
	foundAI := false
	foundUser := false
	foundNew := false
	for _, w := range weapons {
		if w.ID == aiWeapon.ID {
			foundAI = true // Should NOT be found
		}
		if w.ID == userWeapon.ID {
			foundUser = true // Should be found
		}
		if w.WeaponCode == "W02" && w.IsPrimary {
			foundNew = true // New AI weapon
		}
	}

	assert.False(t, foundAI, "Old AI weapon should be deleted")
	assert.True(t, foundUser, "User-modified weapon should be preserved")
	assert.True(t, foundNew, "New AI weapon should be created")
}

func TestTagExperience_LoadsPromptFromDB(t *testing.T) {
	mockAI := &MockLLMProvider{
		response: ai.LLMResponse{
			Content: `{
				"primary_weapon": {
					"code": "W01",
					"confidence": 0.9,
					"reasoning": "Test"
				},
				"secondary_weapons": []
			}`,
		},
	}

	svc, client := newTestWeaponTaggingService(t, mockAI)
	ctx := context.Background()
	userID := createTestUserForWeapon(t, client)
	ensureWeaponCategories(t, client) // This creates the prompt template

	// Get the prompt template created by helper
	prompt, err2 := client.PromptTemplate.Query().
		Where(
			prompttemplate.CategoryEQ("experience_classify"),
			prompttemplate.SubCategoryEQ("weapon_tagging"),
		).
		First(ctx)
	require.NoError(t, err2)

	initialUsageCount := prompt.UsageCount

	exp := client.Experience.Create().
		SetUserID(userID).
		SetTitle("Test Experience").
		SetCategory("프로젝트").
		SetContent("This is a detailed experience content for testing").
		SetStarSituation("Situation").
		SetStarTask("Task").
		SetStarAction("Action").
		SetStarResult("Result").
		SaveX(ctx)

	// Execute tagging
	_, err := svc.TagExperience(ctx, exp.ID, userID)
	require.NoError(t, err)

	// Verify prompt was loaded (check usage stats updated)
	updatedPrompt, err := client.PromptTemplate.Get(ctx, prompt.ID)
	require.NoError(t, err)
	assert.Equal(t, initialUsageCount+1, updatedPrompt.UsageCount, "Prompt usage count should be incremented by 1")
}
