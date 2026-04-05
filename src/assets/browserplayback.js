// local playback, fetching files from a piaf system

const localStoragePlaybackSpeedKey = 'piaf-local-speed'
const localStoragePositionsKey = 'piaf-local-positions'
const localStorageVolumeKey = 'piaf-local-volume'

class BrowserPlayback {
    // The player interface:
    currentTrackDuration = 0

    // Other instance variables:
    currentFile = null
    fetching = false
    howl = null

    playFile(mediaFile) {
        if (this.howl) {
            this.howl.stop()
        }
        const volume = this.getSavedVolume()
        const rate = (document.getElementById('speed-menu-button') === null ? 1 : this.getSavedPlaybackSpeed())
        this.howl = new Howl({
            src: ['/mediafile/' + encodeURIComponent(mediaFile)],
            preload: false,
            html5: true,
            volume: volume / 100,
            rate: rate,
            onloaderror: (soundId, errorCode) => {
                console.log("howlerjs onloaderror: " + errorCode)
            },
            onplayerror: (soundId, errorCode) => {
                console.log("howlerjs reported error code " + errorCode)
            },
            onend: () => {
                this.removeSavedPosition(this.currentFile)
                windowMediaControls.showPlaybackState('stopped')
                windowTrackDisplay.showNoTrackPlaying()
                markFilePlayed(this.currentFile)
                this.currentFile = null
                // Q: Auto play the next??
            },
            onpause: () => {
                if (!this.fetching) {
                    windowMediaControls.showPlaybackState('paused')
                }
            },
        })
        this.howl.once('play', () => {
            this.fetching = false
            windowMediaControls.showPlaybackState('playing')
            const position = this.getSavedPosition(this.currentFile)
            if (position !== null && position > 0) {
                this.howl.seek(position)
            }
        })
        this.currentFile = mediaFile
        this.fetching = true
        this.howl.play()
        navigator.mediaSession.setActionHandler('pause', () => { this.pause() })
        navigator.mediaSession.setActionHandler('play', () => { this.resume() })
        windowMediaControls.showPlaybackState('fetching')
        windowTrackDisplay.showActiveTrack(mediaFile)
    }

    fastBackward() {
        if (this.howl !== null) {
            this.seek(Math.max(this.howl.seek() - 15, 0))
        }
    }

    fastForward() {
        if (this.howl !== null) {
            this.seek(Math.min(this.howl.seek() + 15, this.currentTrackDuration - 2))
        }
    }

    pause() {
        if (this.howl) {
            this.howl.pause()
            this.savePosition(this.currentFile, this.howl.seek())
            windowMediaControls.showPlaybackState('paused')
        } else {
            windowMediaControls.showPlaybackState('stopped')
        }
    }

    resume() {
        this.howl?.play()
    }

    seek(newPos) {
        this.howl?.seek(newPos)
    }

    setSpeed(speed) {
        this.howl?.rate(speed)
        this.savePlaybackSpeed(speed)
        windowMediaControls.showPlaybackSpeed(speed)
    }

    loadSavedPositionMap() {
        let savedPositions = localStorage.getItem(localStoragePositionsKey)
        if (savedPositions === null || savedPositions === "") {
            savedPositions = "{}"
        }
        const savedPositionsObj = JSON.parse(savedPositions)
        return new Map(Object.entries(savedPositionsObj))
    }

    saveSavedPositionMap(savedPositionsMap) {
        localStorage.setItem(localStoragePositionsKey, JSON.stringify(Object.fromEntries(savedPositionsMap)))
    }

    getSavedPosition(mediaFile) {
        const savedPositionsMap = this.loadSavedPositionMap()
        return savedPositionsMap.get(mediaFile) ?? 0
    }

    removeSavedPosition(mediaFile) {
        let savedPositionsMap = this.loadSavedPositionMap()
        savedPositionsMap.delete(mediaFile)
        this.saveSavedPositionMap(savedPositionsMap)
    }

    savePosition(mediaFile, position) {
        // load any existing saved map
        let savedPositionsMap = this.loadSavedPositionMap()
        // update
        savedPositionsMap.set(mediaFile, position)
        // save
        this.saveSavedPositionMap(savedPositionsMap)
    }

    setVolume(volume) {
        // 0 <= volume <= 100
        this.saveVolume(volume)
        this.howl?.volume(volume / 100)
    }

    saveVolume(volume) {
        localStorage.setItem(localStorageVolumeKey, volume)
    }

    getSavedPlaybackSpeed() {
        let speed = localStorage.getItem(localStoragePlaybackSpeedKey)
        if (speed === null || Number.isNaN(speed)) {
            speed = 1
        }

        speed = Number(speed)
        if (speed < 0.5) {
            speed = 1
        } else if (speed > 2) {
            speed = 2
        }
        windowMediaControls.showPlaybackSpeed(speed)
        return speed
    }

    savePlaybackSpeed(speed) {
        localStorage.setItem(localStoragePlaybackSpeedKey, speed)
    }

    getSavedVolume() {
        let volume = localStorage.getItem(localStorageVolumeKey) // 0 <= volume <= 100
        if (volume === null || Number.isNaN(volume)) {
            volume = 50
        }

        volume = Number(volume)
        if (volume < 0) {
            volume = 0
        } else if (volume > 100) {
            volume = 100
        }
        windowMediaControls.showVolume(volume)
        return volume
    }

    async updateNowPlaying() {
        if (this.howl === null) {
            windowMediaControls.showPlaybackState('stopped')
            windowMediaControls.showVolume(this.getSavedVolume())
        } else if (this.fetching) {
            windowMediaControls.showPlaybackState('fetching')
        } else {
            windowMediaControls.showPlaybackState(this.howl.playing() ? 'playing' : 'paused')
            this.currentTrackDuration = this.howl.duration()
            windowMediaControls.showTrackPositionAndDuration(this.howl.seek(), this.currentTrackDuration)
        }
    }
}
