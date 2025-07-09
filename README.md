# DiagnoseQuery

🗄️ **SQL 쿼리 성능 분석 도구** - PostgreSQL 쿼리의 실행 계획을 분석하고 성능을 진단하는 TUI 애플리케이션

## ✨ 주요 기능

- 🔧 **환경변수 설정**: `.env` 파일, 시스템 환경변수, 직접 입력 지원
- 🚀 **쿼리 실행**: SQL 쿼리를 실행하고 결과를 테이블 형태로 표시
- 🔍 **쿼리 분석**: `EXPLAIN ANALYZE`를 통한 상세 성능 분석
- 📊 **진단 리포트**: 쿼리 품질 평가 및 개선 제안
- 📚 **히스토리 관리**: SQLite3 기반 실행 히스토리 저장
- 🎨 **TUI 인터페이스**: BubbleTea 기반의 직관적인 터미널 UI

## 🛠️ 설치 및 빌드

```bash
# 저장소 클론
git clone <repository-url>
cd diagnoseQuery

# 의존성 다운로드
go mod download

# 빌드
go build -o diagnoseQuery ./cmd/diagnoseQuery

# 실행
./diagnoseQuery
```

## 🚀 사용법

### 기본 실행
```bash
./diagnoseQuery
```

### 명령행 옵션
```bash
# 특정 .env 파일 사용
./diagnoseQuery --env-fpath .env.local

# 환경변수 키 지정
./diagnoseQuery --envs DB_URL,DB_SECRET

# 도움말 보기
./diagnoseQuery --help

# 버전 정보
./diagnoseQuery --version
```

## 🔧 환경 설정

### 1. .env 파일 사용
```env
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=username
DB_SECRET=password
DB_DB=database_name
FERNET_KEY=optional_encryption_key
```

### 2. 시스템 환경변수
```bash
export DB_URL=postgresql://user:pass@localhost:5432/dbname
export DB_SECRET=your_password
```

### 3. 직접 입력
프로그램 실행 시 대화형으로 데이터베이스 연결 정보를 입력할 수 있습니다.

## 📊 기능 소개

### 쿼리 실행
- SQL 쿼리를 입력하고 `Enter`로 실행
- 결과를 테이블 형태로 표시
- 암호화된 데이터 자동 복호화 지원

### 쿼리 분석
- `Ctrl+E`로 쿼리 성능 분석 실행
- 실행 계획, 비용 정보, 실제 실행 시간 표시
- 쿼리 품질 평가 및 개선 제안

### 히스토리 관리
- 모든 쿼리 실행 기록을 SQLite3 데이터베이스에 저장
- 환경설정 사용 이력 관리
- 최근 사용한 설정 빠른 접근

## 🎮 키보드 단축키

### 환경 설정 화면
- `↑/↓`: 옵션 선택
- `Enter`: 선택 확인
- `Tab`: 필드 이동 (직접 입력 모드)
- `Esc`: 이전 단계로 돌아가기

### 쿼리 입력 화면
- `Enter`: 쿼리 실행
- `Ctrl+E`: 쿼리 성능 분석
- `Ctrl+C`: 프로그램 종료

### 결과 화면
- `Tab`: 쿼리 입력으로 돌아가기
- `↑/↓`: 스크롤 (분석 결과 화면)
- `Ctrl+C`: 프로그램 종료

## 🏗️ 프로젝트 구조

```
diagnoseQuery/
├── cmd/diagnoseQuery/          # 메인 애플리케이션
│   ├── main.go                 # 진입점
│   └── flags.go                # 명령행 인자 처리
├── internal/
│   ├── analyse/                # 쿼리 분석 엔진
│   │   ├── query_analyzer.go   # 분석 로직
│   │   ├── formatter.go        # 결과 포맷팅
│   │   └── models.go           # 분석 결과 모델
│   ├── database/               # 데이터베이스 연결
│   │   ├── dsn.go              # DSN 생성
│   │   └── models.go           # DB 모델
│   ├── encrypt/                # 암호화/복호화
│   │   └── fernet.go           # Fernet 암호화
│   ├── history/                # 히스토리 관리
│   │   ├── manager.go          # SQLite3 히스토리 매니저
│   │   └── models.go           # 히스토리 모델
│   ├── log/                    # 로깅 및 포맷팅
│   │   ├── log.go              # 로깅 유틸리티
│   │   └── mdFormatting.go     # 마크다운 테이블 포맷
│   ├── parameter/              # 환경변수 관리
│   │   ├── inmemory.go         # 인메모리 스토어
│   │   ├── readEnv.go          # .env 파일 읽기
│   │   ├── uploadEnv.go        # 환경변수 로드
│   │   └── required.go         # 필수 파라미터
│   └── tui/                    # 터미널 UI
│       ├── main_model.go       # 메인 모델
│       ├── env_view.go         # 환경설정 뷰
│       ├── env_update.go       # 환경설정 업데이트
│       ├── model.go            # TUI 모델
│       ├── tui.go              # TUI 초기화
│       ├── update.go           # 업데이트 로직
│       └── view.go             # 뷰 렌더링
├── go.mod                      # Go 모듈 정의
├── go.sum                      # 의존성 체크섬
└── README.md                   # 프로젝트 문서
```

## 🔍 지원 데이터베이스

- ✅ PostgreSQL
- ✅ MySQL/MariaDB (기본 지원)
- 🔄 다른 데이터베이스는 향후 지원 예정

## 📦 의존성

- [BubbleTea](https://github.com/charmbracelet/bubbletea) - TUI 프레임워크
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - 터미널 스타일링
- [PostgreSQL Driver](https://github.com/lib/pq) - PostgreSQL 연결
- [MySQL Driver](https://github.com/go-sql-driver/mysql) - MySQL 연결
- [SQLite3 Driver](https://github.com/mattn/go-sqlite3) - 히스토리 저장
- [Fernet](https://github.com/fernet/fernet-go) - 데이터 암호화
- [GoDotEnv](https://github.com/joho/godotenv) - .env 파일 처리

## 📝 라이선스

이 프로젝트는 MIT 라이선스 하에 배포됩니다.

## 🤝 기여

버그 리포트, 기능 제안, 풀 리퀘스트는 언제든 환영합니다!

## 📞 지원

문제가 있으시면 이슈를 등록해 주세요.