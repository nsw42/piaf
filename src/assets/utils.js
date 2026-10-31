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

function showError(message) {
    console.log(message)
    const container = document.getElementById('error-toast-container')
    if (container === null) {
        return
    }
    const toast = document.createElement('div')
    toast.className = 'toast align-items-center text-bg-danger border-0'
    toast.setAttribute('role', 'alert')
    toast.setAttribute('aria-live', 'assertive')
    toast.setAttribute('aria-atomic', 'true')
    const inner = document.createElement('div')
    inner.className = 'd-flex'
    const msgBody = document.createElement('div')
    msgBody.className = 'toast-body'
    msgBody.textContent = message
    const closeButton = document.createElement('button')
    closeButton.type = 'button'
    closeButton.className = 'btn-close btn-close-white me-2 m-auto'
    closeButton.setAttribute('data-bs-dismiss', 'toast')
    closeButton.setAttribute('aria-label', 'Close')
    inner.append(msgBody, closeButton)
    toast.append(inner)
    container.append(toast)
    toast.addEventListener('hidden.bs.toast', () => toast.remove())
    new bootstrap.Toast(toast, { delay: 10000 }).show()
}

// fetch, reporting any network error or non-2xx status to the user.
// Returns the response if successful, or null otherwise.
async function fetchOrReportError(description, url, options) {
    let response
    try {
        response = await fetch(url, options)
    } catch (error) {
        showError(`${description} failed: ${error.message}`)
        return null
    }
    if (!response.ok) {
        showError(`${description} failed: ${response.status} ${response.statusText}`)
        return null
    }
    return response
}
