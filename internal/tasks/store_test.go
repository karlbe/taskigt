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

func TestStoreSavePreservesArchived(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")
	store := NewStore(path)

	now := time.Date(2026, 4, 16, 13, 0, 0, 0, time.UTC)
	state := NewEmptyState()
	_ = AddTask(&state, "active", now)
	_ = AddTask(&state, "done", now.Add(time.Second))
	_, _ = ToggleDone(&state, 1, now.Add(2*time.Second))
	ArchiveDone(&state, now.Add(3*time.Second))

	if err := store.Save(state); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded.Tasks) != 1 || loaded.Tasks[0].Title != "active" {
		t.Fatalf("unexpected active tasks after load")
	}
	if len(loaded.Archived) != 1 || loaded.Archived[0].Title != "done" {
		t.Fatalf("unexpected archived tasks after load")
	}
}

func TestStoreSavesAndLoadsLang(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")
	store := NewStore(path)

	state := NewEmptyState()
	state.Lang = "sv"

	if err := store.Save(state); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.Lang != "sv" {
		t.Fatalf("expected lang 'sv', got %q", loaded.Lang)
	}
}

func TestStoreLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	if err := os.WriteFile(path, []byte("{not valid json}"), 0600); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	store := NewStore(path)
	if _, err := store.Load(); err == nil {
		t.Fatal("expected error loading invalid JSON")
	}
}
