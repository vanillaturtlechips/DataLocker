// Package service provides business logic for DataLocker.
// This file defines validation interface for files and directories.
package service

import "context"

// ValidationService 파일/디렉터리 검증 서비스
type ValidationService interface {
	// ValidateItem 파일 또는 디렉터리를 검증합니다
	ValidateItem(ctx context.Context, req *ValidationRequest) (*ValidationResult, error)

	// ValidateFile 단일 파일만 검증 (기존 호환성)
	ValidateFile(ctx context.Context, fileName string, fileSize int64, mimeType string) (*FileValidationResult, error)

	// ValidateDirectory 디렉터리 전체를 검증
	ValidateDirectory(ctx context.Context, directoryPath string, files []FileInfo) (*ValidationResult, error)
}
