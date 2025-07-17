let resumeButton;
let pauseButton;
let speedMenuButton;
let volumeSlider;
let volumeText;
let volumeSliderDragActive;

function initPiaf() {
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

    setTimeout(updateNowPlaying, 1000)
}

function disableElements(elements) {
    for (const elt of elements) {
        elt.setAttribute('disabled', 'disabled')
    }
}

function enableElements(elements) {
    for (const elt of elements) {
        elt.removeAttribute('disabled')
    }
}

async function updateNowPlaying() {
    try {
        const response = await fetch("/player/status")
        if (response?.ok) {
            const data = await response.json()
            switch (data['state']) {
                case 'stopped':
                    disableElements([
                        pauseButton,
                        resumeButton,
                        speedMenuButton
                    ])
                    break
                case 'paused':
                    disableElements([pauseButton])
                    enableElements([resumeButton, speedMenuButton])
                    break
                case 'playing':
                    disableElements([resumeButton])
                    enableElements([pauseButton, speedMenuButton])
                    break
            }

            let speed = data['speed']
            if (!speed.endsWith('x')) {
                speed += "x"
            }
            speedMenuButton.innerHTML = speed
            volumeSlider.value = volumeText.innerHTML = data['volume']
        } else {
            console.log(`Fetch failed: ${response?.status}`)
        }
    } catch (error) {
        console.log(error)
        // and try again in a bit
    }

    setTimeout(updateNowPlaying, 1000)
}
