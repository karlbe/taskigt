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

func TestAddTaskEmptyTitleRejected(t *testing.T) {
	state := NewEmptyState()
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	if err := AddTask(&state, "", now); err == nil {
		t.Fatal("expected error for empty title")
	}
	if err := AddTask(&state, "   ", now); err == nil {
		t.Fatal("expected error for whitespace-only title")
	}
	if len(state.Tasks) != 0 {
		t.Fatalf("expected no tasks, got %d", len(state.Tasks))
	}
}

func TestAddTaskTitleTrimmed(t *testing.T) {
	state := NewEmptyState()
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	_ = AddTask(&state, "  hello  ", now)
	if state.Tasks[0].Title != "hello" {
		t.Fatalf("expected trimmed title 'hello', got %q", state.Tasks[0].Title)
	}
}

func TestRemoveTask(t *testing.T) {
	state := seedState("a", "b", "c")

	if err := RemoveTask(&state, 1); err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	if len(state.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(state.Tasks))
	}
	if state.Tasks[0].Title != "a" || state.Tasks[1].Title != "c" {
		t.Fatalf("unexpected tasks after remove: %v %v", state.Tasks[0].Title, state.Tasks[1].Title)
	}
}

func TestRemoveTaskOutOfRange(t *testing.T) {
	state := seedState("a")

	if err := RemoveTask(&state, -1); err == nil {
		t.Fatal("expected error for negative index")
	}
	if err := RemoveTask(&state, 1); err == nil {
		t.Fatal("expected error for out-of-bounds index")
	}
}

func TestRenameTask(t *testing.T) {
	state := seedState("old")
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	if err := RenameTask(&state, 0, "new", now); err != nil {
		t.Fatalf("rename failed: %v", err)
	}
	if state.Tasks[0].Title != "new" {
		t.Fatalf("expected title 'new', got %q", state.Tasks[0].Title)
	}
	if state.Tasks[0].UpdatedAt != now {
		t.Fatal("UpdatedAt not set after rename")
	}
}

func TestRenameTaskEmptyRejected(t *testing.T) {
	state := seedState("keep")
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	if err := RenameTask(&state, 0, "", now); err == nil {
		t.Fatal("expected error for empty rename")
	}
	if state.Tasks[0].Title != "keep" {
		t.Fatal("title should not change on failed rename")
	}
}

func TestTogglePaused(t *testing.T) {
	state := seedState("task")
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	if err := TogglePaused(&state, 0, now); err != nil {
		t.Fatalf("toggle paused failed: %v", err)
	}
	if !state.Tasks[0].Paused {
		t.Fatal("expected task to be paused")
	}
	if state.Tasks[0].PausedAt == nil || *state.Tasks[0].PausedAt != now {
		t.Fatal("PausedAt not set")
	}

	if err := TogglePaused(&state, 0, now.Add(time.Second)); err != nil {
		t.Fatalf("toggle unpaused failed: %v", err)
	}
	if state.Tasks[0].Paused {
		t.Fatal("expected task to be unpaused")
	}
	if state.Tasks[0].PausedAt != nil {
		t.Fatal("PausedAt should be cleared")
	}
}

func TestTogglePausedIgnoresDoneTask(t *testing.T) {
	state := seedState("task")
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	_, _ = ToggleDone(&state, 0, now)
	_ = TogglePaused(&state, 0, now.Add(time.Second))

	if state.Tasks[0].Paused {
		t.Fatal("done task should not become paused")
	}
}

func TestSetDueDate(t *testing.T) {
	state := seedState("task")
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	due := now.Add(48 * time.Hour)

	if err := SetDueDate(&state, 0, &due, now); err != nil {
		t.Fatalf("set due failed: %v", err)
	}
	if state.Tasks[0].DueAt == nil || *state.Tasks[0].DueAt != due {
		t.Fatal("DueAt not set correctly")
	}

	if err := SetDueDate(&state, 0, nil, now); err != nil {
		t.Fatalf("clear due failed: %v", err)
	}
	if state.Tasks[0].DueAt != nil {
		t.Fatal("DueAt should be nil after clear")
	}
}

func TestArchiveTask(t *testing.T) {
	state := seedState("a", "b", "c")
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	task, err := ArchiveTask(&state, 1, now)
	if err != nil {
		t.Fatalf("archive failed: %v", err)
	}
	if task.Title != "b" {
		t.Fatalf("expected archived 'b', got %q", task.Title)
	}
	if len(state.Tasks) != 2 || len(state.Archived) != 1 {
		t.Fatalf("unexpected counts: tasks=%d archived=%d", len(state.Tasks), len(state.Archived))
	}
	if state.Tasks[0].Title != "a" || state.Tasks[1].Title != "c" {
		t.Fatal("remaining tasks in wrong order")
	}
}

func TestArchiveTaskOutOfRange(t *testing.T) {
	state := seedState("a")
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	if _, err := ArchiveTask(&state, -1, now); err == nil {
		t.Fatal("expected error for negative index")
	}
	if _, err := ArchiveTask(&state, 1, now); err == nil {
		t.Fatal("expected error for out-of-bounds index")
	}
}

func TestUnarchiveTask(t *testing.T) {
	state := seedState("a", "b")
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	_, _ = ArchiveTask(&state, 0, now)
	if err := UnarchiveTask(&state, 0, now.Add(time.Second)); err != nil {
		t.Fatalf("unarchive failed: %v", err)
	}
	if len(state.Tasks) != 2 || len(state.Archived) != 0 {
		t.Fatalf("unexpected counts after unarchive: tasks=%d archived=%d", len(state.Tasks), len(state.Archived))
	}
}

func TestDeleteAllArchived(t *testing.T) {
	state := seedState("a", "b", "c")
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	_, _ = ToggleDone(&state, 0, now)
	_, _ = ToggleDone(&state, 1, now)
	ArchiveDone(&state, now.Add(time.Second))

	DeleteAllArchived(&state)

	if len(state.Archived) != 0 {
		t.Fatalf("expected empty archived, got %d", len(state.Archived))
	}
	if len(state.Tasks) != 1 || state.Tasks[0].Title != "c" {
		t.Fatal("active tasks should be unchanged")
	}
}

func TestMoveEdgeCases(t *testing.T) {
	state := seedState("only")

	idx, _ := MoveUp(&state, 0)
	if idx != 0 {
		t.Fatal("MoveUp on first item should stay at 0")
	}

	idx, _ = MoveDown(&state, 0)
	if idx != 0 {
		t.Fatal("MoveDown on last item should stay at 0")
	}

	idx, _ = MoveTop(&state, 0)
	if idx != 0 {
		t.Fatal("MoveTop on first item should stay at 0")
	}

	idx, _ = MoveBottom(&state, 0)
	if idx != 0 {
		t.Fatal("MoveBottom on last item should stay at 0")
	}
}

func TestToggleDoneTimestamps(t *testing.T) {
	state := seedState("task")
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	_, _ = ToggleDone(&state, 0, now)
	if state.Tasks[0].DoneAt == nil || *state.Tasks[0].DoneAt != now {
		t.Fatal("DoneAt not set when marking done")
	}

	_, _ = ToggleDone(&state, 0, now.Add(time.Second))
	if state.Tasks[0].DoneAt != nil {
		t.Fatal("DoneAt should be cleared when unmarking done")
	}
}

func TestArchiveDoneEmptyState(t *testing.T) {
	state := NewEmptyState()
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	count := ArchiveDone(&state, now)
	if count != 0 {
		t.Fatalf("expected 0 archived from empty state, got %d", count)
	}
}

func TestUnarchiveTasksEmpty(t *testing.T) {
	state := seedState("a")
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)

	restored := UnarchiveTasks(&state, nil, now)
	if restored != 0 {
		t.Fatalf("expected 0 restored for nil input, got %d", restored)
	}
}
