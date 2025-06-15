// Package repository provides data access layer for DataLocker application.
// This file implements repository pattern for encryption metadata operations.
package repository

import (
	"fmt"

	"DataLocker/internal/model"

	"gorm.io/gorm"
)

// EncryptionRepository 암호화 메타데이터 저장소 인터페이스
type EncryptionRepository interface {
	Create(metadata *model.EncryptionMetadata) error
	GetByID(id uint) (*model.EncryptionMetadata, error)
	GetByFileID(fileID uint) (*model.EncryptionMetadata, error)
	Update(metadata *model.EncryptionMetadata) error
	DeleteByID(id uint) error
	DeleteByFileID(fileID uint) error
	GetByAlgorithm(algorithm string, offset, limit int) ([]*model.EncryptionMetadata, int64, error)
	Exists(id uint) (bool, error)
	ExistsByFileID(fileID uint) (bool, error)
	Count() (int64, error)
	CountByAlgorithm(algorithm string) (int64, error)
}

// encryptionRepository GORM 기반 암호화 메타데이터 저장소 구현체
type encryptionRepository struct {
	db *gorm.DB
}

// NewEncryptionRepository 새로운 암호화 메타데이터 저장소를 생성합니다
func NewEncryptionRepository(db *gorm.DB) EncryptionRepository {
	if db == nil {
		panic("데이터베이스 연결이 필요합니다")
	}

	return &encryptionRepository{
		db: db,
	}
}

// Create 새로운 암호화 메타데이터 레코드를 생성합니다
func (r *encryptionRepository) Create(metadata *model.EncryptionMetadata) error {
	if metadata == nil {
		return fmt.Errorf("암호화 메타데이터가 없습니다")
	}

	if err := r.db.Create(metadata).Error; err != nil {
		return fmt.Errorf("암호화 메타데이터 생성 실패: %w", err)
	}

	return nil
}

// GetByID ID로 암호화 메타데이터를 조회합니다
func (r *encryptionRepository) GetByID(id uint) (*model.EncryptionMetadata, error) {
	if id == 0 {
		return nil, fmt.Errorf("유효하지 않은 암호화 메타데이터 ID입니다")
	}

	var metadata model.EncryptionMetadata
	err := r.db.Preload("File").First(&metadata, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("암호화 메타데이터를 찾을 수 없습니다: ID %d", id)
		}
		return nil, fmt.Errorf("암호화 메타데이터 조회 실패: %w", err)
	}

	return &metadata, nil
}

// GetByFileID 파일 ID로 암호화 메타데이터를 조회합니다
func (r *encryptionRepository) GetByFileID(fileID uint) (*model.EncryptionMetadata, error) {
	if fileID == 0 {
		return nil, fmt.Errorf("유효하지 않은 파일 ID입니다")
	}

	var metadata model.EncryptionMetadata
	err := r.db.Preload("File").Where("file_id = ?", fileID).First(&metadata).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("파일 ID %d에 대한 암호화 메타데이터를 찾을 수 없습니다", fileID)
		}
		return nil, fmt.Errorf("암호화 메타데이터 조회 실패: %w", err)
	}

	return &metadata, nil
}

// Update 암호화 메타데이터를 업데이트합니다
func (r *encryptionRepository) Update(metadata *model.EncryptionMetadata) error {
	if metadata == nil {
		return fmt.Errorf("암호화 메타데이터가 없습니다")
	}

	if metadata.ID == 0 {
		return fmt.Errorf("유효하지 않은 암호화 메타데이터 ID입니다")
	}

	// 메타데이터 존재 여부 확인
	exists, err := r.Exists(metadata.ID)
	if err != nil {
		return fmt.Errorf("암호화 메타데이터 존재 확인 실패: %w", err)
	}

	if !exists {
		return fmt.Errorf("업데이트할 암호화 메타데이터를 찾을 수 없습니다: ID %d", metadata.ID)
	}

	// 업데이트 실행
	if err := r.db.Save(metadata).Error; err != nil {
		return fmt.Errorf("암호화 메타데이터 업데이트 실패: %w", err)
	}

	return nil
}

// DeleteByID ID로 암호화 메타데이터를 삭제합니다
func (r *encryptionRepository) DeleteByID(id uint) error {
	if id == 0 {
		return fmt.Errorf("유효하지 않은 암호화 메타데이터 ID입니다")
	}

	// 메타데이터 존재 여부 확인
	exists, err := r.Exists(id)
	if err != nil {
		return fmt.Errorf("암호화 메타데이터 존재 확인 실패: %w", err)
	}

	if !exists {
		return fmt.Errorf("삭제할 암호화 메타데이터를 찾을 수 없습니다: ID %d", id)
	}

	// 삭제 실행 (하드 삭제 - 암호화 메타데이터는 보안상 완전 삭제)
	if err := r.db.Unscoped().Delete(&model.EncryptionMetadata{}, id).Error; err != nil {
		return fmt.Errorf("암호화 메타데이터 삭제 실패: %w", err)
	}

	return nil
}

// DeleteByFileID 파일 ID로 암호화 메타데이터를 삭제합니다
func (r *encryptionRepository) DeleteByFileID(fileID uint) error {
	if fileID == 0 {
		return fmt.Errorf("유효하지 않은 파일 ID입니다")
	}

	// 메타데이터 존재 여부 확인
	exists, err := r.ExistsByFileID(fileID)
	if err != nil {
		return fmt.Errorf("암호화 메타데이터 존재 확인 실패: %w", err)
	}

	if !exists {
		return fmt.Errorf("파일 ID %d에 대한 암호화 메타데이터를 찾을 수 없습니다", fileID)
	}

	// 삭제 실행 (하드 삭제)
	if err := r.db.Unscoped().Where("file_id = ?", fileID).Delete(&model.EncryptionMetadata{}).Error; err != nil {
		return fmt.Errorf("암호화 메타데이터 삭제 실패: %w", err)
	}

	return nil
}

// GetByAlgorithm 암호화 알고리즘별로 메타데이터를 조회합니다
func (r *encryptionRepository) GetByAlgorithm(algorithm string, offset, limit int) ([]*model.EncryptionMetadata, int64, error) {
	if algorithm == "" {
		return nil, 0, fmt.Errorf("암호화 알고리즘이 필요합니다")
	}

	if !model.IsValidAlgorithm(algorithm) {
		return nil, 0, fmt.Errorf("유효하지 않은 암호화 알고리즘입니다: %s", algorithm)
	}

	offset, limit = r.normalizePagination(offset, limit)

	var metadataList []*model.EncryptionMetadata
	var total int64

	// 알고리즘별 카운트 조회
	if err := r.db.Model(&model.EncryptionMetadata{}).Where("algorithm = ?", algorithm).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("알고리즘별 암호화 메타데이터 카운트 조회 실패: %w", err)
	}

	// 알고리즘별 메타데이터 목록 조회
	err := r.db.Preload("File").
		Where("algorithm = ?", algorithm).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&metadataList).Error
	if err != nil {
		return nil, 0, fmt.Errorf("알고리즘별 암호화 메타데이터 목록 조회 실패: %w", err)
	}

	return metadataList, total, nil
}

// Exists 암호화 메타데이터 존재 여부를 확인합니다
func (r *encryptionRepository) Exists(id uint) (bool, error) {
	if id == 0 {
		return false, fmt.Errorf("유효하지 않은 암호화 메타데이터 ID입니다")
	}

	var count int64
	err := r.db.Model(&model.EncryptionMetadata{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("암호화 메타데이터 존재 확인 실패: %w", err)
	}

	return count > 0, nil
}

// ExistsByFileID 파일 ID로 암호화 메타데이터 존재 여부를 확인합니다
func (r *encryptionRepository) ExistsByFileID(fileID uint) (bool, error) {
	if fileID == 0 {
		return false, fmt.Errorf("유효하지 않은 파일 ID입니다")
	}

	var count int64
	err := r.db.Model(&model.EncryptionMetadata{}).Where("file_id = ?", fileID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("암호화 메타데이터 존재 확인 실패: %w", err)
	}

	return count > 0, nil
}

// Count 전체 암호화 메타데이터 수를 반환합니다
func (r *encryptionRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.EncryptionMetadata{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("암호화 메타데이터 카운트 조회 실패: %w", err)
	}

	return count, nil
}

// CountByAlgorithm 알고리즘별 암호화 메타데이터 수를 반환합니다
func (r *encryptionRepository) CountByAlgorithm(algorithm string) (int64, error) {
	if algorithm == "" {
		return 0, fmt.Errorf("암호화 알고리즘이 필요합니다")
	}

	if !model.IsValidAlgorithm(algorithm) {
		return 0, fmt.Errorf("유효하지 않은 암호화 알고리즘입니다: %s", algorithm)
	}

	var count int64
	err := r.db.Model(&model.EncryptionMetadata{}).Where("algorithm = ?", algorithm).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("알고리즘별 암호화 메타데이터 카운트 조회 실패: %w", err)
	}

	return count, nil
}

// normalizePagination 페이지네이션 파라미터를 정규화합니다
func (r *encryptionRepository) normalizePagination(offset, limit int) (int, int) {
	if offset < MinOffset {
		offset = MinOffset
	}

	if limit <= 0 || limit > MaxPageSize {
		limit = DefaultPageSize
	}

	return offset, limit
}
