---
title: Render templates in shortcodes and filters
group: Eleventy
---

The easiest way to do this is to define your own instances of the template
renderer, which can then be used wherever you need them. Bear in mind that
most Eleventy templates are first parsed as Liquid or Nunjucks, then as
Markdown.

Eleventy will register all the filters and shortcodes when you call
`setLibrary`.

```javascript
const {Liquid} = require("liquidjs");
const Nunjucks = require('nunjucks');
const markdownIt = require('markdown-it');

const md = markdownIt({
    html: true
})

const liquid = new Liquid({
    extname: ".liquid",
    dynamicPartials: false,
    strictFilters: true,
    root: ["_includes"]
});

const njk = new Nunjucks.Environment(
    new Nunjucks.FileSystemLoader("_includes")
);

const templateify = async (content) => {
    return md.render(await liquid.parseAndRender(content, this));
}

module.exports = function (eleventyConfig) {
    eleventyConfig.setLibrary("liquid", liquid);
    eleventyConfig.setLibrary("njk", njk);
    eleventyConfig.setLibrary("md", md);

    eleventyConfig.addPairedShortcode("templateify", templateify);
};
```