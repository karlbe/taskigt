package tasks

import (
	"testing"
	"time"
)

func seedState(titles ...string) AppState {
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	state := NewEmptyState()
	for _, title := range titles {
		_ = AddTask(&state, title, now)
		now = now.Add(time.Second)
	}
	return state
}

func TestToggleDoneKeepsPosition(t *testing.T) {
	state := seedState("one", "two", "three")
	now := time.Date(2026, 4, 16, 11, 0, 0, 0, time.UTC)

	idx, err := ToggleDone(&state, 1, now)
	if err != nil {
		t.Fatalf("toggle done failed: %v", err)
	}
	if idx != 1 {
		t.Fatalf("expected index 1, got %d", idx)
	}
	if state.Tasks[1].Title != "two" || !state.Tasks[1].Done {
		t.Fatalf("expected task 'two' done at index 1")
	}

	idx, err = ToggleDone(&state, 1, now.Add(time.Second))
	if err != nil {
		t.Fatalf("toggle undo failed: %v", err)
	}
	if idx != 1 {
		t.Fatalf("expected index 1, got %d", idx)
	}
	if state.Tasks[1].Title != "two" || state.Tasks[1].Done {
		t.Fatalf("expected task 'two' restored as undone at same index")
	}
}

func TestMoveOperations(t *testing.T) {
	state := seedState("a", "b", "c")

	idx, _ := MoveDown(&state, 0)
	if idx != 1 || state.Tasks[1].Title != "a" {
		t.Fatalf("move down failed")
	}

	idx, _ = MoveUp(&state, 1)
	if idx != 0 || state.Tasks[0].Title != "a" {
		t.Fatalf("move up failed")
	}

	idx, _ = MoveBottom(&state, 0)
	if idx != 2 || state.Tasks[2].Title != "a" {
		t.Fatalf("move bottom failed")
	}

	idx, _ = MoveTop(&state, 2)
	if idx != 0 || state.Tasks[0].Title != "a" {
		t.Fatalf("move top failed")
	}
}

func TestArchiveDone(t *testing.T) {
	state := seedState("x", "y", "z")
	now := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)

	_, _ = ToggleDone(&state, 0, now)
	moved := ArchiveDone(&state, now.Add(time.Second))

	if moved != 1 {
		t.Fatalf("expected moved=1, got %d", moved)
	}
	if len(state.Tasks) != 2 || len(state.Archived) != 1 {
		t.Fatalf("unexpected active/archive lengths: %d/%d", len(state.Tasks), len(state.Archived))
	}
}

func TestArchiveAndUndo(t *testing.T) {
	state := seedState("x", "y", "z")
	now := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)

	_, _ = ToggleDone(&state, 0, now)
	archived := ArchiveDoneTasks(&state, now.Add(time.Second))
	if len(archived) != 1 {
		t.Fatalf("expected 1 archived task, got %d", len(archived))
	}

	restored := UnarchiveTasks(&state, archived, now.Add(2*time.Second))
	if restored != 1 {
		t.Fatalf("expected restored=1, got %d", restored)
	}
	if len(state.Tasks) != 3 || len(state.Archived) != 0 {
		t.Fatalf("unexpected active/archive lengths after undo: %d/%d", len(state.Tasks), len(state.Archived))
	}
}
