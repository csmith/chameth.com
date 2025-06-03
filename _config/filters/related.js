export default function (collection, tags) {
    try {
        const matches = []
        collection.forEach(post => {
            const postTags = post.data.tags || []
            let score = tags.reduce((s, tag) => s + (postTags.includes(tag) ? 1 : 0), 0)
            if (score > 0 && post.page.url !== this.page.url) {
                matches.push({score, post})
            }
        })

        matches.sort((a, b) => {
            if (a.score > b.score) {
                return -1
            }
            if (a.score < b.score) {
                return 1
            }
            return b.post.page.date - a.post.page.date
        })

        return matches.map((x) => x.post)
    } catch (error) {
        console.error("Failed to calculate related posts", error)
    }
};