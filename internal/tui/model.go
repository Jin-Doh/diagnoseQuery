package tui

import (
	"database/sql"
	"diagnoseQuery/internal/analyse"
	"diagnoseQuery/internal/encrypt"
	logpkg "diagnoseQuery/internal/log"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
)

// TUI의 현재 화면 상태를 나타내는 열거형
type TuiState int

const (
	viewQueryInput TuiState = iota
	viewLoading
	viewQueryResult
	viewAnalysisResult
	viewError
)

// model은 TUI 애플리케이션의 전체 상태를 정의합니다.
type model struct {
	dsn            string
	state          TuiState
	textarea       textarea.Model
	table          table.Model
	viewport       viewport.Model // 분석 결과나 에러 메시지를 보여주기 위함
	spinner        spinner.Model
	analysisResult analyse.QueryAnalysisResult
	err            error
}

// Cmd 실행 후 돌아올 메시지들을 정의합니다.
type queryResultMsg struct {
	columns []string
	rows    [][]string
}

type analysisResultMsg struct {
	result analyse.QueryAnalysisResult
}

type errorMsg struct {
	err error
}

// Query 실행 및 결과 출력
func scanRow(rows *sql.Rows, numCols int) ([]string, error) {
	values := make([]sql.RawBytes, numCols)
	scanArgs := make([]interface{}, numCols)
	for i := range values {
		scanArgs[i] = &values[i]
	}
	if err := rows.Scan(scanArgs...); err != nil {
		return nil, fmt.Errorf("쿼리 결과 스캔 실패: %v", err)
	}

	rowResult := make([]string, numCols)
	for i, col := range values {
		if col == nil {
			rowResult[i] = "NULL"
			continue
		}

		decrypted, err := encrypt.FernetDecrypt(string(col))
		if err == nil && decrypted != "" {
			rowResult[i] = decrypted
		} else {
			rowResult[i] = string(col)
		}
	}
	return rowResult, nil
}

func generateQuery(dsn string, query string) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(fmt.Sprintf("DB 연결 실패: %v", err))
	}
	defer conn.Close()

	rows, err := conn.Query(query)
	if err != nil {
		panic(fmt.Sprintf("쿼리 실행 실패: %v", err))
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		panic(fmt.Sprintf("컬럼 조회 실패: %v", err))
	}
	numCols := len(columns)

	// 결과 저장용
	var results [][]string

	for rows.Next() {
		rowResult, err := scanRow(rows, numCols)
		if err != nil {
			panic(err)
		}
		results = append(results, rowResult)
	}

	if len(results) == 0 {
		fmt.Println("쿼리 결과가 없습니다.")
		return
	}

	// 각 컬럼별 최대 길이 계산 (헤더, 데이터 모두 포함)
	colWidths := make([]int, numCols)
	for i, col := range columns {
		colWidths[i] = len(col)
	}
	for _, row := range results {
		for i, v := range row {
			if l := len(v); l > colWidths[i] {
				colWidths[i] = l
			}
		}
	}

	fmt.Printf("쿼리(%s) 결과:\n", query)
	logpkg.PrintMarkdownTable(columns, results)
}
