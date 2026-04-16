# taskigt

Mean (taskig in swedish) minimal checklist manager in the terminal, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Screenshots


![Main screen](screenshots/main_screen.png)


## Keybindings

### Normal mode

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate |
| `enter` | Toggle done/undone |
| `p` | Pause/unpause task |
| `a` | Add task |
| `e` | Edit task |
| `D` | Set due date |
| `delete` | Delete selected task (confirms first) |
| `m` | Move mode — `↑/↓` reorder, `enter` confirm, `m`/`esc` cancel |
| `x` | Open archive menu |
| `?` | About / language selector |
| `q` / `ctrl+c` | Quit |

### Edit screen

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate fields |
| `enter` | Activate/confirm field |
| `←/→` | Move cursor within field |
| `backspace`/`delete` | Delete characters |
| `esc` | Revert field / discard changes |
| `w` | Save and exit |

### Confirmation dialog

| Key | Action |
|-----|--------|
| `y` | Confirm |
| `n` / `esc` | Cancel |

## Archive

Press `x` to open the archive menu. From there you can:

- Archive the currently selected task (any state)
- Archive all done tasks at once
- Open the archive page to browse and unarchive tasks
- Permanently delete all archived tasks

## Language

Press `?` to open the about dialog. Use `↑/↓` to switch between **English** and **Svenska**. The selection is saved automatically and restored on next launch.

## Command-line flags

The TUI launches by default. When stdout is not a terminal (pipe, redirect) it automatically falls back to `--list`.

| Flag | Description |
|------|-------------|
| `--list` | Print tasks as a human-readable list and exit |
| `--json` | Print full state as indented JSON and exit |
| `--data <path>` | Use a custom tasks JSON file instead of the default |

Examples:

```bash
taskigt --list
taskigt --json
taskigt --list | grep overdue
taskigt | cat          # auto-detects pipe, uses --list behavior
```

## Data

Tasks are stored as human-readable JSON. Location by platform:

| Platform | Path |
|----------|------|
| Windows | `%USERPROFILE%\.taskigt\tasks.json` |
| Linux / macOS | `~/.taskigt/tasks.json` |

Done tasks can be archived into the same file under the `archived` key.

## Build & run

```bash
make run        # run without building
make build      # build to bin/
make install    # build and install to Go bin dir (on PATH)
make test       # run tests
make clean      # remove bin/
```

Or without make:

```bash
go run ./cmd/taskigt
go build -o bin/taskigt ./cmd/taskigt
go install ./cmd/taskigt
```

## Windows build tools

The following tools are needed to build and develop on Windows.

**Go** (required):

```powershell
winget install GoLang.Go
```

**make** (required for `make` targets):

```powershell
winget install GnuWin32.Make
# or
choco install make
```

After installing, restart your terminal so the new tools are on `PATH`.
