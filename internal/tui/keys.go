package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Toggle      key.Binding
	Add         key.Binding
	Delete      key.Binding
	MoveMode    key.Binding
	Pause       key.Binding
	EditTitle   key.Binding
	SetDue      key.Binding
	Write       key.Binding
	ArchiveDone key.Binding
	Quit        key.Binding
	Cancel      key.Binding
	Confirm     key.Binding
	Deny        key.Binding
	Backspace   key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up:          key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
		Down:        key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
		Left:        key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "caret left")),
		Right:       key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "caret right")),
		Toggle:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "toggle done")),
		Add:         key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add task")),
		Delete:      key.NewBinding(key.WithKeys("delete"), key.WithHelp("del", "delete task")),
		Confirm:     key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "confirm")),
		Deny:        key.NewBinding(key.WithKeys("n", "esc"), key.WithHelp("n", "cancel")),
		MoveMode:    key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "move mode")),
		Pause:       key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pause")),
		EditTitle:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit title")),
		SetDue:      key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "set due date")),
		Write:       key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "write/save")),
		ArchiveDone: key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "archive/undo")),
		Quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Cancel:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		Backspace:   key.NewBinding(key.WithKeys("backspace"), key.WithHelp("⌫", "backspace")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.MoveMode,
		k.Toggle,
		k.Add,
		k.Delete,
		k.ArchiveDone,
		k.Quit,
	}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.Up,
			k.Down,
			k.MoveMode,
			k.Toggle,
		},
		{
			k.Add,
			k.Delete,
			k.ArchiveDone,
			k.Quit,
		},
	}
}
