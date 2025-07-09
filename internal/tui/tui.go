// internal/tui/tui.go
package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
)

// NewInitialModel은 TUI의 초기 상태를 설정합니다.
func NewInitialModel(dsn string) model {
	ti := textarea.New()
	ti.Placeholder = "여기에 SQL 쿼리를 입력하세요..."
	ti.Focus()

	return model{
		dsn:      dsn,
		textarea: ti,
		viewport: viewport.New(80, 20),
		state:    viewQueryInput,
		err:      nil,
	}
}
