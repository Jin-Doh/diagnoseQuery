package main

import (
	"flag"
	"fmt"
	"os"
)

// CommandLineOptions holds the parsed command line options
type CommandLineOptions struct {
	EnvPath        string
	EnvKeys        string
	ShowHelp       bool
	ShowVersion    bool
	GenerateSample bool
}

// ParseFlags parses command line flags and returns options
func ParseFlags() *CommandLineOptions {
	opts := &CommandLineOptions{}

	flag.StringVar(&opts.EnvPath, "env-fpath", "", ".env 파일 경로 지정")
	flag.StringVar(&opts.EnvKeys, "envs", "", "환경변수 키 지정 (콤마로 구분)")
	flag.BoolVar(&opts.ShowHelp, "help", false, "도움말 표시")
	flag.BoolVar(&opts.ShowHelp, "h", false, "도움말 표시 (단축)")
	flag.BoolVar(&opts.ShowVersion, "version", false, "버전 정보 표시")
	flag.BoolVar(&opts.ShowVersion, "v", false, "버전 정보 표시 (단축)")
	flag.BoolVar(&opts.GenerateSample, "generate-sample", false, "sample.env 파일 생성")

	flag.Parse()

	return opts
}

// PrintHelp prints help information
func PrintHelp() {
	fmt.Println("DiagnoseQuery - SQL 쿼리 성능 분석 도구")
	fmt.Println()
	fmt.Println("사용법:")
	fmt.Println("  diagnoseQuery [옵션]")
	fmt.Println()
	fmt.Println("옵션:")
	fmt.Println("  --env-fpath PATH       .env 파일 경로를 지정합니다")
	fmt.Println("  --envs KEYS            환경변수 키를 지정합니다 (콤마로 구분)")
	fmt.Println("  --generate-sample      sample.env 파일을 생성합니다")
	fmt.Println("  -h, --help             이 도움말을 표시합니다")
	fmt.Println("  -v, --version          버전 정보를 표시합니다")
	fmt.Println()
	fmt.Println("기능:")
	fmt.Println("  • 환경변수 설정 (.env 파일, 시스템 환경변수, 직접 입력)")
	fmt.Println("  • SQL 쿼리 실행")
	fmt.Println("  • 쿼리 성능 분석")
	fmt.Println("  • 실행 히스토리 관리")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  diagnoseQuery                          # TUI 모드로 실행")
	fmt.Println("  diagnoseQuery --env-fpath .env.local   # 특정 .env 파일 사용")
	fmt.Println("  diagnoseQuery --envs DB_URL,DB_SECRET  # 환경변수 키 지정")
	fmt.Println("  diagnoseQuery --generate-sample        # sample.env 파일 생성")
}

// PrintVersion prints version information
func PrintVersion() {
	fmt.Println("DiagnoseQuery v1.0.0")
	fmt.Println("Go로 개발된 SQL 쿼리 성능 분석 도구")
}

// ShouldShowHelp returns true if help should be shown and program should exit
func (opts *CommandLineOptions) ShouldShowHelp() bool {
	return opts.ShowHelp || len(os.Args) > 1 && (os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h")
}

// ShouldShowVersion returns true if version should be shown and program should exit
func (opts *CommandLineOptions) ShouldShowVersion() bool {
	return opts.ShowVersion || len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version" || os.Args[1] == "-v")
}

// HasEnvPath returns true if env file path is specified
func (opts *CommandLineOptions) HasEnvPath() bool {
	return opts.EnvPath != ""
}

// HasEnvKeys returns true if environment keys are specified
func (opts *CommandLineOptions) HasEnvKeys() bool {
	return opts.EnvKeys != ""
}

// GenerateSampleEnv creates a sample.env file with example values
func GenerateSampleEnv() error {
	sampleContent := `# DiagnoseQuery Sample Environment Configuration
# 이 파일을 복사해서 .env 파일로 이름을 바꾸고 실제 값으로 채워주세요.

# 데이터베이스 유형 (postgresql, mysql, sqlite3 등)
DB_TYPE=postgresql

# 데이터베이스 연결 정보
DB_HOST=localhost
DB_PORT=5432
DB_NAME=your_database_name
DB_USER=your_username
DB_SECRET=your_password
DB_SSL_MODE=disable

# 또는 전체 데이터베이스 URL (위의 개별 설정보다 우선함)
# DB_URL=postgresql://username:password@localhost:5432/database_name?sslmode=disable


# 암호화 키 (선택사항 - Fernet 암호화용)
# FERNET_KEY=your_fernet_encryption_key_here
`

	filename := "sample.env"
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("sample.env 파일이 이미 존재합니다")
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("파일 생성 실패: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(sampleContent); err != nil {
		return fmt.Errorf("파일 쓰기 실패: %w", err)
	}

	return nil
}
