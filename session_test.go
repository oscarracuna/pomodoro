package main

import (
	"strings"
	"testing"
	"time"

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
