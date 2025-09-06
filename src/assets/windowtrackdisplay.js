class WindowTrackDisplay {
    constructor() {
        this.trMediaFiles = document.getElementsByClassName('piaf-media-files')
    }

    showActiveTrack(nowPlayingFile) {
        for (const tr of this.trMediaFiles) {
            const rowPath = tr.getAttribute('data-file')
            if (rowPath === nowPlayingFile) {
                tr.classList.add('table-info')
            } else {
                tr.classList.remove('table-info')
            }
        }
    }

    showNoTrackPlaying() {
        this.showActiveTrack(null)
    }
}
