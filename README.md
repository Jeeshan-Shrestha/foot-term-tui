# foot-term (TUI)

## Build & run

```bash
cd foot-term
go mod tidy   # resolves golang.org/x/* normally on your machine
go build -o foot-term .
./foot-term
```

## Controls

**Menu**
- `↑`/`↓` or `j`/`k` — move
- `enter` / `space` — select
- `1`–`4` — quick-jump (matches your original script's menu numbers)
- `q` / `ctrl+c` — quit

**Any data page (Live / Schedule / History)**
- `esc` / `b` / `backspace` — back to menu
- `r` — refresh (re-fetches from the API)
- `q` / `ctrl+c` — quit

## Files
- `main.go` — the whole app (single file, ~450 lines)
- `go.mod` / `go.sum` — dependencies: `bubbletea`, `bubbles` (spinner), `lipgloss`
