// internal/tui/tui.go
package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// NewInitialModel은 TUI의 초기 상태를 설정합니다.
func NewInitialModel(dsn string) model {
	ti := textarea.New()
	ti.Placeholder = "여기에 SQL 쿼리를 입력하세요..."
	ti.Focus()
	ti.SetWidth(80)
	ti.SetHeight(5)

	s := spinner.New()
	s.Spinner = spinner.Dot

	t := table.New()

	vp := viewport.New(80, 20)

	return model{
		dsn:      dsn,
		textarea: ti,
		table:    t,
		viewport: vp,
		spinner:  s,
		state:    viewQueryInput,
		err:      nil,
	}
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
	)
}
