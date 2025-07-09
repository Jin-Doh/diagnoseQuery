package parameter

import (
	"fmt"
	"os"
	"strings"
)

// ReadEnv reads environment variables from a .env file and returns a map of key-value pairs.
func readEnv(envPath string) (map[string]string, error) {
	if envPath == "" {
		envPath = ".env"
	}

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("환경변수 파일 %s이 존재하지 않습니다", envPath)
	}

	// Read the file content directly
	content, err := os.ReadFile(envPath)
	if err != nil {
		return nil, fmt.Errorf("환경변수 파일 %s 읽기 실패: %v", envPath, err)
	}

	envMap := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // 빈 줄이나 주석은 무시
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("환경변수 형식 오류 (라인 %d): %s", lineNum+1, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		envMap[key] = value
	}

	return envMap, nil
}

// LoadEnvToMemory loads environment variables from the specified path into the InMemoryStore.
func LoadEnvToMemory(envPath string, store *InMemoryStore) error {
	envMap, err := readEnv(envPath)
	if err != nil {
		return err
	}
	store.Load(envMap)
	return nil
}

// LoadEnvToMemoryWithDebug loads environment variables and returns debug info
func LoadEnvToMemoryWithDebug(envPath string, store *InMemoryStore) (map[string]string, error) {
	envMap, err := readEnv(envPath)
	if err != nil {
		return nil, err
	}
	store.Load(envMap)
	return envMap, nil
}

// GetRequiredKeys returns the list of required environment variable keys
func GetRequiredKeys() []string {
	return []string{
		"DB_TYPE",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_SECRET",
		"DB_DB",
		"FERNET_KEY", // Optional
	}
}

// ValidateRequiredKeys checks if all required keys are present in the environment
func ValidateRequiredKeys(envMap map[string]string) []string {
	required := []string{"DB_TYPE", "DB_HOST", "DB_PORT", "DB_USER", "DB_SECRET", "DB_DB"}
	var missing []string

	for _, key := range required {
		if value, exists := envMap[key]; !exists || strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}

	return missing
}
