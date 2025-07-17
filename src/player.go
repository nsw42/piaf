package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"

	"github.com/nsw42/beepm4a/m4a/seekable"
	"github.com/nsw42/piaf/soundtouch_wrapper"
)

type PlayerState int

const (
	PlayerStateStopped PlayerState = iota
	PlayerStatePlaying
	PlayerStatePaused
)

type Player struct {
	State       PlayerState
	NowPlaying  string
	Speed       float64 // ratio, e.g. 1.0, 1.1, etc
	SpeedString string
	Volume      int // 0 <= Volume <= 100

	// The beep streamers
	eofHandler     beep.Streamer
	pauser         *beep.Ctrl
	resampler      *soundtouch_wrapper.TimeStretch
	volumeStreamer *effects.Volume
}

func NewPlayer() *Player {
	return &Player{
		State:       PlayerStateStopped,
		Speed:       1.0,
		SpeedString: "1x",
		Volume:      100,
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

	var streamer beep.StreamSeekCloser
	var format beep.Format

	if strings.HasSuffix(path, ".mp3") {
		streamer, format, err = mp3.Decode(f)
	} else if strings.HasSuffix(path, ".m4a") {
		streamer, format, err = seekable.Decode(f)
	} else {
		err = fmt.Errorf("don't know how to play %s", path)
	}

	if err != nil {
		return err
	}

	player.eofHandler = beep.Seq(streamer, beep.Callback(func() {
		player.State = PlayerStateStopped
		player.NowPlaying = ""
	}))
	player.resampler = soundtouch_wrapper.NewTimeStretch(
		player.eofHandler,
		format.SampleRate,
		player.Speed,
	)
	player.pauser = &beep.Ctrl{
		Streamer: player.resampler,
		Paused:   false,
	}
	silent := false
	if player.Volume == 0 {
		silent = true
	}
	player.volumeStreamer = &effects.Volume{
		Streamer: player.pauser,
		Base:     2,
		Volume:   calculateVolumeRatio(player.Volume),
		Silent:   silent,
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/4))

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

func (player *Player) SetSpeed(newValue string) error {
	speed, err := strconv.ParseFloat(newValue, 64)
	if err != nil || speed < 0 {
		return errors.New("illegal speed string")
	}
	player.SpeedString = newValue
	player.Speed = speed
	speaker.Lock()
	player.resampler.SetTempo(speed)
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
	if player.volumeStreamer != nil {
		speaker.Lock()
		if newValue == 0 {
			player.volumeStreamer.Silent = true
		} else {
			player.volumeStreamer.Silent = false
			player.volumeStreamer.Volume = calculateVolumeRatio(newValue)
		}
		speaker.Unlock()
	}
	return nil
}
