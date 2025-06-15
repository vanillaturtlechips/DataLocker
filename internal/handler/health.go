// Package handler provides HTTP request handlers for DataLocker API endpoints.
// It includes health check, metrics, and various API route handlers.
package handler

import (
	"runtime"
	"time"

	"DataLocker/internal/config"
	"DataLocker/pkg/response"

	"github.com/labstack/echo/v4"
)

// HealthHandler 헬스체크 핸들러
type HealthHandler struct {
	config    *config.Config
	startTime time.Time
}

// NewHealthHandler 새로운 헬스체크 핸들러를 생성합니다
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		config:    cfg,
		startTime: time.Now(),
	}
}

// HealthResponse 헬스체크 응답 구조체
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Version   string                 `json:"version"`
	App       string                 `json:"app"`
	System    SystemInfo             `json:"system"`
	Services  map[string]ServiceInfo `json:"services"`
}

// SystemInfo 시스템 정보 구조체
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
}

// ServiceInfo 서비스 상태 정보 구조체
type ServiceInfo struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Health 기본 헬스체크 엔드포인트
func (h *HealthHandler) Health(c echo.Context) error {
	uptime := time.Since(h.startTime)

	healthData := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
		Version:   h.config.App.Version,
		App:       h.config.App.Name,
		System: SystemInfo{
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
			NumCPU:       runtime.NumCPU(),
			OS:           runtime.GOOS,
			Arch:         runtime.GOARCH,
		},
		Services: map[string]ServiceInfo{
			"api": {
				Status: "healthy",
			},
			"database": {
				Status: "healthy", // TODO: 실제 DB 연결 체크
			},
			"filesystem": {
				Status: "healthy", // TODO: 파일시스템 체크
			},
		},
	}

	return response.Success(c, healthData, "서비스가 정상적으로 동작 중입니다")
}

// Ready 준비 상태 체크 엔드포인트
func (h *HealthHandler) Ready(c echo.Context) error {
	// TODO: 실제 준비 상태 체크 로직 구현
	// - 데이터베이스 연결 확인
	// - 필수 서비스 확인
	// - 설정 파일 로드 확인

	readyData := map[string]interface{}{
		"ready":     true,
		"timestamp": time.Now(),
		"checks": map[string]bool{
			"database":   true, // TODO: 실제 체크
			"filesystem": true, // TODO: 실제 체크
			"config":     true,
		},
	}

	return response.Success(c, readyData, "서비스 준비 완료")
}

// Live 라이브니스 체크 엔드포인트
func (h *HealthHandler) Live(c echo.Context) error {
	// 간단한 라이브니스 체크
	liveData := map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now(),
		"uptime":    time.Since(h.startTime).String(),
	}

	return response.Success(c, liveData, "서비스가 살아있습니다")
}

// Metrics 기본 메트릭 정보 엔드포인트
func (h *HealthHandler) Metrics(c echo.Context) error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metricsData := map[string]interface{}{
		"memory": map[string]interface{}{
			"alloc":           memStats.Alloc,
			"total_alloc":     memStats.TotalAlloc,
			"sys":             memStats.Sys,
			"num_gc":          memStats.NumGC,
			"gc_cpu_fraction": memStats.GCCPUFraction,
		},
		"goroutines": runtime.NumGoroutine(),
		"uptime":     time.Since(h.startTime).String(),
		"timestamp":  time.Now(),
	}

	return response.Success(c, metricsData, "메트릭 정보")
}
