package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coby/colight/apps/backend/ent"
	"github.com/coby/colight/apps/backend/internal/infrastructure/ai"
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/coby/colight/apps/backend/testutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMProvider for controller tests
type MockLLMProviderCtrl struct {
	response ai.LLMResponse
	err      error
}

func (m *MockLLMProviderCtrl) Call(ctx context.Context, req ai.LLMRequest) (ai.LLMResponse, error) {
	return m.response, m.err
}

func setupWeaponTaggingTestRouter(t *testing.T, mockAI *MockLLMProviderCtrl) (*gin.Engine, *ent.Client) {
	t.Helper()
	client := testutil.NewTestClient(t)

	weaponSvc := service.NewWeaponTaggingService(client, ai.NewAIProviderForTest(mockAI))
	weaponCtrl := NewWeaponTaggingController(weaponSvc)

	router := gin.New()
	router.POST("/v1/experiences/:id/tag", func(c *gin.Context) {
		// Simulate auth middleware - extract user_id from X-Test-UserID header
		userIDStr := c.GetHeader("X-Test-UserID")
		if userIDStr != "" {
			userID, _ := uuid.Parse(userIDStr)
			c.Set("user_id", userID)
		}
		weaponCtrl.Tag(c)
	})

	return router, client
}

func TestWeaponTaggingController_Tag(t *testing.T) {
	mockAI := &MockLLMProviderCtrl{
		response: ai.LLMResponse{
			Content: `{
				"primary_weapon": {
					"code": "W01",
					"confidence": 0.9,
					"reasoning": "Test reasoning"
				},
				"secondary_weapons": []
			}`,
		},
	}

	router, client := setupWeaponTaggingTestRouter(t, mockAI)

	// Create user
	user := client.UserProfile.Create().
		SetEmail("wtc-tag-" + uuid.New().String()[:8] + "@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	// Create weapon category
	_ = client.WeaponCategory.Create().
		SetCode("W01").
		SetParentCode("W01").
		SetName("위기극복").
		SetDescription("Test").
		SetKeywords([]string{"test"}).
		SetIcon("🔥").
		SaveX(context.Background())

	// Create prompt template
	_ = client.PromptTemplate.Create().
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
		SaveX(context.Background())

	// Create experience
	exp := client.Experience.Create().
		SetUserID(user.ID).
		SetTitle("Test Experience").
		SetCategory("프로젝트").
		SetContent("This is a test experience with enough content to pass validation").
		SetStarSituation("Situation").
		SetStarTask("Task").
		SetStarAction("Action").
		SetStarResult("Result").
		SaveX(context.Background())

	// Make request
	req := httptest.NewRequest("POST", "/v1/experiences/"+exp.ID.String()+"/tag", nil)
	req.Header.Set("X-Test-UserID", user.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	primaryWeapon := resp["primary_weapon"].(map[string]interface{})
	assert.Equal(t, "W01", primaryWeapon["code"])
	assert.Equal(t, 0.9, primaryWeapon["confidence"])
}

func TestWeaponTaggingController_Unauthorized(t *testing.T) {
	mockAI := &MockLLMProviderCtrl{}
	router, _ := setupWeaponTaggingTestRouter(t, mockAI)

	// Make request without auth header
	req := httptest.NewRequest("POST", "/v1/experiences/"+uuid.New().String()+"/tag", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWeaponTaggingController_Forbidden(t *testing.T) {
	mockAI := &MockLLMProviderCtrl{}
	router, client := setupWeaponTaggingTestRouter(t, mockAI)

	// Create owner
	owner := client.UserProfile.Create().
		SetEmail("wtc-owner-" + uuid.New().String()[:8] + "@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	// Create other user
	otherUser := client.UserProfile.Create().
		SetEmail("wtc-other-" + uuid.New().String()[:8] + "@example.com").
		SetPasswordHash("hash").
		SetRole("user").
		SaveX(context.Background())

	// Create experience owned by owner
	exp := client.Experience.Create().
		SetUserID(owner.ID).
		SetTitle("Owner's Experience").
		SetCategory("프로젝트").
		SetContent("Content").
		SetStarSituation("Situation").
		SetStarTask("Task").
		SetStarAction("Action").
		SetStarResult("Result").
		SaveX(context.Background())

	// Try to tag with other user
	req := httptest.NewRequest("POST", "/v1/experiences/"+exp.ID.String()+"/tag", nil)
	req.Header.Set("X-Test-UserID", otherUser.ID.String())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert 403
	assert.Equal(t, http.StatusForbidden, w.Code)
}
