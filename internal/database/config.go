// Package database provides database configuration and connection management for DataLocker.
// It handles SQLite database connections, connection pooling, and GORM integration.
package database

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"DataLocker/internal/config"
)

// 데이터베이스 관련 상수
const (
	// 연결 풀 설정
	MaxIdleConns    = 10
	MaxOpenConns    = 100
	ConnMaxLifetime = time.Hour
	ConnMaxIdleTime = time.Minute * 30

	// SQLite 설정
	SQLitePragmaJournalMode = "WAL"
	SQLitePragmaSyncMode    = "NORMAL"
	SQLitePragmaCacheSize   = -64000 // 64MB
	SQLitePragmaForeignKeys = "ON"

	// 타임아웃 설정 (밀리초)
	BusyTimeoutMs = 30000

	// 파일 권한
	DBFilePermission = 0600
)

// Database 데이터베이스 연결 관리자
type Database struct {
	DB     *gorm.DB
	config *config.Config
}

// DatabaseConfig 데이터베이스 설정 구조체
type DatabaseConfig struct {
	Path         string
	AutoMigrate  bool
	LogLevel     logger.LogLevel
	MaxIdleConns int
	MaxOpenConns int
}

// NewDatabase 새로운 데이터베이스 인스턴스를 생성합니다
func NewDatabase(cfg *config.Config) (*Database, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config는 필수입니다")
	}

	database := &Database{
		config: cfg,
	}

	if err := database.connect(); err != nil {
		return nil, fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}

	return database, nil
}

// connect 데이터베이스에 연결합니다
func (d *Database) connect() error {
	// SQLite 연결 문자열 구성
	dsn := d.buildConnectionString()

	// GORM 로거 설정
	gormLogger := d.configureLogger()

	// GORM 설정
	gormConfig := &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: false,
		SkipDefaultTransaction:                   false,
	}

	// 데이터베이스 연결
	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("GORM 연결 실패: %w", err)
	}

	// 연결 풀 설정
	if err := d.configureConnectionPool(db); err != nil {
		return fmt.Errorf("연결 풀 설정 실패: %w", err)
	}

	// SQLite 최적화 설정
	if err := d.configureSQLite(db); err != nil {
		return fmt.Errorf("SQLite 설정 실패: %w", err)
	}

	d.DB = db
	return nil
}

// buildConnectionString SQLite 연결 문자열을 구성합니다
func (d *Database) buildConnectionString() string {
	dbPath := d.config.Database.Path

	// 절대 경로 여부 판단 (크로스 플랫폼)
	isAbsolute := filepath.IsAbs(dbPath) || strings.HasPrefix(dbPath, "/")

	// 상대 경로 처리 (명시적으로 현재 디렉토리 표시)
	if !isAbsolute {
		// 이미 ./ 또는 .\ 로 시작하지 않는 경우에만 추가
		if !strings.HasPrefix(dbPath, "./") && !strings.HasPrefix(dbPath, ".\\") {
			// 강제로 ./ 접두사 추가
			dbPath = "./" + dbPath
		}
	}

	// 경로 구분자를 슬래시로 통일 (SQLite는 슬래시 선호)
	dbPath = filepath.ToSlash(dbPath)

	// SQLite 연결 옵션
	options := fmt.Sprintf(
		"?_busy_timeout=%d&_journal_mode=%s&_sync=%s&_cache_size=%d&_foreign_keys=%s",
		BusyTimeoutMs,
		SQLitePragmaJournalMode,
		SQLitePragmaSyncMode,
		SQLitePragmaCacheSize,
		SQLitePragmaForeignKeys,
	)

	return dbPath + options
}

// configureLogger GORM 로거를 설정합니다
func (d *Database) configureLogger() logger.Interface {
	var logLevel logger.LogLevel

	switch d.config.App.LogLevel {
	case "debug":
		logLevel = logger.Info
	case "info":
		logLevel = logger.Warn
	case "warn", "error":
		logLevel = logger.Error
	default:
		logLevel = logger.Silent
	}

	return logger.Default.LogMode(logLevel)
}

// configureConnectionPool 연결 풀을 설정합니다
func (d *Database) configureConnectionPool(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("SQL DB 인스턴스 획득 실패: %w", err)
	}

	// 연결 풀 설정
	sqlDB.SetMaxIdleConns(MaxIdleConns)
	sqlDB.SetMaxOpenConns(MaxOpenConns)
	sqlDB.SetConnMaxLifetime(ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(ConnMaxIdleTime)

	return nil
}

// configureSQLite SQLite 특화 설정을 적용합니다
func (d *Database) configureSQLite(db *gorm.DB) error {
	// SQLite 성능 최적화 PRAGMA 실행
	pragmas := []string{
		"PRAGMA journal_mode = " + SQLitePragmaJournalMode,
		"PRAGMA synchronous = " + SQLitePragmaSyncMode,
		"PRAGMA cache_size = " + fmt.Sprintf("%d", SQLitePragmaCacheSize),
		"PRAGMA foreign_keys = " + SQLitePragmaForeignKeys,
		"PRAGMA temp_store = MEMORY",
		"PRAGMA mmap_size = 268435456", // 256MB
	}

	for _, pragma := range pragmas {
		if err := db.Exec(pragma).Error; err != nil {
			return fmt.Errorf("PRAGMA 실행 실패 [%s]: %w", pragma, err)
		}
	}

	return nil
}

// Close 데이터베이스 연결을 종료합니다
func (d *Database) Close() error {
	if d.DB == nil {
		return nil
	}

	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("SQL DB 인스턴스 획득 실패: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("데이터베이스 연결 종료 실패: %w", err)
	}

	d.DB = nil
	return nil
}

// HealthCheck 데이터베이스 연결 상태를 확인합니다
func (d *Database) HealthCheck() error {
	if d.DB == nil {
		return fmt.Errorf("데이터베이스가 연결되지 않았습니다")
	}

	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("SQL DB 인스턴스 획득 실패: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("데이터베이스 핑 실패: %w", err)
	}

	return nil
}

// GetStats 데이터베이스 연결 통계를 반환합니다
func (d *Database) GetStats() (map[string]interface{}, error) {
	if d.DB == nil {
		return nil, fmt.Errorf("데이터베이스가 연결되지 않았습니다")
	}

	sqlDB, err := d.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("SQL DB 인스턴스 획득 실패: %w", err)
	}

	stats := sqlDB.Stats()

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}
