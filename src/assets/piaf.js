function initPiaf() {
    document.querySelectorAll(".piaf-play-file").forEach(btn => {
        btn.addEventListener("click", () => {
            fetch("/player/play/" + btn.getAttribute("data-file"), {
                method: "PUT"
            })
        })
    })
}
