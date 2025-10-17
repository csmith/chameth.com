document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('*[data-contact]').forEach((el) => {
        el.innerHTML = `
            <form class="contact">
                <input type="text" name="name" id="name" placeholder="Your name/alias" required />
                <input type="email" name="email" id="email" placeholder="Your e-mail address" required />
                <textarea name="message" id="message" placeholder="Your message" required></textarea>
                <button type="submit" class="submit">Submit</button>
            </form>
        `

        el.querySelector('form').addEventListener('submit', (e) => {
            e.preventDefault()

            const payload = {
                page: document.location.href,
                ...Object.fromEntries(new FormData(e.target))
            }

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
})