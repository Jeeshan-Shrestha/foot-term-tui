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

> Note: `go.mod` was built inside a sandboxed environment that couldn't
> reach `golang.org` directly, so it temporarily used `replace` directives
> pointing `golang.org/x/sys|text|sync|term` at their GitHub mirrors. These
> are harmless but unnecessary on a normal machine — running `go mod tidy`
> will clean them up and pull the real modules.

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

## What changed vs. the bash version

- Each menu option is a real page with its own loading spinner, error
  state (with retry via `r`), and empty state, instead of printing once
  and returning to the prompt.
- Status is shown as a colored badge: 🔴 live, 🟢 finished (FT), 🔵 upcoming.
- Team IDs are resolved against `/get/teams` the same way `get_team_name`
  did, including the `"0" → "TBD"` special case.
- The **Schedule** page intentionally keeps your original script's actual
  filter (today's date), even though the header said "tomorrow" — I didn't
  want to silently change behavior. If you actually want it to show
  *tomorrow's* games, say the word and I'll flip the date used in
  `fetchScheduleCmd`.
- Date parsing tries several common layouts since the exact format
  `local_date` comes back in wasn't in your snippet; if kickoff times look
  wrong, tell me what a raw `local_date` value looks like and I'll pin the
  exact layout.

## Files

- `main.go` — the whole app (single file, ~450 lines)
- `go.mod` / `go.sum` — dependencies: `bubbletea`, `bubbles` (spinner), `lipgloss`
