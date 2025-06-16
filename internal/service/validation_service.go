// Package service provides business logic for DataLocker.
// This file implements file and directory validation.
package service

import (
	"context"
	"fmt"
	"strings"
)

// validationService 파일/디렉터리 검증 서비스 구현체
type validationService struct{}

// NewValidationService 새로운 검증 서비스를 생성합니다
func NewValidationService() ValidationService {
	return &validationService{}
}

// ValidateItem 파일 또는 디렉터리를 검증합니다
func (s *validationService) ValidateItem(ctx context.Context, req *ValidationRequest) (*ValidationResult, error) {
	switch req.Type {
	case ItemTypeFile:
		return s.validateSingleFile(req)
	case ItemTypeDirectory:
		return s.validateDirectoryInternal(req)
	default:
		return nil, fmt.Errorf("지원하지 않는 타입입니다: %s", req.Type)
	}
}

// ValidateFile 단일 파일 검증 (기존 호환성)
func (s *validationService) ValidateFile(ctx context.Context, fileName string, fileSize int64, mimeType string) (*FileValidationResult, error) {
	result := &FileValidationResult{
		FileName: fileName,
		IsValid:  true,
		Errors:   make([]string, 0),
	}

	// 기본 검증들
	if fileName == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "파일명이 비어있습니다")
	}

	if fileSize <= MinFileSize {
		result.IsValid = false
		result.Errors = append(result.Errors, "파일이 너무 작습니다")
	}

	if fileSize > MaxFileSize {
		result.IsValid = false
		result.Errors = append(result.Errors, "파일이 너무 큽니다")
	}

	if !s.isAllowedMimeType(mimeType) {
		result.IsValid = false
		result.Errors = append(result.Errors, "지원하지 않는 파일 형식입니다")
	}

	return result, nil
}

// ValidateDirectory 디렉터리 전체를 검증
func (s *validationService) ValidateDirectory(ctx context.Context, directoryPath string, files []FileInfo) (*ValidationResult, error) {
	result := &ValidationResult{
		Type:        ItemTypeDirectory,
		IsValid:     true,
		TotalFiles:  len(files),
		FileResults: make([]FileValidationResult, 0),
		Errors:      make([]string, 0),
	}

	// 1. 디렉터리 기본 검증
	if directoryPath == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "디렉터리 경로가 비어있습니다")
		return result, nil
	}

	// 2. 파일 개수 제한
	if len(files) > MaxFileCount {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("파일이 너무 많습니다 (최대 %d개)", MaxFileCount))
	}

	// 3. 각 파일 검증
	var totalSize int64
	for _, file := range files {
		fileResult, err := s.ValidateFile(ctx, file.Name, file.Size, file.MimeType)
		if err != nil {
			continue // 에러난 파일은 건너뛰기
		}

		fileResult.RelativePath = file.RelativePath
		result.FileResults = append(result.FileResults, *fileResult)

		totalSize += file.Size

		if fileResult.IsValid {
			result.ValidFiles++
		} else {
			result.InvalidFiles++
			result.IsValid = false // 하나라도 실패하면 전체 실패
		}
	}

	// 4. 전체 크기 검증
	result.TotalSize = totalSize
	if totalSize > MaxDirectorySize {
		result.IsValid = false
		result.Errors = append(result.Errors, "디렉터리 전체 크기가 너무 큽니다")
	}

	return result, nil
}

// 내부 헬퍼 메서드들

// validateSingleFile 단일 파일 검증 (내부용)
func (s *validationService) validateSingleFile(req *ValidationRequest) (*ValidationResult, error) {
	fileResult, err := s.ValidateFile(context.Background(), req.FileName, req.FileSize, req.MimeType)
	if err != nil {
		return nil, err
	}

	result := &ValidationResult{
		Type:         ItemTypeFile,
		IsValid:      fileResult.IsValid,
		TotalFiles:   1,
		TotalSize:    req.FileSize,
		ValidFiles:   0,
		InvalidFiles: 0,
		Errors:       fileResult.Errors,
		FileResults:  []FileValidationResult{*fileResult},
	}

	if fileResult.IsValid {
		result.ValidFiles = 1
	} else {
		result.InvalidFiles = 1
	}

	return result, nil
}

// validateDirectoryInternal 디렉터리 검증 (내부용)
func (s *validationService) validateDirectoryInternal(req *ValidationRequest) (*ValidationResult, error) {
	return s.ValidateDirectory(context.Background(), req.DirectoryPath, req.Files)
}

// isAllowedMimeType 허용된 MIME 타입인지 확인
func (s *validationService) isAllowedMimeType(mimeType string) bool {
	for _, allowed := range AllowedMimeTypes {
		if strings.EqualFold(mimeType, allowed) {
			return true
		}
	}
	return false
}
