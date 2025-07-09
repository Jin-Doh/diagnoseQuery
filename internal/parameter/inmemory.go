package parameter

import (
	"sync"
)

type InMemoryStore struct {
	data  map[string]string
	mutex sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]string),
	}
}

func (store *InMemoryStore) Load(envMap map[string]string) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	for key, value := range envMap {
		store.data[key] = value
	}
}

func (store *InMemoryStore) Get(key string) (string, bool) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	value, exists := store.data[key]
	return value, exists
}

func (store *InMemoryStore) GetOrDefault(key string, defaultValue string) string {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	if value, exists := store.data[key]; exists {
		return value
	}
	return defaultValue
}

func (store *InMemoryStore) GetAll() map[string]string {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	copy := make(map[string]string)
	for key, value := range store.data {
		copy[key] = value
	}
	return copy
}
