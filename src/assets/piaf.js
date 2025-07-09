let resumeButton;
let pauseButton;

function initPiaf() {
    document.querySelectorAll(".piaf-play-file").forEach(btn => {
        btn.addEventListener("click", () => {
            fetch("/player/play/" + btn.getAttribute("data-file"), {
                method: "PUT"
            })
        })
    })

    document.querySelector('#pause').addEventListener("click", () => {
        fetch("/player/pause", { method: "PUT" })
    })

    document.querySelector('#resume').addEventListener("click", () => {
        fetch("/player/resume", { method: "PUT" })
    })

    pauseButton = document.getElementById('pause')
    resumeButton = document.getElementById('resume')

    setTimeout(updateNowPlaying, 1000)
}

async function updateNowPlaying() {
    const response = await fetch("/player/status")
    const data = await response.json()
    switch (data['state']) {
        case 'stopped':
            pauseButton.setAttribute('disabled', 'disabled')
            resumeButton.setAttribute('disabled', 'disabled')
            break
        case 'paused':
            pauseButton.setAttribute('disabled', 'disabled')
            resumeButton.removeAttribute('disabled')
            break
        case 'playing':
            pauseButton.removeAttribute('disabled')
            resumeButton.setAttribute('disabled', 'disabled')
            break
    }

    setTimeout(updateNowPlaying, 1000)
}
