package parameter

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// ReadEnv reads environment variables from a .env file and returns a map of key-value pairs.
func readEnv(envPath string) (map[string]string, error) {
	if envPath == "" {
		envPath = ".env"
	}

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("환경변수 파일 %s이 존재하지 않습니다", envPath)
	}
	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("환경변수 파일 %s 읽기 실패, err: %v", envPath, err)
	}
	envMap := make(map[string]string)
	for _, line := range strings.Split(os.Getenv("ENV_CONTENT"), "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue // 빈 줄이나 주석은 무시
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("환경변수 형식 오류: %s", line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
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
