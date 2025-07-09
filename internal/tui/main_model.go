package tui

import (
	"diagnoseQuery/internal/database"
	"diagnoseQuery/internal/encrypt"
	"diagnoseQuery/internal/history"
	"diagnoseQuery/internal/parameter"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))
)

// AppState represents the current state of the TUI application
type AppState int

const (
	StateEnvSetup      AppState = iota
	StateEnvValidation          // New: show loaded env vars and validation
	StateVerifyingConnection
	StateMainMenu
	StateQueryInput
	StateQueryResult
	StateAnalysisResult
	StateHistory
	StateError
)

// EnvSetupType represents different ways to setup environment
type EnvSetupType int

const (
	EnvSetupEnvPath EnvSetupType = iota
	EnvSetupEnvs
	EnvSetupManual
	EnvSetupRecent // New: select from recent configurations
)

// MainModel is the root model for the TUI application
type MainModel struct {
	state         AppState
	envSetup      *EnvSetupModel
	envValidation *EnvValidationModel
	queryApp      *QueryAppModel
	historyMgr    *history.HistoryManager
	store         *parameter.InMemoryStore
	spinner       spinner.Model
	dsn           string
	err           error
}

// EnvSetupModel handles environment variable setup
type EnvSetupModel struct {
	setupType    EnvSetupType
	cursor       int
	textInput    textinput.Model
	recentList   list.Model
	manualFields []textinput.Model
	fieldCursor  int
	stage        int // 0: choose type, 1: input values, 2: confirm
	recentCursor int // cursor for recent configurations
	err          error
}

// EnvValidationModel handles environment variable validation display
type EnvValidationModel struct {
	loadedVars  map[string]string
	missingKeys []string
	cursor      int
	showDetails bool
	err         error
}

// QueryAppModel handles the main query interface
type QueryAppModel struct {
	// Embed the existing model from the tui package
	model
}

// Init initializes the query app model
func (m *QueryAppModel) Init() tea.Cmd {
	return textinput.Blink
}

// NewMainModel creates a new main TUI model
func NewMainModel() (*MainModel, error) {
	historyMgr, err := history.NewHistoryManager()
	if err != nil {
		return nil, err
	}

	store := parameter.NewInMemoryStore()

	envSetup := NewEnvSetupModel(historyMgr)
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return &MainModel{
		state:      StateEnvSetup,
		envSetup:   envSetup,
		historyMgr: historyMgr,
		store:      store,
		spinner:    s,
	}, nil
}

// NewMainModelWithOptions creates a new main TUI model with command line options
func NewMainModelWithOptions(envPath, envKeys string) (*MainModel, error) {
	historyMgr, err := history.NewHistoryManager()
	if err != nil {
		return nil, err
	}

	store := parameter.NewInMemoryStore()
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	// If command line options are provided, pre-populate them
	if envPath != "" || envKeys != "" {
		envSetup := NewEnvSetupModelWithOptions(historyMgr, envPath, envKeys)

		return &MainModel{
			state:      StateEnvSetup,
			envSetup:   envSetup,
			historyMgr: historyMgr,
			store:      store,
			spinner:    s,
		}, nil
	}

	// Otherwise, start with normal environment setup
	envSetup := NewEnvSetupModel(historyMgr)

	return &MainModel{
		state:      StateEnvSetup,
		envSetup:   envSetup,
		historyMgr: historyMgr,
		store:      store,
		spinner:    s,
	}, nil
}

// NewEnvSetupModel creates a new environment setup model
func NewEnvSetupModel(historyMgr *history.HistoryManager) *EnvSetupModel {
	ti := textinput.New()
	ti.Placeholder = "파일 경로 또는 환경변수 키 입력..."
	ti.CharLimit = 256
	ti.Width = 50

	// Create manual input fields
	fields := make([]textinput.Model, 6)
	placeholders := []string{"postgres", "localhost", "5432", "user", "password", "database"}

	for i := range fields {
		fields[i] = textinput.New()
		fields[i].Placeholder = placeholders[i]
		fields[i].CharLimit = 256
		fields[i].Width = 30
	}

	if len(fields) > 0 {
		fields[0].Focus()
	}

	// Create recent configs list
	items := []list.Item{}
	recentConfigs, _ := historyMgr.GetRecentEnvConfigs(10)
	for _, config := range recentConfigs {
		items = append(items, envConfigItem{config})
	}

	recentList := list.New(items, list.NewDefaultDelegate(), 50, 15)
	recentList.Title = "최근 사용한 설정"

	return &EnvSetupModel{
		setupType:    EnvSetupEnvPath,
		textInput:    ti,
		recentList:   recentList,
		manualFields: fields,
		stage:        0,
	}
}

// NewEnvSetupModelWithOptions creates environment setup model with pre-filled options
func NewEnvSetupModelWithOptions(historyMgr *history.HistoryManager, envPath, envKeys string) *EnvSetupModel {
	envSetup := NewEnvSetupModel(historyMgr)

	// Pre-fill based on command line options
	if envPath != "" {
		envSetup.setupType = EnvSetupEnvPath
		envSetup.textInput.SetValue(envPath)
		envSetup.stage = 1 // Go directly to input stage
		envSetup.textInput.Focus()
	} else if envKeys != "" {
		envSetup.setupType = EnvSetupEnvs
		envSetup.textInput.SetValue(envKeys)
		envSetup.stage = 1 // Go directly to input stage
		envSetup.textInput.Focus()
	}

	return envSetup
}

// NewEnvValidationModel creates a new environment validation model
func NewEnvValidationModel(loadedVars map[string]string, missingKeys []string) *EnvValidationModel {
	return &EnvValidationModel{
		loadedVars:  loadedVars,
		missingKeys: missingKeys,
		cursor:      0,
		showDetails: false,
	}
}

// envConfigItem implements list.Item for recent configurations
type envConfigItem struct {
	config history.EnvConfig
}

func (i envConfigItem) FilterValue() string {
	return i.config.ConfigKey
}

func (i envConfigItem) Title() string {
	return i.config.ConfigKey
}

func (i envConfigItem) Description() string {
	return i.config.ConfigType + " - " + i.config.UsedAt.Format("2006-01-02 15:04")
}

// Init initializes the main model
func (m *MainModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		tea.EnterAltScreen,
		m.spinner.Tick,
	)
}

// Update handles messages for the main model
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.historyMgr != nil {
				m.historyMgr.Close()
			}
			return m, tea.Quit
		}
	case connectionVerifiedMsg:
		if msg.err != nil {
			m.state = StateEnvSetup
			m.envSetup.err = msg.err
			return m, nil
		}

		// Connection successful, transition to main query interface
		m.state = StateQueryInput
		m.dsn = msg.dsn
		m.queryApp = &QueryAppModel{
			model: NewInitialModel(msg.dsn),
		}

		// Save environment configuration to history
		for key, value := range msg.envConfig {
			m.historyMgr.SaveEnvConfig(history.EnvConfig{
				ConfigType:  msg.configType,
				ConfigKey:   key,
				ConfigValue: value,
				UsedAt:      time.Now(),
			})
		}

		return m, nil
	}

	// Delegate to appropriate sub-model based on state
	switch m.state {
	case StateEnvSetup:
		return m.updateEnvSetup(msg)
	case StateEnvValidation:
		return m.updateEnvValidation(msg)
	case StateVerifyingConnection:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case StateQueryInput, StateQueryResult, StateAnalysisResult:
		return m.updateQueryApp(msg)
	}

	return m, nil
}

// View renders the main model
func (m *MainModel) View() string {
	switch m.state {
	case StateEnvSetup:
		return m.envSetup.View()
	case StateEnvValidation:
		if m.envValidation != nil {
			return m.envValidation.View()
		}
		return "환경변수 검증 중..."
	case StateVerifyingConnection:
		return fmt.Sprintf("%s 데이터베이스 연결 확인 중...", m.spinner.View())
	case StateQueryInput, StateQueryResult, StateAnalysisResult:
		if m.queryApp != nil {
			return m.queryApp.View()
		}
		return "쿼리 인터페이스 로딩 중..."
	case StateError:
		return "오류: " + m.err.Error() + "\n\nPress 'q' to quit"
	}
	return ""
}

// Messages
type connectionVerifiedMsg struct {
	dsn        string
	configType string
	envConfig  map[string]string
	err        error
}

type envValidationMsg struct {
	loadedVars  map[string]string
	missingKeys []string
	err         error
}

// updateEnvSetup is now defined in env_update.go to reduce complexity

// updateQueryApp handles query application updates
func (m *MainModel) updateQueryApp(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.queryApp == nil {
		return m, nil
	}

	updatedModel, cmd := m.queryApp.model.Update(msg)
	if updatedModel != nil {
		if model, ok := updatedModel.(model); ok {
			m.queryApp.model = model
		}
	}
	return m, cmd
}

// completeEnvSetup processes the environment setup and transitions to the main app
func (m *MainModel) completeEnvSetup() (tea.Model, tea.Cmd) {
	var err error
	var _ string
	envConfig := make(map[string]string)
	m.envSetup.err = nil // Clear previous error
	var loadedVars map[string]string
	var missingKeys []string

	switch m.envSetup.setupType {
	case EnvSetupEnvPath:
		_ = "env-path"
		envPath := m.envSetup.textInput.Value()
		if envPath == "" {
			envPath = ".env"
		}

		// Load with debug info
		loadedVars, err = parameter.LoadEnvToMemoryWithDebug(envPath, m.store)
		if err != nil {
			m.state = StateError
			m.err = err
			return m, nil
		}
		envConfig["path"] = envPath
		// 필수 키 검증
		missingKeys = parameter.ValidateRequiredKeys(loadedVars)

	case EnvSetupEnvs:
		_ = "envs"
		key := m.envSetup.textInput.Value()
		if key == "" {
			m.state = StateError
			m.err = fmt.Errorf("환경변수 키가 필요합니다")
			return m, nil
		}
		err = parameter.LoadKeyToMemory(key, m.store)
		if err != nil {
			m.state = StateError
			m.err = err
			return m, nil
		}
		envConfig["key"] = key

		// 시스템 환경변수에서 로드한 값들 수집
		loadedVars = make(map[string]string)
		keys := strings.Split(key, ",")
		for _, k := range keys {
			k = strings.TrimSpace(k)
			if v, exists := m.store.Get(k); exists {
				loadedVars[k] = v
			}
		}

		// 필수 키 검증
		missingKeys = parameter.ValidateRequiredKeys(loadedVars)

	case EnvSetupManual:
		_ = "manual"
		fieldNames := []string{"DB_TYPE", "DB_HOST", "DB_PORT", "DB_USER", "DB_SECRET", "DB_DB"}
		loadedVars = make(map[string]string)

		for i, field := range m.envSetup.manualFields {
			value := field.Value()
			if value == "" {
				value = field.Placeholder
			}
			envConfig[fieldNames[i]] = value
			parameter.InputEnvToMemory(fieldNames[i], value, m.store)
			loadedVars[fieldNames[i]] = value
		}

		// 필수 키 검증
		missingKeys = parameter.ValidateRequiredKeys(loadedVars)
	}

	// 환경변수 검증 상태로 전환
	m.envValidation = NewEnvValidationModel(loadedVars, missingKeys)
	m.state = StateEnvValidation

	return m, nil
}

// completeRecentSetup processes recent configuration selection and transitions to the main app
func (m *MainModel) completeRecentSetup() (tea.Model, tea.Cmd) {
	if len(m.envSetup.recentList.Items()) == 0 || m.envSetup.recentCursor >= len(m.envSetup.recentList.Items()) {
		m.state = StateError
		m.err = fmt.Errorf("선택된 최근 설정이 유효하지 않습니다")
		return m, nil
	}

	item := m.envSetup.recentList.Items()[m.envSetup.recentCursor]
	envItem, ok := item.(envConfigItem)
	if !ok {
		m.state = StateError
		m.err = fmt.Errorf("설정 정보를 읽을 수 없습니다")
		return m, nil
	}

	config := envItem.config
	configType := config.ConfigType
	envConfig := map[string]string{
		"key": config.ConfigKey,
	}

	var err error

	// Load configuration based on type
	switch config.ConfigType {
	case "env-path":
		err = parameter.LoadEnvToMemory(config.ConfigKey, m.store)
		if err != nil {
			m.state = StateError
			m.err = fmt.Errorf("환경 파일 로드 실패: %v", err)
			return m, nil
		}
	case "envs":
		err = parameter.LoadKeyToMemory(config.ConfigKey, m.store)
		if err != nil {
			m.state = StateError
			m.err = fmt.Errorf("환경변수 로드 실패: %v", err)
			return m, nil
		}
	case "manual":
		// For manual configurations, we might need to recreate them
		// This is a limitation of storing only key in the history
		m.state = StateError
		m.err = fmt.Errorf("수동 설정은 최근 설정에서 복원할 수 없습니다. 다시 입력해주세요.")
		return m, nil
	}

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

	m.state = StateVerifyingConnection

	return m, func() tea.Msg {
		dsn, err := database.ConnectWithString(dbTypeStr, host, port, user, password, dbname)
		return connectionVerifiedMsg{
			dsn:        dsn,
			configType: configType,
			envConfig:  envConfig,
			err:        err,
		}
	}
}
