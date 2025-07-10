// internal/tui/view.go
package tui

import (
	"strings"
)

func (m model) View() string {
	var s strings.Builder

	switch m.state {
	case viewLoading:
		s.WriteString("🔄 처리 중...\n")
		s.WriteString(m.spinner.View())
		s.WriteString("\n잠시만 기다려주세요.")
		return s.String()

	case viewQueryResult:
		s.WriteString("📊 쿼리 실행 결과\n")
		s.WriteString(strings.Repeat("─", 50))
		s.WriteString("\n\n")
		s.WriteString(m.table.View())
		s.WriteString("\n\n")
		s.WriteString("💡 [Tab: 쿼리 입력으로 돌아가기] [Ctrl+C: 종료]")
		return s.String()
	case viewAnalysisResult:
		s.WriteString("🔍 쿼리 분석 결과\n")
		s.WriteString(strings.Repeat("─", 50))
		s.WriteString("\n\n")

		s.WriteString(m.viewport.View())
		s.WriteString("\n\n")
		s.WriteString("💡 [Tab: 쿼리 입력으로 돌아가기] [↑↓: 스크롤] [Ctrl+C: 종료]")
		return s.String()

	case viewError:
		s.WriteString("❌ 오류 발생\n")
		s.WriteString(strings.Repeat("─", 50))
		s.WriteString("\n\n")
		s.WriteString("오류: ")
		s.WriteString(m.err.Error())
		s.WriteString("\n\n")
		s.WriteString("💡 [Tab: 쿼리 입력으로 돌아가기] [Ctrl+C: 종료]")
		return s.String()

	case viewQueryInput:
		fallthrough
	default:
		s.WriteString("🗄️  DiagnoseQuery - SQL 쿼리 분석 도구\n")
		s.WriteString(strings.Repeat("─", 50))
		s.WriteString("\n\n")
		s.WriteString("SQL 쿼리를 입력하세요:\n\n")
		s.WriteString(m.textarea.View())
		s.WriteString("\n\n")
		s.WriteString("💡 사용법:\n")
		s.WriteString("  [Enter: 쿼리 실행] [Ctrl+E: 쿼리 분석] [Ctrl+C: 종료]\n")
		s.WriteString("  [예시: SELECT * FROM users WHERE id = 1]")
		return s.String()
	}
}
