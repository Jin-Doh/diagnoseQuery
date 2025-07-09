// internal/tui/view.go
package tui

import "diagnoseQuery/internal/analyse"

func (m model) View() string {
	switch m.state {
	case viewLoading:
		return "쿼리 실행 중..." // 스피너 등 활용

	case viewQueryResult:
		// bubbletea/table을 사용하여 결과 렌더링
		return m.table.View()

	case viewAnalysisResult:
		// formatter를 사용해 결과를 문자열로 만들고 viewport에 렌더링
		formatter := analyse.NewResultFormatter()
		output := formatter.FormatAnalysisResult(&m.analysisResult)
		m.viewport.SetContent(output)
		return m.viewport.View()

	case viewError:
		return "에러: " + m.err.Error()

	case viewQueryInput:
		fallthrough
	default:
		// 쿼리 입력창과 하단 도움말 렌더링
		return m.textarea.View() + "\n\n[Enter: 쿼리 실행] [Ctrl+E: 쿼리 분석] [Ctrl+C: 종료]"
	}
}
