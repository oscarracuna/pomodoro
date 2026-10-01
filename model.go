package main

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

type tickMsg time.Time

type model struct {
	phase     phase
	durations [3]time.Duration
	remaining time.Duration
	running   bool
	completed int
	pips      int
	bell      bool
	notice    string
	colonOn   bool
	lastTick  time.Time
	themeIdx  int
	width     int
	height    int
}

func newModel() model {
	return model{
		phase: phaseFocus,
		durations: [3]time.Duration{
			25 * time.Minute,
			5 * time.Minute,
			15 * time.Minute,
		},
		remaining: 25 * time.Minute,
		bell:      true,
		colonOn:   true,
		width:     80,
		height:    24,
	}
}

func (m model) theme() Theme {
	return themes[m.themeIdx]
}

func (m model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		extra := m.handleTick(time.Time(msg))
		return m, tea.Batch(tick(), extra)

	case bellFailedMsg:
		m.notice = "bell unavailable"
		return m, tea.Raw("\a")

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "space":
			m.toggle()
		case "s":
			m.skip()
		case "r":
			m.resetTimer()
		case "R", "shift+r":
			m.restart()
		case "t":
			m.cycleTheme(1)
		case "T", "shift+t":
			m.cycleTheme(-1)
		case "b":
			m.bell = !m.bell
		case "B", "shift+b":
			return m, ringBell()
		case "left", "h":
			m.setPhase(m.phase.step(-1))
		case "right", "l":
			m.setPhase(m.phase.step(1))
		case "+", "=", "up":
			m.adjust(time.Minute)
		case "-", "_", "down":
			m.adjust(-time.Minute)
		}
	}
	return m, nil
}

func (m *model) handleTick(now time.Time) tea.Cmd {
	if !m.running {
		m.colonOn = true
		return nil
	}
	if !m.lastTick.IsZero() {
		m.remaining -= now.Sub(m.lastTick)
	}
	m.lastTick = now
	m.colonOn = !m.colonOn
	if m.remaining > 0 {
		return nil
	}
	m.remaining = 0
	m.finishPhase()
	if m.bell {
		return ringBell()
	}
	return nil
}
