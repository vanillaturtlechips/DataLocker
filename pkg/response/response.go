// Package response provides standardized HTTP response utilities for DataLocker API.
// It includes success, error, and various HTTP status code response helpers.
package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response 표준 API 응답 구조체
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 에러 정보 구조체
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Success 성공 응답을 반환합니다
func Success(c echo.Context, data interface{}, message string) error {
	if message == "" {
		message = "요청이 성공적으로 처리되었습니다"
	}

	return c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created 리소스 생성 성공 응답을 반환합니다
func Created(c echo.Context, data interface{}, message string) error {
	if message == "" {
		message = "리소스가 성공적으로 생성되었습니다"
	}

	return c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// BadRequest 잘못된 요청 에러 응답을 반환합니다
func BadRequest(c echo.Context, message string, details string) error {
	if message == "" {
		message = "잘못된 요청입니다"
	}

	return c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    "BAD_REQUEST",
			Message: message,
			Details: details,
		},
	})
}

// InternalError 내부 서버 에러 응답을 반환합니다
func InternalError(c echo.Context, message string, details string) error {
	if message == "" {
		message = "내부 서버 오류가 발생했습니다"
	}

	return c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    "INTERNAL_ERROR",
			Message: message,
			Details: details,
		},
	})
}

// NotFound 리소스를 찾을 수 없음 응답을 반환합니다
func NotFound(c echo.Context, message string) error {
	if message == "" {
		message = "요청한 리소스를 찾을 수 없습니다"
	}

	return c.JSON(http.StatusNotFound, Response{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    "NOT_FOUND",
			Message: message,
		},
	})
}

// Unauthorized 인증되지 않음 응답을 반환합니다
func Unauthorized(c echo.Context, message string) error {
	if message == "" {
		message = "인증이 필요합니다"
	}

	return c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}

// Forbidden 권한 없음 응답을 반환합니다
func Forbidden(c echo.Context, message string) error {
	if message == "" {
		message = "접근 권한이 없습니다"
	}

	return c.JSON(http.StatusForbidden, Response{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    "FORBIDDEN",
			Message: message,
		},
	})
}
