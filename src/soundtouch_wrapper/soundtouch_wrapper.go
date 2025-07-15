package soundtouch_wrapper

/*
#cgo CXXFLAGS: -I/opt/homebrew/include -std=c++11
#cgo LDFLAGS: -L/opt/homebrew/lib -lsoundtouch
#include "soundtouch_wrapper.h"
*/
import "C"
import (
	"time"
	"unsafe"

	"github.com/gopxl/beep"
)

type SoundTouchSample float32

func calculateMaximumReadSamples(tempo float64) int {
	return int(512 * tempo)
}

func calculateMaximumBufferedSamples(sr beep.SampleRate, tempo float64) int {
	return int(float64(sr.N(time.Second/4)) * tempo)
}

type TimeStretch struct {
	src                beep.Streamer
	st                 unsafe.Pointer
	tmp                [][2]float64
	in                 []SoundTouchSample
	out                []SoundTouchSample
	channels           int
	sampleRate         beep.SampleRate
	maxReadSamples     int
	maxBufferedSamples int
}

func NewTimeStretch(src beep.Streamer, sr beep.SampleRate, tempo float64) *TimeStretch {
	channels := 2
	st := C.soundtouch_new(C.int(sr), C.int(channels), C.float(tempo))
	maxRead := calculateMaximumReadSamples(tempo)
	return &TimeStretch{
		src:                src,
		st:                 st,
		tmp:                make([][2]float64, maxRead),
		in:                 make([]SoundTouchSample, 0),
		out:                make([]SoundTouchSample, 0),
		channels:           channels,
		sampleRate:         sr,
		maxReadSamples:     maxRead,
		maxBufferedSamples: calculateMaximumBufferedSamples(sr, tempo),
	}
}

func (ts *TimeStretch) Stream(out [][2]float64) (int, bool) {
	// Stream audio from the original source

	buffered := int(C.soundtouch_num_unprocessed_samples(ts.st))
	ok := true
	toRead := min(ts.maxReadSamples, ts.maxBufferedSamples-buffered)
	if toRead > 0 {
		// We've (partially) drained the soundtouch buffer, so read from the input
		n, _ := ts.src.Stream(ts.tmp[:toRead])
		if n > 0 {
			if n*2 > len(ts.in) {
				ts.in = make([]SoundTouchSample, n*2)
			}
			// Ensure data are arranged correctly, and the correct type
			for i := range n {
				ts.in[i*2] = SoundTouchSample(ts.tmp[i][0])
				ts.in[i*2+1] = SoundTouchSample(ts.tmp[i][1])
			}

			// Feed samples into SoundTouch
			C.soundtouch_put_samples(ts.st, (*C.float)(unsafe.Pointer(&ts.in[0])), C.int(n))
		}
	}

	// Get stretched samples from SoundTouch
	samplesToRead := len(out)
	if samplesToRead*2 > len(ts.out) {
		ts.out = make([]SoundTouchSample, samplesToRead*2)
	}
	got := int(C.soundtouch_receive_samples(ts.st, (*C.float)(unsafe.Pointer(&ts.out[0])), C.int(samplesToRead)))
	if got > 0 {
		// Ensure data are arranged correctly on the way out
		for i := range got {
			out[i][0] = float64(ts.out[i*2])
			out[i][1] = float64(ts.out[i*2+1])
		}
	}
	for i := got; i < len(out); i++ {
		out[i][0] = 0
		out[i][1] = 0
	}

	return len(out), ok
}

func (ts *TimeStretch) SetTempo(tempo float64) {
	ts.maxReadSamples = calculateMaximumReadSamples(tempo)
	ts.maxBufferedSamples = calculateMaximumBufferedSamples(ts.sampleRate, tempo)
	if ts.maxReadSamples > len(ts.tmp) {
		ts.tmp = make([][2]float64, ts.maxReadSamples)
	}
	C.soundtouch_set_tempo(ts.st, C.float(tempo))
}

func (ts *TimeStretch) Close() {
	C.soundtouch_delete(ts.st)
}

func (ts *TimeStretch) Err() error {
	return ts.src.Err()
}
