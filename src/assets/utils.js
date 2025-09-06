function gotoPage(path) {
    location.pathname = path
    history.pushState(null, null, location.toString())
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
