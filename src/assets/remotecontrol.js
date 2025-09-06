// control of a remote piaf playback system

class RemoteControl {
    currentTrackDuration = 0
    currentPosition = 0

    fastBackward() {
        let newPos = this.currentPosition - 15
        if (newPos < 0) {
            newPos = 0
        }
        this.seek(newPos)
    }

    fastForward() {
        let newPos = this.currentPosition  + 15
        if (newPos > this.currentTrackDuration - 2) {
            newPos = this.currentTrackDuration - 2
        }
        this.seek(newPos)
    }

    playFile(mediaFile) {
        // Is this the current file?
        // If so, resume (if necessary)
        // If not, play it
        // and go to the player control page
        const apiEndpoint = (mediaFile == nowPlayingFile) ? "/player/resume" : "/player/play/" + mediaFile
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

        nowPlayingFile = data['now_playing']
        for (const tr of trMediaFiles) {
            const rowPath = tr.getAttribute('data-file')
            if (rowPath == nowPlayingFile) {
                tr.classList.add('table-info')
            } else {
                tr.classList.remove('table-info')
            }
        }
    }
}
