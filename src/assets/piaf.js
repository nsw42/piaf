let contentsDiv;
let navbar;
let pauseButtons;
let resumeButtons;
let fastBackwardButton;
let fastForwardButton;
let speedMenuButton;
let positionSlider;
let positionText;
let currentTrackDuration;
let currentPosition;
let trackDurationText;
let volumeSlider;
let volumeText;
let sliderDragActive;
let trMediaFiles;
let nowPlayingFile;

let disableWhenStopped;
let disableWhenPaused;
let enableWhenPaused;
let disableWhenPlaying;
let enableWhenPlaying;

function initPiaf() {
    contentsDiv = document.getElementById('main-content')
    navbar = document.querySelector('.navbar.fixed-top')
    window.addEventListener('DOMContentLoaded', setContentPadding)
    window.addEventListener('resize', setContentPadding)

    for (const button of document.getElementsByClassName('piaf-play-file')) {
        button.addEventListener("click", () => {
            fetch("/player/play/" + button.getAttribute("data-file"), {
                method: "PUT"
            }).then(() => {
                gotoPage('/player/control')
            })
        })
    }

    pauseButtons = Array.from(document.getElementsByClassName('piaf-btn-pause'))
    for (const button of pauseButtons) {
        button.addEventListener("click", () => {
            fetch("/player/pause", { method: "PUT" })
        })
    }

    resumeButtons = Array.from(document.getElementsByClassName('piaf-btn-resume'))
    for (const button of resumeButtons) {
        button.addEventListener("click", () => {
            fetch("/player/resume", { method: "PUT" })
        })
    }

    fastBackwardButton = document.getElementById('fast-backward')
    fastBackwardButton?.addEventListener('click', () => {
        let newPos = currentPosition - 15
        if (newPos < 0) {
            newPos = 0
        }
        seek(newPos)
    })

    fastForwardButton = document.getElementById('fast-forward')
    fastForwardButton?.addEventListener('click', () => {
        let newPos = currentPosition  + 15
        if (newPos > currentTrackDuration - 2) {
            newPos = currentTrackDuration - 2
        }
        seek(newPos)
    })

    positionSlider = document.getElementById('position-slider')
    positionSlider?.addEventListener('input', (event) => {
        sliderDragActive = true
        const newPos = currentTrackDuration * event.target.value / 100
        positionText.innerHTML = formatDuration(newPos)
    })
    positionSlider?.addEventListener('change', (event) => {
        sliderDragActive = false
        const newPos = currentTrackDuration * event.target.value / 100
        seek(newPos)
    })
    positionText = document.getElementById('position-display')
    trackDurationText = document.getElementById('track-duration')

    volumeSlider = document.getElementById('volume-slider')
    volumeSlider.addEventListener("input", (event) => {
        sliderDragActive = true
        volumeText.innerHTML = event.target.value
        fetch(`/player/volume?v=${event.target.value}`, {
            method: "PUT"
        })
    })
    volumeSlider.addEventListener("change", () => {
        sliderDragActive = false
    })

    volumeText = document.getElementById('volume-text')

    speedMenuButton = document.getElementById('speed-menu-button')
    for (const element of document.getElementsByClassName('speed-menu-item')) {
        element.addEventListener("click", () => {
            fetch(`/player/speed?v=${element.getAttribute('data-speed-value')}`, {
                method: "PUT"
            })
        })
    }

    trMediaFiles = document.getElementsByClassName('piaf-media-files')

    disableWhenStopped = pauseButtons.concat(resumeButtons)
    disableWhenStopped.push(speedMenuButton)
    if (positionSlider) {
        disableWhenStopped.push(positionSlider)
    }

    disableWhenPaused = pauseButtons
    enableWhenPaused = Array.from(resumeButtons)
    enableWhenPaused.push(speedMenuButton)
    if (positionSlider) {
        enableWhenPaused.push(positionSlider)
    }

    disableWhenPlaying = resumeButtons
    enableWhenPlaying = Array.from(pauseButtons)
    enableWhenPlaying.push(speedMenuButton)
    if (positionSlider) {
        enableWhenPlaying.push(positionSlider)
    }

    setTimeout(updateNowPlaying, 1000)
}

function gotoPage(path) {
    location.pathname = path
    history.pushState(null, null, location.toString())
}

function seek(newPos) {
    positionText.innerHTML = formatDuration(newPos)
    fetch(`/player/seek?p=${newPos}`, { method: "PUT" })
}

function setContentPadding() {
    const navbarHeight = navbar.offsetHeight;
    contentsDiv.style.marginTop = `${navbarHeight}px`;
}

function disableElements(elements) {
    for (const elt of elements) {
        elt?.setAttribute('disabled', 'disabled')
    }
}

function enableElements(elements) {
    for (const elt of elements) {
        elt?.removeAttribute('disabled')
    }
}

async function updateNowPlaying() {
    try {
        const response = await fetch("/player/status")
        if (response?.ok) {
            const data = await response.json()
            showNowPlaying(data)
        } else {
            console.log(`Fetch failed: ${response?.status}`)
        }
    } catch (error) {
        console.log(error)
        // and try again in a bit
    }

    setTimeout(updateNowPlaying, 1000)
}

function n02d(n) {
    const s = n.toString()
    return (s.length < 2) ? "0" + s : s
}

function formatDuration(secondsInt) {
    let ss = Math.trunc(secondsInt)
    let mm = Math.trunc(ss / 60)
    ss -= mm*60
    let hh = Math.trunc(mm / 60)
    mm -= hh*60
    return `${n02d(hh)}:${n02d(mm)}:${n02d(ss)}`
}

function showNowPlaying(data) {
    switch (data['state']) {
        case 'stopped':
            if (location.pathname == '/player/control') {
                // not sensible to stay on this page
                let mediaDir = ""
                if (nowPlayingFile) {
                    // return to the media directory
                    let slash = nowPlayingFile.lastIndexOf('/')
                    if (slash < 0) {
                        slash = nowPlayingFile.length
                    }
                    mediaDir = nowPlayingFile.substr(0, slash)
                }
                gotoPage('/media/' + mediaDir)
            } else {
                if (nowPlayingFile) {
                    // refresh the index
                    location.reload()
                } else {
                    disableElements(disableWhenStopped)
                }
            }
            break
        case 'paused':
            disableElements(disableWhenPaused)
            enableElements(enableWhenPaused)
            break
        case 'playing':
            disableElements(disableWhenPlaying)
            enableElements(enableWhenPlaying)
            break
    }

    if (speedMenuButton) {
        const speed = data['speed']
        if (!speed.endsWith('x')) {
            speed += "x"
        }
        speedMenuButton.innerHTML = speed
    }

    if (positionSlider && positionText && trackDurationText && !sliderDragActive) {
        currentTrackDuration = data['duration']
        trackDurationText.innerHTML = formatDuration(currentTrackDuration)
        currentPosition = data['position']
        if (currentPosition) {
            positionSlider.value = currentPosition * 100 / currentTrackDuration
            positionText.innerHTML = formatDuration(currentPosition)
        } else {
            positionSlider.value = 0
            positionText.innerHTML = ""
        }
    }

    volumeSlider.value = volumeText.innerHTML = data['volume']

    nowPlayingFile = data['now_playing']
    for (const tr of trMediaFiles) {
        const rowPath = tr.getAttribute('data-path')
        if (rowPath == nowPlayingFile) {
            tr.classList.add('table-info')
        } else {
            tr.classList.remove('table-info')
        }
    }
}
