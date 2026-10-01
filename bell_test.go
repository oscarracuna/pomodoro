package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestBellChimeDecays(t *testing.T) {
	pcm := bellPCM()
	if len(pcm) < bellSampleRate || len(pcm)%2 != 0 {
		t.Fatalf("pcm length %d", len(pcm))
	}
	samples := make([]float64, len(pcm)/2)
	var peak float64
	for i := range samples {
		v := float64(int16(binary.LittleEndian.Uint16(pcm[i*2:])))
		samples[i] = v
		if a := math.Abs(v); a > peak {
			peak = a
		}
	}
	if peak < 8000 || peak > 32000 {
		t.Fatalf("peak %v", peak)
	}
	window := bellSampleRate / 5
	head := rms(samples[:window])
	tail := rms(samples[len(samples)-window:])
	if head < tail*4 {
		t.Fatalf("head rms %v tail %v", head, tail)
	}
}

func TestBellPlaysTwice(t *testing.T) {
	pcm := bellPCM()
	half := len(pcm) / 2
	if half == 0 || len(pcm) != half*2 || !bytes.Equal(pcm[:half], pcm[half:]) {
		t.Fatal("bell should be the chime played twice back to back")
	}
}

func TestBellOnlyWhenEnabled(t *testing.T) {
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

	m := newModel()
	m.running = true
	m.remaining = time.Second
	m.lastTick = now
	if m.handleTick(now.Add(time.Second)) == nil {
		t.Fatal("expected a bell when the timer finishes")
	}

	m = newModel()
	m.bell = false
	m.running = true
	m.remaining = time.Second
	m.lastTick = now
	if cmd := m.handleTick(now.Add(time.Second)); cmd != nil {
		t.Fatal("disabled bell should stay quiet")
	}
}

func TestShiftBRingsBell(t *testing.T) {
	m := newModel()
	m.bell = false
	_, cmd := m.Update(tea.KeyPressMsg{Text: "B", Code: 'b', ShiftedCode: 'B'})
	if cmd == nil {
		t.Fatal("B should play the bell")
	}
}

func rms(samples []float64) float64 {
	var sum float64
	for _, s := range samples {
		sum += s * s
	}
	return math.Sqrt(sum / float64(len(samples)))
}
