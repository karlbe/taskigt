package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"taskigt/internal/i18n"
	"taskigt/internal/tasks"
	"taskigt/internal/tui"
)

var BuildVersion = "dev"

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	storePath := filepath.Join(home, ".taskigt", "tasks.json")
	store := tasks.NewStore(storePath)

	state, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load tasks: %v\n", err)
		os.Exit(1)
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
