package main

import (
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

type PlayerState int

const (
	PlayerStateStopped PlayerState = iota
	PlayerStatePlaying
	PlayerStatePaused
)

type Player struct {
	State      PlayerState
	Control    *beep.Ctrl
	NowPlaying string
}

func NewPlayer() *Player {
	return &Player{
		State: PlayerStateStopped,
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

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/2))

	speaker.Play(player.Control)

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
