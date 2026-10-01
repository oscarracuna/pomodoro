package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const preferredContentWidth = 52

func (m model) View() tea.View {
	th := m.theme()
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.WindowTitle = "Pomodoro · " + m.phase.title()
	v.BackgroundColor = th.Bg
	v.ForegroundColor = th.Fg
	pct := int(math.Round(m.progress() * 100))
	v.ProgressBar = tea.NewProgressBar(tea.ProgressBarDefault, pct)
	return v
}

func (m model) render() string {
	th := m.theme()
	compact := m.height < 22
	padV, padH := 1, 3
	if compact {
		padV, padH = 0, 2
	}
	// Lip Gloss Width includes padding and border, and wraps inside both.
	// inner is the text column that actually fits in the card.
	inner := m.innerWidth(padH)

	phaseColor := th.phaseColor(m.phase)
	body := lipgloss.JoinVertical(lipgloss.Left, m.sections(th, inner, compact)...)
	card := lipgloss.NewStyle().
		Width(inner+padH*2+2).
		Padding(padV, padH).
		Background(th.Surface).
		Foreground(th.Fg).
		Border(lipgloss.RoundedBorder()).
		BorderBackground(th.Surface).
		BorderForegroundBlend(phaseColor, th.Accent, phaseColor).
		Render(body)

	if m.width <= 0 || m.height <= 0 {
		return card
	}
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		card,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(th.Bg)),
	)
}

func (m model) innerWidth(padH int) int {
	available := m.width - padH*2 - 2
	if available > preferredContentWidth {
		available = preferredContentWidth
	}
	if available < 20 {
		available = 20
	}
	return available
}

func (m model) sections(th Theme, width int, compact bool) []string {
	gap := " "
	if compact {
		gap = ""
	}
	parts := []string{
		m.header(th, width),
	}
	if gap != "" {
		parts = append(parts, "")
	}
	parts = append(parts,
		m.clock(th, width),
		lipgloss.PlaceHorizontal(width, lipgloss.Center, m.phaseLabel(th)),
		m.bar(th, width),
		spread(m.status(th), m.percent(th), width),
	)
	if gap != "" {
		parts = append(parts, "")
	}
	parts = append(parts, lipgloss.PlaceHorizontal(width, lipgloss.Center, m.pipLine(th, width)))
	if !compact {
		parts = append(parts, lipgloss.PlaceHorizontal(width, lipgloss.Center, m.schedule(th)))
	}
	parts = append(parts,
		lipgloss.NewStyle().Foreground(th.Track).Render(strings.Repeat("─", width)),
	)
	for _, line := range m.help(th, width) {
		parts = append(parts, lipgloss.PlaceHorizontal(width, lipgloss.Center, line))
	}
	return parts
}

func (m model) header(th Theme, width int) string {
	mark := lipgloss.NewStyle().Foreground(th.phaseColor(m.phase)).Render("●")
	title := lipgloss.NewStyle().Foreground(th.Accent).Bold(true).Render("pomodoro")
	name := lipgloss.NewStyle().Foreground(th.Muted).Render(th.Name)
	return spread(mark+" "+title, name, width)
}

func (m model) clock(th Theme, width int) string {
	text := fmtClock(m.remaining)
	art := clockArt(text, m.colonOn || !m.running)
	if lipgloss.Width(art) > width {
		return lipgloss.PlaceHorizontal(width, lipgloss.Center,
			lipgloss.NewStyle().Foreground(th.phaseColor(m.phase)).Bold(true).Render(text),
		)
	}
	styled := lipgloss.NewStyle().Foreground(th.phaseColor(m.phase)).Render(art)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, styled)
}

func (m model) phaseLabel(th Theme) string {
	return lipgloss.NewStyle().
		Foreground(th.phaseColor(m.phase)).
		Bold(true).
		Render(tracked(m.phase.label()))
}

func (m model) bar(th Theme, width int) string {
	filled := int(math.Round(m.progress() * float64(width)))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	var b strings.Builder
	if filled > 0 {
		colors := lipgloss.Blend1D(filled, th.phaseColor(m.phase), lipgloss.Lighten(th.phaseColor(m.phase), 0.45))
		for i := range filled {
			b.WriteString(lipgloss.NewStyle().Foreground(colors[i]).Render("█"))
		}
	}
	if rest := width - filled; rest > 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(th.Track).Render(strings.Repeat("░", rest)))
	}
	return b.String()
}

func (m model) status(th Theme) string {
	label := m.statusLabel()
	style := lipgloss.NewStyle().Foreground(th.Muted)
	if m.notice != "" && !m.running {
		style = lipgloss.NewStyle().Foreground(th.Accent).Bold(true)
	}
	return style.Render(label)
}

func (m model) percent(th Theme) string {
	pct := int(math.Round(m.progress() * 100))
	return lipgloss.NewStyle().Foreground(th.Muted).Render(fmt.Sprintf("%d%%", pct))
}

func (m model) pipLine(th Theme, width int) string {
	filled := lipgloss.NewStyle().Foreground(th.phaseColor(m.phase))
	current := lipgloss.NewStyle().Foreground(th.Accent)
	empty := lipgloss.NewStyle().Foreground(th.Track)
	parts := make([]string, 0, cycleLength)
	for i := range cycleLength {
		switch {
		case i < m.pips:
			parts = append(parts, filled.Render("●"))
		case i == m.pips && m.phase == phaseFocus:
			parts = append(parts, current.Render("◉"))
		default:
			parts = append(parts, empty.Render("○"))
		}
	}
	sep := lipgloss.NewStyle().Foreground(th.Track).Render("  ")
	pips := strings.Join(parts, " ")
	done := lipgloss.NewStyle().Foreground(th.Muted)
	line := pips + sep + done.Render(doneLabel(m.completed))
	if lipgloss.Width(line) > width {
		line = pips + sep + done.Render(shortDone(m.completed))
	}
	return line
}

func (m model) schedule(th Theme) string {
	names := []string{"focus", "short", "long"}
	parts := make([]string, len(names))
	for i, name := range names {
		text := fmt.Sprintf("%s %d", name, int(m.durations[i]/time.Minute))
		style := lipgloss.NewStyle().Foreground(th.Muted)
		if phase(i) == m.phase {
			style = lipgloss.NewStyle().Foreground(th.phaseColor(m.phase)).Bold(true)
		}
		parts[i] = style.Render(text)
	}
	sep := lipgloss.NewStyle().Foreground(th.Track).Render(" · ")
	return strings.Join(parts, sep)
}

type hint struct {
	key   string
	label string
}

func (m model) help(th Theme, width int) []string {
	start := "start"
	if m.running {
		start = "pause"
	}
	bell := "bell on"
	if !m.bell {
		bell = "bell off"
	}
	items := []hint{
		{"space", start},
		{"s", "skip"},
		{"r", "reset"},
		{"t", "theme"},
		{"- +", "time"},
		{"←→", "phase"},
		{"b", bell},
		{"B", "test"},
		{"R", "restart"},
		{"q", "quit"},
	}
	return packHints(th, width, items)
}

func packHints(th Theme, width int, items []hint) []string {
	sep := lipgloss.NewStyle().Foreground(th.Track).Render(" · ")
	sepW := lipgloss.Width(sep)
	var lines []string
	var cur []string
	curW := 0
	for _, item := range items {
		piece := lipgloss.NewStyle().Foreground(th.Accent).Bold(true).Render(item.key) +
			" " +
			lipgloss.NewStyle().Foreground(th.Muted).Render(item.label)
		w := lipgloss.Width(piece)
		extra := w
		if len(cur) > 0 {
			extra += sepW
		}
		if len(cur) > 0 && curW+extra > width {
			lines = append(lines, strings.Join(cur, sep))
			cur = nil
			curW = 0
		}
		cur = append(cur, piece)
		if curW == 0 {
			curW = w
		} else {
			curW += sepW + w
		}
	}
	if len(cur) > 0 {
		lines = append(lines, strings.Join(cur, sep))
	}
	return lines
}

func doneLabel(n int) string {
	if n == 1 {
		return "1 focus done"
	}
	return fmt.Sprintf("%d focus done", n)
}

func shortDone(n int) string {
	if n == 1 {
		return "1 done"
	}
	return fmt.Sprintf("%d done", n)
}

func tracked(s string) string {
	runes := []rune(s)
	parts := make([]string, len(runes))
	for i, r := range runes {
		parts[i] = string(r)
	}
	return strings.Join(parts, " ")
}

func spread(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return left + "\n" + right
	}
	return left + strings.Repeat(" ", gap) + right
}
