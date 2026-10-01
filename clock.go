package main

import (
	"strings"
	"time"
)

const clockRows = 5

// Six-wide digits. Every row of a glyph is the same width so the clock
// stays put when the colon blinks.
var digitGlyphs = [10][clockRows]string{
	{" ████ ", "██  ██", "██  ██", "██  ██", " ████ "},
	{"  ██  ", " ███  ", "  ██  ", "  ██  ", "██████"},
	{"██████", "    ██", "██████", "██    ", "██████"},
	{"██████", "    ██", " █████", "    ██", "██████"},
	{"██  ██", "██  ██", "██████", "    ██", "    ██"},
	{"██████", "██    ", "██████", "    ██", "██████"},
	{"██████", "██    ", "██████", "██  ██", "██████"},
	{"██████", "    ██", "   ██ ", "  ██  ", " ██   "},
	{" ████ ", "██  ██", " ████ ", "██  ██", " ████ "},
	{" ████ ", "██  ██", " █████", "    ██", " ████ "},
}

var (
	colonGlyph = [clockRows]string{"  ", "██", "  ", "██", "  "}
	colonBlank = [clockRows]string{"  ", "  ", "  ", "  ", "  "}
)

func clockArt(text string, colonOn bool) string {
	var rows [clockRows]strings.Builder
	var prev rune
	first := true
	for _, r := range text {
		glyph, ok := glyphFor(r, colonOn)
		if !ok {
			continue
		}
		if !first {
			gap := " "
			if r == ':' || prev == ':' {
				gap = "  "
			}
			for i := range rows {
				rows[i].WriteString(gap)
			}
		}
		first = false
		prev = r
		for i := range rows {
			rows[i].WriteString(glyph[i])
		}
	}
	lines := make([]string, clockRows)
	for i := range rows {
		lines[i] = rows[i].String()
	}
	return strings.Join(lines, "\n")
}

func glyphFor(r rune, colonOn bool) ([clockRows]string, bool) {
	switch {
	case r >= '0' && r <= '9':
		return digitGlyphs[r-'0'], true
	case r == ':':
		if colonOn {
			return colonGlyph, true
		}
		return colonBlank, true
	default:
		return [clockRows]string{}, false
	}
}

func fmtClock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Round(time.Second) / time.Second)
	return sprintf02(total/60) + ":" + sprintf02(total%60)
}

func sprintf02(n int) string {
	if n < 0 {
		n = 0
	}
	if n > 99 {
		n = 99
	}
	return string([]byte{byte('0' + n/10), byte('0' + n%10)})
}
