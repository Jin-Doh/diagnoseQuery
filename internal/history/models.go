package history

import "time"

// QueryHistory represents a query history record
type QueryHistory struct {
	ID            int       `json:"id"`
	Query         string    `json:"query"`
	QueryType     string    `json:"query_type"` // "query" or "explain"
	ExecutedAt    time.Time `json:"executed_at"`
	ExecutionTime float64   `json:"execution_time"` // in milliseconds
	Success       bool      `json:"success"`
	ErrorMsg      string    `json:"error_msg,omitempty"`
}

// EnvConfig represents environment configuration history
type EnvConfig struct {
	ID          int       `json:"id"`
	ConfigType  string    `json:"config_type"` // "env-path", "envs", "manual"
	ConfigKey   string    `json:"config_key"`
	ConfigValue string    `json:"config_value,omitempty"`
	UsedAt      time.Time `json:"used_at"`
}
