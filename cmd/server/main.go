package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"DataLocker/internal/config"
	"DataLocker/internal/handler"
	"DataLocker/internal/middleware"
)

func main() {
	// 설정 로드
	cfg := config.Load()

	// 로거 설정
	logger := setupLogger(cfg)

	// Echo 인스턴스 생성
	e := echo.New()

	// 배너 숨기기
	e.HideBanner = true

	// 미들웨어 설정
	middleware.SetupMiddleware(e, cfg, logger)

	// 에러 핸들러 설정
	e.HTTPErrorHandler = middleware.ErrorHandlingMiddleware(logger)

	// 핸들러 초기화
	healthHandler := handler.NewHealthHandler(cfg)

	// 라우트 설정
	setupRoutes(e, healthHandler)

	// 서버 시작
	startServer(e, cfg, logger)
}

// setupLogger 로거를 설정합니다
func setupLogger(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()

	// 로그 레벨 설정
	switch cfg.App.LogLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// 개발환경에서는 텍스트 포맷, 운영환경에서는 JSON 포맷
	if cfg.App.Environment == "development" {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	return logger
}

// setupRoutes 라우트를 설정합니다
func setupRoutes(e *echo.Echo, healthHandler *handler.HealthHandler) {
	// API 버전 그룹
	api := e.Group("/api/v1")

	// 헬스체크 라우트
	health := api.Group("/health")
	health.GET("", healthHandler.Health)
	health.GET("/ready", healthHandler.Ready)
	health.GET("/live", healthHandler.Live)
	health.GET("/metrics", healthHandler.Metrics)

	// 루트 경로
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "DataLocker API Server",
			"version": "2.0.0",
			"status":  "running",
			"docs":    "/api/v1/health",
		})
	})

	// API 문서 경로 (추후 Swagger 연동)
	e.GET("/docs", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "API Documentation",
			"endpoints": map[string]interface{}{
				"health":  "/api/v1/health",
				"ready":   "/api/v1/health/ready",
				"live":    "/api/v1/health/live",
				"metrics": "/api/v1/health/metrics",
			},
		})
	})
}

// startServer 서버를 시작합니다
func startServer(e *echo.Echo, cfg *config.Config, logger *logrus.Logger) {
	// 서버 주소
	address := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	// Graceful Shutdown을 위한 고루틴
	go func() {
		logger.WithFields(logrus.Fields{
			"address":     address,
			"environment": cfg.App.Environment,
			"version":     cfg.App.Version,
		}).Info("서버를 시작합니다")

		// 서버 시작
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("서버 시작에 실패했습니다")
		}
	}()

	// 종료 신호 대기
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("서버를 종료합니다...")

	// Graceful Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("서버 종료 중 오류가 발생했습니다")
	} else {
		logger.Info("서버가 정상적으로 종료되었습니다")
	}
}
