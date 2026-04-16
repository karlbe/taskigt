package tui

import (
	"github.com/charmbracelet/bubbles/key"

	"taskigt/internal/i18n"
)

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
	About       key.Binding
}

func defaultKeyMap(s i18n.Strings) keyMap {
	return keyMap{
		Up:          key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", s.HelpUp)),
		Down:        key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", s.HelpDown)),
		Left:        key.NewBinding(key.WithKeys("left"), key.WithHelp("←", s.HelpLeft)),
		Right:       key.NewBinding(key.WithKeys("right"), key.WithHelp("→", s.HelpRight)),
		Toggle:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", s.HelpToggle)),
		Add:         key.NewBinding(key.WithKeys("a"), key.WithHelp("a", s.HelpAdd)),
		Delete:      key.NewBinding(key.WithKeys("delete"), key.WithHelp("del", s.HelpDelete)),
		Confirm:     key.NewBinding(key.WithKeys("y"), key.WithHelp("y", s.HelpToggle)),
		Deny:        key.NewBinding(key.WithKeys("n", "esc"), key.WithHelp("n", s.HelpCancel)),
		MoveMode:    key.NewBinding(key.WithKeys("m"), key.WithHelp("m", s.HelpMoveMode)),
		Pause:       key.NewBinding(key.WithKeys("p"), key.WithHelp("p", s.HelpPause)),
		EditTitle:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", s.HelpEditTitle)),
		SetDue:      key.NewBinding(key.WithKeys("D"), key.WithHelp("D", s.HelpSetDue)),
		Write:       key.NewBinding(key.WithKeys("w"), key.WithHelp("w", s.HelpWrite)),
		ArchiveDone: key.NewBinding(key.WithKeys("x"), key.WithHelp("x", s.HelpArchive)),
		Quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", s.HelpQuit)),
		Cancel:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", s.HelpCancel)),
		Backspace:   key.NewBinding(key.WithKeys("backspace"), key.WithHelp("⌫", s.HelpBackspace)),
		About:       key.NewBinding(key.WithKeys("?"), key.WithHelp("?", s.HelpAbout)),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.MoveMode, k.Toggle, k.Add, k.Delete, k.ArchiveDone, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.MoveMode, k.Toggle},
		{k.Add, k.Delete, k.ArchiveDone, k.Quit},
	}
}
