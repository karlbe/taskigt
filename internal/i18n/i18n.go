package i18n

import (
	"os"
	"strings"
)

type Strings struct {
	// mode tags
	ModeMove   string
	ModeAdd    string
	ModeEdit   string
	ModeDelete string

	// top bar stats — printf format, args: done, paused, total int, modeTag string
	StatsFmt string

	// edit screen
	EditNewTask    string
	EditEditTask   string
	FieldTitle     string
	FieldDueDate   string
	FieldHintTitle string
	FieldHintDue   string

	// empty list
	EmptyList string

	// delete confirm — printf format, arg: title string
	DeleteConfirmFmt string

	// due date suffixes — printf format, arg: formatted date string
	DueOverdueFmt string

	// about screen
	AboutBuiltBy   string
	AboutTagline   string
	AboutMeaning   string
	AboutNoDates   string
	AboutDataLabel string
	AboutDismiss   string

	// key bar action labels
	KeyConfirm      string
	KeyCancel       string
	KeyNavigate     string
	KeyEditField    string
	KeySave         string
	KeyDiscard      string
	KeyReorder      string
	KeyAbort        string
	KeyConfirmField string
	KeyRevertField  string
	KeyMoveCaret    string
	KeyDeleteChar   string
	KeyToggle       string
	KeyAdd          string
	KeyEdit         string
	KeyDelete       string
	KeyPause        string
	KeyMove         string
	KeyArchive      string
	KeyAbout        string
	KeyAnyKey       string
	KeyQuit         string

	// key binding help text
	HelpUp        string
	HelpDown      string
	HelpLeft      string
	HelpRight     string
	HelpToggle    string
	HelpAdd       string
	HelpDelete    string
	HelpMoveMode  string
	HelpPause     string
	HelpEditTitle string
	HelpSetDue    string
	HelpWrite     string
	HelpArchive   string
	HelpQuit      string
	HelpCancel    string
	HelpBackspace string
	HelpAbout     string

	// archive menu
	ArchiveMenuTitle        string
	ArchiveMenuTaskFmt      string // arg: title
	ArchiveMenuAllFmt       string // arg: count of done tasks
	ArchiveMenuShowFmt      string // arg: count
	ArchiveMenuDeleteFmt    string // arg: count
	ArchiveDeleteConfirmFmt string // arg: count

	// archive page
	ArchivePageTitleFmt  string // arg: count
	ArchivePageEmpty     string
	ArchivePageUnarchive string
	ArchivePageBack      string

	// farewell
	FarewellMsg string

	// status messages
	StatusOrderSaved      string
	StatusNoTaskSelected  string
	StatusMoveCancelled   string
	StatusTitleEmpty      string
	StatusSaved           string
	StatusUnsavedChanges  string
	StatusTaskRemoved     string
	StatusDeleteCancelled string
	StatusSaveErrorFmt    string // arg: error
	StatusRestoredFmt     string // arg: count
	StatusArchivedFmt     string // arg: count
	StatusNoDoneTasks     string
	StatusInvalidDateFmt  string // arg: raw input
}

func English() Strings {
	return Strings{
		ModeMove:   "  [MOVE]",
		ModeAdd:    "  [ADD]",
		ModeEdit:   "  [EDIT]",
		ModeDelete: "  [DELETE?]",

		StatsFmt: "%d done %d paused / %d%s",

		EditNewTask:    "New task",
		EditEditTask:   "Edit task",
		FieldTitle:     "Title",
		FieldDueDate:   "Due date",
		FieldHintTitle: "text",
		FieldHintDue:   "YYYY-MM-DD or +N days, empty to clear",

		EmptyList: "No tasks yet. Press a to add one.",

		DeleteConfirmFmt: `Delete "%s"?`,

		DueOverdueFmt: "overdue %s",

		AboutBuiltBy:   "Built by Karl Bernstål and Claude Sonnet 4.5 - 2026",
		AboutTagline:   "The task manager that judges you for having too many tasks.",
		AboutMeaning:   `Taskigt: Swedish for "mean". It knows what it is.`,
		AboutNoDates:   "No due dates were harmed in the making of this software.",
		AboutDataLabel: "Data file: ",
		AboutDismiss:   "Press any key to save and close",

		ArchiveMenuTitle:        "Archive",
		ArchiveMenuTaskFmt:      `Archive "%s"`,
		ArchiveMenuAllFmt:       "Archive %d done tasks",
		ArchiveMenuShowFmt:      "Show %d archived",
		ArchiveMenuDeleteFmt:    "Delete %d archived",
		ArchiveDeleteConfirmFmt: "Permanently delete all %d archived tasks?",
		ArchivePageTitleFmt:     "Archived (%d)",
		ArchivePageEmpty:        "No archived tasks.",
		ArchivePageUnarchive:    "unarchive",
		ArchivePageBack:         "back",

		FarewellMsg:    "Tack för att du är så taskig! /Karl",

		KeyConfirm:      "confirm",
		KeyCancel:       "cancel",
		KeyNavigate:     "navigate",
		KeyEditField:    "edit field",
		KeySave:         "save",
		KeyDiscard:      "discard",
		KeyReorder:      "reorder",
		KeyAbort:        "abort",
		KeyConfirmField: "confirm field",
		KeyRevertField:  "revert field",
		KeyMoveCaret:    "move caret",
		KeyDeleteChar:   "delete char",
		KeyToggle:       "toggle",
		KeyAdd:          "add",
		KeyEdit:         "edit",
		KeyDelete:       "delete",
		KeyPause:        "pause",
		KeyMove:         "move",
		KeyArchive:      "archive",
		KeyAbout:        "about",
		KeyAnyKey:       "any key",
		KeyQuit:         "quit",

		HelpUp:        "up",
		HelpDown:      "down",
		HelpLeft:      "caret left",
		HelpRight:     "caret right",
		HelpToggle:    "toggle done",
		HelpAdd:       "add task",
		HelpDelete:    "delete task",
		HelpMoveMode:  "move mode",
		HelpPause:     "pause",
		HelpEditTitle: "edit title",
		HelpSetDue:    "set due date",
		HelpWrite:     "write/save",
		HelpArchive:   "archive/undo",
		HelpQuit:      "quit",
		HelpCancel:    "cancel",
		HelpBackspace: "backspace",
		HelpAbout:     "about",

		StatusOrderSaved:      "order saved",
		StatusNoTaskSelected:  "no task selected",
		StatusMoveCancelled:   "move cancelled",
		StatusTitleEmpty:      "title cannot be empty",
		StatusSaved:           "saved",
		StatusUnsavedChanges:  "unsaved changes — press esc again to discard, w to save",
		StatusTaskRemoved:     "task removed",
		StatusDeleteCancelled: "delete cancelled",
		StatusSaveErrorFmt:    "save error: %v",
		StatusRestoredFmt:     "restored %d task(s)",
		StatusArchivedFmt:     "archived %d task(s) (press x to undo)",
		StatusNoDoneTasks:     "no done tasks to archive",
		StatusInvalidDateFmt:  `invalid date "%s" — use YYYY-MM-DD or +N days`,
	}
}

func Swedish() Strings {
	return Strings{
		ModeMove:   "  [FLYTTA]",
		ModeAdd:    "  [LÄGG TILL]",
		ModeEdit:   "  [REDIGERA]",
		ModeDelete: "  [TA BORT?]",

		StatsFmt: "%d klara %d pausade / %d%s",

		EditNewTask:    "Ny uppgift",
		EditEditTask:   "Redigera uppgift",
		FieldTitle:     "Titel",
		FieldDueDate:   "Förfaller",
		FieldHintTitle: "",
		FieldHintDue:   "ÅÅÅÅ-MM-DD eller +N dagar, tom för att rensa",

		EmptyList: "Inga uppgifter än. Tryck a för att lägga till.",

		DeleteConfirmFmt: `Ta bort "%s"?`,

		DueOverdueFmt: "försenad %s",

		AboutBuiltBy:   "Byggt av Karl Bernstål och Claude Sonnet 4.5 - 2026",
		AboutTagline:   "Uppgiftshanteraren som dömer dig för att ha för många uppgifter.",
		AboutMeaning:   `Taskigt: Om du läser det här så kan du svenska.`,
		AboutNoDates:   "Inga förfallodatum skadades i skapandet av denna programvara.",
		AboutDataLabel: "Datafil: ",
		AboutDismiss:   "Tryck på valfri tangent för att spara och stänga",

		ArchiveMenuTitle:        "Arkivera",
		ArchiveMenuTaskFmt:      `Arkivera "%s"`,
		ArchiveMenuAllFmt:       "Arkivera %d klara uppgifter",
		ArchiveMenuShowFmt:      "Visa %d arkiverade",
		ArchiveMenuDeleteFmt:    "Ta bort %d arkiverade",
		ArchiveDeleteConfirmFmt: "Ta bort alla %d arkiverade uppgifter permanent?",
		ArchivePageTitleFmt:     "Arkiverade (%d)",
		ArchivePageEmpty:        "Inga arkiverade uppgifter.",
		ArchivePageUnarchive:    "återställ",
		ArchivePageBack:         "tillbaka",

		FarewellMsg:    "Tack för att du är så taskig! /Karl",

		KeyConfirm:      "bekräfta",
		KeyCancel:       "avbryt",
		KeyNavigate:     "navigera",
		KeyEditField:    "redigera fält",
		KeySave:         "spara",
		KeyDiscard:      "ångra",
		KeyReorder:      "ordna om",
		KeyAbort:        "avbryt",
		KeyConfirmField: "bekräfta fält",
		KeyRevertField:  "återställ fält",
		KeyMoveCaret:    "flytta markör",
		KeyDeleteChar:   "ta bort tecken",
		KeyToggle:       "växla",
		KeyAdd:          "lägg till",
		KeyEdit:         "redigera",
		KeyDelete:       "ta bort",
		KeyPause:        "pausa",
		KeyMove:         "flytta",
		KeyArchive:      "arkivera",
		KeyAbout:        "om",
		KeyAnyKey:       "valfri tangent",
		KeyQuit:         "avsluta",

		HelpUp:        "upp",
		HelpDown:      "ner",
		HelpLeft:      "markör vänster",
		HelpRight:     "markör höger",
		HelpToggle:    "växla klar",
		HelpAdd:       "lägg till uppgift",
		HelpDelete:    "ta bort uppgift",
		HelpMoveMode:  "flyttläge",
		HelpPause:     "pausa",
		HelpEditTitle: "redigera titel",
		HelpSetDue:    "sätt förfallodatum",
		HelpWrite:     "skriv/spara",
		HelpArchive:   "arkivera/ångra",
		HelpQuit:      "avsluta",
		HelpCancel:    "avbryt",
		HelpBackspace: "backsteg",
		HelpAbout:     "om",

		StatusOrderSaved:      "ordning sparad",
		StatusNoTaskSelected:  "ingen uppgift vald",
		StatusMoveCancelled:   "flytt avbruten",
		StatusTitleEmpty:      "titeln kan inte vara tom",
		StatusSaved:           "sparad",
		StatusUnsavedChanges:  "osparade ändringar — tryck esc igen för att ångra, w för att spara",
		StatusTaskRemoved:     "uppgift borttagen",
		StatusDeleteCancelled: "borttagning avbruten",
		StatusSaveErrorFmt:    "sparfel: %v",
		StatusRestoredFmt:     "återställde %d uppgift(er)",
		StatusArchivedFmt:     "arkiverade %d uppgift(er) (tryck x för att ångra)",
		StatusNoDoneTasks:     "inga klara uppgifter att arkivera",
		StatusInvalidDateFmt:  `ogiltigt datum "%s" — använd ÅÅÅÅ-MM-DD eller +N dagar`,
	}
}

// ForCode returns Strings for a language code ("en", "sv"). Empty/unknown → English.
func ForCode(code string) Strings {
	switch strings.ToLower(code) {
	case "sv":
		return Swedish()
	default:
		return English()
	}
}

// Detect picks a language from TASKIGT_LANG or LANG env vars, then English.
func Detect() Strings {
	lang := os.Getenv("TASKIGT_LANG")
	if lang == "" {
		lang = os.Getenv("LANG")
	}
	return ForCode(lang)
}
