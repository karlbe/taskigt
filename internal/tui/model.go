package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"taskigt/internal/tasks"
)

var (
	hintStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	statusStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
	selectedRowStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	doneRowStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Strikethrough(true)
	pausedRowStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	overdueDateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dueDateStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	topBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("63")).
		Foreground(lipgloss.Color("255")).
		Bold(true).
		Padding(0, 1)

	btmBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("244")).
		Padding(0, 1)

	btmBarPieceStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("244"))

	btmBarKeyStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("255")).
		Bold(true)
)

// editDraft holds unsaved field values while the edit screen is open.
type editDraft struct {
	Title  string
	DueRaw string // YYYY-MM-DD, +N, or "" to clear
}

type model struct {
	state          tasks.AppState
	store          *tasks.Store
	keys           keyMap
	buildVersion   string
	storePath      string
	cursor         int
	scroll         int
	width          int
	height         int
	moving         bool
	confirming     bool
	aboutScreen    bool
	// edit screen
	editScreen        bool
	editIsNew         bool // true when creating a new task
	editFieldIdx      int  // focused field row
	editFieldActive   bool // currently typing into a field
	editDirty         bool // unsaved changes
	editDiscardPrompt bool // esc pressed once with dirty data
	editDraft         editDraft
	moveSnapshot   []tasks.Task
	moveCursorOrig int
	input          string
	inputCursor    int  // rune index into input
	status         string
	lastErr        error
	undoSet        []tasks.Task
}

func NewModel(state tasks.AppState, store *tasks.Store, buildVersion string, storePath string) tea.Model {
	m := model{
		state:        state,
		store:        store,
		keys:         defaultKeyMap(),
		buildVersion: buildVersion,
		storePath:    storePath,
		cursor:       0,
	}
	m.clampCursor()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.aboutScreen {
			m.aboutScreen = false
			return m, nil
		}
		if m.confirming {
			return m.handleConfirmMode(msg)
		}
		if m.editScreen {
			return m.handleEditScreen(msg)
		}
		return m.handleNormalMode(msg)
	}
	return m, nil
}

// ── input helpers (rune-safe) ────────────────────────────────────────────────

func (m *model) inputSet(s string) {
	m.input = s
	m.inputCursor = len([]rune(s))
}

func (m *model) inputClear() {
	m.input = ""
	m.inputCursor = 0
}

func (m *model) inputMoveLeft() {
	if m.inputCursor > 0 {
		m.inputCursor--
	}
}

func (m *model) inputMoveRight() {
	if m.inputCursor < len([]rune(m.input)) {
		m.inputCursor++
	}
}

func (m *model) inputInsert(s string) {
	runes := []rune(m.input)
	new := make([]rune, 0, len(runes)+len([]rune(s)))
	new = append(new, runes[:m.inputCursor]...)
	new = append(new, []rune(s)...)
	new = append(new, runes[m.inputCursor:]...)
	m.input = string(new)
	m.inputCursor += len([]rune(s))
}

func (m *model) inputBackspace() {
	if m.inputCursor == 0 {
		return
	}
	runes := []rune(m.input)
	new := make([]rune, 0, len(runes)-1)
	new = append(new, runes[:m.inputCursor-1]...)
	new = append(new, runes[m.inputCursor:]...)
	m.input = string(new)
	m.inputCursor--
}

func (m *model) inputDelete() {
	runes := []rune(m.input)
	if m.inputCursor >= len(runes) {
		return
	}
	new := make([]rune, 0, len(runes)-1)
	new = append(new, runes[:m.inputCursor]...)
	new = append(new, runes[m.inputCursor+1:]...)
	m.input = string(new)
}

func (m *model) inputBefore() string {
	return string([]rune(m.input)[:m.inputCursor])
}

func (m *model) inputAfter() string {
	runes := []rune(m.input)
	if m.inputCursor >= len(runes) {
		return ""
	}
	return string(runes[m.inputCursor:])
}

// ──────────────────────────────────────────────────────────────────────────────

func (m model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.About):
		m.aboutScreen = true
		return m, nil
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Up):
		if m.moving {
			idx, _ := tasks.MoveUp(&m.state, m.cursor)
			m.cursor = idx
			m.adjustScroll()
			break
		}
		if m.cursor > 0 {
			m.cursor--
			m.adjustScroll()
		}
	case key.Matches(msg, m.keys.Down):
		if m.moving {
			idx, _ := tasks.MoveDown(&m.state, m.cursor)
			m.cursor = idx
			m.adjustScroll()
			break
		}
		if m.cursor < len(m.state.Tasks)-1 {
			m.cursor++
			m.adjustScroll()
		}
	case key.Matches(msg, m.keys.Toggle):
		if m.moving {
			// confirm move
			m.moving = false
			m.moveSnapshot = nil
			m.status = "order saved"
			m.lastErr = nil
			m.save()
			break
		}
		if len(m.state.Tasks) == 0 {
			m.status = "no task selected"
			break
		}
		idx, err := tasks.ToggleDone(&m.state, m.cursor, time.Now())
		if err != nil {
			m.status = err.Error()
			m.lastErr = err
			break
		}
		m.cursor = idx
		m.clampCursor()
		m.status = ""
		m.lastErr = nil
		m.undoSet = nil
		m.save()
	case key.Matches(msg, m.keys.MoveMode):
		if m.moving {
			// abort — restore snapshot
			m.state.Tasks = m.moveSnapshot
			m.cursor = m.moveCursorOrig
			m.moveSnapshot = nil
			m.moving = false
			m.status = "move cancelled"
			m.lastErr = nil
			m.save()
			break
		}
		// enter move mode
		snap := make([]tasks.Task, len(m.state.Tasks))
		copy(snap, m.state.Tasks)
		m.moveSnapshot = snap
		m.moveCursorOrig = m.cursor
		m.moving = true
		m.status = ""
		m.lastErr = nil
	case key.Matches(msg, m.keys.Cancel):
		if m.moving {
			// abort — restore snapshot
			m.state.Tasks = m.moveSnapshot
			m.cursor = m.moveCursorOrig
			m.moveSnapshot = nil
			m.moving = false
			m.status = "move cancelled"
			m.lastErr = nil
			m.save()
		}
	case key.Matches(msg, m.keys.Add):
		m.editDraft = editDraft{}
		m.editScreen = true
		m.editIsNew = true
		m.editFieldIdx = 0
		m.editFieldActive = true
		m.editDirty = false
		m.editDiscardPrompt = false
		m.inputClear()
		m.status = ""
		m.lastErr = nil
	case key.Matches(msg, m.keys.Delete):
		if len(m.state.Tasks) == 0 {
			m.status = "no task selected"
			break
		}
		m.confirming = true
		m.status = ""
		m.lastErr = nil
	case key.Matches(msg, m.keys.Pause):
		if len(m.state.Tasks) == 0 {
			break
		}
		_ = tasks.TogglePaused(&m.state, m.cursor, time.Now())
		m.status = ""
		m.lastErr = nil
		m.save()
	case key.Matches(msg, m.keys.EditTitle):
		if len(m.state.Tasks) == 0 {
			break
		}
		t := m.state.Tasks[m.cursor]
		due := ""
		if t.DueAt != nil {
			due = t.DueAt.Format("2006-01-02")
		}
		m.editDraft = editDraft{Title: t.Title, DueRaw: due}
		m.editScreen = true
		m.editIsNew = false
		m.editFieldIdx = 0
		m.editFieldActive = false
		m.editDirty = false
		m.editDiscardPrompt = false
		m.status = ""
		m.lastErr = nil
	case key.Matches(msg, m.keys.ArchiveDone):
		now := time.Now()
		if len(m.undoSet) > 0 {
			restored := tasks.UnarchiveTasks(&m.state, m.undoSet, now)
			m.undoSet = nil
			m.clampCursor()
			m.status = fmt.Sprintf("restored %d task(s)", restored)
			m.lastErr = nil
			m.save()
			break
		}

		archived := tasks.ArchiveDoneTasks(&m.state, now)
		m.clampCursor()
		if len(archived) == 0 {
			m.status = "no done tasks to archive"
			m.lastErr = nil
			break
		}
		m.undoSet = archived
		m.status = fmt.Sprintf("archived %d task(s) (press x to undo)", len(archived))
		m.lastErr = nil
		m.save()
	}

	return m, nil
}

// handleEditScreen handles all input while the edit screen is open.
func (m model) handleEditScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// field definitions (add more here in future)
	numFields := 2 // title, due

	if m.editFieldActive {
		// Typing into a field
		switch {
		case key.Matches(msg, m.keys.Cancel):
			// revert this field to saved value (empty for new tasks)
			if m.editFieldIdx == 0 {
				if m.editIsNew {
					m.editDraft.Title = ""
				} else {
					m.editDraft.Title = m.state.Tasks[m.cursor].Title
				}
			} else {
				if m.editIsNew {
					m.editDraft.DueRaw = ""
				} else if m.state.Tasks[m.cursor].DueAt != nil {
					m.editDraft.DueRaw = m.state.Tasks[m.cursor].DueAt.Format("2006-01-02")
				} else {
					m.editDraft.DueRaw = ""
				}
			}
			m.editFieldActive = false
			m.inputClear()
			return m, nil
		case key.Matches(msg, m.keys.Backspace):
			m.inputBackspace()
			return m, nil
		case key.Matches(msg, m.keys.Delete):
			m.inputDelete()
			return m, nil
		case key.Matches(msg, m.keys.Left):
			m.inputMoveLeft()
			return m, nil
		case key.Matches(msg, m.keys.Right):
			m.inputMoveRight()
			return m, nil
		case key.Matches(msg, m.keys.Toggle): // enter = confirm field
			if m.editFieldIdx == 0 {
				if strings.TrimSpace(m.input) == "" {
					m.status = "title cannot be empty"
					m.lastErr = fmt.Errorf("empty")
					return m, nil
				}
				m.editDraft.Title = m.input
			} else {
				m.editDraft.DueRaw = m.input
			}
			m.editDirty = true
			m.editFieldActive = false
			m.inputClear()
			m.status = ""
			m.lastErr = nil
			return m, nil
		default:
			if msg.Type == tea.KeySpace && m.editFieldIdx == 0 {
				m.inputInsert(" ")
			} else if msg.Type == tea.KeyRunes {
				m.inputInsert(msg.String())
			}
			return m, nil
		}
	}

	// Field list navigation
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.editFieldIdx > 0 {
			m.editFieldIdx--
		}
	case key.Matches(msg, m.keys.Down):
		if m.editFieldIdx < numFields-1 {
			m.editFieldIdx++
		}
	case key.Matches(msg, m.keys.Toggle): // enter = activate field
		if m.editFieldIdx == 0 {
			m.inputSet(m.editDraft.Title)
		} else {
			m.inputSet(m.editDraft.DueRaw)
		}
		m.editFieldActive = true
		m.status = ""
		m.lastErr = nil
	case key.Matches(msg, m.keys.Write): // w = save
		if err := m.applyEditDraft(); err != nil {
			m.status = err.Error()
			m.lastErr = err
			return m, nil
		}
		m.editScreen = false
		m.editIsNew = false
		m.editDirty = false
		m.editDiscardPrompt = false
		m.status = "saved"
		m.lastErr = nil
		m.save()
	case key.Matches(msg, m.keys.Cancel): // esc
		if m.editDirty && !m.editDiscardPrompt {
			m.editDiscardPrompt = true
			m.status = "unsaved changes — press esc again to discard, w to save"
			m.lastErr = nil
			return m, nil
		}
		// double-esc or no changes
		m.editScreen = false
		m.editIsNew = false
		m.editDirty = false
		m.editDiscardPrompt = false
		m.inputClear()
		m.status = ""
		m.lastErr = nil
	default:
		m.editDiscardPrompt = false // any other key resets
	}
	return m, nil
}

func (m *model) applyEditDraft() error {
	now := time.Now()
	if m.editIsNew {
		if err := tasks.AddTask(&m.state, m.editDraft.Title, now); err != nil {
			return err
		}
		m.cursor = len(m.state.Tasks) - 1
	} else {
		if err := tasks.RenameTask(&m.state, m.cursor, m.editDraft.Title, now); err != nil {
			return err
		}
	}
	raw := strings.TrimSpace(m.editDraft.DueRaw)
	if raw == "" {
		if !m.editIsNew {
			_ = tasks.SetDueDate(&m.state, m.cursor, nil, now)
		}
	} else {
		var due time.Time
		var err error
		if strings.HasPrefix(raw, "+") {
			var days int
			_, err = fmt.Sscanf(raw[1:], "%d", &days)
			if err == nil {
				due = now.AddDate(0, 0, days).Truncate(24 * time.Hour)
			}
		} else {
			due, err = time.Parse("2006-01-02", raw)
		}
		if err != nil {
			return fmt.Errorf("invalid date \"" + raw + "\" — use YYYY-MM-DD or +N days")
		}
		_ = tasks.SetDueDate(&m.state, m.cursor, &due, now)
	}
	return nil
}

func (m model) handleConfirmMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		if err := tasks.RemoveTask(&m.state, m.cursor); err != nil {
			m.status = err.Error()
			m.lastErr = err
		} else {
			m.clampCursor()
			m.status = "task removed"
			m.lastErr = nil
			m.undoSet = nil
			m.save()
		}
		m.confirming = false
	case key.Matches(msg, m.keys.Deny):
		m.confirming = false
		m.status = "delete cancelled"
		m.lastErr = nil
	}
	return m, nil
}

func (m *model) save() {
	if err := m.store.Save(m.state); err != nil {
		m.lastErr = err
		m.status = fmt.Sprintf("save error: %v", err)
	}
}

func (m *model) clampCursor() {
	if len(m.state.Tasks) == 0 {
		m.cursor = 0
		m.scroll = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.state.Tasks) {
		m.cursor = len(m.state.Tasks) - 1
	}
	m.adjustScroll()
}

func (m *model) adjustScroll() {
	listH := m.listHeight()
	if listH <= 0 {
		return
	}
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+listH {
		m.scroll = m.cursor - listH + 1
	}
}

func (m *model) listHeight() int {
	// Called during cursor clamping/scroll before View; use a best-effort
	// estimate based on current width. View() re-derives listH from the
	// exact rendered key bar line count to stay in sync.
	w := m.width
	if w == 0 {
		w = 80
	}
	h := m.height - 2 - m.keyBarLineCount(w)
	if h < 1 {
		h = 1
	}
	return h
}

// keyBarItems returns the individual hint chunks for the current mode.
// Each chunk is a pair of rendered strings: [keyPart, labelPart].
func (m *model) keyBarItems() [][2]string {
	if m.confirming {
		return [][2]string{{"y", "confirm"}, {"n/esc", "cancel"}}
	}
	if m.editScreen {
		if m.editFieldActive {
			return [][2]string{{"enter", "confirm field"}, {"esc", "revert field"}, {"←→", "move caret"}, {"⌫/del", "delete char"}}
		}
		return [][2]string{{"↑↓", "navigate"}, {"enter", "edit field"}, {"w", "save"}, {"esc", "discard"}}
	}
	if m.moving {
		return [][2]string{{"↑↓", "reorder"}, {"enter", "confirm"}, {"m/esc", "abort"}}
	}
	return [][2]string{
		{"enter", "toggle"},
		{"a", "add"},
		{"e", "edit"},
		{"del", "delete"},
		{"p", "pause"},
		{"m", "move"},
		{"x", "archive"},
		{"?", "about"},
		{"q", "quit"},
	}
}

// wrapKeyBar groups hint items into lines that each fit within w,
// returning each line pre-rendered.
func (m *model) wrapKeyBar(w int) []string {
	items := m.keyBarItems()
	sep := btmBarPieceStyle.Render(" | ")
	sepW := lipgloss.Width(sep)

	type chunk struct {
		renderedItem string
		displayW     int
	}
	chunks := make([]chunk, 0, len(items))
	for _, it := range items {
		rendered := btmBarKeyStyle.Render(it[0]) + btmBarPieceStyle.Render(" "+it[1])
		chunks = append(chunks, chunk{rendered, lipgloss.Width(rendered)})
	}

	// padding(0,1) on btmBarStyle = 1 char each side
	avail := w - 2
	if avail < 4 {
		avail = 4
	}

	var lines []string
	lineRendered := ""
	lineW := 0
	for i, c := range chunks {
		sepRendered := ""
		sw := 0
		if i > 0 {
			sepRendered = sep
			sw = sepW
		}
		if i > 0 && lineW+sw+c.displayW > avail {
			lines = append(lines, btmBarStyle.Width(w).Render(lineRendered))
			lineRendered = c.renderedItem
			lineW = c.displayW
		} else {
			lineRendered += sepRendered + c.renderedItem
			lineW += sw + c.displayW
		}
	}
	if lineRendered != "" {
		lines = append(lines, btmBarStyle.Width(w).Render(lineRendered))
	}
	if len(lines) == 0 {
		lines = []string{btmBarStyle.Width(w).Render("")}
	}
	return lines
}

func (m *model) keyBarLineCount(w int) int {
	l := len(m.wrapKeyBar(w))
	if l < 1 {
		return 1
	}
	return l
}

func (m model) View() string {
	w := m.width
	if w == 0 {
		w = 80
	}
	h := m.height
	if h == 0 {
		h = 24
	}

	// Compute key bar FIRST so we can derive an exact list height.
	keyBarLines := m.wrapKeyBar(w)
	// top bar (1) + status bar (1) + key bar lines
	listH := h - 2 - len(keyBarLines)
	if listH < 1 {
		listH = 1
	}

	// ── top bar ──────────────────────────────────────────────────────────────
	var modeTag string
	if m.moving {
		modeTag = "  [MOVE]"
	} else if m.editScreen && m.editIsNew {
		modeTag = "  [ADD]"
	} else if m.editScreen {
		modeTag = "  [EDIT]"
	} else if m.confirming {
		modeTag = "  [DELETE?]"
	}

	total := len(m.state.Tasks)
	done := 0
	paused := 0
	for _, t := range m.state.Tasks {
		if t.Done {
			done++
		}
		if t.Paused {
			paused++
		}
	}
	topRight := fmt.Sprintf("%d done %d paused / %d%s", done, paused, total, modeTag)
	topLeft := "taskigt"
	if m.buildVersion != "" {
		topLeft += " " + m.buildVersion
	}
	pad := w - lipgloss.Width(topLeft) - lipgloss.Width(topRight) - 2 // 2 for padding
	if pad < 1 {
		pad = 1
	}
	topBar := topBarStyle.Width(w).Render(
		topLeft + strings.Repeat(" ", pad) + topRight,
	)

	// ── body ─────────────────────────────────────────────────────────────────
	var body strings.Builder

	if m.aboutScreen {
		body.WriteString("\n")
		body.WriteString(selectedRowStyle.Render("  taskigt") + "\n")
		if m.buildVersion != "" {
			body.WriteString(hintStyle.Render("  "+m.buildVersion) + "\n")
		}
		body.WriteString("\n")
		body.WriteString("  Built by Karl Bernstål\n")
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("  The task manager that judges you for having too many tasks.") + "\n")
		body.WriteString(hintStyle.Render("  Taskigt: Swedish for \"mean\". It knows what it is.") + "\n")
		body.WriteString(hintStyle.Render("  No due dates were harmed in the making of this software.") + "\n")
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("  Data: "+m.storePath) + "\n")
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("  Press any key to close") + "\n")
	} else if m.confirming {
		title := ""
		if len(m.state.Tasks) > 0 {
			title = m.state.Tasks[m.cursor].Title
		}
		body.WriteString("\n")
		body.WriteString(errorStyle.Render(fmt.Sprintf("  Delete \"%s\"?", title)) + "\n")
	} else if m.editScreen {
		// ── edit screen (also used for new task) ─────────────────────────────────────────────
		type fieldRow struct{ label, value, hint string }
		dueDisplay := m.editDraft.DueRaw
		if dueDisplay == "" {
			dueDisplay = "(none)"
		}
		fields := []fieldRow{
			{"Title", m.editDraft.Title, "text"},
			{"Due date", dueDisplay, "YYYY-MM-DD or +N days, empty to clear"},
		}
		body.WriteString("\n")
		if m.editIsNew {
			body.WriteString(hintStyle.Render("  New task") + "\n")
		} else {
			body.WriteString(hintStyle.Render("  Edit task") + "\n")
		}
		body.WriteString("\n")
		for i, f := range fields {
			active := i == m.editFieldIdx
			label := fmt.Sprintf("  %-10s", f.label)
			var valueStr string
			if active && m.editFieldActive {
				valueStr = selectedRowStyle.Render(m.inputBefore() + "█" + m.inputAfter())
			} else if active {
				valueStr = selectedRowStyle.Render("▸ " + f.value)
			} else {
				valueStr = hintStyle.Render(f.value)
			}
			body.WriteString(label + valueStr + "\n")
			if active && !m.editFieldActive {
				body.WriteString(hintStyle.Render(fmt.Sprintf("            %s", f.hint)) + "\n")
			}
		}
	} else if len(m.state.Tasks) == 0 {
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("  No tasks yet. Press a to add one.") + "\n")
	} else {
		end := m.scroll + listH
		if end > len(m.state.Tasks) {
			end = len(m.state.Tasks)
		}
		for i := m.scroll; i < end; i++ {
			task := m.state.Tasks[i]
			pointer := "  "
			if i == m.cursor {
				pointer = "▌ "
			}
			mark := "⬜"
			if task.Done {
				mark = "✅"
			} else if task.Paused {
				mark = "⏸️ "
			}
			// due date suffix
			var dueSuffix string
			if task.DueAt != nil {
				now := time.Now()
				overdue := now.After(*task.DueAt) && !task.Done
				formatted := task.DueAt.Format("Jan 02")
				if overdue {
					dueSuffix = " " + overdueDateStyle.Render("(overdue "+formatted+")")
				} else {
					dueSuffix = " " + dueDateStyle.Render("("+formatted+")")
				}
			}
			row := fmt.Sprintf("%s%s  %s", pointer, mark, task.Title)
			// clip to terminal width so long titles don't break line-count
			if lipgloss.Width(row) > w {
				// truncate plain title to fit
				maxLen := w - 8 // generous margin for pointer+mark
				if maxLen < 4 {
					maxLen = 4
				}
				title := []rune(task.Title)
				if len(title) > maxLen {
					title = append(title[:maxLen-1], '…')
				}
				row = fmt.Sprintf("%s%s  %s", pointer, mark, string(title))
			}
			if task.Done && i != m.cursor {
				row = doneRowStyle.Render(row) + dueSuffix
			} else if task.Paused && i != m.cursor {
				row = pausedRowStyle.Render(row) + dueSuffix
			} else if i == m.cursor {
				row = selectedRowStyle.Render(row) + dueSuffix
			} else {
				row = row + dueSuffix
			}
			body.WriteString(row + "\n")
		}
	}

	// pad body to fill space above bottom bar
	bodyLines := strings.Count(body.String(), "\n")
	for i := bodyLines; i < listH; i++ {
		body.WriteString("\n")
	}

	// ── status bar ───────────────────────────────────────────────────────────
	var statusLine string
	if m.status != "" && !m.confirming && !m.editScreen {
		if m.lastErr != nil {
			statusLine = errorStyle.Render(" " + m.status)
		} else {
			statusLine = statusStyle.Render(" " + m.status)
		}
	}
	// always occupy exactly one line
	statusBar := statusLine + strings.Repeat(" ", max(0, w-lipgloss.Width(statusLine)))

	return topBar + "\n" + body.String() + statusBar + "\n" + strings.Join(keyBarLines, "\n")
}
