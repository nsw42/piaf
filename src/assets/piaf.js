let resumeButton;
let pauseButton;
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

    setTimeout(updateNowPlaying, 1000)
}

async function updateNowPlaying() {
    const response = await fetch("/player/status")
    const data = await response.json()
    switch (data['state']) {
        case 'stopped':
            pauseButton.setAttribute('disabled', 'disabled')
            resumeButton.setAttribute('disabled', 'disabled')
            volumeSlider.setAttribute('disabled', 'disabled')
            break
        case 'paused':
            pauseButton.setAttribute('disabled', 'disabled')
            resumeButton.removeAttribute('disabled')
            volumeSlider.removeAttribute('disabled')
            break
        case 'playing':
            pauseButton.removeAttribute('disabled')
            resumeButton.setAttribute('disabled', 'disabled')
            volumeSlider.removeAttribute('disabled')
            break
    }

    volumeSlider.value = data['volume']
    volumeText.innerHTML = data['volume']

    setTimeout(updateNowPlaying, 1000)
}
