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
	State      PlayerState
	NowPlaying string
	Speed      float64 // ratio, e.g. 1.0, 1.1, etc
	Volume     int     // 0 <= Volume <= 100

	// The beep streamers
	pauser         *beep.Ctrl
	resampler      *beep.Resampler
	volumeStreamer *effects.Volume
}

func NewPlayer() *Player {
	return &Player{
		State:  PlayerStateStopped,
		Speed:  1.0,
		Volume: 100,
	}
}

func calculateVolumeRatio(volume int) float64 {
	return (float64(volume - 100)) / 25.0
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
	player.pauser = &beep.Ctrl{
		Streamer: beep.Seq(streamer, beep.Callback(func() {
			player.State = PlayerStateStopped
		})),
		Paused: false,
	}
	player.resampler = beep.ResampleRatio(4, player.Speed, player.pauser)
	player.volumeStreamer = &effects.Volume{
		Streamer: player.resampler,
		Base:     2,
		Volume:   calculateVolumeRatio(player.Volume),
		Silent:   false,
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/2))

	speaker.Play(player.volumeStreamer)

	player.State = PlayerStatePlaying

	return nil
}

func (player *Player) Close() error {
	if player.State != PlayerStateStopped {
		player.pauser.Streamer = nil
	}

	return nil
}

func (player *Player) Pause() error {
	if player.State == PlayerStatePlaying {
		speaker.Lock()
		player.pauser.Paused = true
		speaker.Unlock()
		player.State = PlayerStatePaused
	}
	return nil
}

func (player *Player) Resume() error {
	if player.State == PlayerStatePaused {
		speaker.Lock()
		player.pauser.Paused = false
		speaker.Unlock()
		player.State = PlayerStatePlaying
	}
	return nil
}

func (player *Player) SetSpeed(newValue float64) error {
	speaker.Lock()
	player.resampler.SetRatio(newValue)
	speaker.Unlock()
	return nil
}

func (player *Player) SetVolume(newValue int) error {
	if newValue < 0 {
		newValue = 0
	}
	if newValue > 100 {
		newValue = 100
	}
	player.Volume = newValue
	speaker.Lock()
	if newValue == 0 {
		player.volumeStreamer.Silent = true
	} else {
		player.volumeStreamer.Silent = false
		player.volumeStreamer.Volume = calculateVolumeRatio(newValue)
	}
	speaker.Unlock()
	return nil
}
