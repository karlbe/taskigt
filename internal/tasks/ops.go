package tasks

import (
	"fmt"
	"time"
)

func AddTask(state *AppState, title string, now time.Time) error {
	title = sanitizeTitle(title)
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	id := fmt.Sprintf("%d", now.UnixNano())
	state.Tasks = append(state.Tasks, newTask(title, now, id))
	return nil
}

func RemoveTask(state *AppState, idx int) error {
	if idx < 0 || idx >= len(state.Tasks) {
		return fmt.Errorf("index out of range")
	}
	state.Tasks = append(state.Tasks[:idx], state.Tasks[idx+1:]...)
	return nil
}

func MoveUp(state *AppState, idx int) (int, error) {
	if idx <= 0 || idx >= len(state.Tasks) {
		return idx, nil
	}
	state.Tasks[idx-1], state.Tasks[idx] = state.Tasks[idx], state.Tasks[idx-1]
	return idx - 1, nil
}

func MoveDown(state *AppState, idx int) (int, error) {
	if idx < 0 || idx >= len(state.Tasks)-1 {
		return idx, nil
	}
	state.Tasks[idx], state.Tasks[idx+1] = state.Tasks[idx+1], state.Tasks[idx]
	return idx + 1, nil
}

func MoveTop(state *AppState, idx int) (int, error) {
	if idx <= 0 || idx >= len(state.Tasks) {
		return idx, nil
	}
	task := state.Tasks[idx]
	state.Tasks = append(state.Tasks[:idx], state.Tasks[idx+1:]...)
	state.Tasks = append([]Task{task}, state.Tasks...)
	return 0, nil
}

func MoveBottom(state *AppState, idx int) (int, error) {
	if idx < 0 || idx >= len(state.Tasks)-1 {
		return idx, nil
	}
	task := state.Tasks[idx]
	state.Tasks = append(state.Tasks[:idx], state.Tasks[idx+1:]...)
	state.Tasks = append(state.Tasks, task)
	return len(state.Tasks) - 1, nil
}

func RenameTask(state *AppState, idx int, title string, now time.Time) error {
	title = sanitizeTitle(title)
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if idx < 0 || idx >= len(state.Tasks) {
		return fmt.Errorf("index out of range")
	}
	state.Tasks[idx].Title = title
	state.Tasks[idx].UpdatedAt = now
	return nil
}

func ToggleDone(state *AppState, idx int, now time.Time) (int, error) {
	if idx < 0 || idx >= len(state.Tasks) {
		return idx, fmt.Errorf("index out of range")
	}

	t := state.Tasks[idx]
	if !t.Done {
		t.Done = true
		t.DoneAt = now
		t.UpdatedAt = now
		t.PrevIndex = -1
		state.Tasks[idx] = t
		return idx, nil
	}

	t.Done = false
	t.DoneAt = time.Time{}
	t.UpdatedAt = now
	t.PrevIndex = -1
	state.Tasks[idx] = t
	return idx, nil
}

func TogglePaused(state *AppState, idx int, now time.Time) error {
	if idx < 0 || idx >= len(state.Tasks) {
		return fmt.Errorf("index out of range")
	}
	t := &state.Tasks[idx]
	if t.Done {
		return nil // done tasks cannot be paused
	}
	t.Paused = !t.Paused
	if t.Paused {
		t.PausedAt = now
	} else {
		t.PausedAt = time.Time{}
	}
	t.UpdatedAt = now
	return nil
}

func SetDueDate(state *AppState, idx int, due *time.Time, now time.Time) error {
	if idx < 0 || idx >= len(state.Tasks) {
		return fmt.Errorf("index out of range")
	}
	state.Tasks[idx].DueAt = due
	state.Tasks[idx].UpdatedAt = now
	return nil
}

func ArchiveDone(state *AppState, now time.Time) int {
	return len(ArchiveDoneTasks(state, now))
}

func ArchiveDoneTasks(state *AppState, now time.Time) []Task {
	if len(state.Tasks) == 0 {
		return nil
	}

	remaining := make([]Task, 0, len(state.Tasks))
	archived := make([]Task, 0)
	for _, task := range state.Tasks {
		if task.Done {
			task.UpdatedAt = now
			state.Archived = append(state.Archived, task)
			archived = append(archived, task)
			continue
		}
		remaining = append(remaining, task)
	}
	state.Tasks = remaining
	return archived
}

func UnarchiveTasks(state *AppState, toRestore []Task, now time.Time) int {
	if len(toRestore) == 0 || len(state.Archived) == 0 {
		return 0
	}

	ids := make(map[string]struct{}, len(toRestore))
	for _, task := range toRestore {
		ids[task.ID] = struct{}{}
	}

	remainingArchived := make([]Task, 0, len(state.Archived))
	restored := make([]Task, 0, len(toRestore))
	for _, task := range state.Archived {
		if _, ok := ids[task.ID]; ok {
			task.UpdatedAt = now
			restored = append(restored, task)
			continue
		}
		remainingArchived = append(remainingArchived, task)
	}

	state.Archived = remainingArchived
	state.Tasks = append(state.Tasks, restored...)
	return len(restored)
}
