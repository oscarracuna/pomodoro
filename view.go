package main

import (
	"fmt"
	"image/color"
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
		th.centerOnSurface(width, m.phaseLabel(th)),
		m.bar(th, width),
		spread(m.status(th), m.percent(th), width, th.surfaceFill),
	)
	if gap != "" {
		parts = append(parts, "")
	}
	parts = append(parts, m.pipLine(th, width))
	if !compact {
		parts = append(parts, th.centerOnSurface(width, m.schedule(th)))
	}
	parts = append(parts,
		th.ink(th.Track).Render(strings.Repeat("─", width)),
	)
	for _, line := range m.help(th, width) {
		parts = append(parts, th.centerOnSurface(width, line))
	}
	return parts
}

func (m model) header(th Theme, width int) string {
	mark := th.ink(th.phaseColor(m.phase)).Render("●")
	title := th.ink(th.Accent).Bold(true).Render("pomodoro")
	name := th.ink(th.Muted).Render(th.Name)
	return spread(mark+th.surfaceFill(1)+title, name, width, th.surfaceFill)
}

func (m model) clock(th Theme, width int) string {
	text := fmtClock(m.remaining)
	art := clockArt(text, m.colonOn || !m.running)
	if lipgloss.Width(art) > width {
		return th.centerOnSurface(width, th.ink(th.phaseColor(m.phase)).Bold(true).Render(text))
	}
	return th.centerOnSurface(width, th.ink(th.phaseColor(m.phase)).Render(art))
}

func (m model) phaseLabel(th Theme) string {
	return th.ink(th.phaseColor(m.phase)).Bold(true).Render(tracked(m.phase.label()))
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
			b.WriteString(th.ink(colors[i]).Render("█"))
		}
	}
	if rest := width - filled; rest > 0 {
		b.WriteString(th.ink(th.Track).Render(strings.Repeat("░", rest)))
	}
	return b.String()
}

func (m model) status(th Theme) string {
	label := m.statusLabel()
	style := th.ink(th.Muted)
	if m.notice != "" && !m.running {
		style = th.ink(th.Accent).Bold(true)
	}
	return style.Render(label)
}

func (m model) percent(th Theme) string {
	pct := int(math.Round(m.progress() * 100))
	return th.ink(th.Muted).Render(fmt.Sprintf("%d%%", pct))
}

func (m model) pipLine(th Theme, width int) string {
	filled := th.onShade(th.phaseColor(m.phase))
	current := th.onShade(th.Accent)
	empty := th.onShade(th.Track)
	gap := th.shadeFill(1)
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
	pips := strings.Join(parts, gap)
	sep := th.shadeFill(2)
	done := th.onShade(th.Muted)
	line := pips + sep + done.Render(doneLabel(m.completed))
	if lipgloss.Width(line)+2 > width {
		line = pips + sep + done.Render(shortDone(m.completed))
	}
	// One cell of shade on either side keeps the first pip inside the chip
	// instead of letting the page color run to the card edge and the next row.
	chip := line
	if lipgloss.Width(line)+2 <= width {
		chip = th.shadeFill(1) + line + th.shadeFill(1)
	}
	return th.centerOnSurface(width, chip)
}

func (m model) schedule(th Theme) string {
	names := []string{"focus", "short", "long"}
	parts := make([]string, len(names))
	for i, name := range names {
		text := fmt.Sprintf("%s %d", name, int(m.durations[i]/time.Minute))
		style := th.ink(th.Muted)
		if phase(i) == m.phase {
			style = th.ink(th.phaseColor(m.phase)).Bold(true)
		}
		parts[i] = style.Render(text)
	}
	sep := th.ink(th.Track).Render(" · ")
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
	sep := th.ink(th.Track).Render(" · ")
	sepW := lipgloss.Width(sep)
	sp := th.surfaceFill(1)
	var lines []string
	var cur []string
	curW := 0
	for _, item := range items {
		piece := th.ink(th.Accent).Bold(true).Render(item.key) +
			sp +
			th.ink(th.Muted).Render(item.label)
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

func spread(left, right string, width int, fill func(int) string) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return left + "\n" + right
	}
	return left + fill(gap) + right
}

// ink draws on the card. Foreground-only styles reset the background, which
// lets the darker page color show through the rest of the line.
func (t Theme) ink(c color.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c).Background(t.Surface)
}

// onShade draws on the focus-counter chip.
func (t Theme) onShade(c color.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c).Background(t.Bg)
}

func (t Theme) surfaceFill(n int) string {
	if n < 1 {
		return ""
	}
	return lipgloss.NewStyle().Background(t.Surface).Render(strings.Repeat(" ", n))
}

func (t Theme) shadeFill(n int) string {
	if n < 1 {
		return ""
	}
	return lipgloss.NewStyle().Background(t.Bg).Render(strings.Repeat(" ", n))
}

func (t Theme) centerOnSurface(width int, s string) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, s,
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(t.Surface)),
	)
}
