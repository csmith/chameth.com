---
title: Deploy assets in page bundles
group: Eleventy
---

If you keep assets alongside the content they're used in (what Hugo calls
"page bundles"), you can use a rename filter on `addPassthroughCopy` to
deploy them into the right place, respecting permalinks:

```javascript
const fs = require('fs');
const path = require('path');
const matter = require("gray-matter");

const rename = function (original) {
    // See if there's an index.md alongside the asset, if so figure out the permalink and put the file there.
    // This emulates Hugo's "page bundles" behaviour.
    const content = path.join(path.dirname(original), 'index.md');
    if (fs.existsSync(content)) {
        const data = matter(fs.readFileSync(content)).data;
        if (data && data.permalink) {
            return path.join(data.permalink, path.basename(original));
        }
    }
    return original;
};

module.exports = function (eleventyConfig) {
    eleventyConfig.addPassthroughCopy(".", {
        filter: [
            "content/**/*.jpg",
            "content/**/*.webp",
            // etc
        ],
        rename,
    });
};
```