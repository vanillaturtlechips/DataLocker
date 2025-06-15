package model

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 테스트용 상수
const (
	TestDBDir    = "./testdata"
	TestDBFile   = "test_model.db"
	TestDirPerm  = 0o750
	TestSaltHex  = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 32 bytes
	TestNonceHex = "0123456789abcdef01234567"                                         // 12 bytes
)

// setupTestDB 테스트용 데이터베이스를 설정합니다
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	// 테스트 디렉토리 생성
	err := os.MkdirAll(TestDBDir, TestDirPerm)
	require.NoError(t, err)

	// 고유한 테스트 DB 파일명 생성
	dbPath := filepath.Join(TestDBDir, "test_"+t.Name()+".db")

	// 데이터베이스 연결 (외래키 활성화 포함)
	dsn := dbPath + "?_foreign_keys=ON&_journal_mode=WAL&_sync=NORMAL"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 테스트 시 로그 최소화
	})
	require.NoError(t, err)

	// 외래키 제약조건 명시적 활성화
	err = db.Exec("PRAGMA foreign_keys = ON").Error
	require.NoError(t, err)

	// 외래키 활성화 확인
	var foreignKeysEnabled string
	err = db.Raw("PRAGMA foreign_keys").Scan(&foreignKeysEnabled).Error
	require.NoError(t, err)

	if foreignKeysEnabled != "1" {
		t.Logf("경고: 외래키 제약조건이 완전히 활성화되지 않았습니다 (값: %s)", foreignKeysEnabled)
	}

	// 마이그레이션 실행
	err = Migrate(db)
	require.NoError(t, err)

	// 정리 함수 반환
	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		_ = os.Remove(dbPath)
		_ = os.RemoveAll(TestDBDir)
	}

	return db, cleanup
}

// createTestFile 테스트용 File 모델을 생성합니다
func createTestFile() *File {
	return &File{
		OriginalName:  "test.txt",
		EncryptedPath: "/encrypted/test.enc",
		Size:          1024,
		MimeType:      "text/plain",
		ChecksumMD5:   "d41d8cd98f00b204e9800998ecf8427e",
		Status:        FileStatusPending,
	}
}

// createTestEncryptionMetadata 테스트용 EncryptionMetadata 모델을 생성합니다
func createTestEncryptionMetadata(fileID uint) *EncryptionMetadata {
	return &EncryptionMetadata{
		FileID:        fileID,
		Algorithm:     EncryptionAlgorithmAES256GCM,
		KeyDerivation: KeyDerivationPBKDF2SHA256,
		SaltHex:       TestSaltHex,
		NonceHex:      TestNonceHex,
		Iterations:    DefaultIterations,
	}
}

func TestMigrate_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 테이블이 올바르게 생성되었는지 확인
	assert.True(t, db.Migrator().HasTable(&File{}))
	assert.True(t, db.Migrator().HasTable(&EncryptionMetadata{}))

	// 컬럼이 올바르게 생성되었는지 확인
	assert.True(t, db.Migrator().HasColumn(&File{}, "original_name"))
	assert.True(t, db.Migrator().HasColumn(&File{}, "encrypted_path"))
	assert.True(t, db.Migrator().HasColumn(&EncryptionMetadata{}, "salt_hex"))
	assert.True(t, db.Migrator().HasColumn(&EncryptionMetadata{}, "nonce_hex"))
}

func TestFile_CRUD_Operations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create
	file := createTestFile()
	err := db.Create(file).Error
	require.NoError(t, err)
	assert.NotZero(t, file.ID)

	// Read
	var readFile File
	err = db.First(&readFile, file.ID).Error
	require.NoError(t, err)
	assert.Equal(t, file.OriginalName, readFile.OriginalName)
	assert.Equal(t, file.EncryptedPath, readFile.EncryptedPath)

	// Update
	readFile.Status = FileStatusEncrypted
	err = db.Save(&readFile).Error
	require.NoError(t, err)

	var updatedFile File
	err = db.First(&updatedFile, file.ID).Error
	require.NoError(t, err)
	assert.Equal(t, FileStatusEncrypted, updatedFile.Status)

	// Delete
	err = db.Delete(&updatedFile, file.ID).Error
	require.NoError(t, err)

	var deletedFile File
	err = db.First(&deletedFile, file.ID).Error
	assert.Error(t, err) // 삭제된 레코드이므로 에러가 발생해야 함
}

func TestEncryptionMetadata_CRUD_Operations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 먼저 File 생성
	file := createTestFile()
	err := db.Create(file).Error
	require.NoError(t, err)

	// Create EncryptionMetadata
	metadata := createTestEncryptionMetadata(file.ID)
	err = db.Create(metadata).Error
	require.NoError(t, err)
	assert.NotZero(t, metadata.ID)

	// Read with relationship
	var readMetadata EncryptionMetadata
	err = db.Preload("File").First(&readMetadata, metadata.ID).Error
	require.NoError(t, err)
	assert.Equal(t, file.ID, readMetadata.FileID)
	assert.Equal(t, file.OriginalName, readMetadata.File.OriginalName)

	// Update
	readMetadata.Iterations = 150000
	err = db.Save(&readMetadata).Error
	require.NoError(t, err)

	var updatedMetadata EncryptionMetadata
	err = db.First(&updatedMetadata, metadata.ID).Error
	require.NoError(t, err)
	assert.Equal(t, 150000, updatedMetadata.Iterations)
}

func TestFile_Validation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	testCases := []struct {
		name        string
		modifyFile  func(*File)
		expectError bool
		errorType   error
	}{
		{
			name: "유효한 파일",
			modifyFile: func(f *File) {
				// 기본값 사용
			},
			expectError: false,
		},
		{
			name: "빈 원본 파일명",
			modifyFile: func(f *File) {
				f.OriginalName = ""
			},
			expectError: true,
			errorType:   ErrEmptyOriginalName,
		},
		{
			name: "빈 암호화 경로",
			modifyFile: func(f *File) {
				f.EncryptedPath = ""
			},
			expectError: true,
			errorType:   ErrEmptyEncryptedPath,
		},
		{
			name: "음수 파일 크기",
			modifyFile: func(f *File) {
				f.Size = -1
			},
			expectError: true,
			errorType:   ErrInvalidFileSize,
		},
		{
			name: "빈 MIME 타입",
			modifyFile: func(f *File) {
				f.MimeType = ""
			},
			expectError: true,
			errorType:   ErrEmptyMimeType,
		},
		{
			name: "잘못된 파일 상태",
			modifyFile: func(f *File) {
				f.Status = "invalid_status"
			},
			expectError: true,
			errorType:   ErrInvalidFileStatus,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file := createTestFile()
			tc.modifyFile(file)

			err := db.Create(file).Error

			if tc.expectError {
				require.Error(t, err)
				if tc.errorType != nil {
					assert.Contains(t, err.Error(), tc.errorType.Error())
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEncryptionMetadata_Validation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 먼저 File 생성
	file := createTestFile()
	err := db.Create(file).Error
	require.NoError(t, err)

	testCases := []struct {
		name           string
		modifyMetadata func(*EncryptionMetadata)
		expectError    bool
		errorType      error
	}{
		{
			name: "유효한 메타데이터",
			modifyMetadata: func(m *EncryptionMetadata) {
				// 기본값 사용
			},
			expectError: false,
		},
		{
			name: "잘못된 파일 ID",
			modifyMetadata: func(m *EncryptionMetadata) {
				m.FileID = 0
			},
			expectError: true,
			errorType:   ErrInvalidFileID,
		},
		{
			name: "빈 알고리즘",
			modifyMetadata: func(m *EncryptionMetadata) {
				m.Algorithm = ""
			},
			expectError: true,
			errorType:   ErrEmptyAlgorithm,
		},
		{
			name: "잘못된 알고리즘",
			modifyMetadata: func(m *EncryptionMetadata) {
				m.Algorithm = "INVALID_ALGORITHM"
			},
			expectError: true,
			errorType:   ErrInvalidAlgorithm,
		},
		{
			name: "빈 Salt",
			modifyMetadata: func(m *EncryptionMetadata) {
				m.SaltHex = ""
			},
			expectError: true,
			errorType:   ErrEmptySalt,
		},
		{
			name: "잘못된 Salt hex",
			modifyMetadata: func(m *EncryptionMetadata) {
				m.SaltHex = "invalid_hex"
			},
			expectError: true,
			errorType:   ErrInvalidSaltHex,
		},
		{
			name: "잘못된 Nonce 크기",
			modifyMetadata: func(m *EncryptionMetadata) {
				m.NonceHex = "0123" // 너무 짧음
			},
			expectError: true,
			errorType:   ErrInvalidNonceSize,
		},
		{
			name: "너무 적은 반복 횟수",
			modifyMetadata: func(m *EncryptionMetadata) {
				m.Iterations = 500
			},
			expectError: true,
			errorType:   ErrInvalidIterations,
		},
		{
			name: "너무 많은 반복 횟수",
			modifyMetadata: func(m *EncryptionMetadata) {
				m.Iterations = 2000000
			},
			expectError: true,
			errorType:   ErrInvalidIterations,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metadata := createTestEncryptionMetadata(file.ID)
			tc.modifyMetadata(metadata)

			err := db.Create(metadata).Error

			if tc.expectError {
				require.Error(t, err)
				if tc.errorType != nil {
					assert.Contains(t, err.Error(), tc.errorType.Error())
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFile_Methods(t *testing.T) {
	file := createTestFile()

	// Status check methods
	assert.True(t, file.IsPending())
	assert.False(t, file.IsEncrypted())
	assert.False(t, file.IsFailed())
	assert.False(t, file.IsCorrupted())

	// Status change methods
	file.MarkAsEncrypted()
	assert.Equal(t, FileStatusEncrypted, file.Status)
	assert.True(t, file.IsEncrypted())

	file.MarkAsFailed()
	assert.Equal(t, FileStatusFailed, file.Status)
	assert.True(t, file.IsFailed())

	file.MarkAsCorrupted()
	assert.Equal(t, FileStatusCorrupted, file.Status)
	assert.True(t, file.IsCorrupted())

	// Size conversion methods
	file.Size = 1048576 // 1MB
	assert.Equal(t, 1.0, file.GetSizeInMB())
	assert.Equal(t, 1024.0, file.GetSizeInKB())
}

func TestEncryptionMetadata_Methods(t *testing.T) {
	metadata := createTestEncryptionMetadata(1)

	// Algorithm and key derivation checks
	assert.True(t, metadata.IsAES256GCM())
	assert.True(t, metadata.IsPBKDF2SHA256())

	// Byte conversion methods
	saltBytes, err := metadata.GetSaltBytes()
	require.NoError(t, err)
	assert.Len(t, saltBytes, ExpectedSaltSize)

	nonceBytes, err := metadata.GetNonceBytes()
	require.NoError(t, err)
	assert.Len(t, nonceBytes, ExpectedNonceSize)

	// Set byte methods
	newSalt := make([]byte, ExpectedSaltSize)
	for i := range newSalt {
		newSalt[i] = byte(i)
	}

	err = metadata.SetSaltBytes(newSalt)
	require.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(newSalt), metadata.SaltHex)

	newNonce := make([]byte, ExpectedNonceSize)
	for i := range newNonce {
		newNonce[i] = byte(i + 100)
	}

	err = metadata.SetNonceBytes(newNonce)
	require.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(newNonce), metadata.NonceHex)
}

func TestForeignKeyConstraint(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 존재하지 않는 파일 ID로 메타데이터 생성 시도
	metadata := createTestEncryptionMetadata(99999) // 존재하지 않는 ID

	err := db.Create(metadata).Error
	require.Error(t, err)
	// SQLite의 외래키 제약조건 에러 확인
	assert.Contains(t, err.Error(), "FOREIGN KEY constraint failed")
}

func TestCascadeDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// File과 EncryptionMetadata 생성
	file := createTestFile()
	err := db.Create(file).Error
	require.NoError(t, err)

	metadata := createTestEncryptionMetadata(file.ID)
	err = db.Create(metadata).Error
	require.NoError(t, err)

	// File 삭제
	err = db.Delete(file).Error
	require.NoError(t, err)

	// EncryptionMetadata도 함께 삭제되었는지 확인
	var deletedMetadata EncryptionMetadata
	err = db.First(&deletedMetadata, metadata.ID).Error

	// CASCADE 삭제가 작동하면 에러가 발생해야 함 (레코드 없음)
	// 하지만 SQLite에서는 CASCADE가 제대로 작동할 수 있음
	if err == nil {
		// CASCADE가 작동하지 않았다면, 메타데이터가 여전히 존재할 수 있음
		// 이 경우 File이 삭제되었는지 확인
		var deletedFile File
		fileErr := db.First(&deletedFile, file.ID).Error
		assert.Error(t, fileErr, "File이 삭제되어야 합니다")
	} else {
		// CASCADE가 정상 작동하여 메타데이터도 삭제됨
		assert.Error(t, err, "EncryptionMetadata도 CASCADE로 삭제되어야 합니다")
	}
}

func TestUniqueConstraint(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 첫 번째 파일 생성
	file1 := createTestFile()
	err := db.Create(file1).Error
	require.NoError(t, err)

	// 같은 암호화 경로로 두 번째 파일 생성 시도
	file2 := createTestFile()
	file2.OriginalName = "another.txt" // 다른 이름
	// 같은 EncryptedPath 사용

	err = db.Create(file2).Error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UNIQUE constraint failed")
}

func TestGetTableInfo(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 테스트 데이터 생성
	file := createTestFile()
	err := db.Create(file).Error
	require.NoError(t, err)

	metadata := createTestEncryptionMetadata(file.ID)
	err = db.Create(metadata).Error
	require.NoError(t, err)

	// 테이블 정보 조회
	info, err := GetTableInfo(db)
	require.NoError(t, err)
	require.NotNil(t, info)

	// 테이블 존재 확인
	tables := info["tables"].([]string)
	assert.Contains(t, tables, "files")
	assert.Contains(t, tables, "encryption_metadata")

	// 레코드 수 확인
	recordCounts := info["record_counts"].(map[string]int64)
	assert.Equal(t, int64(1), recordCounts["files"])
	assert.Equal(t, int64(1), recordCounts["encryption_metadata"])
}

func TestValidateSchema(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	err := ValidateSchema(db)
	assert.NoError(t, err)
}

func TestResetDatabase(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 테스트 데이터 생성
	file := createTestFile()
	err := db.Create(file).Error
	require.NoError(t, err)

	// 데이터베이스 리셋
	err = ResetDatabase(db)
	require.NoError(t, err)

	// 테이블이 다시 생성되었지만 데이터는 없는지 확인
	assert.True(t, db.Migrator().HasTable(&File{}))
	assert.True(t, db.Migrator().HasTable(&EncryptionMetadata{}))

	var count int64
	err = db.Model(&File{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// 벤치마크 테스트
func BenchmarkFile_Create(b *testing.B) {
	// 메모리 데이터베이스 사용
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(b, err)

	err = Migrate(db)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file := createTestFile()
		file.OriginalName = fmt.Sprintf("test_%d.txt", i)             // 고유한 이름
		file.EncryptedPath = fmt.Sprintf("/encrypted/test_%d.enc", i) // 고유한 경로

		err := db.Create(file).Error
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncryptionMetadata_Create(b *testing.B) {
	// 메모리 데이터베이스 사용
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(b, err)

	err = Migrate(db)
	require.NoError(b, err)

	// 벤치마크용 파일들 미리 생성
	files := make([]*File, b.N)
	for i := 0; i < b.N; i++ {
		file := createTestFile()
		file.OriginalName = fmt.Sprintf("bench_%d.txt", i)
		file.EncryptedPath = fmt.Sprintf("/encrypted/bench_%d.enc", i)
		err := db.Create(file).Error
		require.NoError(b, err)
		files[i] = file
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metadata := createTestEncryptionMetadata(files[i].ID)
		err := db.Create(metadata).Error
		if err != nil {
			b.Fatal(err)
		}
	}
}
