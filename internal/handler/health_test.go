package handler

import (
	"DataLocker/internal/config"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Health(t *testing.T) {
	// 테스트용 설정
	cfg := &config.Config{
		App: config.AppConfig{
			Name:    "DataLocker",
			Version: "2.0.0",
		},
	}

	// 핸들러 생성
	handler := NewHealthHandler(cfg)

	// Echo 인스턴스 생성
	e := echo.New()

	// 테스트 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 핸들러 실행
	err := handler.Health(c)

	// 검증
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 응답 JSON 파싱
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 응답 검증
	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	// 데이터 상세 검증
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "healthy", data["status"])
	assert.Equal(t, "DataLocker", data["app"])
	assert.Equal(t, "2.0.0", data["version"])
}

func TestHealthHandler_Ready(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:    "DataLocker",
			Version: "2.0.0",
		},
	}

	handler := NewHealthHandler(cfg)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Ready(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.True(t, data["ready"].(bool))
}

func TestHealthHandler_Live(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:    "DataLocker",
			Version: "2.0.0",
		},
	}

	handler := NewHealthHandler(cfg)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Live(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.True(t, data["alive"].(bool))
}

func TestHealthHandler_Metrics(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:    "DataLocker",
			Version: "2.0.0",
		},
	}

	handler := NewHealthHandler(cfg)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Metrics(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["memory"])
	assert.NotNil(t, data["goroutines"])
	assert.NotNil(t, data["uptime"])
}
