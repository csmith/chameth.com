const site = require('./site');

module.exports = function (eleventyConfig) {
    site.plugins.forEach((p) => {
        if (Array.isArray(p)) {
            eleventyConfig.addPlugin(p[0], p[1]);
        } else {
            eleventyConfig.addPlugin(p);
        }
    });

    Object.entries(site.filters).forEach(([name, fn]) => {
        eleventyConfig.addFilter(name, fn);
    });

    Object.entries(site.dataExtensions).forEach(([name, fn]) => {
        eleventyConfig.addDataExtension(name, fn);
    });

    eleventyConfig.setFrontMatterParsingOptions({
        excerpt: true,
        excerpt_separator: "<!--more-->"
    });

    Object.entries(site.shortcodes.single).forEach(([name, fn]) => {
        eleventyConfig.addShortcode(name, fn);
    });

    Object.entries(site.shortcodes.paired).forEach(([name, fn]) => {
        eleventyConfig.addPairedShortcode(name, fn);
    });

    eleventyConfig.addPassthroughCopy(".", {
        filter: site.assets.filter,
        rename: site.assets.renamer,
    });

    Object.entries(site.libraries).forEach(([name, fn]) => {
        eleventyConfig.setLibrary(name, fn);
    });

    Object.entries(site.transforms).forEach(([name, fn]) => {
        eleventyConfig.addTransform(name, fn);
    });
};