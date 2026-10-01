package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ebitengine/oto/v3"
)

const bellSampleRate = 44100

type bellFailedMsg struct{}

// ringBell plays the end-of-phase chime. A failure comes back as
// bellFailedMsg so the timer can fall back to the terminal bell.
func ringBell() tea.Cmd {
	return func() tea.Msg {
		if err := playBell(); err != nil {
			return bellFailedMsg{}
		}
		return nil
	}
}

var (
	bellPCMOnce sync.Once
	bellPCMData []byte

	audioOnce sync.Once
	audioCtx  *oto.Context
	audioErr  error
	audioMu   sync.Mutex
)

func bellPCM() []byte {
	bellPCMOnce.Do(func() {
		once := synthesizeBell()
		bellPCMData = append(append([]byte{}, once...), once...)
	})
	return bellPCMData
}

func playBell() error {
	ctx, err := bellContext()
	if err != nil {
		return err
	}

	audioMu.Lock()
	defer audioMu.Unlock()

	player := ctx.NewPlayer(bytes.NewReader(bellPCM()))
	player.SetVolume(0.85)
	player.Play()

	deadline := time.Now().Add(6 * time.Second)
	for player.IsPlaying() && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	// The device still holds one buffer after the player reports drained.
	time.Sleep(80 * time.Millisecond)
	if err := player.Err(); err != nil {
		player.Close()
		return err
	}
	return player.Close()
}

func bellContext() (*oto.Context, error) {
	audioOnce.Do(func() {
		ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
			SampleRate:      bellSampleRate,
			ChannelCount:    1,
			Format:          oto.FormatSignedInt16LE,
			BufferSize:      50 * time.Millisecond,
			ApplicationName: "pomodoro",
		})
		if err != nil {
			audioErr = err
			return
		}
		<-ready
		if err := ctx.Err(); err != nil {
			audioErr = err
			return
		}
		audioCtx = ctx
	})
	if audioErr != nil {
		return nil, audioErr
	}
	return audioCtx, nil
}

// synthesizeBell builds a two-note FM chime: a strike, then a fifth above it.
func synthesizeBell() []byte {
	const dur = 1.15
	n := int(bellSampleRate * dur)
	samples := make([]float64, n)
	var peak float64
	for i := range samples {
		s := bellSample(float64(i) / bellSampleRate)
		samples[i] = s
		if a := math.Abs(s); a > peak {
			peak = a
		}
	}
	if peak == 0 {
		peak = 1
	}
	gain := 0.72 / peak
	buf := make([]byte, n*2)
	for i, s := range samples {
		v := int16(math.Round(s * gain * 32767))
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(v))
	}
	return buf
}

func bellSample(t float64) float64 {
	strikes := []struct {
		at, freq, amp float64
	}{
		{0, 784.0, 1},
		{0.18, 1174.7, 0.92},
	}
	var sum float64
	for _, s := range strikes {
		dt := t - s.at
		if dt < 0 {
			continue
		}
		env := math.Exp(-dt * 3.6)
		index := 3.4 * math.Exp(-dt*7)
		mod := math.Sin(2 * math.Pi * s.freq * 1.414 * dt)
		sum += s.amp * env * math.Sin(2*math.Pi*s.freq*dt+index*mod)
	}
	return sum
}
