# DataLocker v2.0

파일 및 폴더 DRM/암복호화 솔루션

## 🏗️ 프로젝트 구조

```
DataLocker/
├── cmd/server/              # 서버 진입점
│   └── main.go
├── internal/                # 내부 패키지
│   ├── config/             # 설정 관리
│   ├── handler/            # HTTP 핸들러
│   ├── middleware/         # 미들웨어
│   ├── service/            # 비즈니스 로직
│   ├── repository/         # 데이터 접근 계층
│   └── model/              # 데이터 모델
├── pkg/                    # 공용 패키지
│   ├── crypto/             # 암호화 유틸리티 ⭐ NEW
│   ├── fileutil/          # 파일 유틸리티
│   └── response/          # API 응답 유틸리티
├── frontend/               # Wails React 프론트엔드
├── test/                   # 테스트 파일
├── docs/                   # 문서
└── build/                  # 빌드 결과물
```

## 🚀 빠른 시작

### 1. 의존성 설치
```bash
make deps
```

### 2. 개발 서버 실행
```bash
make dev
```

### 3. 헬스체크 확인
```bash
curl http://localhost:8080/api/v1/health
```

## 📋 사용 가능한 명령어

```bash
make dev             # 개발 서버 실행
make build           # 애플리케이션 빌드
make test            # 전체 테스트 실행
make crypto-test     # 암호화 모듈 테스트만 실행 ⭐ NEW
make test-coverage   # 전체 테스트 커버리지
make crypto-coverage # 암호화 모듈 커버리지 ⭐ NEW
make bench           # 벤치마크 테스트 ⭐ NEW
make clean           # 빌드 파일 정리
make help            # 전체 명령어 보기
```

## 🔧 개발 도구

### Air (핫 리로드)
```bash
make install-tools  # 개발 도구 설치
make air           # 핫 리로드 서버 실행
```

## 🔐 암호화 모듈

### 핵심 기능
- **AES-256-GCM**: 고급 암호화 표준
- **PBKDF2**: 패스워드 기반 키 유도
- **스트림 처리**: 대용량 파일 암복호화
- **안전한 랜덤**: Salt/Nonce 생성

### 사용 예시
```go
import "DataLocker/pkg/crypto"

engine := crypto.NewCryptoEngine()

// 간단한 데이터 암호화
data := []byte("비밀 데이터")
encrypted, err := engine.Encrypt(data, "mypassword")

// 복호화
decrypted, err := engine.Decrypt(encrypted, "mypassword")

// 스트림 암호화 (대용량 파일)
err = engine.EncryptStream(reader, writer, "mypassword")
```

### 테스트 실행
```bash
# 암호화 모듈만 테스트
make crypto-test

# 커버리지 확인
make crypto-coverage

# 성능 벤치마크
make bench
```

## 📡 API 엔드포인트

### 헬스체크
- `GET /api/v1/health` - 전체 헬스체크
- `GET /api/v1/health/ready` - 준비 상태 확인
- `GET /api/v1/health/live` - 라이브니스 확인
- `GET /api/v1/health/metrics` - 시스템 메트릭

### 기본
- `GET /` - 서버 정보
- `GET /docs` - API 문서

## 🧪 테스트

```bash
# 전체 테스트 실행
make test

# 암호화 모듈만 테스트
make crypto-test

# 커버리지 포함 테스트
make test-coverage

# 특정 패키지 테스트
go test ./pkg/crypto/...
```

## 📦 기술 스택

- **Backend**: Go + Echo Framework
- **Frontend**: Wails + React + TypeScript
- **Database**: SQLite + GORM
- **Crypto**: AES-256-GCM + PBKDF2
- **Build**: Make + Air (핫 리로드)

## 🔧 환경 변수

```bash
PORT=8080                    # 서버 포트
HOST=localhost               # 서버 호스트
LOG_LEVEL=info              # 로그 레벨
ENVIRONMENT=development      # 환경 설정
MAX_FILE_SIZE=1073741824    # 최대 파일 크기 (1GB)
DB_PATH=./datalocker.db     # 데이터베이스 경로
```

## 📝 개발 진행 상황

### ✅ 완료된 작업 (이슈 #1)
- [x] Go 프로젝트 구조 설정
- [x] Echo 프레임워크 설치 및 기본 서버 설정
- [x] CORS, 로깅 미들웨어 설정
- [x] Health check 엔드포인트 구현
- [x] 기본 에러 핸들링 미들웨어
- [x] 설정 관리 시스템
- [x] 응답 유틸리티 패키지
- [x] 기본 테스트 설정

### ✅ 완료된 작업 (이슈 #2) ⭐ NEW
- [x] AES-256-GCM 암호화 엔진 구현
- [x] PBKDF2 키 유도 함수 구현
- [x] 안전한 Salt/Nonce 생성
- [x] 스트림 방식 대용량 파일 처리
- [x] 포괄적인 단위 테스트 (25개 테스트 케이스)
- [x] 에러 처리 및 검증 로직
- [x] 성능 벤치마크 테스트
- [x] 완전한 테스트 커버리지

### 🔄 다음 작업 (이슈 #3)
- [ ] 파일 시스템 유틸리티 구현
- [ ] 메타데이터 관리
- [ ] 파일 무결성 검증
- [ ] 압축 기능 통합

## 🔒 보안 고려사항

- **AES-256-GCM**: 인증된 암호화로 무결성 보장
- **PBKDF2**: 100,000 반복으로 무차별 대입 공격 방어
- **랜덤 Salt/Nonce**: 각 암호화마다 고유한 값 사용
- **메모리 보안**: 민감한 데이터 즉시 삭제
- **스트림 처리**: 메모리 사용량 최적화

## 🤝 기여

1. 이슈 확인 및 브랜치 생성
2. 기능 구현 및 테스트 작성
3. 커밋 및 푸시
4. Pull Request 생성

## 📄 라이선스

MIT License