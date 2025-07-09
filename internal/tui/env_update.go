package tui

import (
	"diagnoseQuery/internal/database"
	"diagnoseQuery/internal/encrypt"

	tea "github.com/charmbracelet/bubbletea"
)

// updateEnvSetup handles environment setup updates
func (m *MainModel) updateEnvSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Handle control keys first
		switch keyMsg.String() {
		case "enter", "tab", "down", "shift+tab", "up", "esc":
			return m.handleEnvSetupKeys(keyMsg)
		}

		// If we're in stage 0 (selection), don't pass to input fields
		if m.envSetup.stage == 0 {
			return m, nil
		}
	}

	// Only update inputs if we're in stage 1 (input stage)
	if m.envSetup.stage == 1 {
		return m.updateEnvSetupInputs(msg)
	}

	return m, nil
}

// handleEnvSetupKeys handles keyboard input for environment setup
func (m *MainModel) handleEnvSetupKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return m.handleEnvSetupEnter()
	case "tab", "down":
		return m.handleEnvSetupNavigation(1)
	case "shift+tab", "up":
		return m.handleEnvSetupNavigation(-1)
	case "esc":
		return m.handleEnvSetupEscape()
	}
	return m, nil
}

// handleEnvSetupEnter handles Enter key press in environment setup
func (m *MainModel) handleEnvSetupEnter() (tea.Model, tea.Cmd) {
	if m.envSetup.stage == 0 {
		// Move to input stage
		m.envSetup.stage = 1
		switch m.envSetup.setupType {
		case EnvSetupEnvPath, EnvSetupEnvs:
			m.envSetup.textInput.Focus()
		case EnvSetupManual:
			if len(m.envSetup.manualFields) > 0 {
				m.envSetup.manualFields[0].Focus()
			}
		case EnvSetupRecent:
			// No additional setup needed for recent selection
		}
	} else if m.envSetup.stage == 1 {
		// Process the input and complete setup
		if m.envSetup.setupType == EnvSetupRecent {
			return m.completeRecentSetup()
		}
		return m.completeEnvSetup()
	}
	return m, nil
}

// handleEnvSetupNavigation handles navigation in environment setup
func (m *MainModel) handleEnvSetupNavigation(direction int) (tea.Model, tea.Cmd) {
	if m.envSetup.stage == 0 {
		// Navigate setup type selection
		if direction > 0 {
			m.envSetup.cursor = (m.envSetup.cursor + 1) % 4 // Now we have 4 options
		} else {
			m.envSetup.cursor = (m.envSetup.cursor - 1 + 4) % 4
		}
		m.envSetup.setupType = EnvSetupType(m.envSetup.cursor)
	} else if m.envSetup.stage == 1 {
		switch m.envSetup.setupType {
		case EnvSetupManual:
			// Navigate manual input fields
			m.envSetup.manualFields[m.envSetup.fieldCursor].Blur()
			if direction > 0 {
				m.envSetup.fieldCursor = (m.envSetup.fieldCursor + 1) % len(m.envSetup.manualFields)
			} else {
				m.envSetup.fieldCursor = (m.envSetup.fieldCursor - 1 + len(m.envSetup.manualFields)) % len(m.envSetup.manualFields)
			}
			m.envSetup.manualFields[m.envSetup.fieldCursor].Focus()
		case EnvSetupRecent:
			// Navigate recent configurations
			if len(m.envSetup.recentList.Items()) > 0 {
				if direction > 0 {
					m.envSetup.recentCursor = (m.envSetup.recentCursor + 1) % len(m.envSetup.recentList.Items())
				} else {
					m.envSetup.recentCursor = (m.envSetup.recentCursor - 1 + len(m.envSetup.recentList.Items())) % len(m.envSetup.recentList.Items())
				}
			}
		}
	}
	return m, nil
}

// handleEnvSetupEscape handles Escape key press in environment setup
func (m *MainModel) handleEnvSetupEscape() (tea.Model, tea.Cmd) {
	if m.envSetup.stage == 1 {
		// Go back to setup type selection
		m.envSetup.stage = 0
		m.envSetup.textInput.Blur()
		for i := range m.envSetup.manualFields {
			m.envSetup.manualFields[i].Blur()
		}
	}
	return m, nil
}

// updateEnvSetupInputs updates input components in environment setup
func (m *MainModel) updateEnvSetupInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m.envSetup.setupType {
	case EnvSetupEnvPath, EnvSetupEnvs:
		m.envSetup.textInput, cmd = m.envSetup.textInput.Update(msg)
		cmds = append(cmds, cmd)
	case EnvSetupManual:
		if m.envSetup.fieldCursor < len(m.envSetup.manualFields) {
			m.envSetup.manualFields[m.envSetup.fieldCursor], cmd = m.envSetup.manualFields[m.envSetup.fieldCursor].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// updateEnvValidation handles environment validation updates
func (m *MainModel) updateEnvValidation(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			// 검증 통과 후 연결 진행
			m.state = StateVerifyingConnection

			// Get connection parameters
			dbTypeStr := m.store.GetOrDefault("DB_TYPE", "postgres")
			host := m.store.GetOrDefault("DB_HOST", "localhost")
			port := m.store.GetOrDefault("DB_PORT", "5432")
			user := m.store.GetOrDefault("DB_USER", "user")
			password := m.store.GetOrDefault("DB_SECRET", "password")
			dbname := m.store.GetOrDefault("DB_DB", "database")

			// Initialize Fernet key if provided
			if fernetKey, exists := m.store.Get("FERNET_KEY"); exists {
				encrypt.InjectFernetKey(fernetKey)
			}

			var configType string
			envConfig := make(map[string]string)

			switch m.envSetup.setupType {
			case EnvSetupEnvPath:
				configType = "env-path"
				envConfig["path"] = m.envSetup.textInput.Value()
			case EnvSetupEnvs:
				configType = "envs"
				envConfig["key"] = m.envSetup.textInput.Value()
			case EnvSetupManual:
				configType = "manual"
				fieldNames := []string{"DB_TYPE", "DB_HOST", "DB_PORT", "DB_USER", "DB_SECRET", "DB_DB"}
				for i, field := range m.envSetup.manualFields {
					value := field.Value()
					if value == "" {
						value = field.Placeholder
					}
					envConfig[fieldNames[i]] = value
				}
			case EnvSetupRecent:
				if len(m.envSetup.recentList.Items()) > 0 && m.envSetup.recentCursor < len(m.envSetup.recentList.Items()) {
					item := m.envSetup.recentList.Items()[m.envSetup.recentCursor]
					if envItem, ok := item.(envConfigItem); ok {
						configType = envItem.config.ConfigType
						envConfig["key"] = envItem.config.ConfigKey
					}
				}
			}

			return m, func() tea.Msg {
				dsn, err := database.ConnectWithString(dbTypeStr, host, port, user, password, dbname)
				return connectionVerifiedMsg{
					dsn:        dsn,
					configType: configType,
					envConfig:  envConfig,
					err:        err,
				}
			}

		case "esc", "backspace":
			// 환경 설정으로 돌아가기
			m.state = StateEnvSetup
			return m, nil

		case "d", "D":
			// 세부 정보 토글
			if m.envValidation != nil {
				m.envValidation.showDetails = !m.envValidation.showDetails
			}
			return m, nil
		}
	}

	return m, nil
}
