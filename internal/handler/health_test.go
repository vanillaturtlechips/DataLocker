package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"DataLocker/internal/config"
)

// createTestConfig creates a test configuration
func createTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:    "DataLocker",
			Version: "2.0.0",
		},
	}
}

// createTestContext creates a test Echo context
func createTestContext(method, path string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, http.NoBody)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// assertSuccessResponse asserts that the response is successful
func assertSuccessResponse(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	return response
}

func TestHealthHandler_Health(t *testing.T) {
	handler := NewHealthHandler(createTestConfig())
	c, rec := createTestContext(http.MethodGet, "/health")

	err := handler.Health(c)
	assert.NoError(t, err)

	response := assertSuccessResponse(t, rec)

	// 응답 데이터 검증
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "healthy", data["status"])
	assert.Equal(t, "DataLocker", data["app"])
	assert.Equal(t, "2.0.0", data["version"])
}

func TestHealthHandler_Ready(t *testing.T) {
	handler := NewHealthHandler(createTestConfig())
	c, rec := createTestContext(http.MethodGet, "/ready")

	err := handler.Ready(c)
	assert.NoError(t, err)

	response := assertSuccessResponse(t, rec)

	data := response["data"].(map[string]interface{})
	assert.True(t, data["ready"].(bool))
}

func TestHealthHandler_Live(t *testing.T) {
	handler := NewHealthHandler(createTestConfig())
	c, rec := createTestContext(http.MethodGet, "/live")

	err := handler.Live(c)
	assert.NoError(t, err)

	response := assertSuccessResponse(t, rec)

	data := response["data"].(map[string]interface{})
	assert.True(t, data["alive"].(bool))
}

func TestHealthHandler_Metrics(t *testing.T) {
	handler := NewHealthHandler(createTestConfig())
	c, rec := createTestContext(http.MethodGet, "/metrics")

	err := handler.Metrics(c)
	assert.NoError(t, err)

	response := assertSuccessResponse(t, rec)

	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["memory"])
	assert.NotNil(t, data["goroutines"])
	assert.NotNil(t, data["uptime"])
}
