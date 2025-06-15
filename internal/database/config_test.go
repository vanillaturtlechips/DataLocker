package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"DataLocker/internal/config"
)

// 테스트용 상수
const (
	TestDBPath  = "./test_datalocker.db"
	TestDBDir   = "./testdata"
	TestTimeout = 5 * time.Second
)

// createTestConfig 테스트용 설정을 생성합니다
func createTestConfig(dbPath string) *config.Config {
	return &config.Config{
		Database: config.DatabaseConfig{
			Path:        dbPath,
			AutoMigrate: true,
		},
		App: config.AppConfig{
			LogLevel: "error", // 테스트 시 로그 최소화
		},
	}
}

// setupTestDB 테스트용 데이터베이스를 설정합니다
func setupTestDB(t *testing.T) (*Database, func()) {
	// 테스트 디렉토리 생성
	err := os.MkdirAll(TestDBDir, 0755)
	require.NoError(t, err)

	// 고유한 테스트 DB 파일명 생성
	dbPath := filepath.Join(TestDBDir, "test_"+t.Name()+".db")

	cfg := createTestConfig(dbPath)

	db, err := NewDatabase(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)

	// 정리 함수 반환
	cleanup := func() {
		if db != nil {
			_ = db.Close()
		}
		_ = os.Remove(dbPath)
		_ = os.RemoveAll(TestDBDir)
	}

	return db, cleanup
}

func TestNewDatabase_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 데이터베이스가 올바르게 생성되었는지 확인
	assert.NotNil(t, db)
	assert.NotNil(t, db.DB)
	assert.NotNil(t, db.config)
}

func TestNewDatabase_NilConfig(t *testing.T) {
	_, err := NewDatabase(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config는 필수입니다")
}

func TestDatabase_HealthCheck(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 헬스체크 성공
	err := db.HealthCheck()
	assert.NoError(t, err)
}

func TestDatabase_HealthCheck_ClosedConnection(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 연결 종료
	err := db.Close()
	require.NoError(t, err)

	// 헬스체크 실패 확인
	err = db.HealthCheck()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "데이터베이스가 연결되지 않았습니다")
}

func TestDatabase_Close(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 정상 종료
	err := db.Close()
	assert.NoError(t, err)
	assert.Nil(t, db.DB)

	// 중복 종료 시도 (에러 없이 처리)
	err = db.Close()
	assert.NoError(t, err)
}

func TestDatabase_GetStats(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	stats, err := db.GetStats()
	require.NoError(t, err)
	require.NotNil(t, stats)

	// 통계 필드 확인
	expectedFields := []string{
		"max_open_connections",
		"open_connections",
		"in_use",
		"idle",
		"wait_count",
		"wait_duration",
		"max_idle_closed",
		"max_idle_time_closed",
		"max_lifetime_closed",
	}

	for _, field := range expectedFields {
		assert.Contains(t, stats, field)
	}

	// 연결 수 확인
	assert.Equal(t, MaxOpenConns, stats["max_open_connections"])
}

func TestDatabase_GetStats_ClosedConnection(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 연결 종료
	err := db.Close()
	require.NoError(t, err)

	// 통계 조회 실패 확인
	_, err = db.GetStats()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "데이터베이스가 연결되지 않았습니다")
}

func TestDatabase_ConnectionString(t *testing.T) {
	testCases := []struct {
		name         string
		dbPath       string
		expectedPath string
	}{
		{
			name:         "상대 경로",
			dbPath:       "test.db",
			expectedPath: "./test.db",
		},
		{
			name:         "절대 경로 Unix",
			dbPath:       "/tmp/test.db",
			expectedPath: "/tmp/test.db",
		},
		{
			name:         "절대 경로 Windows",
			dbPath:       "C:\\tmp\\test.db",
			expectedPath: "C:/tmp/test.db",
		},
		{
			name:         "현재 디렉토리",
			dbPath:       "./test.db",
			expectedPath: "./test.db",
		},
		{
			name:         "Windows 상대 경로",
			dbPath:       ".\\test.db",
			expectedPath: "./test.db",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := createTestConfig(tc.dbPath)
			db := &Database{config: cfg}

			connStr := db.buildConnectionString()

			// 경로 부분 확인 (옵션 제거)
			pathPart := strings.Split(connStr, "?")[0]
			assert.Equal(t, tc.expectedPath, pathPart)

			// SQLite 옵션 확인
			assert.Contains(t, connStr, "_busy_timeout=")
			assert.Contains(t, connStr, "_journal_mode=")
			assert.Contains(t, connStr, "_sync=")
			assert.Contains(t, connStr, "_cache_size=")
			assert.Contains(t, connStr, "_foreign_keys=")
		})
	}
}

func TestDatabase_SQLitePragmas(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// PRAGMA 설정 확인
	testCases := []struct {
		pragma   string
		expected string
	}{
		{"PRAGMA journal_mode", "wal"},
		{"PRAGMA synchronous", "1"},  // NORMAL = 1
		{"PRAGMA foreign_keys", "1"}, // ON = 1
	}

	for _, tc := range testCases {
		t.Run(tc.pragma, func(t *testing.T) {
			var result string
			err := db.DB.Raw(tc.pragma).Scan(&result).Error
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDatabase_ConnectionPool(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	stats, err := db.GetStats()
	require.NoError(t, err)

	// 연결 풀 설정 확인
	assert.Equal(t, MaxOpenConns, stats["max_open_connections"])
	assert.GreaterOrEqual(t, stats["open_connections"].(int), 1)
}

func TestDatabase_MultipleConnections(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// 동시에 여러 쿼리 실행
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()

			var result int
			err := db.DB.Raw("SELECT ? as id", id).Scan(&result).Error
			assert.NoError(t, err)
			assert.Equal(t, id, result)
		}(i)
	}

	// 모든 고루틴 완료 대기
	for i := 0; i < 5; i++ {
		select {
		case <-done:
			// 성공
		case <-time.After(TestTimeout):
			t.Fatal("타임아웃: 동시 연결 테스트 실패")
		}
	}
}

// 벤치마크 테스트
func BenchmarkDatabase_HealthCheck(b *testing.B) {
	cfg := createTestConfig(":memory:")
	db, err := NewDatabase(cfg)
	require.NoError(b, err)
	defer func() { _ = db.Close() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := db.HealthCheck()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDatabase_GetStats(b *testing.B) {
	cfg := createTestConfig(":memory:")
	db, err := NewDatabase(cfg)
	require.NoError(b, err)
	defer func() { _ = db.Close() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.GetStats()
		if err != nil {
			b.Fatal(err)
		}
	}
}
