# taskigt

Minimal checklist manager in terminal with Bubble Tea.

## Keybindings

- `↑/↓`: navigate
- `m`: toggle move mode (then use `↑/↓` to move selected task up/down)
- `enter`: toggle done/undone
  - keeps task in current position
- `a`: add task
- `d`: delete selected task
- `K` / `J`: move selected task one step up/down
- `g` / `G`: move selected task to top/bottom
- `x`: archive all done tasks (press `x` again to undo last archive)
- `q` or `ctrl+c`: quit

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
