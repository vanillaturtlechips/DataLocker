// Package repository provides data access layer for DataLocker application.
// It implements repository pattern for database operations with GORM.
package repository

import (
	"fmt"

	"DataLocker/internal/model"

	"gorm.io/gorm"
)

// 페이지네이션 관련 상수
const (
	// DefaultPageSize 기본 페이지 크기
	DefaultPageSize = 10

	// MaxPageSize 최대 페이지 크기
	MaxPageSize = 100

	// MinOffset 최소 오프셋
	MinOffset = 0
)

// FileRepository 파일 메타데이터 저장소 인터페이스
type FileRepository interface {
	Create(file *model.File) error
	GetByID(id uint) (*model.File, error)
	GetAll(offset, limit int) ([]*model.File, int64, error)
	Update(file *model.File) error
	Delete(id uint) error
	GetByStatus(status string, offset, limit int) ([]*model.File, int64, error)
	GetByChecksumMD5(checksum string) (*model.File, error)
	Exists(id uint) (bool, error)
	Count() (int64, error)
}

// fileRepository GORM 기반 파일 저장소 구현체
type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository 새로운 파일 저장소를 생성합니다
func NewFileRepository(db *gorm.DB) FileRepository {
	if db == nil {
		panic("데이터베이스 연결이 필요합니다")
	}

	return &fileRepository{
		db: db,
	}
}

// Create 새로운 파일 레코드를 생성합니다
func (r *fileRepository) Create(file *model.File) error {
	if file == nil {
		return fmt.Errorf("파일 데이터가 없습니다")
	}

	if err := r.db.Create(file).Error; err != nil {
		return fmt.Errorf("파일 생성 실패: %w", err)
	}

	return nil
}

// GetByID ID로 파일을 조회합니다
func (r *fileRepository) GetByID(id uint) (*model.File, error) {
	if id == 0 {
		return nil, fmt.Errorf("유효하지 않은 파일 ID입니다")
	}

	var file model.File
	err := r.db.Preload("EncryptionMetadata").First(&file, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("파일을 찾을 수 없습니다: ID %d", id)
		}
		return nil, fmt.Errorf("파일 조회 실패: %w", err)
	}

	return &file, nil
}

// GetAll 모든 파일을 페이지네이션으로 조회합니다
func (r *fileRepository) GetAll(offset, limit int) ([]*model.File, int64, error) {
	offset, limit = r.normalizePagination(offset, limit)

	var files []*model.File
	var total int64

	// 전체 카운트 조회
	if err := r.db.Model(&model.File{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("파일 카운트 조회 실패: %w", err)
	}

	// 페이지네이션된 데이터 조회
	err := r.db.Preload("EncryptionMetadata").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&files).Error
	if err != nil {
		return nil, 0, fmt.Errorf("파일 목록 조회 실패: %w", err)
	}

	return files, total, nil
}

// Update 파일 정보를 업데이트합니다
func (r *fileRepository) Update(file *model.File) error {
	if file == nil {
		return fmt.Errorf("파일 데이터가 없습니다")
	}

	if file.ID == 0 {
		return fmt.Errorf("유효하지 않은 파일 ID입니다")
	}

	// 파일 존재 여부 확인
	exists, err := r.Exists(file.ID)
	if err != nil {
		return fmt.Errorf("파일 존재 확인 실패: %w", err)
	}

	if !exists {
		return fmt.Errorf("업데이트할 파일을 찾을 수 없습니다: ID %d", file.ID)
	}

	// 업데이트 실행
	if err := r.db.Save(file).Error; err != nil {
		return fmt.Errorf("파일 업데이트 실패: %w", err)
	}

	return nil
}

// Delete 파일을 삭제합니다 (소프트 삭제)
func (r *fileRepository) Delete(id uint) error {
	if id == 0 {
		return fmt.Errorf("유효하지 않은 파일 ID입니다")
	}

	// 파일 존재 여부 확인
	exists, err := r.Exists(id)
	if err != nil {
		return fmt.Errorf("파일 존재 확인 실패: %w", err)
	}

	if !exists {
		return fmt.Errorf("삭제할 파일을 찾을 수 없습니다: ID %d", id)
	}

	// 소프트 삭제 실행
	if err := r.db.Delete(&model.File{}, id).Error; err != nil {
		return fmt.Errorf("파일 삭제 실패: %w", err)
	}

	return nil
}

// GetByStatus 상태별로 파일을 조회합니다
func (r *fileRepository) GetByStatus(status string, offset, limit int) ([]*model.File, int64, error) {
	if status == "" {
		return nil, 0, fmt.Errorf("상태 값이 필요합니다")
	}

	if !model.IsValidFileStatus(status) {
		return nil, 0, fmt.Errorf("유효하지 않은 파일 상태입니다: %s", status)
	}

	offset, limit = r.normalizePagination(offset, limit)

	var files []*model.File
	var total int64

	// 상태별 카운트 조회
	if err := r.db.Model(&model.File{}).Where("status = ?", status).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("상태별 파일 카운트 조회 실패: %w", err)
	}

	// 상태별 파일 목록 조회
	err := r.db.Preload("EncryptionMetadata").
		Where("status = ?", status).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&files).Error
	if err != nil {
		return nil, 0, fmt.Errorf("상태별 파일 목록 조회 실패: %w", err)
	}

	return files, total, nil
}

// GetByChecksumMD5 MD5 체크섬으로 파일을 조회합니다 (중복 검사용)
func (r *fileRepository) GetByChecksumMD5(checksum string) (*model.File, error) {
	if checksum == "" {
		return nil, fmt.Errorf("체크섬 값이 필요합니다")
	}

	var file model.File
	err := r.db.Preload("EncryptionMetadata").
		Where("checksum_md5 = ?", checksum).
		First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 중복이 없음을 나타내기 위해 nil 반환
		}
		return nil, fmt.Errorf("체크섬 조회 실패: %w", err)
	}

	return &file, nil
}

// Exists 파일 존재 여부를 확인합니다
func (r *fileRepository) Exists(id uint) (bool, error) {
	if id == 0 {
		return false, fmt.Errorf("유효하지 않은 파일 ID입니다")
	}

	var count int64
	err := r.db.Model(&model.File{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("파일 존재 확인 실패: %w", err)
	}

	return count > 0, nil
}

// Count 전체 파일 수를 반환합니다
func (r *fileRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.File{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("파일 카운트 조회 실패: %w", err)
	}

	return count, nil
}

// normalizePagination 페이지네이션 파라미터를 정규화합니다
func (r *fileRepository) normalizePagination(offset, limit int) (int, int) {
	if offset < MinOffset {
		offset = MinOffset
	}

	if limit <= 0 || limit > MaxPageSize {
		limit = DefaultPageSize
	}

	return offset, limit
}
