package tasks

import (
	"strings"
	"time"
)

type Task struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Done      bool       `json:"done"`
	Paused    bool       `json:"paused,omitempty"`
	DueAt     *time.Time `json:"due_at,omitempty"`
	PrevIndex int        `json:"prev_index,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DoneAt    *time.Time `json:"done_at,omitempty"`
	PausedAt  *time.Time `json:"paused_at,omitempty"`
}

type AppState struct {
	Tasks    []Task `json:"tasks"`
	Archived []Task `json:"archived"`
	Lang     string `json:"lang,omitempty"`
}

func NewEmptyState() AppState {
	return AppState{Tasks: []Task{}, Archived: []Task{}}
}

func sanitizeTitle(title string) string {
	return strings.TrimSpace(title)
}

func newTask(title string, now time.Time, id string) Task {
	return Task{
		ID:        id,
		Title:     sanitizeTitle(title),
		Done:      false,
		Paused:    false,
		PrevIndex: -1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
