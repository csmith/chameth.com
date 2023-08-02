export default function (eleventyConfig) {
    eleventyConfig.setFrontMatterParsingOptions({
        excerpt: true,
        excerpt_separator: "<!--more-->"
    });
}