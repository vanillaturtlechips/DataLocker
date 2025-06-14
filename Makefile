# DataLocker Makefile

# 기본 변수
APP_NAME := DataLocker
VERSION := 2.0.0
BUILD_DIR := build
CMD_DIR := cmd/server

# Go 관련 변수
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# 빌드 플래그
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(shell date -u +%Y%m%d.%H%M%S)"

# 기본 타겟
.PHONY: all build clean test run dev deps help

# 기본 명령어
all: deps test build

# 개발 서버 실행
dev:
	@echo "🚀 개발 서버를 시작합니다..."
	@$(GOCMD) run $(CMD_DIR)/main.go

# 서버 실행 (빌드된 바이너리)
run: build
	@echo "🏃 서버를 실행합니다..."
	@./$(BUILD_DIR)/server

# 빌드
build:
	@echo "🔨 애플리케이션을 빌드합니다..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/server $(CMD_DIR)/main.go
	@echo "✅ 빌드 완료: $(BUILD_DIR)/server"

# 의존성 설치
deps:
	@echo "📦 의존성을 설치합니다..."
	@$(GOGET) github.com/labstack/echo/v4
	@$(GOGET) github.com/labstack/echo/v4/middleware
	@$(GOGET) github.com/sirupsen/logrus
	@$(GOGET) gorm.io/gorm
	@$(GOGET) gorm.io/driver/sqlite
	@$(GOMOD) tidy
	@echo "✅ 의존성 설치 완료"

# 테스트 실행
test:
	@echo "🧪 테스트를 실행합니다..."
	@$(GOTEST) -v ./...

# 테스트 커버리지
test-coverage:
	@echo "📊 테스트 커버리지를 확인합니다..."
	@$(GOTEST) -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ 커버리지 리포트: coverage.html"

# 린트 검사
lint:
	@echo "🔍 코드 린트를 실행합니다..."
	@golangci-lint run

# 포맷팅
fmt:
	@echo "✨ 코드를 포맷팅합니다..."
	@$(GOCMD) fmt ./...

# 정리
clean:
	@echo "🧹 빌드 파일을 정리합니다..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ 정리 완료"

# 헬스체크 테스트
health-check:
	@echo "🩺 헬스체크를 테스트합니다..."
	@curl -s http://localhost:8080/api/v1/health | jq .

# 개발 도구 설치
install-tools:
	@echo "🛠️ 개발 도구를 설치합니다..."
	@$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GOGET) github.com/air-verse/air@latest
	@echo "✅ 개발 도구 설치 완료"

# 핫 리로드 개발 서버
air:
	@echo "🔥 핫 리로드 개발 서버를 시작합니다..."
	@air

# 도움말
help:
	@echo "DataLocker Build Commands:"
	@echo ""
	@echo "  make dev           - 개발 서버 실행"
	@echo "  make build         - 애플리케이션 빌드"
	@echo "  make run           - 빌드된 서버 실행"
	@echo "  make test          - 테스트 실행"
	@echo "  make test-coverage - 테스트 커버리지 확인"
	@echo "  make deps          - 의존성 설치"
	@echo "  make clean         - 빌드 파일 정리"
	@echo "  make fmt           - 코드 포맷팅"
	@echo "  make lint          - 린트 검사"
	@echo "  make health-check  - 헬스체크 테스트"
	@echo "  make install-tools - 개발 도구 설치"
	@echo "  make air           - 핫 리로드 개발 서버"
	@echo "  make help          - 이 도움말 표시"