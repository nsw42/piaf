// A class to show the playback status (paused, position, etc), as well as
// manage the events for the controls

class WindowMediaControls {
    constructor() {

        this.bodyElement = document.getElementsByTagName('body')[0]

        this.fetchingIndicator = document.getElementById('fetching')

        this.pauseButtons = Array.from(document.getElementsByClassName('piaf-btn-pause'))
        for (const button of this.pauseButtons) {
            button.addEventListener("click", () => { currentPlayer.pause() })
        }

        this.resumeButtons = Array.from(document.getElementsByClassName('piaf-btn-resume'))
        for (const button of this.resumeButtons) {
            button.addEventListener("click", () => { currentPlayer.resume() })
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
        this.trackDurationText = document.getElementById('track-duration')

        this.sliderDragActive = false

        this.positionSlider?.addEventListener('input', (event) => {
            this.sliderDragActive = true
            const newPos = currentPlayer?.currentTrackDuration * event.target.value / 100
            this.showTrackPosition(newPos)
        })

        this.positionSlider?.addEventListener('change', (event) => {
            this.sliderDragActive = false
            const newPos = currentPlayer?.currentTrackDuration * event.target.value / 100
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
                currentPlayer.setSpeed(element.getAttribute('data-speed-value'))
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
        allControls.push(this.speedMenuButton)
        allControls.push(this.positionSlider)
        allControls.push(...this.fastBackwardButtons)
        allControls.push(...this.fastForwardButtons)
        allControls.push(this.volumeSlider)
        allControls = allControls.filter(c => c)

        this.disableWhenStopped = allControls
        this.disableWhenInitialising = allControls

        this.disableWhenPaused = this.pauseButtons.concat(...this.fastBackwardButtons, ...this.fastForwardButtons)
        this.enableWhenPaused = allControls.filter(c => this.disableWhenPaused.indexOf(c) == -1)

        this.disableWhenPlaying = this.resumeButtons
        this.enableWhenPlaying = allControls.filter(c => this.disableWhenPlaying.indexOf(c) == -1)
    }

    showPlaybackSpeed(speed) {
        if (this.speedMenuButton) {
            if (!speed.endsWith('x')) {
                speed += "x"
            }
            this.speedMenuButton.innerHTML = speed
        }
    }

    showPlaybackState(state) {
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
        if (this.positionSlider && this.positionText && this.trackDurationText && !this.sliderDragActive) {
            this.trackDurationText.innerHTML = formatDuration(duration)
            if (position === null || position === undefined) {
                this.positionSlider.value = 0
                this.positionText.innerHTML = ""
            } else {
                this.positionSlider.value = position * 100 / duration
                this.showTrackPosition(position)
            }
        }
    }

    showVolume(volume) {
        // 0 <= volume <= 100
        this.volumeSlider.value = this.volumeText.innerHTML = volume
    }
}
