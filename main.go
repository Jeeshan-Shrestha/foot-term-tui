package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*f = flexString(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err == nil {
		*f = flexString(n.String())
		return nil
	}
	*f = ""
	return nil
}

type apiGame struct {
	HomeTeamID  flexString `json:"home_team_id"`
	AwayTeamID  flexString `json:"away_team_id"`
	HomeScore   flexString `json:"home_score"`
	AwayScore   flexString `json:"away_score"`
	LocalDate   string     `json:"local_date"`
	TimeElapsed string     `json:"time_elapsed"`
	Finished    flexString `json:"finished"`
}

type gamesResp struct {
	Games []apiGame `json:"games"`
}

type apiTeam struct {
	ID     flexString `json:"id"`
	NameEn string     `json:"name_en"`
}

type teamsResp struct {
	Teams []apiTeam `json:"teams"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func fetchJSON(url string, v interface{}) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func fetchAll() (gamesResp, teamsResp, error) {
	var g gamesResp
	var t teamsResp
	if err := fetchJSON("https://worldcup26.ir/get/games", &g); err != nil {
		return g, t, fmt.Errorf("fetching games: %w", err)
	}
	if err := fetchJSON("https://worldcup26.ir/get/teams", &t); err != nil {
		return g, t, fmt.Errorf("fetching teams: %w", err)
	}
	return g, t, nil
}

func teamNameMap(t teamsResp) map[string]string {
	m := map[string]string{"0": "TBD"}
	for _, tm := range t.Teams {
		m[string(tm.ID)] = tm.NameEn
	}
	return m
}
func todayStr() string     { return time.Now().Format("01/02/2006") }
func yesterdayStr() string { return time.Now().AddDate(0, 0, -1).Format("01/02/2006") }

var dateLayouts = []string{
	"01/02/2006 15:04:05",
	"01/02/2006 15:04",
	"01/02/2006",
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

func parseLocalDate(s string) (time.Time, bool) {
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
type statusKind int

const (
	statusUpcoming statusKind = iota
	statusLive
	statusFinished
)

type row struct {
	home, away           string
	scoreHome, scoreAway string
	timeText             string
	statusText           string
	kind                 statusKind
}

func buildRow(gm apiGame, names map[string]string) row {
	home := names[string(gm.HomeTeamID)]
	if home == "" {
		home = "Unknown"
	}
	away := names[string(gm.AwayTeamID)]
	if away == "" {
		away = "Unknown"
	}

	timeText := gm.LocalDate
	if tm, ok := parseLocalDate(gm.LocalDate); ok {
		timeText = tm.Format("Jan 02, 15:04")
	}

	statusText := gm.TimeElapsed
	lm := strings.ToLower(statusText)
	kind := statusUpcoming
	switch {
	case strings.Contains(lm, "finish") || strings.Contains(strings.ToUpper(string(gm.Finished)), "TRUE"):
		kind = statusFinished
		if statusText == "" {
			statusText = "FT"
		}
	case lm == "" || lm == "notstarted":
		kind = statusUpcoming
		if statusText == "" {
			statusText = "Upcoming"
		}
	default:
		kind = statusLive
	}

	hs := string(gm.HomeScore)
	as := string(gm.AwayScore)
	if hs == "" {
		hs = "-"
	}
	if as == "" {
		as = "-"
	}

	return row{
		home: home, away: away,
		scoreHome: hs, scoreAway: as,
		timeText: timeText, statusText: statusText, kind: kind,
	}
}
type liveMsg struct {
	rows []row
	err  error
}
type scheduleMsg struct {
	rows []row
	err  error
}
type historyMsg struct {
	rows []row
	err  error
}

func fetchLiveCmd() tea.Msg {
	g, t, err := fetchAll()
	if err != nil {
		return liveMsg{err: err}
	}
	names := teamNameMap(t)
	today := todayStr()
	var rows []row
	for _, gm := range g.Games {
		if !strings.HasPrefix(gm.LocalDate, today) {
			continue
		}
		if gm.TimeElapsed == "notstarted" || gm.TimeElapsed == "" {
			continue
		}
		rows = append(rows, buildRow(gm, names))
	}
	return liveMsg{rows: rows}
}

func fetchScheduleCmd() tea.Msg {
	g, t, err := fetchAll()
	if err != nil {
		return scheduleMsg{err: err}
	}
	names := teamNameMap(t)
	today := todayStr()
	var rows []row
	for _, gm := range g.Games {
		if !strings.HasPrefix(gm.LocalDate, today) {
			continue
		}
		rows = append(rows, buildRow(gm, names))
	}
	return scheduleMsg{rows: rows}
}

func fetchHistoryCmd() tea.Msg {
	g, t, err := fetchAll()
	if err != nil {
		return historyMsg{err: err}
	}
	names := teamNameMap(t)
	yesterday := yesterdayStr()
	var rows []row
	for _, gm := range g.Games {
		if !strings.HasPrefix(gm.LocalDate, yesterday) {
			continue
		}
		if !strings.Contains(strings.ToUpper(string(gm.Finished)), "TRUE") {
			continue
		}
		rows = append(rows, buildRow(gm, names))
	}
	return historyMsg{rows: rows}
}
var (
	colBg       = lipgloss.Color("#11121a")
	colBorder   = lipgloss.Color("#89b4fa")
	colAccent   = lipgloss.Color("#f9e2af")
	colGreen    = lipgloss.Color("#a6e3a1")
	colRed      = lipgloss.Color("#f38ba8")
	colBlue     = lipgloss.Color("#89dceb")
	colMuted    = lipgloss.Color("#7f849c")
	colText     = lipgloss.Color("#cdd6f4")
	colSelBg    = lipgloss.Color("#313244")
	colTitleFg  = lipgloss.Color("#1e1e2e")
	colTitleBg1 = lipgloss.Color("#94e2d5")

	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleBarStyle = lipgloss.NewStyle().
			Background(colTitleBg1).
			Foreground(colTitleFg).
			Bold(true).
			Padding(0, 2).
			MarginBottom(1)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorder).
			Padding(1, 2)

	menuItemStyle = lipgloss.NewStyle().
			Foreground(colText).
			Padding(0, 2)

	menuItemSelectedStyle = lipgloss.NewStyle().
				Foreground(colTitleFg).
				Background(colAccent).
				Bold(true).
				Padding(0, 2)

	headerRowStyle = lipgloss.NewStyle().
			Foreground(colMuted).
			Bold(true).
			Underline(true)

	teamStyle    = lipgloss.NewStyle().Foreground(colText).Bold(true)
	scoreStyle   = lipgloss.NewStyle().Foreground(colAccent).Bold(true)
	timeStyle    = lipgloss.NewStyle().Foreground(colMuted)
	footerStyle  = lipgloss.NewStyle().Foreground(colMuted).MarginTop(1)
	errStyle     = lipgloss.NewStyle().Foreground(colRed).Bold(true)
	emptyStyle   = lipgloss.NewStyle().Foreground(colMuted).Italic(true)
	spinnerStyle = lipgloss.NewStyle().Foreground(colAccent)

	badgeLive     = lipgloss.NewStyle().Background(colRed).Foreground(colTitleFg).Bold(true).Padding(0, 1)
	badgeFinished = lipgloss.NewStyle().Background(colGreen).Foreground(colTitleFg).Bold(true).Padding(0, 1)
	badgeUpcoming = lipgloss.NewStyle().Background(colBlue).Foreground(colTitleFg).Bold(true).Padding(0, 1)
)

func badge(k statusKind, text string) string {
	if text == "" {
		text = "—"
	}
	switch k {
	case statusLive:
		return badgeLive.Render("● " + strings.ToUpper(text))
	case statusFinished:
		return badgeFinished.Render("✓ FT")
	default:
		return badgeUpcoming.Render("◷ " + text)
	}
}


type view int

const (
	viewMenu view = iota
	viewLive
	viewSchedule
	viewHistory
)

type model struct {
	view    view
	cursor  int
	choices []string

	spinner spinner.Model
	loading bool
	err     error

	liveRows     []row
	scheduleRows []row
	historyRows  []row

	width, height int
	quitting      bool
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle
	return model{
		view: viewMenu,
		choices: []string{
			"⚽  Live Scores",
			"📅  Today's Schedule",
			"📜  Yesterday's Results",
			"🚪  Exit",
		},
		spinner: s,
	}
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
		if m.view == viewMenu {
			return m.updateMenu(msg)
		}
		return m.updateSubpage(msg)

	case liveMsg:
		m.loading = false
		m.err = msg.err
		m.liveRows = msg.rows
		return m, nil

	case scheduleMsg:
		m.loading = false
		m.err = msg.err
		m.scheduleRows = msg.rows
		return m, nil

	case historyMsg:
		m.loading = false
		m.err = msg.err
		m.historyRows = msg.rows
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.choices)-1 {
			m.cursor++
		}
	case "1":
		m.cursor = 0
		return m.enterPage(viewLive)
	case "2":
		m.cursor = 1
		return m.enterPage(viewSchedule)
	case "3":
		m.cursor = 2
		return m.enterPage(viewHistory)
	case "4":
		m.quitting = true
		return m, tea.Quit
	case "enter", " ":
		switch m.cursor {
		case 0:
			return m.enterPage(viewLive)
		case 1:
			return m.enterPage(viewSchedule)
		case 2:
			return m.enterPage(viewHistory)
		case 3:
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) enterPage(v view) (tea.Model, tea.Cmd) {
	m.view = v
	m.loading = true
	m.err = nil
	var cmd tea.Cmd
	switch v {
	case viewLive:
		cmd = fetchLiveCmd
	case viewSchedule:
		cmd = fetchScheduleCmd
	case viewHistory:
		cmd = fetchHistoryCmd
	}
	return m, tea.Batch(cmd, m.spinner.Tick)
}

func (m model) updateSubpage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	case "esc", "b", "backspace":
		m.view = viewMenu
		m.err = nil
		return m, nil
	case "r":
		return m.enterPage(m.view)
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return lipgloss.NewStyle().Foreground(colAccent).Padding(1, 2).Render(
			"CLOSING PROGRAM\nBYE BYE C YA AGAIN! 👋\n")
	}

	var body string
	switch m.view {
	case viewMenu:
		body = m.renderMenu()
	case viewLive:
		body = m.renderPage("⚽ LIVE SCORES", m.liveRows, true)
	case viewSchedule:
		body = m.renderPage("📅 TODAY'S SCHEDULE", m.scheduleRows, false)
	case viewHistory:
		body = m.renderPage("📜 YESTERDAY'S RESULTS", m.historyRows, false)
	}

	body = appStyle.Render(body)
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
	}
	return body
}

func (m model) renderMenu() string {
	title := titleBarStyle.Render("⚽  FOOT-TERM — World Cup 2026 Live Terminal")

	var items strings.Builder
	for i, choice := range m.choices {
		cursor := "  "
		style := menuItemStyle
		if m.cursor == i {
			cursor = "▶ "
			style = menuItemSelectedStyle
		}
		items.WriteString(style.Render(cursor+choice) + "\n")
	}

	card := cardStyle.Render(items.String())
	footer := footerStyle.Render("↑/↓ or j/k move · enter select · 1-4 quick jump · q quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, card, footer)
}

func (m model) renderPage(title string, rows []row, showLiveHint bool) string {
	bar := titleBarStyle.Render(title)

	var content string
	switch {
	case m.loading:
		content = fmt.Sprintf("%s Fetching latest data from worldcup26.ir...", m.spinner.View())
	case m.err != nil:
		content = errStyle.Render("⚠ Error: "+m.err.Error()) + "\n" + emptyStyle.Render("press r to retry")
	case len(rows) == 0:
		content = emptyStyle.Render("No matches found.")
	default:
		content = renderTable(rows)
	}

	card := cardStyle.Render(content)

	hint := "esc/b back · r refresh · q quit"
	if showLiveHint {
		hint = "esc/b back · r refresh (scores update on refresh) · q quit"
	}
	footer := footerStyle.Render(hint)

	return lipgloss.JoinVertical(lipgloss.Left, bar, card, footer)
}

func renderTable(rows []row) string {
	const homeW, awayW = 24, 24

	var b strings.Builder
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(homeW).Align(lipgloss.Right).Render(headerRowStyle.Render("HOME")),
		lipgloss.NewStyle().Width(9).Align(lipgloss.Center).Render(headerRowStyle.Render("SCORE")),
		lipgloss.NewStyle().Width(awayW).Render(headerRowStyle.Render("AWAY")),
		"  ",
		headerRowStyle.Render("KICKOFF"),
		"   ",
		headerRowStyle.Render("STATUS"),
	)
	b.WriteString(header + "\n\n")

	for _, r := range rows {
		home := lipgloss.NewStyle().Width(homeW).Align(lipgloss.Right).Render(teamStyle.Render(truncate(r.home, homeW)))
		score := lipgloss.NewStyle().Width(9).Align(lipgloss.Center).Render(
			scoreStyle.Render(fmt.Sprintf("%s - %s", r.scoreHome, r.scoreAway)))
		away := lipgloss.NewStyle().Width(awayW).Render(teamStyle.Render(truncate(r.away, awayW)))
		kickoff := lipgloss.NewStyle().Width(16).Render(timeStyle.Render(r.timeText))
		status := badge(r.kind, r.statusText)

		line := lipgloss.JoinHorizontal(lipgloss.Top, home, score, away, "  ", kickoff, "  ", status)
		b.WriteString(line + "\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func truncate(s string, w int) string {
	if len(s) <= w {
		return s
	}
	if w <= 1 {
		return s[:w]
	}
	return s[:w-1] + "…"
}
func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running foot-term:", err)
	}
}

