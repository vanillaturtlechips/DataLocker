package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"DataLocker/internal/model"
)

// 테스트용 상수
const (
	TestDBDir           = "./testdata"
	TestDirPerm         = 0750
	TestValidChecksum   = "d41d8cd98f00b204e9800998ecf8427e"
	TestLargeFileSize   = 1048576 // 1MB
	TestSmallFileSize   = 1024    // 1KB
	TestPageSize        = 5
	TestLargePageSize   = 200 // MaxPageSize보다 큰 값
	TestNegativeOffset  = -1
	TestValidFileID     = 1
	TestInvalidFileID   = 0
	TestNonExistentID   = 999999
	TestBulkCreateCount = 50
)

// setupTestDB 테스트용 데이터베이스를 설정합니다
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	// 테스트 디렉토리 생성
	err := os.MkdirAll(TestDBDir, TestDirPerm)
	require.NoError(t, err)

	// 고유한 테스트 DB 파일명 생성
	dbPath := filepath.Join(TestDBDir, "test_file_repo_"+t.Name()+".db")

	// 데이터베이스 연결
	dsn := dbPath + "?_foreign_keys=ON&_journal_mode=WAL&_sync=NORMAL"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 테스트 시 로그 최소화
	})
	require.NoError(t, err)

	// 마이그레이션 실행
	err = model.Migrate(db)
	require.NoError(t, err)

	// 정리 함수 반환
	cleanup := func() {
		sqlDB, dbErr := db.DB()
		if dbErr == nil {
			_ = sqlDB.Close()
		}
		_ = os.Remove(dbPath)
		_ = os.RemoveAll(TestDBDir)
	}

	return db, cleanup
}

// createTestFile 테스트용 파일 모델을 생성합니다
func createTestFile(suffix string) *model.File {
	return &model.File{
		OriginalName:  fmt.Sprintf("test%s.txt", suffix),
		EncryptedPath: fmt.Sprintf("/encrypted/test%s.enc", suffix),
		Size:          TestSmallFileSize,
		MimeType:      "text/plain",
		ChecksumMD5:   TestValidChecksum,
		Status:        model.FileStatusPending,
	}
}

func TestNewFileRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 정상 케이스
	repo := NewFileRepository(db)
	assert.NotNil(t, repo)

	// nil DB 패닉 테스트
	assert.Panics(t, func() {
		NewFileRepository(nil)
	})
}

func TestFileRepository_Create_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)
	file := createTestFile("_create")

	err := repo.Create(file)
	require.NoError(t, err)
	assert.NotZero(t, file.ID)
	assert.NotZero(t, file.CreatedAt)
	assert.NotZero(t, file.UpdatedAt)
}

func TestFileRepository_Create_ErrorCases(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	testCases := []struct {
		name    string
		file    *model.File
		wantErr string
	}{
		{
			name:    "nil 파일",
			file:    nil,
			wantErr: "파일 데이터가 없습니다",
		},
		{
			name: "빈 원본 파일명",
			file: &model.File{
				OriginalName:  "",
				EncryptedPath: "/test.enc",
				Size:          TestSmallFileSize,
				MimeType:      "text/plain",
				ChecksumMD5:   TestValidChecksum,
			},
			wantErr: "파일 생성 실패",
		},
		{
			name: "중복된 암호화 경로",
			file: func() *model.File {
				// 먼저 파일 하나 생성
				firstFile := createTestFile("_first")
				_ = repo.Create(firstFile)

				// 같은 암호화 경로로 두 번째 파일 생성 시도
				secondFile := createTestFile("_second")
				secondFile.EncryptedPath = firstFile.EncryptedPath
				return secondFile
			}(),
			wantErr: "파일 생성 실패",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Create(tc.file)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestFileRepository_GetByID_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)
	originalFile := createTestFile("_getbyid")

	// 파일 생성
	err := repo.Create(originalFile)
	require.NoError(t, err)

	// 조회 테스트
	retrievedFile, err := repo.GetByID(originalFile.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedFile)

	assert.Equal(t, originalFile.ID, retrievedFile.ID)
	assert.Equal(t, originalFile.OriginalName, retrievedFile.OriginalName)
	assert.Equal(t, originalFile.EncryptedPath, retrievedFile.EncryptedPath)
	assert.Equal(t, originalFile.Size, retrievedFile.Size)
	assert.Equal(t, originalFile.MimeType, retrievedFile.MimeType)
	assert.Equal(t, originalFile.ChecksumMD5, retrievedFile.ChecksumMD5)
	assert.Equal(t, originalFile.Status, retrievedFile.Status)
}

func TestFileRepository_GetByID_ErrorCases(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	testCases := []struct {
		name    string
		id      uint
		wantErr string
	}{
		{
			name:    "잘못된 ID (0)",
			id:      TestInvalidFileID,
			wantErr: "유효하지 않은 파일 ID입니다",
		},
		{
			name:    "존재하지 않는 ID",
			id:      TestNonExistentID,
			wantErr: "파일을 찾을 수 없습니다",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := repo.GetByID(tc.id)
			require.Error(t, err)
			assert.Nil(t, file)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestFileRepository_GetAll_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	// 테스트 파일들 생성
	testFiles := make([]*model.File, TestPageSize)
	for i := 0; i < TestPageSize; i++ {
		file := createTestFile(fmt.Sprintf("_getall_%d", i))
		err := repo.Create(file)
		require.NoError(t, err)
		testFiles[i] = file

		// 생성 시간 차이를 두기 위해 약간의 지연
		time.Sleep(time.Millisecond)
	}

	// 전체 조회 테스트
	files, total, err := repo.GetAll(0, TestPageSize)
	require.NoError(t, err)
	assert.Len(t, files, TestPageSize)
	assert.Equal(t, int64(TestPageSize), total)

	// 최신 순으로 정렬되었는지 확인
	for i := 0; i < len(files)-1; i++ {
		assert.True(t, files[i].CreatedAt.After(files[i+1].CreatedAt) ||
			files[i].CreatedAt.Equal(files[i+1].CreatedAt))
	}
}

func TestFileRepository_GetAll_Pagination(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	// 테스트 파일들 생성 (10개)
	for i := 0; i < 10; i++ {
		file := createTestFile(fmt.Sprintf("_pagination_%d", i))
		err := repo.Create(file)
		require.NoError(t, err)
	}

	testCases := []struct {
		name           string
		offset         int
		limit          int
		expectedCount  int
		expectedTotal  int64
		expectedOffset int
		expectedLimit  int
	}{
		{
			name:           "첫 번째 페이지",
			offset:         0,
			limit:          TestPageSize,
			expectedCount:  TestPageSize,
			expectedTotal:  10,
			expectedOffset: 0,
			expectedLimit:  TestPageSize,
		},
		{
			name:           "두 번째 페이지",
			offset:         TestPageSize,
			limit:          TestPageSize,
			expectedCount:  TestPageSize,
			expectedTotal:  10,
			expectedOffset: TestPageSize,
			expectedLimit:  TestPageSize,
		},
		{
			name:           "음수 offset 정규화",
			offset:         TestNegativeOffset,
			limit:          TestPageSize,
			expectedCount:  TestPageSize,
			expectedTotal:  10,
			expectedOffset: 0,
			expectedLimit:  TestPageSize,
		},
		{
			name:           "큰 limit 정규화",
			offset:         0,
			limit:          TestLargePageSize,
			expectedCount:  10,
			expectedTotal:  10,
			expectedOffset: 0,
			expectedLimit:  DefaultPageSize,
		},
		{
			name:           "0 limit 정규화",
			offset:         0,
			limit:          0,
			expectedCount:  10,
			expectedTotal:  10,
			expectedOffset: 0,
			expectedLimit:  DefaultPageSize,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files, total, err := repo.GetAll(tc.offset, tc.limit)
			require.NoError(t, err)
			assert.Len(t, files, tc.expectedCount)
			assert.Equal(t, tc.expectedTotal, total)
		})
	}
}

func TestFileRepository_Update_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)
	file := createTestFile("_update")

	// 파일 생성
	err := repo.Create(file)
	require.NoError(t, err)

	// 파일 수정
	file.Status = model.FileStatusEncrypted
	file.Size = TestLargeFileSize

	err = repo.Update(file)
	require.NoError(t, err)

	// 수정 확인
	updatedFile, err := repo.GetByID(file.ID)
	require.NoError(t, err)
	assert.Equal(t, model.FileStatusEncrypted, updatedFile.Status)
	assert.Equal(t, int64(TestLargeFileSize), updatedFile.Size)
}

func TestFileRepository_Update_ErrorCases(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	testCases := []struct {
		name    string
		file    *model.File
		wantErr string
	}{
		{
			name:    "nil 파일",
			file:    nil,
			wantErr: "파일 데이터가 없습니다",
		},
		{
			name: "잘못된 ID",
			file: &model.File{
				ID: TestInvalidFileID,
			},
			wantErr: "유효하지 않은 파일 ID입니다",
		},
		{
			name: "존재하지 않는 파일",
			file: &model.File{
				ID: TestNonExistentID,
			},
			wantErr: "업데이트할 파일을 찾을 수 없습니다",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Update(tc.file)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestFileRepository_Delete_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)
	file := createTestFile("_delete")

	// 파일 생성
	err := repo.Create(file)
	require.NoError(t, err)

	// 파일 삭제
	err = repo.Delete(file.ID)
	require.NoError(t, err)

	// 삭제 확인 (소프트 삭제이므로 GetByID는 에러 반환)
	_, err = repo.GetByID(file.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "파일을 찾을 수 없습니다")
}

func TestFileRepository_Delete_ErrorCases(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	testCases := []struct {
		name    string
		id      uint
		wantErr string
	}{
		{
			name:    "잘못된 ID",
			id:      TestInvalidFileID,
			wantErr: "유효하지 않은 파일 ID입니다",
		},
		{
			name:    "존재하지 않는 파일",
			id:      TestNonExistentID,
			wantErr: "삭제할 파일을 찾을 수 없습니다",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Delete(tc.id)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestFileRepository_GetByStatus_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	// 다양한 상태의 파일들 생성
	statuses := []string{
		model.FileStatusPending,
		model.FileStatusEncrypted,
		model.FileStatusFailed,
	}

	filesPerStatus := 3
	for _, status := range statuses {
		for i := 0; i < filesPerStatus; i++ {
			file := createTestFile(fmt.Sprintf("_%s_%d", status, i))
			file.Status = status
			err := repo.Create(file)
			require.NoError(t, err)
		}
	}

	// 각 상태별 조회 테스트
	for _, status := range statuses {
		t.Run(fmt.Sprintf("status_%s", status), func(t *testing.T) {
			files, total, err := repo.GetByStatus(status, 0, DefaultPageSize)
			require.NoError(t, err)
			assert.Len(t, files, filesPerStatus)
			assert.Equal(t, int64(filesPerStatus), total)

			// 모든 파일이 요청한 상태인지 확인
			for _, file := range files {
				assert.Equal(t, status, file.Status)
			}
		})
	}
}

func TestFileRepository_GetByStatus_ErrorCases(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	testCases := []struct {
		name    string
		status  string
		wantErr string
	}{
		{
			name:    "빈 상태",
			status:  "",
			wantErr: "상태 값이 필요합니다",
		},
		{
			name:    "잘못된 상태",
			status:  "invalid_status",
			wantErr: "유효하지 않은 파일 상태입니다",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files, total, err := repo.GetByStatus(tc.status, 0, DefaultPageSize)
			require.Error(t, err)
			assert.Nil(t, files)
			assert.Zero(t, total)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestFileRepository_GetByChecksumMD5_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)
	file := createTestFile("_checksum")
	uniqueChecksum := "unique123456789abcdef0123456789ab"
	file.ChecksumMD5 = uniqueChecksum

	// 파일 생성
	err := repo.Create(file)
	require.NoError(t, err)

	// 체크섬으로 조회
	foundFile, err := repo.GetByChecksumMD5(uniqueChecksum)
	require.NoError(t, err)
	require.NotNil(t, foundFile)
	assert.Equal(t, file.ID, foundFile.ID)
	assert.Equal(t, uniqueChecksum, foundFile.ChecksumMD5)

	// 존재하지 않는 체크섬 조회
	notFoundFile, err := repo.GetByChecksumMD5("nonexistent_checksum")
	require.NoError(t, err)
	assert.Nil(t, notFoundFile)
}

func TestFileRepository_GetByChecksumMD5_ErrorCases(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	// 빈 체크섬
	file, err := repo.GetByChecksumMD5("")
	require.Error(t, err)
	assert.Nil(t, file)
	assert.Contains(t, err.Error(), "체크섬 값이 필요합니다")
}

func TestFileRepository_Exists_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)
	file := createTestFile("_exists")

	// 파일 생성 전 - 존재하지 않음
	exists, err := repo.Exists(TestValidFileID)
	require.NoError(t, err)
	assert.False(t, exists)

	// 파일 생성
	err = repo.Create(file)
	require.NoError(t, err)

	// 파일 생성 후 - 존재함
	exists, err = repo.Exists(file.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	// 파일 삭제
	err = repo.Delete(file.ID)
	require.NoError(t, err)

	// 파일 삭제 후 - 존재하지 않음 (소프트 삭제)
	exists, err = repo.Exists(file.ID)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFileRepository_Exists_ErrorCases(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	// 잘못된 ID
	exists, err := repo.Exists(TestInvalidFileID)
	require.Error(t, err)
	assert.False(t, exists)
	assert.Contains(t, err.Error(), "유효하지 않은 파일 ID입니다")
}

func TestFileRepository_Count_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	// 초기 카운트 확인
	count, err := repo.Count()
	require.NoError(t, err)
	assert.Zero(t, count)

	// 파일들 생성
	fileCount := 5
	for i := 0; i < fileCount; i++ {
		file := createTestFile(fmt.Sprintf("_count_%d", i))
		createErr := repo.Create(file)
		require.NoError(t, createErr)
	}

	// 생성 후 카운트 확인
	count, err = repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(fileCount), count)

	// 파일 하나 삭제
	deleteErr := repo.Delete(TestValidFileID)
	require.NoError(t, deleteErr)

	// 삭제 후 카운트 확인 (소프트 삭제이므로 카운트 감소)
	count, err = repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(fileCount-1), count)
}

func TestFileRepository_NormalizePagination(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db).(*fileRepository)

	testCases := []struct {
		name           string
		inputOffset    int
		inputLimit     int
		expectedOffset int
		expectedLimit  int
	}{
		{
			name:           "정상 값",
			inputOffset:    10,
			inputLimit:     20,
			expectedOffset: 10,
			expectedLimit:  20,
		},
		{
			name:           "음수 offset",
			inputOffset:    -5,
			inputLimit:     10,
			expectedOffset: 0,
			expectedLimit:  10,
		},
		{
			name:           "0 limit",
			inputOffset:    0,
			inputLimit:     0,
			expectedOffset: 0,
			expectedLimit:  DefaultPageSize,
		},
		{
			name:           "음수 limit",
			inputOffset:    0,
			inputLimit:     -1,
			expectedOffset: 0,
			expectedLimit:  DefaultPageSize,
		},
		{
			name:           "큰 limit",
			inputOffset:    0,
			inputLimit:     TestLargePageSize,
			expectedOffset: 0,
			expectedLimit:  DefaultPageSize,
		},
		{
			name:           "MaxPageSize 경계값",
			inputOffset:    0,
			inputLimit:     MaxPageSize,
			expectedOffset: 0,
			expectedLimit:  MaxPageSize,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOffset, actualLimit := repo.normalizePagination(tc.inputOffset, tc.inputLimit)
			assert.Equal(t, tc.expectedOffset, actualOffset)
			assert.Equal(t, tc.expectedLimit, actualLimit)
		})
	}
}

func TestFileRepository_ConcurrentOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	// 동시에 파일 생성
	const goroutineCount = 10
	results := make(chan error, goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func(index int) {
			file := createTestFile(fmt.Sprintf("_concurrent_%d", index))
			file.EncryptedPath = fmt.Sprintf("/encrypted/concurrent_%d.enc", index) // 고유한 경로
			err := repo.Create(file)
			results <- err
		}(i)
	}

	// 모든 고루틴 완료 대기
	for i := 0; i < goroutineCount; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	// 최종 카운트 확인
	count, err := repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(goroutineCount), count)
}

func TestFileRepository_BulkOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	// 대량 파일 생성
	for i := 0; i < TestBulkCreateCount; i++ {
		file := createTestFile(fmt.Sprintf("_bulk_%d", i))
		file.EncryptedPath = fmt.Sprintf("/encrypted/bulk_%d.enc", i)
		err := repo.Create(file)
		require.NoError(t, err)
	}

	// 전체 조회 성능 테스트
	files, total, err := repo.GetAll(0, TestBulkCreateCount)
	require.NoError(t, err)
	assert.Len(t, files, TestBulkCreateCount)
	assert.Equal(t, int64(TestBulkCreateCount), total)

	// 페이지네이션 테스트
	pageSize := 10
	expectedPages := TestBulkCreateCount / pageSize

	totalRetrieved := 0
	for page := 0; page < expectedPages; page++ {
		offset := page * pageSize
		pageFiles, _, err := repo.GetAll(offset, pageSize)
		require.NoError(t, err)
		totalRetrieved += len(pageFiles)
	}

	assert.Equal(t, TestBulkCreateCount, totalRetrieved)
}

func TestFileRepository_EdgeCases(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	t.Run("빈 데이터베이스에서 GetAll", func(t *testing.T) {
		files, total, err := repo.GetAll(0, DefaultPageSize)
		require.NoError(t, err)
		assert.Empty(t, files)
		assert.Zero(t, total)
	})

	t.Run("빈 데이터베이스에서 GetByStatus", func(t *testing.T) {
		files, total, err := repo.GetByStatus(model.FileStatusPending, 0, DefaultPageSize)
		require.NoError(t, err)
		assert.Empty(t, files)
		assert.Zero(t, total)
	})

	t.Run("큰 offset으로 GetAll", func(t *testing.T) {
		// 파일 몇 개 생성
		for i := 0; i < 5; i++ {
			file := createTestFile(fmt.Sprintf("_edge_%d", i))
			err := repo.Create(file)
			require.NoError(t, err)
		}

		// 전체 파일 수보다 큰 offset으로 조회
		files, total, err := repo.GetAll(100, DefaultPageSize)
		require.NoError(t, err)
		assert.Empty(t, files)
		assert.Equal(t, int64(5), total) // 전체 수는 여전히 5
	})

	t.Run("파일 상태 변경 추적", func(t *testing.T) {
		file := createTestFile("_status_change")
		err := repo.Create(file)
		require.NoError(t, err)

		// 상태를 여러 번 변경
		statuses := []string{
			model.FileStatusEncrypted,
			model.FileStatusFailed,
			model.FileStatusCorrupted,
		}

		for _, status := range statuses {
			file.Status = status
			err := repo.Update(file)
			require.NoError(t, err)

			// 업데이트된 상태 확인
			updatedFile, err := repo.GetByID(file.ID)
			require.NoError(t, err)
			assert.Equal(t, status, updatedFile.Status)
		}
	})
}

func TestFileRepository_DataIntegrity(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewFileRepository(db)

	t.Run("체크섬 중복 검사", func(t *testing.T) {
		checksum := "duplicate_checksum_test_123456789abc"

		// 첫 번째 파일 생성
		file1 := createTestFile("_integrity1")
		file1.ChecksumMD5 = checksum
		file1.EncryptedPath = "/encrypted/integrity1.enc"
		err1 := repo.Create(file1)
		require.NoError(t, err1)

		// 같은 체크섬으로 두 번째 파일 생성
		file2 := createTestFile("_integrity2")
		file2.ChecksumMD5 = checksum
		file2.EncryptedPath = "/encrypted/integrity2.enc"
		err2 := repo.Create(file2)
		require.NoError(t, err2) // 체크섬 중복은 허용

		// 체크섬으로 조회 시 첫 번째 파일이 반환되는지 확인
		foundFile, err3 := repo.GetByChecksumMD5(checksum)
		require.NoError(t, err3)
		require.NotNil(t, foundFile)
		assert.Equal(t, file1.ID, foundFile.ID) // 첫 번째 생성된 파일
	})

	t.Run("파일 크기 경계값 테스트", func(t *testing.T) {
		testCases := []struct {
			name string
			size int64
		}{
			{"0 바이트", 0},
			{"1 바이트", 1},
			{"1KB", 1024},
			{"1MB", 1024 * 1024},
			{"1GB", 1024 * 1024 * 1024},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				file := createTestFile(fmt.Sprintf("_size_%d", tc.size))
				file.Size = tc.size
				file.EncryptedPath = fmt.Sprintf("/encrypted/size_%d.enc", tc.size)

				createErr := repo.Create(file)
				require.NoError(t, createErr)

				retrievedFile, getErr := repo.GetByID(file.ID)
				require.NoError(t, getErr)
				assert.Equal(t, tc.size, retrievedFile.Size)
			})
		}
	})
}

// 벤치마크 테스트
func BenchmarkFileRepository_Create(b *testing.B) {
	// 메모리 데이터베이스 사용
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(b, err)

	migrationErr := model.Migrate(db)
	require.NoError(b, migrationErr)

	repo := NewFileRepository(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file := createTestFile(fmt.Sprintf("_bench_%d", i))
		file.EncryptedPath = fmt.Sprintf("/encrypted/bench_%d.enc", i)
		createErr := repo.Create(file)
		if createErr != nil {
			b.Fatal(createErr)
		}
	}
}

func BenchmarkFileRepository_GetByID(b *testing.B) {
	// 메모리 데이터베이스 사용
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(b, err)

	migrationErr := model.Migrate(db)
	require.NoError(b, migrationErr)

	repo := NewFileRepository(db)

	// 벤치마크용 파일들 미리 생성
	fileIDs := make([]uint, b.N)
	for i := 0; i < b.N; i++ {
		file := createTestFile(fmt.Sprintf("_bench_get_%d", i))
		file.EncryptedPath = fmt.Sprintf("/encrypted/bench_get_%d.enc", i)
		createErr := repo.Create(file)
		require.NoError(b, createErr)
		fileIDs[i] = file.ID
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, getErr := repo.GetByID(fileIDs[i])
		if getErr != nil {
			b.Fatal(getErr)
		}
	}
}

func BenchmarkFileRepository_GetAll(b *testing.B) {
	// 메모리 데이터베이스 사용
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(b, err)

	migrationErr := model.Migrate(db)
	require.NoError(b, migrationErr)

	repo := NewFileRepository(db)

	// 벤치마크용 파일들 미리 생성
	for i := 0; i < TestBulkCreateCount; i++ {
		file := createTestFile(fmt.Sprintf("_bench_getall_%d", i))
		file.EncryptedPath = fmt.Sprintf("/encrypted/bench_getall_%d.enc", i)
		createErr := repo.Create(file)
		require.NoError(b, createErr)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, getAllErr := repo.GetAll(0, DefaultPageSize)
		if getAllErr != nil {
			b.Fatal(getAllErr)
		}
	}
}
