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
.PHONY: all build clean test run dev deps help crypto-test db-test db-coverage db-init db-status

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
	@$(GOGET) golang.org/x/crypto
	@$(GOGET) github.com/stretchr/testify
	@$(GOMOD) tidy
	@echo "✅ 의존성 설치 완료"

# 테스트 실행
test:
	@echo "🧪 테스트를 실행합니다..."
	@$(GOTEST) -v ./...

# 암호화 모듈 테스트만 실행
crypto-test:
	@echo "🔐 암호화 모듈 테스트를 실행합니다..."
	@$(GOTEST) -v ./pkg/crypto/...

# 데이터베이스 모듈 테스트만 실행
db-test:
	@echo "🗄️ 데이터베이스 모듈 테스트를 실행합니다..."
	@$(GOTEST) -v ./internal/database/...

# 테스트 커버리지
test-coverage:
	@echo "📊 테스트 커버리지를 확인합니다..."
	@$(GOTEST) -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ 커버리지 리포트: coverage.html"

# 암호화 모듈 커버리지
crypto-coverage:
	@echo "🔐 암호화 모듈 커버리지를 확인합니다..."
	@$(GOTEST) -coverprofile=crypto-coverage.out ./pkg/crypto/...
	@$(GOCMD) tool cover -html=crypto-coverage.out -o crypto-coverage.html
	@echo "✅ 암호화 모듈 커버리지: crypto-coverage.html"

# 데이터베이스 모듈 커버리지
db-coverage:
	@echo "🗄️ 데이터베이스 모듈 커버리지를 확인합니다..."
	@$(GOTEST) -coverprofile=db-coverage.out ./internal/database/...
	@$(GOCMD) tool cover -html=db-coverage.out -o db-coverage.html
	@echo "✅ 데이터베이스 모듈 커버리지: db-coverage.html"

# 벤치마크 테스트
bench:
	@echo "⚡ 벤치마크 테스트를 실행합니다..."
	@$(GOTEST) -bench=. -benchmem ./pkg/crypto/...
	@$(GOTEST) -bench=. -benchmem ./internal/database/...

# 린트 검사
lint:
	@echo "🔍 코드 린트를 실행합니다..."
	@golangci-lint run

lint-fix:
	@echo "🔧 린트 오류를 자동 수정합니다..."
	@golangci-lint run --fix

# 포맷팅 (go fmt + goimports 사용)
fmt:
	@echo "✨ 코드를 포맷팅합니다..."
	@go fmt ./...
	@goimports -local DataLocker -w .
	@echo "✅ 포맷팅 완료"

# 고급 포맷팅 (gofumpt 사용 - 선택적)
fmt-strict:
	@echo "✨ 엄격한 코드 포맷팅을 실행합니다..."
	@gofumpt -w .
	@goimports -local DataLocker -w .
	@echo "✅ 엄격한 포맷팅 완료"

fmt-check:
	@echo "🔍 포맷팅 검사를 실행합니다..."
	@test -z $(shell gofmt -l .) || (echo "다음 파일들이 포맷팅이 필요합니다:" && gofmt -l . && exit 1)

# 정리
clean:
	@echo "🧹 빌드 파일을 정리합니다..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -f crypto-coverage.out crypto-coverage.html
	@rm -f db-coverage.out db-coverage.html
	@rm -rf ./testdata
	@rm -f ./datalocker.db
	@rm -f ./test*.db
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
	@$(GOGET) mvdan.cc/gofumpt@latest
	@$(GOGET) golang.org/x/tools/cmd/goimports@latest
	@echo "✅ 개발 도구 설치 완료"

# 핫 리로드 개발 서버
air:
	@echo "🔥 핫 리로드 개발 서버를 시작합니다..."
	@air

# 데이터베이스 초기화
db-init:
	@echo "🗄️ 데이터베이스를 초기화합니다..."
	@rm -f ./datalocker.db
	@rm -f ./test*.db
	@rm -rf ./testdata
	@echo "✅ 데이터베이스 초기화 완료"

# 데이터베이스 상태 확인
db-status:
	@echo "🗄️ 데이터베이스 상태를 확인합니다..."
	@if [ -f "./datalocker.db" ]; then \
		echo "📁 datalocker.db 파일 존재"; \
		sqlite3 ./datalocker.db ".tables" 2>/dev/null | head -10; \
	else \
		echo "❌ datalocker.db 파일이 없습니다"; \
	fi

# 데이터베이스 스키마 확인
db-schema:
	@echo "🗄️ 데이터베이스 스키마를 확인합니다..."
	@if [ -f "./datalocker.db" ]; then \
		sqlite3 ./datalocker.db ".schema" 2>/dev/null; \
	else \
		echo "❌ datalocker.db 파일이 없습니다"; \
	fi

# 전체 테스트 (모든 모듈)
test-all:
	@echo "🧪 전체 모듈 테스트를 실행합니다..."
	@make crypto-test
	@make db-test

# 전체 커버리지 (모든 모듈)
coverage-all:
	@echo "📊 전체 모듈 커버리지를 확인합니다..."
	@make crypto-coverage
	@make db-coverage
	@make test-coverage

# 개발 환경 설정
setup-dev:
	@echo "🛠️ 개발 환경을 설정합니다..."
	@make deps
	@make install-tools
	@make db-init
	@echo "✅ 개발 환경 설정 완료"

# CI/CD 테스트 (GitHub Actions와 동일한 테스트)
ci-test:
	@echo "🔄 CI/CD 테스트를 실행합니다..."
	@make fmt
	@make lint
	@make test
	@make build
	@echo "✅ CI/CD 테스트 완료"

# 포맷팅 후 린트 실행 (개발 시 유용)
format-and-lint:
	@echo "✨ 포맷팅 후 린트를 실행합니다..."
	@make fmt
	@make lint

# 도움말
help:
	@echo "📋 DataLocker Build Commands:"
	@echo ""
	@echo "🚀 Development:"
	@echo "  make dev             - 개발 서버 실행"
	@echo "  make air             - 핫 리로드 개발 서버"
	@echo "  make setup-dev       - 개발 환경 초기 설정"
	@echo ""
	@echo "🔨 Build & Run:"
	@echo "  make build           - 애플리케이션 빌드"
	@echo "  make run             - 빌드된 서버 실행"
	@echo "  make clean           - 빌드 파일 정리"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  make test            - 전체 테스트 실행"
	@echo "  make crypto-test     - 암호화 모듈 테스트만 실행"
	@echo "  make db-test         - 데이터베이스 모듈 테스트만 실행"
	@echo "  make test-all        - 모든 모듈 테스트 실행"
	@echo "  make bench           - 벤치마크 테스트 실행"
	@echo "  make ci-test         - CI/CD 스타일 테스트"
	@echo ""
	@echo "📊 Coverage:"
	@echo "  make test-coverage   - 전체 테스트 커버리지 확인"
	@echo "  make crypto-coverage - 암호화 모듈 커버리지 확인"
	@echo "  make db-coverage     - 데이터베이스 모듈 커버리지 확인"
	@echo "  make coverage-all    - 모든 모듈 커버리지 확인"
	@echo ""
	@echo "🗄️ Database:"
	@echo "  make db-init         - 데이터베이스 초기화"
	@echo "  make db-status       - 데이터베이스 상태 확인"
	@echo "  make db-schema       - 데이터베이스 스키마 확인"
	@echo ""
	@echo "🛠️ Tools:"
	@echo "  make deps            - 의존성 설치"
	@echo "  make install-tools   - 개발 도구 설치"
	@echo "  make fmt             - 기본 코드 포맷팅 (go fmt + goimports)"
	@echo "  make fmt-strict      - 엄격한 코드 포맷팅 (gofumpt + goimports)"
	@echo "  make fmt-check       - 포맷팅 검사"
	@echo "  make lint            - 린트 검사"
	@echo "  make lint-fix        - 린트 오류 자동 수정"
	@echo "  make format-and-lint - 포맷팅 후 린트 실행"
	@echo "  make health-check    - 헬스체크 테스트"
	@echo ""
	@echo "ℹ️  Help:"
	@echo "  make help            - 이 도움말 표시"