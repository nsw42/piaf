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
	"github.com/nsw42/piaf/mediadir"
	"github.com/nsw42/piaf/soundtouch_wrapper"
)

type PlayerState int

const (
	PlayerStateStopped PlayerState = iota
	PlayerStateInitialising
	PlayerStatePlaying
	PlayerStatePaused
)

func (state PlayerState) String() string {
	switch state {
	case PlayerStateStopped:
		return "stopped"
	case PlayerStateInitialising:
		return "initialising"
	case PlayerStatePlaying:
		return "playing"
	case PlayerStatePaused:
		return "paused"
	default:
		return "unknown"
	}
}

type Player struct {
	State       PlayerState
	NowPlaying  *mediadir.MediaFile
	Speed       float64 // ratio, e.g. 1.0, 1.1, etc
	SpeedString string
	Volume      int // 0 <= Volume <= 100

	// The beep streamers and other pertinent information
	format         beep.Format
	seeker         beep.StreamSeekCloser
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
		Volume:      50,
	}
}

func calculateVolumeRatio(volume int) float64 {
	return (float64(volume - 100)) / 25.0
}

func (player *Player) Play(file *mediadir.MediaFile, enableSpeedControl bool, eofCallback func()) error {
	player.Close()

	player.State = PlayerStateInitialising // Decoding may take a while
	player.NowPlaying = file

	f, err := os.Open(file.Path)
	if err != nil {
		return err
	}

	if strings.HasSuffix(file.Path, ".mp3") {
		player.seeker, player.format, err = mp3.Decode(f)
	} else if strings.HasSuffix(file.Path, ".m4a") {
		player.seeker, player.format, err = seekable.Decode(f)
	} else {
		err = fmt.Errorf("don't know how to play %s", file.Path)
	}

	if err != nil {
		player.State = PlayerStateStopped
		player.NowPlaying = nil
		return err
	}

	player.eofHandler = beep.Seq(player.seeker, beep.Callback(func() {
		player.State = PlayerStateStopped
		player.NowPlaying = nil
		eofCallback()
	}))

	streamer := player.eofHandler
	if enableSpeedControl {
		player.resampler = soundtouch_wrapper.NewTimeStretch(
			player.eofHandler,
			player.format.SampleRate,
			player.Speed,
		)
		streamer = player.resampler
	} else {
		player.resampler = nil
	}
	player.pauser = &beep.Ctrl{
		Streamer: streamer,
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

	speaker.Init(player.format.SampleRate, player.format.SampleRate.N(time.Second/4))

	speaker.Play(player.volumeStreamer)

	player.State = PlayerStatePlaying

	return nil
}

func (player *Player) Close() error {
	if player.State != PlayerStateStopped {
		player.pauser.Streamer = nil
	}

	if player.seeker != nil {
		speaker.Lock()
		player.seeker.Close()
		speaker.Unlock()
	}

	return nil
}

func (player *Player) Pause() error {
	if player.State == PlayerStatePlaying {
		speaker.Lock()
		defer speaker.Unlock()
		player.pauser.Paused = true
		player.State = PlayerStatePaused
	}
	return nil
}

func (player *Player) Resume() error {
	if player.State == PlayerStatePaused {
		speaker.Lock()
		defer speaker.Unlock()
		player.pauser.Paused = false
		player.State = PlayerStatePlaying
	}
	return nil
}

func (player *Player) GetPosition() time.Duration {
	if player.seeker == nil {
		return 0
	}
	return player.format.SampleRate.D(player.seeker.Position())
}

func (player *Player) SetPosition(p time.Duration) error {
	if player.seeker == nil {
		return errors.New("not playing or invalid state")
	}
	speaker.Lock()
	defer speaker.Unlock()
	return player.seeker.Seek(player.format.SampleRate.N(p))
}

func (player *Player) SetSpeed(newValue string) error {
	if player.resampler == nil {
		return errors.New("speed control disabled for this player")
	}
	speed, err := strconv.ParseFloat(newValue, 64)
	if err != nil || speed < 0 {
		return errors.New("illegal speed string")
	}
	player.SpeedString = newValue
	player.Speed = speed
	speaker.Lock()
	defer speaker.Unlock()
	player.resampler.SetTempo(speed)
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
		defer speaker.Unlock()
		if newValue == 0 {
			player.volumeStreamer.Silent = true
		} else {
			player.volumeStreamer.Silent = false
			player.volumeStreamer.Volume = calculateVolumeRatio(newValue)
		}
	}
	return nil
}
