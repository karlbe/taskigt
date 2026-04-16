package tasks

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreLoadCreatesFileWhenMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")
	store := NewStore(path)

	state, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(state.Tasks) != 0 || len(state.Archived) != 0 {
		t.Fatalf("expected empty state")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to be created: %v", err)
	}
}

func TestStoreSaveAndLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")
	store := NewStore(path)

	now := time.Date(2026, 4, 16, 13, 0, 0, 0, time.UTC)
	state := NewEmptyState()
	if err := AddTask(&state, "hello", now); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	if err := store.Save(state); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded.Tasks) != 1 || loaded.Tasks[0].Title != "hello" {
		t.Fatalf("unexpected loaded data")
	}
}
