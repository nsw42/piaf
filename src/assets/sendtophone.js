const localStorageKeyPhoneAddress = 'piaf-phone-address'
let currentThreeDotsMenu = null

document.addEventListener("DOMContentLoaded", (event) => {
    const fileToUploadElt = document.getElementById('file-to-upload')
    const phoneAddressElt = document.getElementById('phone-ip-address')
    const sendToPhoneForm = document.getElementById('send-to-phone-form')
    const sendToPhoneModal = document.getElementById('send-to-phone-modal')
    const sendToPhoneAddressMenuItems = document.getElementsByClassName('send-to-phone-address-menu-item')

    for (const form of document.querySelectorAll('.needs-validation')) {
        form.addEventListener('submit', (event) => {
            if (!form.checkValidity()) {
                event.preventDefault()
                event.stopImmediatePropagation()
            }

            form.classList.add('was-validated')
        })
    }

    sendToPhoneModal.addEventListener('show.bs.modal', (event) => {
        // Set upload information in the form
        phoneUploadMenuInitiated(event.relatedTarget, fileToUploadElt)

        let phoneAddress = localStorage.getItem(localStorageKeyPhoneAddress)
        if (phoneAddress === null) {
            phoneAddress = '192.168.0.240'
        }
        phoneAddressElt.value = phoneAddress

        // Clear any previous state to ensure the modal displays correctly
        sendToPhoneModal.classList.remove('send-failed', 'sending')
    })

    sendToPhoneModal.addEventListener('shown.bs.modal', () => {
        phoneAddressElt.focus()
    })

    sendToPhoneForm.addEventListener('submit', async (event) => {
        event.preventDefault()
        event.stopImmediatePropagation()  // Prevent the page from being reloaded

        if (sendToPhoneModal.classList.contains('sending')) {
            return  // prevent double submission
        }

        doPhoneUpload(sendToPhoneModal, phoneAddressElt.value, fileToUploadElt)
    })

    for (const menuItem of sendToPhoneAddressMenuItems) {
        menuItem.addEventListener('click', async (event) => {
            phoneUploadMenuInitiated(event.currentTarget, fileToUploadElt)
            currentThreeDotsMenu?.classList.remove('bi-three-dots-vertical', 'bi-arrow-repeat', 'bi-check2', 'bi-x')
            currentThreeDotsMenu?.classList.add('bi-arrow-repeat')
            let ok = await doPhoneUpload(sendToPhoneModal, event.currentTarget.dataset.phoneAddress, fileToUploadElt)
            currentThreeDotsMenu?.classList.remove('bi-arrow-repeat')
            currentThreeDotsMenu?.classList.add(ok ? 'bi-check2' : 'bi-x')
        })
    }
})

function phoneUploadMenuInitiated(menuItem, fileToUploadElt) {
    currentThreeDotsMenu = menuItem.closest('ul').previousElementSibling
    if (currentThreeDotsMenu.tagName !== 'I') {
        console.log(`Error finding three dot menu node: found a ${currentThreeDotsMenu.tagName}`)
        currentThreeDotsMenu = null
    }
    fileToUploadElt.value = getDataFileFromContainingTR(menuItem)
}

async function doPhoneUpload(sendToPhoneModal, phoneAddress, fileToUploadElt) {
    sendToPhoneModal.classList.remove('send-failed')
    sendToPhoneModal.classList.add('sending')

    localStorage.setItem(localStorageKeyPhoneAddress, phoneAddress)

    const response = await fetch('/phone', {
        method: "POST",
        body: JSON.stringify({
            file: fileToUploadElt.value,
            phone: phoneAddress,
        }),
        headers: {
            "Content-Type": "application/json",
        }
    })
    sendToPhoneModal.classList.remove('sending')
    if (response.ok) {
        currentThreeDotsMenu?.classList.remove('bi-three-dots-vertical', 'bi-arrow-repeat', 'bi-check2', 'bi-x')
        currentThreeDotsMenu?.classList.add('bi-check2')
        bootstrap.Modal.getInstance(sendToPhoneModal)?.hide()
    } else {
        sendToPhoneModal.classList.add('send-failed')
    }
    return response.ok
}
