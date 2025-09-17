const modeRemoteControl = 'remote';
const modeLocalPlayback = 'local';

let currentMode;
let modeButtonRemoteControl;
let modeButtonLocalPlayback;

let body;
let contentsDiv;
let navbar;
let controlLinkButton;
let nowPlayingFile;
let indexPositionSlider;

let playerLocalPlaback;
let playerRemoteControl;
let currentPlayer;

let windowMediaControls;
let windowTrackDisplay;

function initPiaf() {
    const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]')
    tooltipTriggerList.forEach(tooltipTriggerEl => new bootstrap.Tooltip(tooltipTriggerEl))

    windowMediaControls = new WindowMediaControls()
    windowTrackDisplay = new WindowTrackDisplay()

    body = document.getElementsByTagName('body')[0]
    contentsDiv = document.getElementById('main-content')
    navbar = document.querySelector('.navbar.fixed-top')
    window.addEventListener('DOMContentLoaded', setContentPadding)
    window.addEventListener('resize', setContentPadding)

    modeButtonRemoteControl = document.getElementById('mode-indicator-remote')
    modeButtonLocalPlayback = document.getElementById('mode-indicator-local')
    modeButtonRemoteControl.addEventListener('click', () => { setMode(modeLocalPlayback) })
    modeButtonLocalPlayback.addEventListener('click', () => { setMode(modeRemoteControl) })

    indexPositionSlider = document.getElementById('index-position-slider')
    controlLinkButton = document.getElementById('control-link')

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
    playerLocalPlaback = new LocalPlayback()

    currentMode = localStorage.getItem('piaf-current-mode');
    if (currentMode === null) {
        currentMode = modeRemoteControl
    }
    setMode(currentMode)

    setTimeout(updateNowPlaying, 1)
}

function getDataFileFromContainingTR(button) {
    const tr = button.closest('tr')
    return tr.getAttribute('data-file')
}

function markFilePlayed(mediaFile) {
    fetch(`/mediafile/${mediaFile}`, { method: "DELETE" }).then(() => {
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
        modeButtonRemoteControl.classList.remove('d-none')
        modeButtonLocalPlayback.classList.add('d-none')
        controlLinkButton?.classList.remove('d-none')
        indexPositionSlider?.classList.remove('position-relative')  // needs to be display-block for the animation to work
        indexPositionSlider?.classList.add('animate__animated', 'animate__slideOutDown')
        indexPositionSlider?.addEventListener('animationend', () => {
            body.classList.remove('footer-includes-slider')
        }, {'once': true})
    } else {
        currentPlayer = playerLocalPlaback
        modeButtonRemoteControl.classList.add('d-none')
        modeButtonLocalPlayback.classList.remove('d-none')
        controlLinkButton?.classList.add('d-none')
        body.classList.add('footer-includes-slider')
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
