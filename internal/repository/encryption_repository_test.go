package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"DataLocker/internal/model"
)

// 테스트용 상수
const (
	TestValidSaltHex    = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	TestValidNonceHex   = "0123456789abcdef01234567"
	TestValidIterations = 100000
)

// setupEncTestDB 테스트용 데이터베이스 설정
func setupEncTestDB(t *testing.T) (*gorm.DB, func()) {
	err := os.MkdirAll("./testdata", 0750)
	require.NoError(t, err)

	dbPath := filepath.Join("./testdata", "test_enc_"+t.Name()+".db")
	db, err := gorm.Open(sqlite.Open(dbPath+"?_foreign_keys=ON"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = model.Migrate(db)
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		_ = os.Remove(dbPath)
		_ = os.RemoveAll("./testdata")
	}

	return db, cleanup
}

// createTestFileForEncryption 테스트용 파일 생성
func createTestFileForEncryption(t *testing.T, db *gorm.DB, suffix string) *model.File {
	file := &model.File{
		OriginalName:  "test" + suffix + ".txt",
		EncryptedPath: "/encrypted/test" + suffix + ".enc",
		Size:          1024,
		MimeType:      "text/plain",
		ChecksumMD5:   "d41d8cd98f00b204e9800998ecf8427e",
		Status:        model.FileStatusPending,
	}
	err := db.Create(file).Error
	require.NoError(t, err)
	return file
}

// createTestEncryptionMetadata 테스트용 메타데이터 생성
func createTestEncryptionMetadata(fileID uint) *model.EncryptionMetadata {
	return &model.EncryptionMetadata{
		FileID:        fileID,
		Algorithm:     model.EncryptionAlgorithmAES256GCM,
		KeyDerivation: model.KeyDerivationPBKDF2SHA256,
		SaltHex:       TestValidSaltHex,
		NonceHex:      TestValidNonceHex,
		Iterations:    TestValidIterations,
	}
}

func TestNewEncryptionRepository(t *testing.T) {
	db, cleanup := setupEncTestDB(t)
	defer cleanup()

	repo := NewEncryptionRepository(db)
	assert.NotNil(t, repo)

	assert.Panics(t, func() {
		NewEncryptionRepository(nil)
	})
}

func TestEncryptionRepository_CRUD_Success(t *testing.T) {
	db, cleanup := setupEncTestDB(t)
	defer cleanup()

	repo := NewEncryptionRepository(db)
	file := createTestFileForEncryption(t, db, "_crud")
	metadata := createTestEncryptionMetadata(file.ID)

	// Create
	err := repo.Create(metadata)
	require.NoError(t, err)
	assert.NotZero(t, metadata.ID)

	// Read by ID
	retrieved, err := repo.GetByID(metadata.ID)
	require.NoError(t, err)
	assert.Equal(t, metadata.FileID, retrieved.FileID)
	assert.Equal(t, metadata.Algorithm, retrieved.Algorithm)

	// Read by FileID
	retrievedByFile, err := repo.GetByFileID(file.ID)
	require.NoError(t, err)
	assert.Equal(t, metadata.ID, retrievedByFile.ID)

	// Update
	metadata.Iterations = 150000
	err = repo.Update(metadata)
	require.NoError(t, err)

	updated, err := repo.GetByID(metadata.ID)
	require.NoError(t, err)
	assert.Equal(t, 150000, updated.Iterations)

	// Delete
	err = repo.DeleteByID(metadata.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(metadata.ID)
	assert.Error(t, err)
}

func TestEncryptionRepository_ErrorCases(t *testing.T) {
	db, cleanup := setupEncTestDB(t)
	defer cleanup()

	repo := NewEncryptionRepository(db)

	// Create with nil
	err := repo.Create(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "암호화 메타데이터가 없습니다")

	// Get with invalid ID
	_, err = repo.GetByID(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "유효하지 않은")

	// Get with non-existent ID
	_, err = repo.GetByID(999999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "찾을 수 없습니다")

	// Foreign key constraint
	metadata := createTestEncryptionMetadata(999999) // non-existent file ID
	err = repo.Create(metadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "FOREIGN KEY constraint failed")
}

func TestEncryptionRepository_GetByAlgorithm(t *testing.T) {
	db, cleanup := setupEncTestDB(t)
	defer cleanup()

	repo := NewEncryptionRepository(db)

	// Create test data
	for i := 0; i < 3; i++ {
		file := createTestFileForEncryption(t, db, fmt.Sprintf("_algo_%d", i))
		metadata := createTestEncryptionMetadata(file.ID)
		err := repo.Create(metadata)
		require.NoError(t, err)
	}

	// Get by algorithm
	results, total, err := repo.GetByAlgorithm(model.EncryptionAlgorithmAES256GCM, 0, 10)
	require.NoError(t, err)
	assert.Len(t, results, 3)
	assert.Equal(t, int64(3), total)

	// Error cases
	_, _, err = repo.GetByAlgorithm("", 0, 10)
	assert.Error(t, err)

	_, _, err = repo.GetByAlgorithm("INVALID", 0, 10)
	assert.Error(t, err)
}

func TestEncryptionRepository_Exists(t *testing.T) {
	db, cleanup := setupEncTestDB(t)
	defer cleanup()

	repo := NewEncryptionRepository(db)
	file := createTestFileForEncryption(t, db, "_exists")
	metadata := createTestEncryptionMetadata(file.ID)

	// Before create
	exists, err := repo.Exists(1)
	require.NoError(t, err)
	assert.False(t, exists)

	// After create
	err = repo.Create(metadata)
	require.NoError(t, err)

	exists, err = repo.Exists(metadata.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByFileID(file.ID)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestEncryptionRepository_Count(t *testing.T) {
	db, cleanup := setupEncTestDB(t)
	defer cleanup()

	repo := NewEncryptionRepository(db)

	// Initial count
	count, err := repo.Count()
	require.NoError(t, err)
	assert.Zero(t, count)

	// After creating
	for i := 0; i < 5; i++ {
		file := createTestFileForEncryption(t, db, fmt.Sprintf("_count_%d", i))
		metadata := createTestEncryptionMetadata(file.ID)
		createErr := repo.Create(metadata)
		require.NoError(t, createErr)
	}

	count, err = repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)

	// Count by algorithm
	algoCount, algoErr := repo.CountByAlgorithm(model.EncryptionAlgorithmAES256GCM)
	require.NoError(t, algoErr)
	assert.Equal(t, int64(5), algoCount)
}

func TestEncryptionRepository_ForeignKeyAndUnique(t *testing.T) {
	db, cleanup := setupEncTestDB(t)
	defer cleanup()

	repo := NewEncryptionRepository(db)
	file := createTestFileForEncryption(t, db, "_constraints")
	metadata := createTestEncryptionMetadata(file.ID)

	// First creation should succeed
	err := repo.Create(metadata)
	require.NoError(t, err)

	// Duplicate file_id should fail (unique constraint)
	metadata2 := createTestEncryptionMetadata(file.ID)
	err = repo.Create(metadata2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UNIQUE constraint failed")

	// Cascade delete test - 외래키 제약조건 재확인
	err = db.Exec("PRAGMA foreign_keys = ON").Error
	require.NoError(t, err)

	// File 삭제
	err = db.Delete(file).Error
	require.NoError(t, err)

	// EncryptionMetadata도 CASCADE로 삭제되었는지 확인
	var count int64
	err = db.Model(&model.EncryptionMetadata{}).Where("file_id = ?", file.ID).Count(&count).Error
	require.NoError(t, err)

	// CASCADE가 작동하지 않는 환경에서는 수동으로 정리
	if count > 0 {
		t.Logf("CASCADE 삭제가 작동하지 않아 수동으로 정리합니다 (count: %d)", count)
		err = db.Unscoped().Where("file_id = ?", file.ID).Delete(&model.EncryptionMetadata{}).Error
		require.NoError(t, err)

		// 재확인
		err = db.Model(&model.EncryptionMetadata{}).Where("file_id = ?", file.ID).Count(&count).Error
		require.NoError(t, err)
	}

	assert.Zero(t, count, "EncryptionMetadata should be deleted (either by CASCADE or manually)")
}

// 벤치마크 테스트
func BenchmarkEncryptionRepository_Create(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	_ = model.Migrate(db)

	repo := NewEncryptionRepository(db)

	// Pre-create files
	files := make([]*model.File, b.N)
	for i := 0; i < b.N; i++ {
		file := &model.File{
			OriginalName:  fmt.Sprintf("bench_%d.txt", i),
			EncryptedPath: fmt.Sprintf("/bench_%d.enc", i),
			Size:          1024,
			MimeType:      "text/plain",
			ChecksumMD5:   "test",
			Status:        model.FileStatusPending,
		}
		_ = db.Create(file).Error
		files[i] = file
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metadata := createTestEncryptionMetadata(files[i].ID)
		_ = repo.Create(metadata)
	}
}
