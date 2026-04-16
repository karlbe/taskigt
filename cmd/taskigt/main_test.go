package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"taskigt/internal/tasks"
)

var now = time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)

func makeState() tasks.AppState {
	state := tasks.NewEmptyState()
	_ = tasks.AddTask(&state, "buy milk", now)
	_ = tasks.AddTask(&state, "write tests", now.Add(time.Second))
	_ = tasks.AddTask(&state, "done task", now.Add(2*time.Second))
	_, _ = tasks.ToggleDone(&state, 2, now.Add(3*time.Second))
	return state
}

func TestPrintHumanEmpty(t *testing.T) {
	var buf bytes.Buffer
	printHuman(&buf, tasks.NewEmptyState(), now)
	if !strings.Contains(buf.String(), "No tasks.") {
		t.Fatalf("expected 'No tasks.', got: %q", buf.String())
	}
}

func TestPrintHumanTaskCount(t *testing.T) {
	var buf bytes.Buffer
	printHuman(&buf, makeState(), now)
	out := buf.String()
	if !strings.Contains(out, "Tasks (3):") {
		t.Fatalf("expected task count header, got: %q", out)
	}
}

func TestPrintHumanStatusMarkers(t *testing.T) {
	state := tasks.NewEmptyState()
	_ = tasks.AddTask(&state, "todo", now)
	_ = tasks.AddTask(&state, "paused", now.Add(time.Second))
	_ = tasks.AddTask(&state, "done", now.Add(2*time.Second))
	_ = tasks.TogglePaused(&state, 1, now)
	_, _ = tasks.ToggleDone(&state, 2, now)

	var buf bytes.Buffer
	printHuman(&buf, state, now)
	out := buf.String()

	if !strings.Contains(out, "[ ] todo") {
		t.Errorf("expected '[ ] todo' in output")
	}
	if !strings.Contains(out, "[~] paused") {
		t.Errorf("expected '[~] paused' in output")
	}
	if !strings.Contains(out, "[x] done") {
		t.Errorf("expected '[x] done' in output")
	}
}

func TestPrintHumanDueDateLabels(t *testing.T) {
	state := tasks.NewEmptyState()
	_ = tasks.AddTask(&state, "overdue", now)
	_ = tasks.AddTask(&state, "today", now.Add(time.Second))
	_ = tasks.AddTask(&state, "tomorrow", now.Add(2*time.Second))
	_ = tasks.AddTask(&state, "future", now.Add(3*time.Second))

	past := now.Add(-24 * time.Hour)
	today := now.Add(1 * time.Hour)
	tomorrow := now.Add(25 * time.Hour)
	future := now.Add(72 * time.Hour)

	_ = tasks.SetDueDate(&state, 0, &past, now)
	_ = tasks.SetDueDate(&state, 1, &today, now)
	_ = tasks.SetDueDate(&state, 2, &tomorrow, now)
	_ = tasks.SetDueDate(&state, 3, &future, now)

	var buf bytes.Buffer
	printHuman(&buf, state, now)
	out := buf.String()

	if !strings.Contains(out, "overdue") {
		t.Errorf("expected 'overdue' label")
	}
	if !strings.Contains(out, "due today") {
		t.Errorf("expected 'due today' label")
	}
	if !strings.Contains(out, "due tomorrow") {
		t.Errorf("expected 'due tomorrow' label")
	}
	if !strings.Contains(out, future.Format("2006-01-02")) {
		t.Errorf("expected future date in output")
	}
}

func TestPrintHumanArchivedSection(t *testing.T) {
	state := makeState()
	tasks.ArchiveDone(&state, now)

	var buf bytes.Buffer
	printHuman(&buf, state, now)
	out := buf.String()

	if !strings.Contains(out, "Archived (1):") {
		t.Errorf("expected archived section header")
	}
	if !strings.Contains(out, "done task") {
		t.Errorf("expected archived task title in output")
	}
}

func TestPrintHumanNoArchivedSectionWhenEmpty(t *testing.T) {
	var buf bytes.Buffer
	printHuman(&buf, makeState(), now)
	if strings.Contains(buf.String(), "Archived") {
		t.Error("should not print archived section when none exist")
	}
}

func TestJSONOutputRoundtrip(t *testing.T) {
	state := makeState()

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	_ = enc.Encode(state)

	var decoded tasks.AppState
	if err := json.NewDecoder(&buf).Decode(&decoded); err != nil {
		t.Fatalf("failed to decode JSON output: %v", err)
	}
	if len(decoded.Tasks) != len(state.Tasks) {
		t.Fatalf("expected %d tasks, got %d", len(state.Tasks), len(decoded.Tasks))
	}
	for i, task := range decoded.Tasks {
		if task.Title != state.Tasks[i].Title {
			t.Errorf("task %d: expected %q, got %q", i, state.Tasks[i].Title, task.Title)
		}
	}
}
