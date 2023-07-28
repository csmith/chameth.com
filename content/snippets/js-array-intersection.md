---
title: Array intersection
group: JavaScript
---

To find elements of `arrays` that contain all elements of a `target`:

```javascript
arrays.filter(i => target.every(j => i.includes(j)))
```