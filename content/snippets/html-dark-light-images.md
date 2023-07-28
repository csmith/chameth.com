---
title: Provide dark and light mode versions of images
group: HTML
---

```html
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="dark.png"/>
  <source media="(prefers-color-scheme: light)" srcset="light.png"/>
  <img src="default.png" alt="...">
</picture>
```