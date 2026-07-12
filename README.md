# foot-term (TUI)

A Bubble Tea port of your `foot-term` bash script — same data source
(`worldcup26.ir`), same three views (Live / Schedule / History), each now
its own scrollable page with a glossy dark rounded-border UI instead of
`echo`/`printf`.

## Build & run

```bash
cd foot-term
go mod tidy   # resolves golang.org/x/* normally on your machine
go build -o foot-term .
./foot-term
```

Requires Go 1.22+. No `jq`/`curl`/`bash date` needed — HTTP + JSON parsing
and date handling are done natively in Go.

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
