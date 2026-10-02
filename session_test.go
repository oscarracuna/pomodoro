package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
)

func TestFinishFocusThenLongBreak(t *testing.T) {
	m := newModel()
	for range 3 {
		m.finishPhase()
		if m.phase != phaseShort {
			t.Fatalf("after focus, phase = %s", m.phase.label())
		}
		m.finishPhase()
		if m.phase != phaseFocus {
			t.Fatalf("after short break, phase = %s", m.phase.label())
		}
	}
	m.finishPhase()
	if m.phase != phaseLong || m.completed != 4 || m.pips != 4 || m.running {
		t.Fatalf("phase %s completed %d pips %d running %v", m.phase.label(), m.completed, m.pips, m.running)
	}
	if m.remaining != m.durations[phaseLong] {
		t.Fatalf("remaining = %s", m.remaining)
	}
	m.finishPhase()
	if m.phase != phaseFocus || m.pips != 0 {
		t.Fatalf("after long break, phase %s pips %d", m.phase.label(), m.pips)
	}
}

func TestSkipFocusDoesNotCount(t *testing.T) {
	m := newModel()
	m.skip()
	if m.phase != phaseShort || m.completed != 0 || m.pips != 0 {
		t.Fatalf("phase %s completed %d pips %d", m.phase.label(), m.completed, m.pips)
	}
	m.skip()
	if m.phase != phaseFocus {
		t.Fatalf("phase %s", m.phase.label())
	}
}

func TestAdjustClampsDuration(t *testing.T) {
	m := newModel()
	m.adjust(-time.Hour)
	if m.durations[phaseFocus] != minDuration || m.remaining != minDuration {
		t.Fatalf("duration %s remaining %s", m.durations[phaseFocus], m.remaining)
	}
	m.adjust(100 * time.Hour)
	if m.durations[phaseFocus] != maxDuration || m.remaining != maxDuration {
		t.Fatalf("duration %s remaining %s", m.durations[phaseFocus], m.remaining)
	}
}

func TestTickCompletesRunningFocus(t *testing.T) {
	m := newModel()
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	m.remaining = time.Second
	m.running = true
	m.lastTick = now

	next, cmd := m.Update(tickMsg(now.Add(time.Second)))
	got := next.(model)
	if got.phase != phaseShort || got.completed != 1 || got.running {
		t.Fatalf("phase %s completed %d running %v", got.phase.label(), got.completed, got.running)
	}
	if cmd == nil {
		t.Fatal("expected a follow-up command")
	}
}

func TestThemeCycleWraps(t *testing.T) {
	m := newModel()
	if m.theme().Name != "Mocha" {
		t.Fatalf("default theme %s", m.theme().Name)
	}
	m.cycleTheme(-1)
	if m.theme().Name != "Charm" {
		t.Fatalf("previous theme %s", m.theme().Name)
	}
	m.cycleTheme(1)
	if m.theme().Name != "Mocha" {
		t.Fatalf("next theme %s", m.theme().Name)
	}
}

func TestClockArtIsStable(t *testing.T) {
	on := clockArt("25:00", true)
	off := clockArt("25:00", false)
	if lipgloss.Width(on) != lipgloss.Width(off) {
		t.Fatalf("colon blink changes width: %d vs %d", lipgloss.Width(on), lipgloss.Width(off))
	}
	lines := strings.Split(on, "\n")
	if len(lines) != clockRows {
		t.Fatalf("rows = %d", len(lines))
	}
	width := lipgloss.Width(lines[0])
	for _, line := range lines {
		if lipgloss.Width(line) != width {
			t.Fatalf("uneven clock row %q", line)
		}
	}
	for _, glyph := range digitGlyphs {
		for _, row := range glyph {
			if len([]rune(row)) != 6 {
				t.Fatalf("digit row %q", row)
			}
		}
	}
}

func TestFocusCounterShadeStaysWithThePips(t *testing.T) {
	m := newModel()
	m.themeIdx = 3 // Gruvbox, the palette in the reported layout
	m.width, m.height = 80, 24
	m.running = true
	cells := paintCells(m.render())
	for y, row := range cells {
		for x, c := range row {
			if !c.set {
				t.Fatalf("background unset at %d,%d %q", x, y, string(c.r))
			}
		}
	}
	var counter, schedule []cell
	for _, row := range cells {
		text := rowText(row)
		switch {
		case strings.Contains(text, "focus done"):
			counter = row
		case strings.Contains(text, "focus 25"):
			schedule = row
		}
	}
	if counter == nil || schedule == nil {
		t.Fatal("missing counter or schedule row")
	}

	pip := -1
	for i, c := range counter {
		if c.r == '◉' {
			pip = i
			break
		}
	}
	if pip < 0 {
		t.Fatal("missing current focus pip")
	}
	shade := mustHex("#282828")
	surface := mustHex("#3C3836")
	if !sameRGB(counter[pip].bg, shade) {
		t.Fatalf("first pip background = %v, want the shade", counter[pip].bg)
	}

	start, end := pip, pip
	for start > 0 && sameRGB(counter[start-1].bg, shade) {
		start--
	}
	for end+1 < len(counter) && sameRGB(counter[end+1].bg, shade) {
		end++
	}
	if start == 0 || end == len(counter)-1 || pip <= start || pip >= end {
		t.Fatalf("pip %d shade run %d-%d of %d\n%s", pip, start, end, len(counter), describeRow(counter))
	}
	if !sameRGB(counter[start-1].bg, surface) || !sameRGB(counter[end+1].bg, surface) {
		t.Fatalf("chip is not inset on the card\n%s", describeRow(counter))
	}
	for _, c := range schedule {
		if c.r == ' ' || c.r == '\u00a0' {
			continue
		}
		if sameRGB(c.bg, shade) {
			t.Fatalf("schedule glyph %q sits on the shade\n%s", string(c.r), describeRow(schedule))
		}
	}
}

func describeRow(row []cell) string {
	var b strings.Builder
	var prev [3]uint8
	var have bool
	for _, c := range row {
		if !have || c.bg != prev {
			fmt.Fprintf(&b, "[%02x%02x%02x]", c.bg[0], c.bg[1], c.bg[2])
			prev = c.bg
			have = true
		}
		b.WriteRune(c.r)
	}
	return b.String()
}

func TestRenderIncludesTimer(t *testing.T) {
	m := newModel()
	m.width, m.height = 80, 24
	view := m.View()
	if !view.AltScreen {
		t.Fatal("expected alternate screen")
	}
	plain := stripANSI(view.Content)
	for _, want := range []string{"pomodoro", "Mocha", "F O C U S", "paused", "0 focus done", "focus 25", "test"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("render missing %q\n%s", want, plain)
		}
	}
}

func TestRenderDoesNotPanicAtSmallSizes(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {100, 40}, {40, 18}, {28, 12}} {
		m := newModel()
		m.width, m.height = sz[0], sz[1]
		if m.render() == "" {
			t.Fatalf("empty render at %v", sz)
		}
	}
}

type cell struct {
	r   rune
	bg  [3]uint8
	set bool
}

func paintCells(s string) [][]cell {
	var rows [][]cell
	var row []cell
	var bg [3]uint8
	set := false
	flush := func() {
		rows = append(rows, row)
		row = nil
	}
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && !((s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z')) {
				j++
			}
			if j < len(s) && s[j] == 'm' {
				bg, set = applySGR(s[i+2:j], bg, set)
			}
			if j < len(s) {
				i = j + 1
			} else {
				i = j
			}
			continue
		}
		if s[i] == '\n' {
			flush()
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		row = append(row, cell{r: r, bg: bg, set: set})
		i += size
	}
	if len(row) > 0 {
		flush()
	}
	return rows
}

func applySGR(seq string, bg [3]uint8, set bool) ([3]uint8, bool) {
	if seq == "" || seq == "0" {
		return bg, false
	}
	parts := strings.Split(seq, ";")
	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "0", "":
			set = false
		case "49":
			set = false
		case "38":
			i += skipColor(parts[i+1:])
		case "48":
			n := skipColor(parts[i+1:])
			if n == 4 && i+4 < len(parts) {
				bg = [3]uint8{parseByte(parts[i+2]), parseByte(parts[i+3]), parseByte(parts[i+4])}
				set = true
			} else if n > 0 {
				set = true
			}
			i += n
		}
	}
	return bg, set
}

func skipColor(parts []string) int {
	if len(parts) == 0 {
		return 0
	}
	switch parts[0] {
	case "2":
		if len(parts) >= 4 {
			return 4
		}
	case "5":
		if len(parts) >= 2 {
			return 2
		}
	}
	return 0
}

func parseByte(s string) uint8 {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return uint8(n)
}

func rowText(row []cell) string {
	var b strings.Builder
	for _, c := range row {
		b.WriteRune(c.r)
	}
	return b.String()
}

func mustHex(s string) [3]uint8 {
	var r, g, b uint8
	fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
	return [3]uint8{r, g, b}
}

func sameRGB(a, b [3]uint8) bool {
	return a == b
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != 0x1b {
			b.WriteByte(s[i])
			continue
		}
		if i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && !((s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z')) {
				i++
			}
		}
	}
	return b.String()
}
