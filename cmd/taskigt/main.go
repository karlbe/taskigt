package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"taskigt/internal/i18n"
	"taskigt/internal/tasks"
	"taskigt/internal/tui"
)

var BuildVersion = "dev"

func main() {
	printJSON := flag.Bool("json", false, "print tasks as JSON and exit")
	printList := flag.Bool("list", false, "print tasks as human-readable list and exit")
	dataFlag := flag.String("data", "", "path to tasks JSON file (overrides default)")
	flag.Parse()

	storePath := *dataFlag
	if storePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get home directory: %v\n", err)
			os.Exit(1)
		}
		storePath = filepath.Join(home, ".taskigt", "tasks.json")
	}
	store := tasks.NewStore(storePath)

	state, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load tasks: %v\n", err)
		os.Exit(1)
	}

	if *printJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(state)
		return
	}

	if *printList || !isTTY() {
		printHuman(os.Stdout, state, time.Now())
		return
	}

	model := tui.NewModel(state, store, BuildVersion, storePath, i18n.ForCode(state.Lang))
	program := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := program.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
		os.Exit(1)
	}
	type fareweller interface{ Farewell() string }
	if f, ok := finalModel.(fareweller); ok {
		fmt.Println(f.Farewell())
	}
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func printHuman(w io.Writer, state tasks.AppState, now time.Time) {
	if len(state.Tasks) == 0 {
		fmt.Fprintln(w, "No tasks.")
	} else {
		fmt.Fprintf(w, "Tasks (%d):\n", len(state.Tasks))
		for i, t := range state.Tasks {
			status := "[ ]"
			if t.Done {
				status = "[x]"
			} else if t.Paused {
				status = "[~]"
			}
			due := ""
			if t.DueAt != nil {
				days := int(t.DueAt.Sub(now).Hours() / 24)
				switch {
				case days < 0:
					due = fmt.Sprintf(" (overdue %s)", t.DueAt.Format("2006-01-02"))
				case days == 0:
					due = " (due today)"
				case days == 1:
					due = " (due tomorrow)"
				default:
					due = fmt.Sprintf(" (due %s)", t.DueAt.Format("2006-01-02"))
				}
			}
			fmt.Fprintf(w, "  %d. %s %s%s\n", i+1, status, t.Title, due)
		}
	}

	if len(state.Archived) > 0 {
		fmt.Fprintf(w, "\nArchived (%d):\n", len(state.Archived))
		for i, t := range state.Archived {
			fmt.Fprintf(w, "  %d. %s  (%s)\n", i+1, t.Title, t.UpdatedAt.Format("2006-01-02"))
		}
	}
}
