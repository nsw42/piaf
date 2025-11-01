const localStorageKeyPhoneAddress = 'piaf-phone-address'
let currentThreeDotsMenu = null

document.addEventListener("DOMContentLoaded", (event) => {
    const fileToUploadElt = document.getElementById('file-to-upload')
    const phoneAddressElt = document.getElementById('phone-ip-address')
    const sendToPhoneForm = document.getElementById('send-to-phone-form')
    const sendToPhoneModal = document.getElementById('send-to-phone-modal')

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
        currentThreeDotsMenu = event.relatedTarget.closest('ul').previousElementSibling
        if (currentThreeDotsMenu.tagName !== 'I') {
            console.log(`Error finding three dot menu node: found a ${currentThreeDotsMenu.tagName}`)
            currentThreeDotsMenu = null
        }
        fileToUploadElt.value = getDataFileFromContainingTR(event.relatedTarget)

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

        sendToPhoneModal.classList.remove('send-failed')
        sendToPhoneModal.classList.add('sending')

        let phoneAddress = phoneAddressElt.value
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
            currentThreeDotsMenu?.classList.remove('bi-three-dots-vertical')
            currentThreeDotsMenu?.classList.add('bi-check2')
            bootstrap.Modal.getInstance(sendToPhoneModal).hide()
        } else {
            sendToPhoneModal.classList.add('send-failed')
        }
    })
})
