package parameter

import (
	"fmt"
	"os"
)

func findEnv(key string) (string, bool) {
	value, exists := os.LookupEnv(key)
	if !exists {
		return "", false
	}
	return value, true
}

func InputEnvToMemory(key string, value string, store *InMemoryStore) error {
	store.Load(map[string]string{key: value})
	return nil
}

func LoadKeyToMemory(key string, store *InMemoryStore) error {
	value, exists := findEnv(key)
	if !exists {
		return fmt.Errorf("환경변수 %s이(가) 존재하지 않습니다", key)
	}
	store.Load(map[string]string{key: value})
	return nil
}
