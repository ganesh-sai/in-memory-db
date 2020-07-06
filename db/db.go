package db

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type memoryDB struct {
	items map[string]string
	mu    sync.RWMutex
}

func newDB() memoryDB {
	file, err := os.Open("db.json")
	if err != nil {
		return memoryDB{items: map[string]string{}}
	}
	items := map[string]string{}
	if err := json.NewDecoder(file).Decode(&items); err != nil {
		fmt.Println("cannot decode", err.Error())
		return memoryDB{items: map[string]string{}}
	}
	return memoryDB{items: items}
}

func (m *memoryDB) save() {
	file, err := os.Create("db.json")
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(file).Encode(&m.items); err != nil {
		log.Fatal(err)
	}
}

func (m *memoryDB) set(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = value
}

func (m *memoryDB) get(key string) (value string, found bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, found = m.items[key]
	return
}

func (m *memoryDB) delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
}
