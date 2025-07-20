let contentsDiv;
let navbar;
let resumeButton;
let pauseButton;
let speedMenuButton;
let volumeSlider;
let volumeText;
let volumeSliderDragActive;
let trMediaFiles;
let wasPlaying = false;

function initPiaf() {
    contentsDiv = document.getElementById('main-content')
    navbar = document.querySelector('.navbar.fixed-top')
    window.addEventListener('DOMContentLoaded', setContentPadding)
    window.addEventListener('resize', setContentPadding)

    document.querySelectorAll(".piaf-play-file").forEach(btn => {
        btn.addEventListener("click", () => {
            fetch("/player/play/" + btn.getAttribute("data-file"), {
                method: "PUT"
            })
        })
    })

    pauseButton = document.getElementById('pause')
    pauseButton.addEventListener("click", () => {
        fetch("/player/pause", { method: "PUT" })
    })

    resumeButton = document.getElementById('resume')
    resumeButton.addEventListener("click", () => {
        fetch("/player/resume", { method: "PUT" })
    })

    volumeSlider = document.getElementById('volume-slider')
    volumeSlider.addEventListener("input", (event) => {
        volumeText.innerHTML = event.target.value
        fetch(`/player/volume?v=${event.target.value}`, {
            method: "PUT"
        })
    })
    volumeSlider.addEventListener("onmousedown", () => {
        volumeSliderDragActive = true
    })
    volumeSlider.addEventListener("onmouseup", () => {
        volumeSliderDragActive = false
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

    setTimeout(updateNowPlaying, 1000)
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


function showNowPlaying(data) {
    switch (data['state']) {
        case 'stopped':
            if (wasPlaying) {
                location.reload()
            } else {
                disableElements([
                    pauseButton,
                    resumeButton,
                    speedMenuButton
                ])
            }
            break
        case 'paused':
            disableElements([pauseButton])
            enableElements([resumeButton, speedMenuButton])
            break
        case 'playing':
            wasPlaying = true;
            disableElements([resumeButton])
            enableElements([pauseButton, speedMenuButton])
            break
    }

    if (speedMenuButton) {
        const speed = data['speed']
        if (!speed.endsWith('x')) {
            speed += "x"
        }
        speedMenuButton.innerHTML = speed
    }

    volumeSlider.value = volumeText.innerHTML = data['volume']

    const nowPlaying = data['now_playing']
    for (const tr of trMediaFiles) {
        const rowPath = tr.getAttribute('data-path')
        if (rowPath == nowPlaying) {
            tr.classList.add('table-info')
        } else {
            tr.classList.remove('table-info')
        }
    }
}
