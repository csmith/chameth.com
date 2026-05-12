document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('.contact-container').forEach((el) => {
        const form = el.querySelector('form')

        form.addEventListener('submit', (e) => {
            e.preventDefault()

            const payload = Object.fromEntries(new FormData(form))

            const loadingContainer = document.createElement('div')
            loadingContainer.classList.add('loading')
            loadingContainer.appendChild(document.createElement('span'))
            el.appendChild(loadingContainer)

            const resultContainer = document.createElement('div')
            resultContainer.classList.add('result')
            const resultText = document.createElement('p')
            resultContainer.appendChild(resultText)

            fetch("/api/contact", {
                method: "POST",
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            }).then(response => {
                if (response.ok) {
                    resultText.innerText = 'Message sent! Thanks for getting in touch!'
                    el.appendChild(resultContainer)
                    el.removeChild(loadingContainer)
                } else {
                    resultText.innerText = 'Sorry, something went wrong. Your message was not sent.'
                    el.appendChild(resultContainer)
                    el.removeChild(loadingContainer)
                }
            }).catch(() => {
                resultText.innerText = 'Sorry, something went wrong. Your message was not sent.'
                el.appendChild(resultContainer)
                el.removeChild(loadingContainer)
            });
        })
    })

    document.querySelectorAll('.nod-container').forEach((el) => {
        const form = el.querySelector('form')

        form.addEventListener('click', (e) => {
            if (!e.target.classList.contains('submit')) {
                form.querySelector('.submit').click()
            }
        })

        form.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                form.querySelector('.submit').click()
            }
        })

        form.addEventListener('submit', (e) => {
            e.preventDefault()

            const payload = {
                page: form.querySelector('input[name="page"]').value
            }

            const loadingContainer = document.createElement('div')
            loadingContainer.classList.add('loading')
            loadingContainer.appendChild(document.createElement('span'))
            el.appendChild(loadingContainer)

            const resultContainer = document.createElement('div')
            resultContainer.classList.add('result')
            const resultText = document.createElement('p')
            resultContainer.appendChild(resultText)

            fetch("/api/nod", {
                method: "POST",
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            }).then(response => {
                if (response.ok) {
                    resultText.innerText = 'Your nod has been received and is appreciated'
                    el.appendChild(resultContainer)
                    el.removeChild(loadingContainer)
                } else {
                    resultText.innerText = 'Sorry, something went wrong.'
                    el.appendChild(resultContainer)
                    el.removeChild(loadingContainer)
                }
            }).catch(() => {
                resultText.innerText = 'Sorry, something went wrong.'
                el.appendChild(resultContainer)
                el.removeChild(loadingContainer)
            });
        })
    })
})