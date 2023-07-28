---
title: Load JSON
group: JavaScript
---

To load JSON from another file/endpoint, just use the `fetch` API:

```javascript
fetch('data.json')
  .then(response => response.json())
  .then(json => {
  })
```