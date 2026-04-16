package tasks

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Store struct {
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Load() (AppState, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			state := NewEmptyState()
			if saveErr := s.Save(state); saveErr != nil {
				return AppState{}, saveErr
			}
			return state, nil
		}
		return AppState{}, err
	}

	if len(data) == 0 {
		return NewEmptyState(), nil
	}

	var state AppState
	if err := json.Unmarshal(data, &state); err != nil {
		return AppState{}, err
	}
	if state.Tasks == nil {
		state.Tasks = []Task{}
	}
	if state.Archived == nil {
		state.Archived = []Task{}
	}
	return state, nil
}

func (s *Store) Save(state AppState) error {
	if state.Tasks == nil {
		state.Tasks = []Task{}
	}
	if state.Archived == nil {
		state.Archived = []Task{}
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "tasks-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, s.path)
}
