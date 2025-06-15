// Package model provides database models for DataLocker application.
// This file contains all model definitions to avoid circular import issues.
package model

import (
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// 파일 상태 관련 상수
const (
	// FileStatusPending 파일 처리 대기 중
	FileStatusPending = "pending"

	// FileStatusEncrypted 파일 암호화 완료
	FileStatusEncrypted = "encrypted"

	// FileStatusFailed 파일 처리 실패
	FileStatusFailed = "failed"

	// FileStatusCorrupted 파일 손상됨
	FileStatusCorrupted = "corrupted"
)

// 암호화 알고리즘 관련 상수
const (
	// EncryptionAlgorithmAES256GCM AES-256-GCM 암호화 알고리즘
	EncryptionAlgorithmAES256GCM = "AES-256-GCM"

	// KeyDerivationPBKDF2SHA256 PBKDF2-SHA256 키 유도 방식
	KeyDerivationPBKDF2SHA256 = "PBKDF2-SHA256"

	// DefaultIterations 기본 PBKDF2 반복 횟수
	DefaultIterations = 100000
)

// 필드 길이 제한 상수
const (
	// MaxOriginalNameLength 원본 파일명 최대 길이
	MaxOriginalNameLength = 255

	// MaxEncryptedPathLength 암호화된 파일 경로 최대 길이
	MaxEncryptedPathLength = 500

	// MaxMimeTypeLength MIME 타입 최대 길이
	MaxMimeTypeLength = 100

	// MaxChecksumLength 체크섬 최대 길이 (MD5 = 32자)
	MaxChecksumLength = 64

	// MaxStatusLength 상태 최대 길이
	MaxStatusLength = 20

	// MaxAlgorithmLength 알고리즘명 최대 길이
	MaxAlgorithmLength = 50

	// MaxKeyDerivationLength 키 유도 방식 최대 길이
	MaxKeyDerivationLength = 50

	// MaxSaltHexLength Salt hex 문자열 최대 길이 (32bytes * 2 = 64)
	MaxSaltHexLength = 64

	// MaxNonceHexLength Nonce hex 문자열 최대 길이 (12bytes * 2 = 24)
	MaxNonceHexLength = 24

	// MinIterations 최소 반복 횟수
	MinIterations = 1000

	// MaxIterations 최대 반복 횟수
	MaxIterations = 1000000
)

// 바이트 크기 상수 (암호화 모듈과 일치)
const (
	// ExpectedSaltSize 예상 Salt 크기 (32 바이트)
	ExpectedSaltSize = 32

	// ExpectedNonceSize 예상 Nonce 크기 (12 바이트)
	ExpectedNonceSize = 12
)

// File 암호화된 파일의 기본 정보를 저장하는 모델
type File struct {
	// 기본 필드
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"not null;index:idx_files_created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 파일 정보 필드
	OriginalName  string `gorm:"type:varchar(255);not null;index:idx_files_original_name" json:"original_name"`
	EncryptedPath string `gorm:"type:varchar(500);not null;unique" json:"encrypted_path"`
	Size          int64  `gorm:"not null;check:size >= 0" json:"size"`
	MimeType      string `gorm:"type:varchar(100);not null" json:"mime_type"`
	ChecksumMD5   string `gorm:"type:varchar(64);not null;index:idx_files_checksum" json:"checksum_md5"`
	Status        string `gorm:"type:varchar(20);not null;default:'pending';index:idx_files_status" json:"status"`

	// 관계: 1:1 (File has one EncryptionMetadata)
	EncryptionMetadata *EncryptionMetadata `gorm:"foreignKey:FileID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"encryption_metadata,omitempty"`
}

// EncryptionMetadata 암호화에 사용된 설정과 키 정보를 저장하는 모델
type EncryptionMetadata struct {
	// 기본 필드
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"not null;index:idx_encryption_metadata_created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`

	// 외래키 필드
	FileID uint `gorm:"not null;uniqueIndex:idx_encryption_metadata_file_id" json:"file_id"`

	// 암호화 설정 필드
	Algorithm     string `gorm:"type:varchar(50);not null;default:'AES-256-GCM';index:idx_encryption_metadata_algorithm" json:"algorithm"`
	KeyDerivation string `gorm:"type:varchar(50);not null;default:'PBKDF2-SHA256'" json:"key_derivation"`
	SaltHex       string `gorm:"type:varchar(64);not null" json:"salt_hex"`
	NonceHex      string `gorm:"type:varchar(24);not null" json:"nonce_hex"`
	Iterations    int    `gorm:"not null;default:100000;check:iterations >= 1000 AND iterations <= 1000000" json:"iterations"`

	// 관계: N:1 (EncryptionMetadata belongs to File)
	File *File `gorm:"foreignKey:FileID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

// TableName GORM 테이블명을 명시적으로 지정
func (File) TableName() string {
	return "files"
}

// TableName GORM 테이블명을 명시적으로 지정
func (EncryptionMetadata) TableName() string {
	return "encryption_metadata"
}

// BeforeCreate 생성 전 검증 로직
func (f *File) BeforeCreate(tx *gorm.DB) error {
	if err := f.validate(); err != nil {
		return err
	}

	// 기본 상태 설정
	if f.Status == "" {
		f.Status = FileStatusPending
	}

	return nil
}

// BeforeUpdate 수정 전 검증 로직
func (f *File) BeforeUpdate(tx *gorm.DB) error {
	return f.validate()
}

// validate 파일 모델 데이터 검증
func (f *File) validate() error {
	if f.OriginalName == "" {
		return ErrEmptyOriginalName
	}

	if len(f.OriginalName) > MaxOriginalNameLength {
		return ErrOriginalNameTooLong
	}

	if f.EncryptedPath == "" {
		return ErrEmptyEncryptedPath
	}

	if len(f.EncryptedPath) > MaxEncryptedPathLength {
		return ErrEncryptedPathTooLong
	}

	if f.Size < 0 {
		return ErrInvalidFileSize
	}

	if f.MimeType == "" {
		return ErrEmptyMimeType
	}

	if len(f.MimeType) > MaxMimeTypeLength {
		return ErrMimeTypeTooLong
	}

	if f.ChecksumMD5 == "" {
		return ErrEmptyChecksum
	}

	if len(f.ChecksumMD5) > MaxChecksumLength {
		return ErrChecksumTooLong
	}

	if !IsValidFileStatus(f.Status) {
		return ErrInvalidFileStatus
	}

	return nil
}

// BeforeCreate 생성 전 검증 로직
func (em *EncryptionMetadata) BeforeCreate(tx *gorm.DB) error {
	if err := em.validate(); err != nil {
		return err
	}

	// 기본값 설정
	if em.Algorithm == "" {
		em.Algorithm = EncryptionAlgorithmAES256GCM
	}

	if em.KeyDerivation == "" {
		em.KeyDerivation = KeyDerivationPBKDF2SHA256
	}

	if em.Iterations == 0 {
		em.Iterations = DefaultIterations
	}

	return nil
}

// BeforeUpdate 수정 전 검증 로직
func (em *EncryptionMetadata) BeforeUpdate(tx *gorm.DB) error {
	return em.validate()
}

// validate 암호화 메타데이터 검증
func (em *EncryptionMetadata) validate() error {
	if em.FileID == 0 {
		return ErrInvalidFileID
	}

	if em.Algorithm == "" {
		return ErrEmptyAlgorithm
	}

	if len(em.Algorithm) > MaxAlgorithmLength {
		return ErrAlgorithmTooLong
	}

	if !IsValidAlgorithm(em.Algorithm) {
		return ErrInvalidAlgorithm
	}

	if em.KeyDerivation == "" {
		return ErrEmptyKeyDerivation
	}

	if len(em.KeyDerivation) > MaxKeyDerivationLength {
		return ErrKeyDerivationTooLong
	}

	if !IsValidKeyDerivation(em.KeyDerivation) {
		return ErrInvalidKeyDerivation
	}

	if em.SaltHex == "" {
		return ErrEmptySalt
	}

	if len(em.SaltHex) > MaxSaltHexLength {
		return ErrSaltTooLong
	}

	if !IsValidHex(em.SaltHex) {
		return ErrInvalidSaltHex
	}

	if em.NonceHex == "" {
		return ErrEmptyNonce
	}

	if len(em.NonceHex) > MaxNonceHexLength {
		return ErrNonceTooLong
	}

	if !IsValidHex(em.NonceHex) {
		return ErrInvalidNonceHex
	}

	if em.Iterations < MinIterations || em.Iterations > MaxIterations {
		return ErrInvalidIterations
	}

	// Salt와 Nonce 바이트 크기 검증
	if err := em.validateCryptoSizes(); err != nil {
		return err
	}

	return nil
}

// validateCryptoSizes Salt와 Nonce의 실제 바이트 크기 검증
func (em *EncryptionMetadata) validateCryptoSizes() error {
	// Salt 크기 검증
	saltBytes, err := hex.DecodeString(em.SaltHex)
	if err != nil {
		return ErrInvalidSaltHex
	}

	if len(saltBytes) != ExpectedSaltSize {
		return ErrInvalidSaltSize
	}

	// Nonce 크기 검증
	nonceBytes, err := hex.DecodeString(em.NonceHex)
	if err != nil {
		return ErrInvalidNonceHex
	}

	if len(nonceBytes) != ExpectedNonceSize {
		return ErrInvalidNonceSize
	}

	return nil
}

// IsValidFileStatus 유효한 파일 상태인지 확인
func IsValidFileStatus(status string) bool {
	validStatuses := map[string]bool{
		FileStatusPending:   true,
		FileStatusEncrypted: true,
		FileStatusFailed:    true,
		FileStatusCorrupted: true,
	}

	return validStatuses[status]
}

// IsValidAlgorithm 유효한 암호화 알고리즘인지 확인
func IsValidAlgorithm(algorithm string) bool {
	validAlgorithms := map[string]bool{
		EncryptionAlgorithmAES256GCM: true,
	}

	return validAlgorithms[algorithm]
}

// IsValidKeyDerivation 유효한 키 유도 방식인지 확인
func IsValidKeyDerivation(keyDerivation string) bool {
	validDerivations := map[string]bool{
		KeyDerivationPBKDF2SHA256: true,
	}

	return validDerivations[keyDerivation]
}

// IsValidHex 유효한 16진수 문자열인지 확인
func IsValidHex(hexStr string) bool {
	_, err := hex.DecodeString(hexStr)
	return err == nil
}

// File 메소드들
// IsEncrypted 파일이 암호화되었는지 확인
func (f *File) IsEncrypted() bool {
	return f.Status == FileStatusEncrypted
}

// IsFailed 파일 처리가 실패했는지 확인
func (f *File) IsFailed() bool {
	return f.Status == FileStatusFailed
}

// IsPending 파일이 처리 대기 중인지 확인
func (f *File) IsPending() bool {
	return f.Status == FileStatusPending
}

// IsCorrupted 파일이 손상되었는지 확인
func (f *File) IsCorrupted() bool {
	return f.Status == FileStatusCorrupted
}

// MarkAsEncrypted 파일을 암호화 완료 상태로 변경
func (f *File) MarkAsEncrypted() {
	f.Status = FileStatusEncrypted
}

// MarkAsFailed 파일을 처리 실패 상태로 변경
func (f *File) MarkAsFailed() {
	f.Status = FileStatusFailed
}

// MarkAsCorrupted 파일을 손상됨 상태로 변경
func (f *File) MarkAsCorrupted() {
	f.Status = FileStatusCorrupted
}

// GetSizeInMB 파일 크기를 MB 단위로 반환
func (f *File) GetSizeInMB() float64 {
	const bytesPerMB = 1024 * 1024
	return float64(f.Size) / bytesPerMB
}

// GetSizeInKB 파일 크기를 KB 단위로 반환
func (f *File) GetSizeInKB() float64 {
	const bytesPerKB = 1024
	return float64(f.Size) / bytesPerKB
}

// EncryptionMetadata 메소드들
// GetSaltBytes Salt를 바이트 배열로 반환
func (em *EncryptionMetadata) GetSaltBytes() ([]byte, error) {
	return hex.DecodeString(em.SaltHex)
}

// GetNonceBytes Nonce를 바이트 배열로 반환
func (em *EncryptionMetadata) GetNonceBytes() ([]byte, error) {
	return hex.DecodeString(em.NonceHex)
}

// SetSaltBytes 바이트 배열을 Salt hex 문자열로 설정
func (em *EncryptionMetadata) SetSaltBytes(saltBytes []byte) error {
	if len(saltBytes) != ExpectedSaltSize {
		return ErrInvalidSaltSize
	}

	em.SaltHex = hex.EncodeToString(saltBytes)
	return nil
}

// SetNonceBytes 바이트 배열을 Nonce hex 문자열로 설정
func (em *EncryptionMetadata) SetNonceBytes(nonceBytes []byte) error {
	if len(nonceBytes) != ExpectedNonceSize {
		return ErrInvalidNonceSize
	}

	em.NonceHex = hex.EncodeToString(nonceBytes)
	return nil
}

// IsAES256GCM AES-256-GCM 알고리즘을 사용하는지 확인
func (em *EncryptionMetadata) IsAES256GCM() bool {
	return em.Algorithm == EncryptionAlgorithmAES256GCM
}

// IsPBKDF2SHA256 PBKDF2-SHA256 키 유도를 사용하는지 확인
func (em *EncryptionMetadata) IsPBKDF2SHA256() bool {
	return em.KeyDerivation == KeyDerivationPBKDF2SHA256
}

// GetIterationsString 반복 횟수를 문자열로 반환 (K 단위)
func (em *EncryptionMetadata) GetIterationsString() string {
	iterations := em.Iterations / 1000
	return fmt.Sprintf("%dK", iterations)
}
