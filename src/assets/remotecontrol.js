// control of a remote piaf playback system

class RemoteControl {
    // The player interface:
    currentTrackDuration = 0

    // remote-control-specific variables
    currentPosition = 0

    fastBackward() {
        this.seek(Math.max(this.currentPosition - 15, 0))
    }

    fastForward() {
        this.seek(Math.min(this.currentPosition + 15, this.currentTrackDuration - 2))
    }

    playFile(mediaFile) {
        // Is this the current file?
        // If so, resume (if necessary)
        // If not, play it
        // and go to the player control page
        const apiEndpoint = (mediaFile == nowPlayingFile) ? "/player/resume" : "/player/play/" + encodeURIComponent(mediaFile)
        fetch(apiEndpoint, { method: "PUT" })
        // Don't wait for the fetch to return - that only happens when playback actually starts,
        // which might be a while in the case of having to convert a long .m4a to .wav. This way
        // there's immediate visual feedback that the click has been seen.
        gotoPage('/player/control')
    }

    pause() {
        fetch("/player/pause", { method: "PUT" })
    }

    resume() {
        fetch("/player/resume", { method: "PUT" })
    }

    seek(newPos) {
        fetch(`/player/seek?p=${newPos}`, { method: "PUT" })
    }

    setSpeed(speed) {
        fetch(`/player/speed?v=${speed}`, { method: "PUT" })
    }

    setVolume(volume) {
        fetch(`/player/volume?v=${volume}`, { method: "PUT" })
    }

    async updateNowPlaying() {
        const response = await fetch("/player/status")
        if (response?.ok) {
            const data = await response.json()
            this.showNowPlaying(data)
        } else {
            console.log(`Fetch failed: ${response?.status}`)
        }
    }

    showNowPlaying(data) {
        windowMediaControls.showPlaybackState(data['state'])
        windowMediaControls.showPlaybackSpeed(data['speed'])
        this.currentPosition = data['position']
        this.currentTrackDuration = data['duration']
        windowMediaControls.showTrackPositionAndDuration(this.currentPosition, this.currentTrackDuration)
        windowMediaControls.showVolume(data['volume'])
        windowTrackDisplay.showActiveTrack(data['now_playing'])
    }
}
