package middleware

import (
	"DataLocker/internal/config"
	"DataLocker/pkg/response"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

// SetupMiddleware 모든 미들웨어를 설정합니다
func SetupMiddleware(e *echo.Echo, cfg *config.Config, logger *logrus.Logger) {
	// Recovery 미들웨어 - 패닉 복구
	e.Use(RecoveryMiddleware(logger))

	// CORS 미들웨어
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.Security.AllowedOrigins,
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
		MaxAge:           86400, // 24시간
	}))

	// 요청 로깅 미들웨어
	e.Use(RequestLoggingMiddleware(logger))

	// 응답 시간 측정 미들웨어
	e.Use(ResponseTimeMiddleware(logger))

	// Body Limit 미들웨어
	e.Use(middleware.BodyLimit(fmt.Sprintf("%d", cfg.Security.MaxFileSize)))

	// Rate Limiting (개발환경에서는 비활성화)
	if cfg.App.Environment == "production" {
		e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(100)))
	}

	// 보안 헤더 미들웨어
	e.Use(SecurityHeadersMiddleware())
}

// RecoveryMiddleware 패닉을 복구하고 로깅합니다
func RecoveryMiddleware(logger *logrus.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					logger.WithFields(logrus.Fields{
						"panic":      r,
						"stack":      string(debug.Stack()),
						"method":     c.Request().Method,
						"uri":        c.Request().RequestURI,
						"ip":         c.RealIP(),
						"user_agent": c.Request().UserAgent(),
					}).Error("패닉이 발생했습니다")

					// 클라이언트에게 에러 응답 전송
					response.InternalError(c, "서버에서 예상치 못한 오류가 발생했습니다", "")
				}
			}()
			return next(c)
		}
	}
}

// RequestLoggingMiddleware 요청을 로깅합니다
func RequestLoggingMiddleware(logger *logrus.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start)

			entry := logger.WithFields(logrus.Fields{
				"method":      c.Request().Method,
				"uri":         c.Request().RequestURI,
				"status":      c.Response().Status,
				"ip":          c.RealIP(),
				"user_agent":  c.Request().UserAgent(),
				"duration_ms": duration.Milliseconds(),
				"bytes_in":    c.Request().ContentLength,
				"bytes_out":   c.Response().Size,
			})

			if err != nil {
				entry.WithError(err).Error("요청 처리 중 오류가 발생했습니다")
			} else {
				if c.Response().Status >= 400 {
					entry.Warn("클라이언트 오류 응답")
				} else {
					entry.Info("요청 처리 완료")
				}
			}

			return err
		}
	}
}

// ResponseTimeMiddleware 응답 시간을 헤더에 추가합니다
func ResponseTimeMiddleware(logger *logrus.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start)
			c.Response().Header().Set("X-Response-Time", duration.String())

			// 느린 요청 경고 (1초 이상)
			if duration > time.Second {
				logger.WithFields(logrus.Fields{
					"method":      c.Request().Method,
					"uri":         c.Request().RequestURI,
					"duration_ms": duration.Milliseconds(),
				}).Warn("느린 요청이 감지되었습니다")
			}

			return err
		}
	}
}

// SecurityHeadersMiddleware 보안 헤더를 추가합니다
func SecurityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// XSS Protection
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			// Content Type Options
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")

			// Frame Options
			c.Response().Header().Set("X-Frame-Options", "DENY")

			// Content Security Policy
			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'")

			// Referrer Policy
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			return next(c)
		}
	}
}

// ErrorHandlingMiddleware 전역 에러 핸들링 미들웨어
func ErrorHandlingMiddleware(logger *logrus.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		// Echo HTTP Error 처리
		if he, ok := err.(*echo.HTTPError); ok {
			switch he.Code {
			case 400:
				response.BadRequest(c, fmt.Sprintf("%v", he.Message), "")
			case 401:
				response.Unauthorized(c, fmt.Sprintf("%v", he.Message))
			case 403:
				response.Forbidden(c, fmt.Sprintf("%v", he.Message))
			case 404:
				response.NotFound(c, fmt.Sprintf("%v", he.Message))
			default:
				response.InternalError(c, fmt.Sprintf("%v", he.Message), "")
			}
		} else {
			// 일반 에러 처리
			logger.WithFields(logrus.Fields{
				"error":  err.Error(),
				"method": c.Request().Method,
				"uri":    c.Request().RequestURI,
				"ip":     c.RealIP(),
			}).Error("처리되지 않은 에러가 발생했습니다")

			response.InternalError(c, "내부 서버 오류가 발생했습니다", err.Error())
		}
	}
}
