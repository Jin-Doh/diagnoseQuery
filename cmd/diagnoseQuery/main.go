package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"

	"diagnoseQuery/internal/analyse"
	"diagnoseQuery/internal/encrypt"
	"diagnoseQuery/internal/parameter"
)

// getQueryCost는 EXPLAIN (ANALYZE, BUFFERS)를 실행하여 쿼리를 심층 분석합니다.
func getQueryCost(dsn string, query string) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(fmt.Sprintf("DB 연결 실패: %v", err))
	}
	defer conn.Close()

	// Create analyzer and formatter
	analyzer := analyse.NewQueryAnalyzer(conn)
	formatter := analyse.NewResultFormatter()

	// Analyze query
	result, err := analyzer.AnalyzeQuery(query)
	if err != nil {
		fmt.Printf("쿼리 분석 실패: %v\n", err)
		return
	}

	// Format and display results
	output := formatter.FormatAnalysisResult(result)
	fmt.Print(output)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		PrintHelp()
		return
	}

	store := parameter.NewInMemoryStore()
	if err := ProcessFlags(args, store); err != nil {
		panic(fmt.Sprintf("플래그 처리 실패: %v", err))
	}

	requiredParams := parameter.RequiredParameter{
		DBType:   store.GetOrDefault("DB_TYPE", "postgres"),
		Host:     store.GetOrDefault("DB_HOST", "localhost"),
		Port:     store.GetOrDefault("DB_PORT", "5432"),
		User:     store.GetOrDefault("DB_USER", "user"),
		Password: store.GetOrDefault("DB_SECRET", "password"),
		Database: store.GetOrDefault("DB_DB", "database"),
	}

	dsn := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable",
		requiredParams.DBType,
		requiredParams.User,
		requiredParams.Password,
		requiredParams.Host,
		requiredParams.Port,
		requiredParams.Database)

	fernetKey, exists := store.Get("FERNET_KEY")
	if !exists {
		panic("FERNET_KEY 환경변수가 필요합니다")
	}

	if err := encrypt.InjectFernetKey(fernetKey); err != nil {
		panic(fmt.Sprintf("Fernet 키 초기화 실패: %v", err))
	}

	if len(args) == 0 {
		PrintHelp()
		return
	}

	switch args[0] {
	case "--query":
		if len(args) < 2 {
			fmt.Println("--query 사용 시 SQL문을 인자로 입력하세요.")
			return
		}
		query := strings.Join(args[1:], " ")
		generateQuery(dsn, query)
	case "--explain":
		if len(args) < 2 {
			fmt.Println("--explain 사용 시 SQL문을 인자로 입력하세요.")
			return
		}
		query := strings.Join(args[1:], " ")
		getQueryCost(dsn, query)
	case "--help", "-h":
		PrintHelp()
	default:
		fmt.Println("알 수 없는 옵션:", args[0])
		PrintHelp()
	}
}
