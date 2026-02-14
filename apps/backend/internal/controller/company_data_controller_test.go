package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCompanyDataTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	companyDataSvc := service.NewCompanyDataService()
	ctrl := NewCompanyDataController(companyDataSvc)

	router := gin.New()
	router.GET("/v1/company-data", ctrl.GetCompanyData)

	return router
}

func TestGetCompanyData_MissingName(t *testing.T) {
	router := setupCompanyDataTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/company-data", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "VALID_001", errObj["code"])
}

func TestGetCompanyData_EmptyName(t *testing.T) {
	router := setupCompanyDataTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/company-data?name=", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCompanyData_ReturnsJSON(t *testing.T) {
	router := setupCompanyDataTestRouter(t)

	// Uses real DART/Naver which may fail in CI, so just verify response shape
	req := httptest.NewRequest(http.MethodGet, "/v1/company-data?name=삼성전자", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 (external APIs may fail gracefully) or 500
	if w.Code == http.StatusOK {
		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp, "basic_info")
		assert.Contains(t, resp, "news")
	}
}

func TestGetCompanyData_ResponseStructure(t *testing.T) {
	router := setupCompanyDataTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/company-data?name=테스트회사", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Even for unknown companies, service returns partial data (never nil)
	if w.Code == http.StatusOK {
		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

		// basic_info should exist in response
		assert.Contains(t, resp, "basic_info")
		assert.Contains(t, resp, "news")
	}
}
