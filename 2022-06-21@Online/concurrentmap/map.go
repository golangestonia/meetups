package concurrentmap

//go:generate go run github.com/thepudds/fzgen/cmd/fzgen@main -chain -parallel .

// To see the failing output:
//   unix:
//     export FZDEBUG=repro=1
//
//   windows:
//     set FZDEBUG=repro=1

import "sync"

// Map a broken concurrent map.
type Map struct {
	mu   sync.RWMutex
	data map[string]string
}

func New() *Map { return &Map{} }

func (m *Map) Set(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.data == nil {
		m.data = make(map[string]string)
	}

	m.data[key] = value
}

func (m *Map) Get(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.data == nil {
		m.data = make(map[string]string)
	}

	v, ok := m.data[key]
	return v, ok
}
