package tui

import (
	"database/sql"
	"diagnoseQuery/internal/analyse"

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
				m.state = viewLoading // 로딩 상태로 변경
				query := m.textarea.Value()
				// Cmd를 통해 비동기로 쿼리 실행
				return m, executeQueryCmd(m.dsn, query)
			}
		case tea.KeyCtrlE: // Ctrl+E로 분석 실행
			if m.state == viewQueryInput {
				m.state = viewLoading
				query := m.textarea.Value()
				// Cmd를 통해 비동기로 쿼리 분석
				return m, analyzeQueryCmd(m.dsn, query)
			}
		}

	// 비동기 작업 결과 처리
	case queryResultMsg:
		m.state = viewQueryResult
		// ... table 모델에 결과 데이터 설정 ...
		return m, nil
	case analysisResultMsg:
		m.state = viewAnalysisResult
		m.analysisResult = msg.result
		// ... viewport에 포맷팅된 결과 설정 ...
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
		// ... 다른 상태에 대한 처리 ...
	}
	return m, cmd
}

// --- Commands ---

func analyzeQueryCmd(dsn, query string) tea.Cmd {
	return func() tea.Msg {
		conn, _ := sql.Open("postgres", dsn) // 실제로는 에러 처리 필요
		defer conn.Close()
		analyzer := analyse.NewQueryAnalyzer(conn)
		res, err := analyzer.AnalyzeQuery(query)
		if err != nil {
			return errorMsg{err}
		}
		return analysisResultMsg{res}
	}
}

// TODO: executeQueryCmd 구현
