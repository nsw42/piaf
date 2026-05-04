// A class to show the playback status (paused, position, etc), as well as
// manage the events for the controls

const positionSliderResolution = 1000

class WindowMediaControls {
    constructor() {

        this.currentState = 'stopped'
        this.bodyElement = document.getElementsByTagName('body')[0]
        this.speedMultiplier = 1

        this.fetchingIndicator = document.getElementById('fetching')

        this.pauseButtons = Array.from(document.getElementsByClassName('piaf-btn-pause'))
        for (const button of this.pauseButtons) {
            button.addEventListener("click", () => { currentPlayer.pause() })
        }

        this.resumeButtons = Array.from(document.getElementsByClassName('piaf-btn-resume'))
        for (const button of this.resumeButtons) {
            button.addEventListener("click", () => {
                if (this.currentState === 'stopped' || this.currentState === 'uninitialised') {
                    const firstFile = this.getFirstFileOnPage()
                    if (firstFile !== null) {
                        // Should never be null, because we should have disabled the
                        // 'resume' button if there are no files
                        currentPlayer.playFile(firstFile)
                    }
                } else {
                    currentPlayer.resume()
                }
            })
        }

        this.fastBackwardButtons = document.getElementsByClassName('piaf-fast-backward')
        for (const button of this.fastBackwardButtons) {
            button.addEventListener('click', () => { currentPlayer.fastBackward() })
        }

        this.fastForwardButtons = document.getElementsByClassName('piaf-fast-forward')
        for (const button of this.fastForwardButtons) {
            button.addEventListener('click', () => { currentPlayer.fastForward() })
        }

        this.positionSlider = document.getElementById('position-slider')
        this.positionText = document.getElementById('position-display')
        this.trackRemainingText = document.getElementById('track-remaining')

        this.sliderDragActive = false

        this.positionSlider?.addEventListener('input', (event) => {
            this.sliderDragActive = true
            const newPos = currentPlayer?.currentTrackDuration * event.target.value / positionSliderResolution
            this.showTrackPosition(newPos)
        })

        this.positionSlider?.addEventListener('change', (event) => {
            this.sliderDragActive = false
            const newPos = currentPlayer?.currentTrackDuration * event.target.value / positionSliderResolution
            this.showTrackPosition(newPos)
            currentPlayer.seek(newPos)
        })

        this.volumeSlider = document.getElementById('volume-slider')
        this.volumeSlider?.addEventListener("input", (event) => {
            this.sliderDragActive = true
            this.showVolume(event.target.value)
            currentPlayer.setVolume(event.target.value)
        })
        this.volumeSlider?.addEventListener("change", () => {
            this.sliderDragActive = false
        })

        this.volumeText = document.getElementById('volume-text')

        this.speedMenuButton = document.getElementById('speed-menu-button')
        for (const element of document.getElementsByClassName('speed-menu-item')) {
            element.addEventListener("click", () => {
                currentPlayer.setSpeed(element.dataset.speedValue)
            })
        }

        // Prepare the lists of buttons:
        // for each player state (uninitialised / stopped, initialising, paused, playing)
        // need to ensure that each button goes into either disable or
        // enable. The list of buttons:
        //  - control link
        //  - pause
        //  - resume
        //  - speed (0 or 1)
        //  - position (0 or 1)
        //  - fast backward (0 or 1)
        //  - fast forward (0 or 1)
        //  - volume

        let allControls = this.pauseButtons.concat(this.resumeButtons)
        // allControls.push(controlLinkButton)  TODO
        allControls.push(this.speedMenuButton,
                         this.positionSlider,
                         ...this.fastBackwardButtons,
                         ...this.fastForwardButtons,
                         this.volumeSlider)
        allControls = allControls.filter(Boolean)

        this.disableWhenStopped = allControls

        this.disableWhenInitialising = allControls

        this.disableWhenPaused = this.pauseButtons.concat(...this.fastBackwardButtons, ...this.fastForwardButtons)
        this.enableWhenPaused = allControls.filter(c => !this.disableWhenPaused.includes(c))

        this.disableWhenPlaying = this.resumeButtons
        this.enableWhenPlaying = allControls.filter(c => !this.disableWhenPlaying.includes(c))
    }

    getFirstFileOnPage() {
        const trs = document.getElementsByClassName('piaf-media-files')
        return (trs.length == 0) ? null : trs[0].dataset.file
    }

    showPlaybackSpeed(speed) {
        if (this.speedMenuButton) {
            if (speed instanceof Number || typeof(speed) === 'number') {
                this.speedMultiplier = speed
                speed = speed.toString()
            } else {
                let speedString = speed;
                if (speedString.endsWith('x')) {
                    speedString = speedString.substring(0, speedString.length-1)
                }
                this.speedMultiplier = Number(speedString)
            }
            if (!speed.endsWith('x')) {
                speed += "x"
            }
            if (speed == '1.0x') {
                speed = '1x'
            }
            this.speedMenuButton.innerHTML = speed
        }
    }

    showPlaybackState(state) {
        this.currentState = state;

        if (state === 'fetching') {
            this.bodyElement.classList.add('fetching')
        } else {
            this.bodyElement.classList.remove('fetching')
        }

        switch (state) {
            case 'stopped':
            case 'uninitialised':
                if (location.pathname == '/player/control') {
                    // not sensible to stay on this page
                    let mediaDir = ""
                    if (nowPlayingFile) {
                        // return to the media directory
                        const slash = nowPlayingFile.lastIndexOf('/')
                        mediaDir = nowPlayingFile.substr(0, slash)
                    }
                    gotoPage('/media/' + mediaDir)
                } else {
                    if (nowPlayingFile) {
                        // refresh the index
                        location.reload()
                    } else {
                        disableElements(this.disableWhenStopped)
                        if (this.getFirstFileOnPage()) {
                            // it makes sense to enable the 'play' button
                            enableElements(this.resumeButtons)
                        }
                    }
                }
                break
            case 'initialising':
                disableElements(this.disableWhenInitialising)
                break
            case 'fetching':
                disableElements(this.disableWhenStopped)
                break
            case 'paused':
                disableElements(this.disableWhenPaused)
                enableElements(this.enableWhenPaused)
                break
            case 'playing':
                disableElements(this.disableWhenPlaying)
                enableElements(this.enableWhenPlaying)
                break
        }
    }

    showTrackPosition(pos) {
        this.positionText.innerHTML = formatDuration(pos)
    }

    showTrackPositionAndDuration(position, duration) {
        if (this.positionSlider && this.positionText && this.trackRemainingText && !this.sliderDragActive) {
            if (position === null || position === undefined) {
                this.positionSlider.value = 0
                this.positionText.innerHTML = ""
                this.trackRemainingText.innerHTML = formatDuration(duration)
            } else {
                this.positionSlider.value = position * positionSliderResolution / duration
                this.showTrackPosition(position)
                this.trackRemainingText.innerHTML = "-" + formatDuration((duration - position) / this.speedMultiplier)
            }
        }
    }

    showVolume(volume) {
        // 0 <= volume <= 100
        this.volumeSlider.value = this.volumeText.innerHTML = volume
    }
}
