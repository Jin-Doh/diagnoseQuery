package history

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbFileName = "diagnose_query_history.db"
)

type HistoryManager struct {
	db *sql.DB
}

// NewHistoryManager creates a new history manager with SQLite database
func NewHistoryManager() (*HistoryManager, error) {
	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("사용자 홈 디렉토리를 찾을 수 없습니다: %v", err)
	}

	// Create .diagnosequery directory if it doesn't exist
	configDir := filepath.Join(homeDir, ".diagnosequery")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("설정 디렉토리 생성 실패: %v", err)
	}

	// Open SQLite database
	dbPath := filepath.Join(configDir, dbFileName)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("데이터베이스 연결 실패: %v", err)
	}

	hm := &HistoryManager{db: db}

	// Initialize tables
	if err := hm.initTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("테이블 초기화 실패: %v", err)
	}

	return hm, nil
}

// Close closes the database connection
func (hm *HistoryManager) Close() error {
	return hm.db.Close()
}

// initTables creates necessary tables if they don't exist
func (hm *HistoryManager) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS query_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			query TEXT NOT NULL,
			query_type TEXT NOT NULL,
			executed_at DATETIME NOT NULL,
			execution_time REAL DEFAULT 0,
			success BOOLEAN NOT NULL,
			error_msg TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS env_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			config_type TEXT NOT NULL,
			config_key TEXT NOT NULL,
			config_value TEXT,
			used_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_query_history_executed_at ON query_history(executed_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_env_config_used_at ON env_config(used_at DESC)`,
	}

	for _, query := range queries {
		if _, err := hm.db.Exec(query); err != nil {
			return fmt.Errorf("테이블 생성 실패: %v", err)
		}
	}

	return nil
}

// SaveQueryHistory saves a query execution history
func (hm *HistoryManager) SaveQueryHistory(history QueryHistory) error {
	query := `INSERT INTO query_history (query, query_type, executed_at, execution_time, success, error_msg)
			  VALUES (?, ?, ?, ?, ?, ?)`

	_, err := hm.db.Exec(query, history.Query, history.QueryType, history.ExecutedAt,
		history.ExecutionTime, history.Success, history.ErrorMsg)
	if err != nil {
		return fmt.Errorf("쿼리 히스토리 저장 실패: %v", err)
	}

	return nil
}

// SaveEnvConfig saves environment configuration history
func (hm *HistoryManager) SaveEnvConfig(config EnvConfig) error {
	query := `INSERT INTO env_config (config_type, config_key, config_value, used_at)
			  VALUES (?, ?, ?, ?)`

	_, err := hm.db.Exec(query, config.ConfigType, config.ConfigKey, config.ConfigValue, config.UsedAt)
	if err != nil {
		return fmt.Errorf("환경설정 히스토리 저장 실패: %v", err)
	}

	return nil
}

// GetRecentQueries returns recent query history
func (hm *HistoryManager) GetRecentQueries(limit int) ([]QueryHistory, error) {
	query := `SELECT id, query, query_type, executed_at, execution_time, success, error_msg
			  FROM query_history
			  ORDER BY executed_at DESC
			  LIMIT ?`

	rows, err := hm.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("최근 쿼리 조회 실패: %v", err)
	}
	defer rows.Close()

	var histories []QueryHistory
	for rows.Next() {
		var h QueryHistory
		var errorMsg sql.NullString

		err := rows.Scan(&h.ID, &h.Query, &h.QueryType, &h.ExecutedAt,
			&h.ExecutionTime, &h.Success, &errorMsg)
		if err != nil {
			return nil, fmt.Errorf("쿼리 히스토리 스캔 실패: %v", err)
		}

		if errorMsg.Valid {
			h.ErrorMsg = errorMsg.String
		}

		histories = append(histories, h)
	}

	return histories, nil
}

// GetRecentEnvConfigs returns recent environment configurations
func (hm *HistoryManager) GetRecentEnvConfigs(limit int) ([]EnvConfig, error) {
	query := `SELECT id, config_type, config_key, config_value, used_at
			  FROM env_config
			  ORDER BY used_at DESC
			  LIMIT ?`

	rows, err := hm.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("최근 환경설정 조회 실패: %v", err)
	}
	defer rows.Close()

	var configs []EnvConfig
	for rows.Next() {
		var c EnvConfig
		var configValue sql.NullString

		err := rows.Scan(&c.ID, &c.ConfigType, &c.ConfigKey, &configValue, &c.UsedAt)
		if err != nil {
			return nil, fmt.Errorf("환경설정 히스토리 스캔 실패: %v", err)
		}

		if configValue.Valid {
			c.ConfigValue = configValue.String
		}

		configs = append(configs, c)
	}

	return configs, nil
}

// ClearOldHistory removes old history records
func (hm *HistoryManager) ClearOldHistory(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	queries := []string{
		`DELETE FROM query_history WHERE executed_at < ?`,
		`DELETE FROM env_config WHERE used_at < ?`,
	}

	for _, query := range queries {
		if _, err := hm.db.Exec(query, cutoffDate); err != nil {
			return fmt.Errorf("오래된 히스토리 삭제 실패: %v", err)
		}
	}

	return nil
}
