package main

import "time"

const (
	cycleLength = 4
	minDuration = time.Minute
	maxDuration = 90 * time.Minute
)

type phase int

const (
	phaseFocus phase = iota
	phaseShort
	phaseLong
)

func (p phase) label() string {
	switch p {
	case phaseShort:
		return "SHORT BREAK"
	case phaseLong:
		return "LONG BREAK"
	default:
		return "FOCUS"
	}
}

func (p phase) title() string {
	switch p {
	case phaseShort:
		return "Short break"
	case phaseLong:
		return "Long break"
	default:
		return "Focus"
	}
}

func (p phase) step(dir int) phase {
	n := int(phaseLong) + 1
	return phase((int(p) + dir%n + n) % n)
}

func (m *model) toggle() {
	if m.running {
		m.running = false
		m.lastTick = time.Time{}
		m.colonOn = true
		return
	}
	if m.remaining <= 0 {
		m.remaining = m.durations[m.phase]
	}
	if m.remaining <= 0 {
		return
	}
	m.running = true
	m.notice = ""
	m.lastTick = time.Now()
	m.colonOn = true
}

func (m *model) resetTimer() {
	m.remaining = m.durations[m.phase]
	m.running = false
	m.lastTick = time.Time{}
	m.colonOn = true
	m.notice = ""
}

func (m *model) restart() {
	m.phase = phaseFocus
	m.remaining = m.durations[phaseFocus]
	m.running = false
	m.lastTick = time.Time{}
	m.colonOn = true
	m.completed = 0
	m.pips = 0
	m.notice = "session restarted"
}

func (m *model) setPhase(p phase) {
	if p == m.phase {
		return
	}
	m.phase = p
	m.remaining = m.durations[p]
	m.running = false
	m.lastTick = time.Time{}
	m.colonOn = true
	m.notice = ""
}

// skip moves to the next phase without crediting a finished focus.
func (m *model) skip() {
	m.running = false
	m.lastTick = time.Time{}
	m.colonOn = true
	m.notice = ""
	if m.phase == phaseFocus {
		m.phase = phaseShort
	} else {
		if m.phase == phaseLong {
			m.pips = 0
		}
		m.phase = phaseFocus
	}
	m.remaining = m.durations[m.phase]
}

// finishPhase ends the current phase, credits a completed focus, and
// pauses on the following phase so the transition is easy to catch.
func (m *model) finishPhase() {
	m.running = false
	m.lastTick = time.Time{}
	m.colonOn = true
	switch m.phase {
	case phaseFocus:
		m.completed++
		m.pips++
		if m.pips >= cycleLength {
			m.phase = phaseLong
		} else {
			m.phase = phaseShort
		}
		m.notice = "focus complete"
	case phaseLong:
		m.pips = 0
		m.phase = phaseFocus
		m.notice = "back to focus"
	default:
		m.phase = phaseFocus
		m.notice = "back to focus"
	}
	m.remaining = m.durations[m.phase]
}

func (m *model) adjust(delta time.Duration) {
	next := m.durations[m.phase] + delta
	if next < minDuration {
		next = minDuration
	}
	if next > maxDuration {
		next = maxDuration
	}
	applied := next - m.durations[m.phase]
	m.durations[m.phase] = next
	m.remaining += applied
	if m.remaining < 0 {
		m.remaining = 0
	}
	if m.remaining > next {
		m.remaining = next
	}
}

func (m *model) cycleTheme(dir int) {
	n := len(themes)
	m.themeIdx = (m.themeIdx + dir) % n
	if m.themeIdx < 0 {
		m.themeIdx += n
	}
}

func (m model) progress() float64 {
	total := m.durations[m.phase]
	if total <= 0 {
		return 1
	}
	done := total - m.remaining
	if done < 0 {
		done = 0
	}
	if done > total {
		done = total
	}
	return float64(done) / float64(total)
}

func (m model) statusLabel() string {
	if !m.running && m.notice != "" {
		return m.notice
	}
	if m.running {
		if m.phase == phaseFocus {
			return "focusing"
		}
		return "resting"
	}
	return "paused"
}
