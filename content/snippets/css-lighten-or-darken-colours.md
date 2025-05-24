---
title: Lighten or darken colours
group: CSS
---

Use [colour-mix](https://caniuse.com/mdn-css_types_color_color-mix):

```css
:root {
    --colour: #3c658d;
    --colour-dark: color-mix(in srgb, var(--colour), black 20%);
    --colour-light: color-mix(in srgb, var(--colour), white 20%);
}
```