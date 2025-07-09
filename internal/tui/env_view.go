package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Define styles locally to avoid conflicts
var (
	envPrimaryColor = "#7C3AED"
	envGrayColor    = "#666666"
	envWhiteColor   = "#FFFFFF"
	envBlackColor   = "#000000"
	envErrorColor   = "#EF4444"
)

var (
	envTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(envPrimaryColor)).
			MarginBottom(1)

	envOptionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(envWhiteColor))

	envSelectedOptionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(envBlackColor)).
				Background(lipgloss.Color(envPrimaryColor)).
				Padding(0, 1)

	envInputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(envPrimaryColor)).
			Padding(1).
			Width(60)

	envFieldLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(envPrimaryColor)).
				Width(12)

	envHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(envGrayColor)).
			MarginTop(1)

	envErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(envErrorColor)).
			Bold(true)
)

// View renders the environment setup interface
func (m *EnvSetupModel) View() string {
	var s strings.Builder

	s.WriteString(envTitleStyle.Render("🔧 DiagnoseQuery 환경 설정"))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(envErrorStyle.Render(fmt.Sprintf("오류: %s", m.err)))
		s.WriteString("\n\n")
	}

	switch m.stage {
	case 0:
		// Stage 0: Choose setup type
		s.WriteString("환경변수 설정 방법을 선택하세요:\n\n")

		options := []string{
			"📁 .env 파일에서 읽기",
			"🔑 시스템 환경변수에서 읽기",
			"✏️  직접 입력",
			"🕰️  최근 사용한 설정",
		}

		for i, option := range options {
			if i == m.cursor {
				s.WriteString(envSelectedOptionStyle.Render("> " + option))
			} else {
				s.WriteString(envOptionStyle.Render("  " + option))
			}
			s.WriteString("\n")
		}

		s.WriteString("\n")

		// Show recent configurations if available
		if len(m.recentList.Items()) > 0 {
			s.WriteString(envHelpStyle.Render("최근 사용한 설정:"))
			s.WriteString("\n")
			for i, item := range m.recentList.Items() {
				if envItem, ok := item.(envConfigItem); ok {
					if i < 3 { // Show only first 3 items
						s.WriteString(envHelpStyle.Render(fmt.Sprintf("  • %s (%s)", envItem.config.ConfigKey, envItem.config.ConfigType)))
						s.WriteString("\n")
					}
				}
			}
			s.WriteString("\n")
		}

		s.WriteString(envHelpStyle.Render("↑/↓: 선택, Enter: 확인, Ctrl+C: 종료"))

	case 1:
		// Stage 1: Input based on selected setup type
		switch m.setupType {
		case EnvSetupEnvPath:
			s.WriteString("📁 .env 파일 경로를 입력하세요:\n\n")
			s.WriteString(envInputStyle.Render(m.textInput.View()))
			s.WriteString("\n\n")
			s.WriteString(envHelpStyle.Render("Enter: 확인, Esc: 뒤로가기"))

		case EnvSetupEnvs:
			s.WriteString("🔑 환경변수 키를 입력하세요 (콤마로 구분):\n\n")
			s.WriteString(envInputStyle.Render(m.textInput.View()))
			s.WriteString("\n\n")
			s.WriteString(envHelpStyle.Render("예시: DB_URL,DB_SECRET"))
			s.WriteString("\n")
			s.WriteString(envHelpStyle.Render("Enter: 확인, Esc: 뒤로가기"))

		case EnvSetupManual:
			s.WriteString("✏️  데이터베이스 연결 정보를 입력하세요:\n\n")

			fieldNames := []string{"DB 타입", "호스트", "포트", "사용자명", "비밀번호", "데이터베이스"}

			for i, field := range m.manualFields {
				label := envFieldLabelStyle.Render(fieldNames[i] + ":")

				if i == m.fieldCursor {
					s.WriteString("> ")
				} else {
					s.WriteString("  ")
				}

				s.WriteString(label)
				s.WriteString(" ")
				s.WriteString(field.View())
				s.WriteString("\n")
			}

			s.WriteString("\n")
			s.WriteString(envHelpStyle.Render("Tab/↑↓: 필드 이동, Enter: 확인, Esc: 뒤로가기"))

		case EnvSetupRecent:
			s.WriteString("🕰️  최근 사용한 설정을 선택하세요:\n\n")

			if len(m.recentList.Items()) == 0 {
				s.WriteString(envHelpStyle.Render("사용 가능한 최근 설정이 없습니다."))
				s.WriteString("\n\n")
				s.WriteString(envHelpStyle.Render("Esc: 뒤로가기"))
			} else {
				for i, item := range m.recentList.Items() {
					if envItem, ok := item.(envConfigItem); ok {
						configInfo := fmt.Sprintf("%s (%s) - %s",
							envItem.config.ConfigKey,
							envItem.config.ConfigType,
							envItem.config.UsedAt.Format("2006-01-02 15:04"))

						if i == m.recentCursor {
							s.WriteString(envSelectedOptionStyle.Render("> " + configInfo))
						} else {
							s.WriteString(envOptionStyle.Render("  " + configInfo))
						}
						s.WriteString("\n")
					}
				}
				s.WriteString("\n")
				s.WriteString(envHelpStyle.Render("↑/↓: 선택, Enter: 사용, Esc: 뒤로가기"))
			}
		}
	}

	return s.String()
}

// View renders the environment validation interface
func (m *EnvValidationModel) View() string {
	var s strings.Builder

	s.WriteString(envTitleStyle.Render("🔍 환경변수 검증"))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(envErrorStyle.Render(fmt.Sprintf("오류: %s", m.err)))
		s.WriteString("\n\n")
	}

	// 로드된 환경변수 개수 표시
	loadedCount := len(m.loadedVars)
	missingCount := len(m.missingKeys)

	if loadedCount > 0 {
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true).
			Render(fmt.Sprintf("✅ %d개의 환경변수가 로드되었습니다.", loadedCount)))
		s.WriteString("\n")
	}

	if missingCount > 0 {
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true).
			Render(fmt.Sprintf("⚠️ %d개의 필수 환경변수가 누락되었습니다.", missingCount)))
		s.WriteString("\n")
	} else {
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true).
			Render("✅ 모든 필수 환경변수가 설정되었습니다."))
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// 세부 정보 표시
	if m.showDetails {
		s.WriteString(envTitleStyle.Render("📋 로드된 환경변수:"))
		s.WriteString("\n\n")

		if loadedCount == 0 {
			s.WriteString(envErrorStyle.Render("로드된 환경변수가 없습니다."))
		} else {
			// 환경변수 목록 정렬
			var keys []string
			for k := range m.loadedVars {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				v := m.loadedVars[k]
				// 암호 등 민감한 정보는 마스킹
				if strings.Contains(strings.ToLower(k), "secret") ||
					strings.Contains(strings.ToLower(k), "password") ||
					strings.Contains(strings.ToLower(k), "key") {
					// 앞에 두 자리, 뒤에 두 자리 표시
					if len(v) > 6 {
						v = v[:2] + strings.Repeat("*", len(v)-4) + v[len(v)-2:]
					} else if len(v) > 2 {
						// 앞에 두 자리 표시
						v = v[:2] + strings.Repeat("*", len(v)-2)
					} else {
						// 2자리 이하인 경우 그대로 표시
						v = strings.Repeat("*", len(v))
					}
				}

				s.WriteString(lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color(envPrimaryColor)).
					Render(k))
				s.WriteString(": ")
				s.WriteString(v)
				s.WriteString("\n")
			}
		}

		if missingCount > 0 {
			s.WriteString("\n")
			s.WriteString(envErrorStyle.Render("⚠️ 누락된 필수 환경변수:"))
			s.WriteString("\n\n")

			for _, k := range m.missingKeys {
				s.WriteString("• ")
				s.WriteString(envErrorStyle.Render(k))
				s.WriteString("\n")
			}
		}
	}

	s.WriteString("\n")

	if missingCount > 0 {
		s.WriteString(envHelpStyle.Render("💡 연결은 가능하지만 일부 기능이 제한될 수 있습니다."))
		s.WriteString("\n\n")
	}

	s.WriteString(envHelpStyle.Render("D: 세부정보 토글, Enter: 연결 시도, Esc: 뒤로가기"))

	return s.String()
}
