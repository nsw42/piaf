// local playback, fetching files from a piaf system

const localStorageVolumeKey = 'piaf-local-volume'
const localStoragePositionsKey = 'piaf-local-positions'

class LocalPlayback {
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
        this.howl = new Howl({
            src: ['/mediafile/' + mediaFile],
            preload: false,
            html5: true,
            volume: volume / 100.0,
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
                fetch(`/mediafile/${this.currentFile}`, { method: "DELETE" }).then(() => {
                    location.reload()
                })
                this.currentFile = null
                // Q: Auto play the next??
            },
            onpause: () => {
                if (!this.fetching) {
                    windowMediaControls.showPlaybackState('paused')
                }
            },
            onplay: () => {
                this.fetching = false
                windowMediaControls.showPlaybackState('playing')
                const position = this.getSavedPosition(this.currentFile)
                if (position !== null && position > 0) {
                    this.howl.seek(position)
                }
            },
        })
        this.currentFile = mediaFile
        this.fetching = true;
        this.howl.play()
        windowMediaControls.showPlaybackState('fetching')
        windowTrackDisplay.showActiveTrack(mediaFile)
    }

    fastBackward() {
        console.log("LocalPlayback.fastBackward: NYI")
    }

    fastForward() {
        console.log("LocalPlayback.fastForward: NYI")
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
        console.log("LocalPlayback.setSpeed: NYI")
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
        this.howl?.volume(volume / 100.0)
    }

    saveVolume(volume) {
        localStorage.setItem(localStorageVolumeKey, volume)
    }

    getSavedVolume() {
        let volume = localStorage.getItem(localStorageVolumeKey) // 0 <= volume <= 100
        if (volume === null || isNaN(volume)) {
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
