package main

import (
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/v2/effects"
)

type PlayerState int

const (
	PlayerStateStopped PlayerState = iota
	PlayerStatePlaying
	PlayerStatePaused
)

type Player struct {
	State         PlayerState
	Control       *beep.Ctrl
	NowPlaying    string
	DisplayVolume int // 0 <= DisplayVolume <= 100
	Volume        *effects.Volume
}

func NewPlayer() *Player {
	return &Player{
		State:         PlayerStateStopped,
		DisplayVolume: 100,
	}
}

func (player *Player) Play(path string) error {
	player.Close()

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return err
	}
	player.Control = &beep.Ctrl{
		Streamer: beep.Seq(streamer, beep.Callback(func() {
			player.State = PlayerStateStopped
		})),
		Paused: false,
	}
	player.Volume = &effects.Volume{
		Streamer: player.Control,
		Base:     2,
		Volume:   calculateVolumeRatio(player.DisplayVolume),
		Silent:   false,
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/2))

	speaker.Play(player.Volume)

	player.State = PlayerStatePlaying

	return nil
}

func (player *Player) Close() error {
	if player.State != PlayerStateStopped {
		player.Control.Streamer = nil
	}

	return nil
}

func (player *Player) Pause() error {
	if player.State == PlayerStatePlaying {
		speaker.Lock()
		player.Control.Paused = true
		speaker.Unlock()
		player.State = PlayerStatePaused
	}
	return nil
}

func (player *Player) Resume() error {
	if player.State == PlayerStatePaused {
		speaker.Lock()
		player.Control.Paused = false
		speaker.Unlock()
		player.State = PlayerStatePlaying
	}
	return nil
}

func (player *Player) SetVolume(newValue int) error {
	if newValue < 0 {
		newValue = 0
	}
	if newValue > 100 {
		newValue = 100
	}
	player.DisplayVolume = newValue
	speaker.Lock()
	if newValue == 0 {
		player.Volume.Silent = true
	} else {
		player.Volume.Silent = false
		player.Volume.Volume = calculateVolumeRatio(newValue)
	}
	speaker.Unlock()
	return nil
}

func calculateVolumeRatio(volume int) float64 {
	return (float64(volume - 100)) / 25.0
}
