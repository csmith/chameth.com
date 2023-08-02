---
title: Use ES modules
group: Eleventy
---

Eleventy doesn't currently support importing ES modules, and you can't easily
bridge the gap between CJS and ESM because ES importing is async and CJS is
synchronous.

To work around it, you can use the [require-esm-in-cjs](https://www.npmjs.com/package/require-esm-in-cjs)
package:

eleventy.config.cjs:

```javascript
const req = require('require-esm-in-cjs');
module.exports = req(`${__dirname}/eleventy.config.mjs`)
```

eleventy.config.mjs:

```javascript
export default function (eleventyConfig) {
    // ...
}
```