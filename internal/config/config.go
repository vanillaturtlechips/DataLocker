// Package config provides configuration management for DataLocker application.
// It handles environment variables, default values, and application settings.
package config

import (
	"os"
	"strconv"
)

// 서버 설정 관련 상수
const (
	// 기본 타임아웃 설정 (초)
	DefaultReadTimeoutSeconds  = 30
	DefaultWriteTimeoutSeconds = 30
)

// 파일 크기 관련 상수
const (
	BytesPerKB = 1024
	BytesPerMB = BytesPerKB * 1024
	BytesPerGB = BytesPerMB * 1024

	// 기본 최대 파일 크기 (1GB)
	DefaultMaxFileSizeBytes = 1 * BytesPerGB
)

// Config 애플리케이션 설정 구조체
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Security SecurityConfig `json:"security"`
	App      AppConfig      `json:"app"`
}

// ServerConfig 서버 관련 설정
type ServerConfig struct {
	Port         string `json:"port"`
	Host         string `json:"host"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
}

// DatabaseConfig 데이터베이스 설정
type DatabaseConfig struct {
	Path        string `json:"path"`
	AutoMigrate bool   `json:"auto_migrate"`
}

// SecurityConfig 보안 설정
type SecurityConfig struct {
	AllowedOrigins []string `json:"allowed_origins"`
	MaxFileSize    int64    `json:"max_file_size"`
}

// AppConfig 앱 관련 설정
type AppConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	LogLevel    string `json:"log_level"`
}

// Load 환경변수에서 설정을 로드합니다
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Host:         getEnv("HOST", "localhost"),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", DefaultReadTimeoutSeconds),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", DefaultWriteTimeoutSeconds),
		},
		Database: DatabaseConfig{
			Path:        getEnv("DB_PATH", "./datalocker.db"),
			AutoMigrate: getEnvAsBool("DB_AUTO_MIGRATE", true),
		},
		Security: SecurityConfig{
			AllowedOrigins: []string{
				getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
				"http://localhost:34115", // Wails dev server
			},
			MaxFileSize: getEnvAsInt64("MAX_FILE_SIZE", DefaultMaxFileSizeBytes),
		},
		App: AppConfig{
			Name:        "DataLocker",
			Version:     "2.0.0",
			Environment: getEnv("ENVIRONMENT", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
	}
}

// getEnv 환경변수를 가져오고, 없으면 기본값을 반환
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 환경변수를 int로 변환
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 환경변수를 int64로 변환
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 환경변수를 bool로 변환
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
