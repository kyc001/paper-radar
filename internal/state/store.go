package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyc001/paper-radar/internal/model"
)

type FileState struct {
	SeenIDs map[string]bool     `json:"seen_ids"`
	Pending []model.ScoredPaper `json:"pending"`
}

type Store struct {
	path string
}

func New(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Load() (FileState, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return emptyState(), nil
		}
		return FileState{}, err
	}

	var st FileState
	if err := json.Unmarshal(data, &st); err != nil {
		return FileState{}, err
	}

	if st.SeenIDs == nil {
		st.SeenIDs = map[string]bool{}
	}
	if st.Pending == nil {
		st.Pending = []model.ScoredPaper{}
	}

	return st, nil
}

func (s *Store) Save(st FileState) error {
	if st.SeenIDs == nil {
		st.SeenIDs = map[string]bool{}
	}
	if st.Pending == nil {
		st.Pending = []model.ScoredPaper{}
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmpFile := s.path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmpFile, s.path); err != nil {
		return fmt.Errorf("replace state file: %w", err)
	}

	return nil
}

func emptyState() FileState {
	return FileState{
		SeenIDs: map[string]bool{},
		Pending: []model.ScoredPaper{},
	}
}
