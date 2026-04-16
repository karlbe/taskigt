# taskigt

Mean minimal checklist manager in terminal with Bubble Tea.

## Keybindings

### Normal mode

- `↑/↓`: navigate
- `enter`: toggle task done/undone
- `p`: pause/unpause task
- `a`: add task
- `e`: edit task title
- `D`: set due date
- `delete`: delete selected task (prompts for confirmation)
- `m`: enter move mode — use `↑/↓` to reorder, `enter` to confirm, `m`/`esc` to cancel
- `x`: archive all done tasks (press `x` again to undo last archive)
- `?`: about
- `q` or `ctrl+c`: quit

### Edit screen

- `↑/↓`: navigate fields
- `enter`: activate/confirm field
- `←/→`: move cursor within field
- `backspace`/`delete`: delete characters
- `esc`: revert field / discard changes
- `w`: save and exit

### Confirmation dialog

- `y`: confirm
- `n` or `esc`: cancel

## Persistence

Data is stored in `tasks.json` (human-readable, deterministic JSON).
Done tasks can be archived into the same file under `archived`.

## Run

```bash
go run ./cmd/taskigt
```

## Make targets

```bash
make run
make test
make build
make clean
```
