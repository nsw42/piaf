const modeRemoteControl = 'remote';
const modeLocalPlayback = 'local';

let currentMode;
let modeButtonRemoteControl;
let modeButtonLocalPlayback;

let contentsDiv;
let navbar;
let controlLinkButton;
let trMediaFiles;
let nowPlayingFile;

let playerLocalPlaback;
let playerRemoteControl;
let currentPlayer;

let windowMediaControls;

function initPiaf() {
    windowMediaControls = new WindowMediaControls()

    contentsDiv = document.getElementById('main-content')
    navbar = document.querySelector('.navbar.fixed-top')
    window.addEventListener('DOMContentLoaded', setContentPadding)
    window.addEventListener('resize', setContentPadding)

    modeButtonRemoteControl = document.getElementById('mode-indicator-remote')
    modeButtonLocalPlayback = document.getElementById('mode-indicator-local')
    modeButtonRemoteControl.addEventListener('click', () => { setMode(modeLocalPlayback) })
    modeButtonLocalPlayback.addEventListener('click', () => { setMode(modeRemoteControl) })

    controlLinkButton = document.getElementById('control-link')

    for (const button of document.getElementsByClassName('piaf-play-file')) {
        button.addEventListener("click", () => {
            const tr = button.closest('tr')
            const mediaFile = tr.getAttribute("data-file")
            currentPlayer.playFile(mediaFile)
        })
    }

    trMediaFiles = document.getElementsByClassName('piaf-media-files')

    playerRemoteControl = new RemoteControl()
    playerLocalPlaback = new LocalPlayback()

    currentMode = localStorage.getItem('piaf-current-mode');
    if (currentMode === null) {
        currentMode = modeRemoteControl
    }
    setMode(currentMode)

    setTimeout(updateNowPlaying, 1)
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
    } else {
        currentPlayer = playerLocalPlaback
        modeButtonRemoteControl.classList.add('d-none')
        modeButtonLocalPlayback.classList.remove('d-none')
        controlLinkButton?.classList.add('d-none')
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
