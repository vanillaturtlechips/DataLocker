// Package service provides business logic for DataLocker.
// This file defines validation DTOs for files and directories.
package service

// ItemType 검증 대상 타입
type ItemType string

const (
	ItemTypeFile      ItemType = "file"
	ItemTypeDirectory ItemType = "directory"
)

// ValidationRequest 파일 또는 디렉터리 검증 요청
type ValidationRequest struct {
	Type ItemType `json:"type"` // "file" 또는 "directory"
	Path string   `json:"path"` // 파일 경로 또는 디렉터리 경로

	// 파일인 경우
	FileName string `json:"file_name,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
	MimeType string `json:"mime_type,omitempty"`

	// 디렉터리인 경우
	DirectoryPath string     `json:"directory_path,omitempty"`
	Files         []FileInfo `json:"files,omitempty"` // 디렉터리 내 파일들
}

// FileInfo 파일 정보
type FileInfo struct {
	Name         string `json:"name"`
	RelativePath string `json:"relative_path"` // 디렉터리 기준 상대 경로
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
}

// ValidationResult 검증 결과
type ValidationResult struct {
	IsValid bool     `json:"is_valid"`
	Type    ItemType `json:"type"`

	// 전체 요약
	TotalFiles   int      `json:"total_files"`
	TotalSize    int64    `json:"total_size"`
	ValidFiles   int      `json:"valid_files"`
	InvalidFiles int      `json:"invalid_files"`
	Errors       []string `json:"errors,omitempty"`

	// 개별 파일 결과 (디렉터리인 경우)
	FileResults []FileValidationResult `json:"file_results,omitempty"`
}

// FileValidationResult 개별 파일 검증 결과
type FileValidationResult struct {
	FileName     string   `json:"file_name"`
	RelativePath string   `json:"relative_path"`
	IsValid      bool     `json:"is_valid"`
	Errors       []string `json:"errors,omitempty"`
}

// 제한 상수들
const (
	MaxFileSize      = 100 * 1024 * 1024  // 100MB
	MaxDirectorySize = 1024 * 1024 * 1024 // 1GB (디렉터리 전체)
	MaxFileCount     = 1000               // 디렉터리당 최대 파일 수
	MinFileSize      = 1
)

// 허용된 MIME 타입 (기본적인 것만)
var AllowedMimeTypes = []string{
	"text/plain",
	"application/pdf",
	"image/jpeg",
	"image/png",
}
