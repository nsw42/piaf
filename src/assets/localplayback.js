// local playback, fetching files from a piaf system

const localStorageVolumeKey = 'piaf-local-volume'

class LocalPlayback {
    fetching = false
    howl = null

    playFile(mediaFile) {
        const volume = this.getSavedVolume()
        this.howl = new Howl({
            src: ['/mediafile/' + mediaFile],
            preload: false,
            html5: true,
            volume: volume,
            onloaderror: (soundId, errorCode) => {
                console.log("howlerjs onloaderror: " + errorCode)
            },
            onplayerror: (soundId, errorCode) => {
                console.log("howlerjs reported error code " + errorCode)
            },
            onend: () => {
                windowMediaControls.showPlaybackState('stopped')
                // TODO
                // $('#track_'+trackId).removeClass('active-track');
                // if (localTrackIndex + 1 < playlistTrackIds.length) {
                //     localPlay(localTrackIndex + 1);
                // } else {
                //     hideButtons(['#local-previous', '#local-pause', '#local-fetching', '#local-resume', '#local-next', '#local-volume']);
                //     currentTrackId = localTrackIndex = null;
                // }
            },
            onpause: () => {
                if (!this.fetching) {
                    windowMediaControls.showPlaybackState('paused')
                }
            },
            onplay: () => {
                this.fetching = false
                windowMediaControls.showPlaybackState('playing')
            },
        })
        this.fetching = true;
        this.howl.play()
        windowMediaControls.showPlaybackState('fetching')
    }

    fastBackward() {
        console.log("LocalPlayback.fastBackward: NYI")
    }

    fastForward() {
        console.log("LocalPlayback.fastForward: NYI")
    }

    pause() {
        this.howl?.pause()
        windowMediaControls.showPlaybackState('paused')
    }

    resume() {
        this.howl?.play()
    }

    seek(newPos) {
        console.log("LocalPlayback.seek: NYI")
    }

    setSpeed(speed) {
        console.log("LocalPlayback.setSpeed: NYI")
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
        let volume = localStorage.getItem(localStorageVolumeKey)
        if (volume === null || isNaN(volume)) {
            volume = 1.0
        } else {
            volume = Number(volume)
            if (volume < 0) {
                volume = 0
            } else if (volume > 1.0) {
                volume = 1.0
            }
        }
        windowMediaControls.showVolume(volume * 100.0)
        return volume
    }

    updateNowPlaying() {
        if (this.howl === null) {
            windowMediaControls.showPlaybackState('stopped')
        } else if (this.fetching) {
            windowMediaControls.showPlaybackState('fetching')
        } else {
            windowMediaControls.showPlaybackState(this.howl.playing() ? 'playing' : 'paused')
        }
    }
}
