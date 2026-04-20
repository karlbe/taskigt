package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"taskigt/internal/i18n"
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

type langOption struct {
	code  string
	label string
	get   func() i18n.Strings
}

var availableLangs = []langOption{
	{"en", "English", i18n.English},
	{"sv", "Svenska", i18n.Swedish},
}

// editDraft holds unsaved field values while the edit screen is open.
type editDraft struct {
	Title  string
	DueRaw string // YYYY-MM-DD, +N, or "" to clear
}

type model struct {
	state          tasks.AppState
	store          *tasks.Store
	keys           keyMap
	str            i18n.Strings
	buildVersion   string
	storePath      string
	cursor         int
	scroll         int
	width          int
	height         int
	moving         bool
	confirming     bool
	aboutScreen    bool
	aboutLangIdx   int
	archiveMenu          bool
	archiveMenuIdx       int
	archivePage          bool
	archivePageCursor    int
	archivePageScroll    int
	archiveConfirmDelete bool
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

func NewModel(state tasks.AppState, store *tasks.Store, buildVersion string, storePath string, str i18n.Strings) tea.Model {
	langIdx := 0
	for i, l := range availableLangs {
		if l.code == state.Lang {
			langIdx = i
			break
		}
	}
	m := model{
		state:        state,
		store:        store,
		keys:         defaultKeyMap(str),
		str:          str,
		buildVersion: buildVersion,
		storePath:    storePath,
		aboutLangIdx: langIdx,
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
			return m.handleAboutScreen(msg)
		}
		if m.archiveConfirmDelete {
			return m.handleArchiveConfirmDelete(msg)
		}
		if m.archivePage {
			return m.handleArchivePage(msg)
		}
		if m.archiveMenu {
			return m.handleArchiveMenu(msg)
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
			m.status = m.str.StatusOrderSaved
			m.lastErr = nil
			m.save()
			break
		}
		if len(m.state.Tasks) == 0 {
			m.status = m.str.StatusNoTaskSelected
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
			m.status = m.str.StatusMoveCancelled
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
			m.status = m.str.StatusMoveCancelled
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
			m.status = m.str.StatusNoTaskSelected
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
		m.archiveMenu = true
		m.archiveMenuIdx = 0
		m.status = ""
		m.lastErr = nil
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
					m.status = m.str.StatusTitleEmpty
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
		m.status = m.str.StatusSaved
		m.lastErr = nil
		m.save()
	case key.Matches(msg, m.keys.Cancel): // esc
		if m.editDirty && !m.editDiscardPrompt {
			m.editDiscardPrompt = true
			m.status = m.str.StatusUnsavedChanges
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
			return fmt.Errorf(m.str.StatusInvalidDateFmt, raw)
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
			m.status = m.str.StatusTaskRemoved
			m.lastErr = nil
			m.undoSet = nil
			m.save()
		}
		m.confirming = false
	case key.Matches(msg, m.keys.Deny):
		m.confirming = false
		m.status = m.str.StatusDeleteCancelled
		m.lastErr = nil
	}
	return m, nil
}

func (m model) handleAboutScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.aboutLangIdx > 0 {
			m.aboutLangIdx--
			m.str = availableLangs[m.aboutLangIdx].get()
			m.keys = defaultKeyMap(m.str)
			m.state.Lang = availableLangs[m.aboutLangIdx].code
			m.save()
		}
	case key.Matches(msg, m.keys.Down):
		if m.aboutLangIdx < len(availableLangs)-1 {
			m.aboutLangIdx++
			m.str = availableLangs[m.aboutLangIdx].get()
			m.keys = defaultKeyMap(m.str)
			m.state.Lang = availableLangs[m.aboutLangIdx].code
			m.save()
		}
	default:
		m.aboutScreen = false
	}
	return m, nil
}

func (m model) Farewell() string {
	return m.str.FarewellMsg
}

func (m model) handleArchiveMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	const numItems = 4
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.archiveMenuIdx > 0 {
			m.archiveMenuIdx--
		}
	case key.Matches(msg, m.keys.Down):
		if m.archiveMenuIdx < numItems-1 {
			m.archiveMenuIdx++
		}
	case key.Matches(msg, m.keys.Toggle):
		switch m.archiveMenuIdx {
		case 0: // archive selected task (any state)
			if len(m.state.Tasks) == 0 {
				break
			}
			_, err := tasks.ArchiveTask(&m.state, m.cursor, time.Now())
			if err != nil {
				m.status = err.Error()
				m.lastErr = err
			} else {
				m.clampCursor()
				m.status = fmt.Sprintf(m.str.StatusArchivedFmt, 1)
				m.lastErr = nil
				m.save()
			}
			m.archiveMenu = false
		case 1: // archive all done tasks
			archived := tasks.ArchiveDoneTasks(&m.state, time.Now())
			if len(archived) == 0 {
				break
			}
			m.clampCursor()
			m.status = fmt.Sprintf(m.str.StatusArchivedFmt, len(archived))
			m.lastErr = nil
			m.save()
			m.archiveMenu = false
		case 2: // show archived page
			if len(m.state.Archived) == 0 {
				break
			}
			m.archiveMenu = false
			m.archivePage = true
			m.archivePageCursor = 0
			m.archivePageScroll = 0
		case 3: // delete all archived
			if len(m.state.Archived) == 0 {
				break
			}
			m.archiveConfirmDelete = true
		}
	case key.Matches(msg, m.keys.Cancel):
		m.archiveMenu = false
	}
	return m, nil
}

func (m model) handleArchivePage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.archivePageCursor > 0 {
			m.archivePageCursor--
			m.adjustArchiveScroll()
		}
	case key.Matches(msg, m.keys.Down):
		if m.archivePageCursor < len(m.state.Archived)-1 {
			m.archivePageCursor++
			m.adjustArchiveScroll()
		}
	case key.Matches(msg, m.keys.Toggle):
		if len(m.state.Archived) == 0 {
			break
		}
		if err := tasks.UnarchiveTask(&m.state, m.archivePageCursor, time.Now()); err != nil {
			m.status = err.Error()
			m.lastErr = err
		} else {
			if m.archivePageCursor >= len(m.state.Archived) && m.archivePageCursor > 0 {
				m.archivePageCursor--
			}
			m.adjustArchiveScroll()
			m.save()
			if len(m.state.Archived) == 0 {
				m.archivePage = false
			}
		}
	case key.Matches(msg, m.keys.Cancel):
		m.archivePage = false
	}
	return m, nil
}

func (m model) handleArchiveConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		tasks.DeleteAllArchived(&m.state)
		m.archiveConfirmDelete = false
		m.archiveMenu = false
		m.save()
	case key.Matches(msg, m.keys.Deny):
		m.archiveConfirmDelete = false
	}
	return m, nil
}

func (m *model) adjustArchiveScroll() {
	listH := m.listHeight()
	if listH <= 0 {
		return
	}
	if m.archivePageCursor < m.archivePageScroll {
		m.archivePageScroll = m.archivePageCursor
	}
	if m.archivePageCursor >= m.archivePageScroll+listH {
		m.archivePageScroll = m.archivePageCursor - listH + 1
	}
}

func (m *model) save() {
	if err := m.store.Save(m.state); err != nil {
		m.lastErr = err
		m.status = fmt.Sprintf(m.str.StatusSaveErrorFmt, err)
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
	s := m.str
	if m.aboutScreen {
		return [][2]string{{"↑↓", s.KeyNavigate}, {s.KeyAnyKey, s.KeyCancel}}
	}
	if m.archiveConfirmDelete {
		return [][2]string{{"y", s.KeyConfirm}, {"n/esc", s.KeyCancel}}
	}
	if m.archivePage {
		return [][2]string{{"↑↓", s.KeyNavigate}, {"enter", s.ArchivePageUnarchive}, {"esc", s.ArchivePageBack}}
	}
	if m.archiveMenu {
		return [][2]string{{"↑↓", s.KeyNavigate}, {"enter", s.KeyConfirm}, {"esc", s.KeyCancel}}
	}
	if m.confirming {
		return [][2]string{{"y", s.KeyConfirm}, {"n/esc", s.KeyCancel}}
	}
	if m.editScreen {
		if m.editFieldActive {
			return [][2]string{{"enter", s.KeyConfirmField}, {"esc", s.KeyRevertField}, {"←→", s.KeyMoveCaret}, {"⌫/del", s.KeyDeleteChar}}
		}
		return [][2]string{{"↑↓", s.KeyNavigate}, {"enter", s.KeyEditField}, {"w", s.KeySave}, {"esc", s.KeyDiscard}}
	}
	if m.moving {
		return [][2]string{{"↑↓", s.KeyReorder}, {"enter", s.KeyConfirm}, {"m/esc", s.KeyAbort}}
	}
	return [][2]string{
		{"a", s.KeyAdd},
		{"e", s.KeyEdit},
		{"p", s.KeyPause},
		{"m", s.KeyMove},
		{"x", s.KeyArchive},
		{"q", s.KeyQuit},
	}
}

// keyBarRightItem returns a right-anchored hint rendered for the current mode.
func (m *model) keyBarRightItem() (string, int) {
	if m.aboutScreen || m.confirming || m.editScreen || m.moving {
		return "", 0
	}
	rendered := btmBarKeyStyle.Render("?") + btmBarPieceStyle.Render(" "+m.str.KeyAbout)
	return rendered, lipgloss.Width(rendered)
}

// wrapKeyBar groups hint items into lines that each fit within w,
// returning each line pre-rendered. The last line has the about hint right-anchored.
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

	rightRendered, rightW := m.keyBarRightItem()

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
	// last line: right-anchor the about hint if there's room
	if rightRendered != "" {
		space := avail - lineW - rightW
		if space >= 1 {
			lineRendered += btmBarPieceStyle.Render(strings.Repeat(" ", space)) + rightRendered
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

	// Always use normal-mode key bar height for listH so total render height
	// stays constant across modes — prevents bubbletea leaving stale lines.
	normalM := m
	normalM.aboutScreen = false
	normalM.confirming = false
	normalM.editScreen = false
	normalM.moving = false
	normalM.archiveMenu = false
	normalM.archivePage = false
	normalM.archiveConfirmDelete = false
	baseKeyBarH := normalM.keyBarLineCount(w)

	listH := h - 2 - baseKeyBarH
	if listH < 1 {
		listH = 1
	}
	for len(keyBarLines) < baseKeyBarH {
		keyBarLines = append(keyBarLines, btmBarStyle.Width(w).Render(""))
	}

	// ── top bar ──────────────────────────────────────────────────────────────
	var modeTag string
	if m.moving {
		modeTag = m.str.ModeMove
	} else if m.editScreen && m.editIsNew {
		modeTag = m.str.ModeAdd
	} else if m.editScreen {
		modeTag = m.str.ModeEdit
	} else if m.confirming {
		modeTag = m.str.ModeDelete
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
	topRight := fmt.Sprintf(m.str.StatsFmt, done, paused, total, modeTag)
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
		body.WriteString("  " + m.str.AboutBuiltBy + "\n")
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("  "+m.str.AboutTagline) + "\n")
		body.WriteString(hintStyle.Render("  "+m.str.AboutMeaning) + "\n")
		body.WriteString(hintStyle.Render("  "+m.str.AboutNoDates) + "\n")
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("  "+m.str.AboutDataLabel+m.storePath) + "\n")
		body.WriteString("\n")
		for i, l := range availableLangs {
			if i == m.aboutLangIdx {
				body.WriteString(selectedRowStyle.Render("  ▸ "+l.label) + "\n")
			} else {
				body.WriteString(hintStyle.Render("    "+l.label) + "\n")
			}
		}
		body.WriteString("\n")
		body.WriteString(hintStyle.Render("  "+m.str.AboutDismiss) + "\n")
	} else if m.archiveConfirmDelete {
		body.WriteString("\n")
		body.WriteString(errorStyle.Render("  "+fmt.Sprintf(m.str.ArchiveDeleteConfirmFmt, len(m.state.Archived))) + "\n")
	} else if m.archivePage {
		body.WriteString("\n")
		body.WriteString(selectedRowStyle.Render("  "+fmt.Sprintf(m.str.ArchivePageTitleFmt, len(m.state.Archived))) + "\n")
		body.WriteString("\n")
		if len(m.state.Archived) == 0 {
			body.WriteString(hintStyle.Render("  "+m.str.ArchivePageEmpty) + "\n")
		} else {
			end := m.archivePageScroll + listH - 3 // 3 lines used by header above
			if end > len(m.state.Archived) {
				end = len(m.state.Archived)
			}
			showBelow := end < len(m.state.Archived)
			for i := m.archivePageScroll; i < end; i++ {
				t := m.state.Archived[i]
				var scrollCol string
				if i == m.archivePageScroll && m.archivePageScroll > 0 {
					scrollCol = hintStyle.Render("↑ ")
				} else if i == end-1 && showBelow {
					scrollCol = hintStyle.Render("↓ ")
				} else {
					scrollCol = "  "
				}
				pointer := "  "
				if i == m.archivePageCursor {
					pointer = "▌ "
				}
				row := fmt.Sprintf("%s%s  ✅  %s", pointer, "  ", t.Title)
				if i == m.archivePageCursor {
					row = scrollCol + selectedRowStyle.Render(row)
				} else {
					row = scrollCol + doneRowStyle.Render(row)
				}
				body.WriteString(row + "\n")
			}
		}
	} else if m.archiveMenu {
		body.WriteString("\n")
		body.WriteString(selectedRowStyle.Render("  "+m.str.ArchiveMenuTitle) + "\n")
		body.WriteString("\n")
		// item 0: archive selected task
		taskTitle := ""
		if len(m.state.Tasks) > 0 {
			taskTitle = m.state.Tasks[m.cursor].Title
		}
		doneCount := 0
		for _, t := range m.state.Tasks {
			if t.Done {
				doneCount++
			}
		}
		items := []struct {
			label   string
			enabled bool
		}{
			{fmt.Sprintf(m.str.ArchiveMenuTaskFmt, taskTitle), len(m.state.Tasks) > 0},
			{fmt.Sprintf(m.str.ArchiveMenuAllFmt, doneCount), doneCount > 0},
			{fmt.Sprintf(m.str.ArchiveMenuShowFmt, len(m.state.Archived)), len(m.state.Archived) > 0},
			{fmt.Sprintf(m.str.ArchiveMenuDeleteFmt, len(m.state.Archived)), len(m.state.Archived) > 0},
		}
		for i, item := range items {
			prefix := "    "
			if i == m.archiveMenuIdx {
				prefix = "  ▸ "
			}
			if i == m.archiveMenuIdx && item.enabled {
				body.WriteString(selectedRowStyle.Render(prefix+item.label) + "\n")
			} else if item.enabled {
				body.WriteString(prefix + item.label + "\n")
			} else {
				body.WriteString(hintStyle.Render(prefix+item.label) + "\n")
			}
		}
	} else if m.confirming {
		title := ""
		if len(m.state.Tasks) > 0 {
			title = m.state.Tasks[m.cursor].Title
		}
		body.WriteString("\n")
		body.WriteString(errorStyle.Render("  "+fmt.Sprintf(m.str.DeleteConfirmFmt, title)) + "\n")
	} else if m.editScreen {
		// ── edit screen (also used for new task) ─────────────────────────────────────────────
		type fieldRow struct{ label, value, hint string }
		dueDisplay := m.editDraft.DueRaw
		if dueDisplay == "" {
			dueDisplay = "(none)"
		}
		fields := []fieldRow{
			{m.str.FieldTitle, m.editDraft.Title, m.str.FieldHintTitle},
			{m.str.FieldDueDate, dueDisplay, m.str.FieldHintDue},
		}
		body.WriteString("\n")
		if m.editIsNew {
			body.WriteString(hintStyle.Render("  "+m.str.EditNewTask) + "\n")
		} else {
			body.WriteString(hintStyle.Render("  "+m.str.EditEditTask) + "\n")
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
		body.WriteString(hintStyle.Render("  "+m.str.EmptyList) + "\n")
	} else {
		showAbove := m.scroll > 0
		end := m.scroll + listH
		if end > len(m.state.Tasks) {
			end = len(m.state.Tasks)
		}
		showBelow := end < len(m.state.Tasks)
		for i := m.scroll; i < end; i++ {
			// scroll indicator column
			var scrollCol string
			if i == m.scroll && showAbove {
				scrollCol = hintStyle.Render("↑ ")
			} else if i == end-1 && showBelow {
				scrollCol = hintStyle.Render("↓ ")
			} else {
				scrollCol = "  "
			}
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
					dueSuffix = " " + overdueDateStyle.Render("("+fmt.Sprintf(m.str.DueOverdueFmt, formatted)+")")
				} else {
					dueSuffix = " " + dueDateStyle.Render("("+formatted+")")
				}
			}
			row := fmt.Sprintf("%s%s  %s", pointer, mark, task.Title)
			// clip to terminal width so long titles don't break line-count
			if lipgloss.Width(scrollCol)+lipgloss.Width(row) > w {
				maxLen := w - 10 // margin for scrollCol+pointer+mark
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
				row = scrollCol + doneRowStyle.Render(row) + dueSuffix
			} else if task.Paused && i != m.cursor {
				row = scrollCol + pausedRowStyle.Render(row) + dueSuffix
			} else if i == m.cursor {
				row = scrollCol + selectedRowStyle.Render(row) + dueSuffix
			} else {
				row = scrollCol + row + dueSuffix
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
