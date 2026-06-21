package fsm

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type Command struct {
	Op   string `json:"op"` // "shorten"
	Code string `json:"code"`
	URL  string `json:"url"`
}

type URLStore struct {
	mu   sync.RWMutex
	data map[string]string // short_code -> long_url
}

func New() *URLStore {
	return &URLStore{data: make(map[string]string)}
}

// Apply is called by Raft to save data
func (s *URLStore) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if cmd.Op == "shorten" {
		s.data[cmd.Code] = cmd.URL
	}
	return nil
}

// Resolve reads a URL locally
func (s *URLStore) Resolve(code string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[code]
	return val, ok
}

// Raft requires Snapshot/Restore (keeping it simple)
func (s *URLStore) Snapshot() (raft.FSMSnapshot, error) { return &snap{}, nil }
func (s *URLStore) Restore(io.ReadCloser) error         { return nil }

type snap struct{}

func (s *snap) Persist(sink raft.SnapshotSink) error { return nil }
func (s *snap) Release()                             {}