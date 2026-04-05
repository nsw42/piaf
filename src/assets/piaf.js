const modeRemoteControl = 'remote';
const modeBrowserPlayback = 'browser';

let currentMode;
let modeButtonRemoteControl;
let modeButtonBrowserPlayback;

let body;
let contentsDiv;
let navbar;
let nowPlayingFile;
let indexPositionSlider;
let pageContainsTracks;

let playerBrowserPlayback;
let playerRemoteControl;
let currentPlayer;

let windowMediaControls;
let windowTrackDisplay;

function initPiaf(enableBrowserPlayback, enableRemoteControl) {
    const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]')
    for (const tooltipTriggerEl of tooltipTriggerList) {
        new bootstrap.Tooltip(tooltipTriggerEl)  // NOSONARQA: S1848 false-positive
    }

    windowMediaControls = new WindowMediaControls()
    windowTrackDisplay = new WindowTrackDisplay()

    body = document.getElementsByTagName('body')[0]
    contentsDiv = document.getElementById('main-content')
    navbar = document.querySelector('.navbar.fixed-top')
    window.addEventListener('DOMContentLoaded', setContentPadding)
    window.addEventListener('resize', setContentPadding)

    modeButtonRemoteControl = document.getElementById('mode-indicator-remote')
    modeButtonBrowserPlayback = document.getElementById('mode-indicator-local')
    // modeButtonX do not exist when only one control type is enabled
    modeButtonRemoteControl?.addEventListener('click', () => { setMode(modeBrowserPlayback) })
    modeButtonBrowserPlayback?.addEventListener('click', () => { setMode(modeRemoteControl) })

    indexPositionSlider = document.getElementById('index-position-slider')

    pageContainsTracks = document.getElementById('index-table') && document.getElementsByClassName('piaf-media-files').length > 0

    for (const button of document.getElementsByClassName('piaf-play-file')) {
        button.addEventListener('click', () => {
            currentPlayer.playFile(getDataFileFromContainingTR(button))
        })
    }

    for (const button of document.getElementsByClassName('piaf-mark-played')) {
        button.addEventListener('click', () => {
            markFilePlayed(getDataFileFromContainingTR(button))
        })
    }

    playerRemoteControl = new RemoteControl()
    playerBrowserPlayback = new BrowserPlayback()

    currentMode = localStorage.getItem('piaf-current-mode')
    if (currentMode === null) {
        currentMode = modeRemoteControl
    }
    if (currentMode === modeRemoteControl && !enableRemoteControl) {
        currentMode = modeBrowserPlayback
    } else if (currentMode === modeBrowserPlayback && !enableBrowserPlayback) {
        currentMode = modeRemoteControl
    }

    setMode(currentMode)

    setTimeout(updateNowPlaying, 1)
}

function getDataFileFromContainingTR(button) {
    const tr = button.closest('tr')
    return tr.dataset.file
}

function markFilePlayed(mediaFile) {
    fetch(`/mediafile/${encodeURIComponent(mediaFile)}`, { method: "DELETE" }).then(() => {
        location.reload()
    })
}

function setContentPadding() {
    const navbarHeight = navbar.offsetHeight;
    contentsDiv.style.marginTop = `${navbarHeight}px`;
}

function setMode(newMode) {
    currentPlayer?.pause()
    currentMode = newMode
    if (currentMode === modeRemoteControl) {
        currentPlayer = playerRemoteControl
        body.classList.add('mode-remote-control')
        body.classList.remove('mode-local-playback', 'no-footer')
        indexPositionSlider?.classList.remove('position-relative')  // needs to be display-block for the animation to work
        indexPositionSlider?.classList.add('animate__animated', 'animate__slideOutDown')
        indexPositionSlider?.addEventListener('animationend', () => {
            body.classList.remove('footer-includes-slider')
        }, {'once': true})
    } else {
        currentPlayer = playerBrowserPlayback
        body.classList.add('mode-local-playback')
        body.classList.remove('mode-remote-control')
        body.classList.add('footer-includes-slider')
        if (!pageContainsTracks) {
            // Hide the footer on index pages that have no tracks to play - typically just the root folder
            body.classList.add('no-footer')
        }

        indexPositionSlider?.classList.remove('animate__slideOutDown', 'position-relative') // needs to be visible, and display-block for the animation to work
        indexPositionSlider?.classList.add('animate__animated', 'animate__slideInUp')
        indexPositionSlider?.addEventListener('animationend', () => {
            body.classList.add('footer-includes-slider')  // in case of spurious trigger of the other animationend
            indexPositionSlider?.classList.add('position-relative') // needs to be position-relative for its display to be correct
        }, {'once': true})
    }

    localStorage.setItem('piaf-current-mode', newMode)
}

async function updateNowPlaying() {
    try {
        await currentPlayer.updateNowPlaying()
    } catch (error) {
        console.log(error)
        // and try again in a bit
    }
    setTimeout(updateNowPlaying, 1000)
}
