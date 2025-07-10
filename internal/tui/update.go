package tui

import (
	"database/sql"
	"diagnoseQuery/internal/analyse"
	"diagnoseQuery/internal/encrypt"
	"diagnoseQuery/internal/history"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.state == viewQueryInput {
				m.state = viewLoading
				query := m.textarea.Value()
				return m, executeQueryCmd(m.dsn, query)
			}
		case tea.KeyCtrlE: // Ctrl+E로 분석 실행
			if m.state == viewQueryInput {
				m.state = viewLoading
				query := m.textarea.Value()
				return m, analyzeQueryCmd(m.dsn, query)
			}
		case tea.KeyTab:
			// 결과 화면에서 입력 화면으로 돌아가기
			if m.state == viewQueryResult || m.state == viewAnalysisResult || m.state == viewError {
				m.state = viewQueryInput
				m.textarea.Focus()
				return m, nil
			}
		}

	// 비동기 작업 결과 처리
	case queryResultMsg:
		m.state = viewQueryResult
		// 테이블 설정
		columns := []table.Column{}
		for _, col := range msg.columns {
			columns = append(columns, table.Column{Title: col, Width: 15})
		}

		rows := []table.Row{}
		for _, row := range msg.rows {
			rows = append(rows, table.Row(row))
		}

		m.table = table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(10),
		)
		return m, nil

	case analysisResultMsg:
		m.state = viewAnalysisResult
		m.analysisResult = msg.result
		// Set viewport content when we receive the analysis result
		formatter := analyse.NewResultFormatter()
		output := formatter.FormatAnalysisResult(&msg.result)
		m.viewport.SetContent(output)
		return m, nil

	case errorMsg:
		m.state = viewError
		m.err = msg.err
		return m, nil
	}

	// 현재 활성화된 컴포넌트에 이벤트 전달
	var cmd tea.Cmd
	switch m.state {
	case viewQueryInput:
		m.textarea, cmd = m.textarea.Update(msg)
	case viewQueryResult:
		m.table, cmd = m.table.Update(msg)
	case viewAnalysisResult:
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

// getDriverNameFromDSN derives the driver name from the DSN string.
func getDriverNameFromDSN(dsn string) string {
	// Example logic: Parse DSN to extract the database type
	if len(dsn) > 0 {
		if dsn[:8] == "postgres" {
			return "postgres"
		} else if dsn[:5] == "mysql" {
			return "mysql"
		} else if dsn[:6] == "sqlite" {
			return "sqlite3"
		}
	}
	// Default to "postgres" if no match is found
	return "postgres"
}

// executeQueryCmd executes a SQL query and returns the result
func executeQueryCmd(dsn, query string) tea.Cmd {
	return func() tea.Msg {
		startTime := time.Now()

		driverName := getDriverNameFromDSN(dsn)
		conn, err := sql.Open(driverName, dsn)
		if err != nil {
			return errorMsg{fmt.Errorf("데이터베이스 연결 실패: %v", err)}
		}
		defer conn.Close()

		rows, err := conn.Query(query)
		if err != nil {
			// 실패한 쿼리도 히스토리에 저장
			saveQueryHistory(query, "query", startTime, false, err.Error())
			return errorMsg{fmt.Errorf("쿼리 실행 실패: %v", err)}
		}
		defer rows.Close()

		// 컬럼 정보 가져오기
		columns, err := rows.Columns()
		if err != nil {
			return errorMsg{fmt.Errorf("컬럼 정보 조회 실패: %v", err)}
		}

		// 결과 데이터 읽기
		var results [][]string
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return errorMsg{fmt.Errorf("결과 스캔 실패: %v", err)}
			}

			row := make([]string, len(columns))
			for i, v := range values {
				if v == nil {
					row[i] = "NULL"
				} else {
					valueStr := fmt.Sprintf("%s", v)
					// Try to decrypt if it looks like encrypted data
					if decrypted, err := encrypt.FernetDecrypt(valueStr); err == nil && decrypted != "" {
						row[i] = decrypted
					} else {
						row[i] = valueStr
					}
				}
			}
			results = append(results, row)
		}

		// 성공한 쿼리 히스토리 저장
		saveQueryHistory(query, "query", startTime, true, "")

		return queryResultMsg{
			columns: columns,
			rows:    results,
		}
	}
}

// analyzeQueryCmd executes query analysis
func analyzeQueryCmd(dsn, query string) tea.Cmd {
	return func() tea.Msg {
		startTime := time.Now()

		conn, err := sql.Open("postgres", dsn)
		if err != nil {
			return errorMsg{fmt.Errorf("데이터베이스 연결 실패: %v", err)}
		}
		defer conn.Close()

		analyzer := analyse.NewQueryAnalyzer(conn)
		result, err := analyzer.AnalyzeQuery(query)
		if err != nil {
			// 실패한 분석도 히스토리에 저장
			saveQueryHistory(query, "analyze", startTime, false, err.Error())
			return errorMsg{fmt.Errorf("쿼리 분석 실패: %v", err)}
		}

		// 성공한 분석 히스토리 저장
		saveQueryHistory(query, "analyze", startTime, true, "")

		return analysisResultMsg{result}
	}
}

// saveQueryHistory saves query execution history
func saveQueryHistory(query, queryType string, startTime time.Time, success bool, errorMsg string) {
	historyMgr, err := history.NewHistoryManager()
	if err != nil {
		return // 히스토리 저장 실패는 무시
	}
	defer historyMgr.Close()

	executionTime := time.Since(startTime).Seconds() * 1000 // ms로 변환

	historyRecord := history.QueryHistory{
		Query:         query,
		QueryType:     queryType,
		ExecutedAt:    startTime,
		ExecutionTime: executionTime,
		Success:       success,
		ErrorMsg:      errorMsg,
	}

	historyMgr.SaveQueryHistory(historyRecord)
}
